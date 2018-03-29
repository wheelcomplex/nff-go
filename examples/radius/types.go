package radius

import ()

type DirectionType uint8

const (
	UPLINK   DirectionType = 0
	DOWNLINK DirectionType = 1
)

type ActionType uint8

const (
	DROP    ActionType = 0
	PROCESS ActionType = 1
)

type InetAddr uint32

type BngAclProfile struct {
	AclRuleIdx     uint8
	DIRECTION      DirectionType
	AclRemoteIP    InetAddr
	AclRemotePort  uint16
	AclRemoteProto uint16
	AclLocalIP     InetAddr
	AclLocalPort   uint16
	AclLocalProto  uint16
	ACTION         ActionType
	RatingGrp      uint8
	Quota          uint16
}

type BngQoSProfile struct {
	SlaProfileID string
	QoSIdx       uint8
	QoSCir       uint16
	QoSMbr       uint16
}

type SlaProfileDef struct {
	BngID        uint16
	SlaProfile   string
	SlaProfileID uint16
	AclProfile   BngAclProfile
	QoSProfile   BngQoSProfile
}

type HomeSubnet InetAddr

const MAX_HOME_NETS = 1024

type IMSIBng struct {
	BngID         uint16
	IMSI          uint64
	HomeSubnetLst [MAX_HOME_NETS]HomeSubnet
}

const ALL_IMSI = 100

var IMSILDAP [ALL_IMSI]IMSIBng

type SlaBng struct {
	BngId      uint16
	IMSI       uint64
	SlaProfile string
}

type IMSI_profile struct {
	IMSI         uint64
	IMSIBngLocal IMSIBng
	SlaProfile   string
}
