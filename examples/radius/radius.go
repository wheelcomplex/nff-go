package radius

import (
	"sync"
)

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

func lookupInRADIUS(BngID uint16, profile string) (*SlaProfileDef, bool) {
	value, ok := SlaProfileDB[BngID][profile]
	return value, ok
}

// TODO: End hardcoded variables which should be replaced with RADIUS lookup
// ----------------------------------------------------------------------

// CachingRADIUSMap is a caching map abstract interface
type CachingRADIUSMap interface {
	Get(BngID uint16, profile string) (value *SlaProfileDef, ok bool)
}

type threadUnsafeCachingMap struct {
	lookup map[string]*SlaProfileDef
}

func (this threadUnsafeCachingMap) Get(BngID uint16, profile string) (*SlaProfileDef, bool) {
	value, ok := this.lookup[profile]

	if ok {
		return value, ok
	}

	// TODO: Replace with RADIUS lookup
	value, ok = lookupInRADIUS(BngID, profile)
	if ok {
		// Cache the value
		this.lookup[profile] = value
		return value, true
	}
	return nil, false
}

type threadSafeCachingMap struct {
	lookup sync.Map
}

func (this threadSafeCachingMap) Get(BngID uint16, profile string) (*SlaProfileDef, bool) {
	value, ok := this.lookup.Load(profile)

	if ok {
		return value.(*SlaProfileDef), ok
	}

	// TODO: Replace with RADIUS lookup
	value, ok = lookupInRADIUS(BngID, profile)
	if ok {
		// Cache the value
		this.lookup.Store(profile, value)
		return value.(*SlaProfileDef), true
	}
	return nil, false
}

// NewCachingRADIUSMap returns caching map instance. If mt parameter
// is true, internal map structure is thread safe, if it is false,
// internal map is not thread safe.
func NewCachingRADIUSMap(mt bool) (m CachingRADIUSMap) {
	if mt {
		return threadSafeCachingMap{lookup: sync.Map{}}
	} else {
		return threadUnsafeCachingMap{lookup: map[string]*SlaProfileDef{}}
	}
}
