package radius

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type serverInfo struct {
	addr   string
	secret string
}

// ----------------------------------------------------------------------
// TODO: Start hardcoded variables which should be replaced with RADIUS lookup
const MAX_BNG_ID = 10
const MAX_BNG_PRFL = 1024

var SlaProfileDB [MAX_BNG_ID]map[string]*SlaProfileDef

func init() {
	for i := range SlaProfileDB {
		SlaProfileDB[i] = map[string]*SlaProfileDef{}
	}
}

func lookupInRADIUS(BngID uint16, profile string) (*SlaProfileDef, error) {
	value, ok := SlaProfileDB[BngID][profile]
	if !ok {
		return nil, errors.New("Not found: " + profile)
	}
	return value, nil
}

// TODO: End hardcoded variables which should be replaced with RADIUS lookup
// ----------------------------------------------------------------------

// Main RADIUS lookup routine
func (si *serverInfo) lookupInRADIUS(BngID uint16, profile string) (*SlaProfileDef, error) {
	packet := radius.New(radius.CodeAccessRequest, []byte(si.secret))
	rfc2865.UserName_SetString(packet, profile)
	BNGID_Set(packet, BNGID(BngID))

	response, err := radius.Exchange(context.Background(), packet, si.addr)
	if err != nil {
		return nil, err
	}

	if response.Code != radius.CodeAccessAccept {
		return nil, errors.New(fmt.Sprintf("Server did not authorize user %s. Returned code is %+v", profile, response.Code))
	}

	var value SlaProfileDef
	bid, err := BNGID_Lookup(response)
	if err != nil {
		return nil, err
	}
	value.BngID = uint16(bid)

	pid, err := BNGProfileID_Lookup(response)
	if err != nil {
		return nil, err
	}
	value.SlaProfileID = uint16(pid)

	return &value, nil
}

// CachingRADIUSMap is a caching map abstract interface
type CachingRADIUSMap interface {
	Get(BngID uint16, profile string) (value *SlaProfileDef, err error)
}

type threadUnsafeCachingMap struct {
	serverInfo
	lookup map[string]*SlaProfileDef
}

func (this threadUnsafeCachingMap) Get(BngID uint16, profile string) (*SlaProfileDef, error) {
	value, ok := this.lookup[profile]

	if ok {
		return value, nil
	}

	var err error
	value, err = (&this.serverInfo).lookupInRADIUS(BngID, profile)
	if err != nil {
		return nil, err
	}

	// Cache the value
	this.lookup[profile] = value
	return value, nil
}

type threadSafeCachingMap struct {
	serverInfo
	lookup sync.Map
}

func (this threadSafeCachingMap) Get(BngID uint16, profile string) (*SlaProfileDef, error) {
	value, ok := this.lookup.Load(profile)

	if ok {
		return value.(*SlaProfileDef), nil
	}

	var err error
	value, err = (&this.serverInfo).lookupInRADIUS(BngID, profile)
	if err != nil {
		return nil, err
	}

	// Cache the value
	this.lookup.Store(profile, value)
	return value.(*SlaProfileDef), nil
}

// NewCachingRADIUSMap returns caching map instance. If mt parameter
// is true, internal map structure is thread safe, if it is false,
// internal map is not thread safe.
func NewCachingRADIUSMap(mt bool, serverAddr string, serverSecret string) (m CachingRADIUSMap) {
	si := serverInfo{
		addr:   serverAddr,
		secret: serverSecret,
	}
	if mt {
		return threadSafeCachingMap{
			serverInfo: si,
			lookup:     sync.Map{},
		}
	} else {
		return threadUnsafeCachingMap{
			serverInfo: si,
			lookup:     map[string]*SlaProfileDef{},
		}
	}
}
