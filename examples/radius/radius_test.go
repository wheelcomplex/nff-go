// Copyright 2017 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package radius

import (
	"testing"

	"github.com/intel-go/nff-go/common"
	"github.com/intel-go/nff-go/packet"
)

func init() {
	SlaProfileDB[3]["testuser1"] = &SlaProfileDef{
		BngID:        3,
		SlaProfile:   "testuser1",
		SlaProfileID: 42,
		AclProfile: BngAclProfile{
			AclRuleIdx:     0,
			DIRECTION:      UPLINK,
			AclRemoteIP:    InetAddr(packet.BytesToIPv4(13, 91, 95, 74)),
			AclRemotePort:  80,
			AclRemoteProto: common.TCPNumber,
			AclLocalIP:     InetAddr(packet.BytesToIPv4(192, 168, 111, 42)),
			AclLocalPort:   0,
			AclLocalProto:  common.TCPNumber,
			ACTION:         PROCESS,
			RatingGrp:      0,
			Quota:          0,
		},
		QoSProfile: BngQoSProfile{
			SlaProfileID: "testuser1",
			QoSIdx:       0,
			QoSCir:       0,
			QoSMbr:       0,
		},
	}
}

func testCommon(cm CachingRADIUSMap, t *testing.T) {
	// Test valid lookup
	value, ok := cm.Get(3, "testuser1")
	if !ok || value == nil {
		t.Errorf("Bad returned values for profile testuser1: %+v, %+v", value, ok)
		t.Fail()
	}

	// Test invalid lookup
	value, ok = cm.Get(0, "nosuchuser")
	if ok || value != nil {
		t.Errorf("Bad returned values for profile nosuchuser: %+v, %+v", value, ok)
		t.Fail()
	}

	// Test cached lookup
	value, ok = cm.Get(3, "testuser1")
	if !ok || value == nil {
		t.Errorf("Bad returned values for profile testuser1: %+v, %+v", value, ok)
		t.Fail()
	}
}

func TestThreadSafe(t *testing.T) {
	cm := NewCachingRADIUSMap(true)
	testCommon(cm, t)
}

func TestThreadUnsafe(t *testing.T) {
	cm := NewCachingRADIUSMap(false)
	testCommon(cm, t)
}
