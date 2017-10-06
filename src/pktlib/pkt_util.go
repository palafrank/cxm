package pktutil

import (
	"fmt"
)

const SERVER_PORT_LINK = 41000
const SERVER_PORT_CXM = 40000
const SERVER_PORT_PROBE = 41110
const SERVER_PORT_WIRE = 41100

//Packet HDR type
const PKT_HDR_TYPE_LINK_REG = 30
const PKT_HDR_TYPE_LINK_STATUS = 32

//CPM message types
const PKT_HDR_TYPE_COMP_STATUS_UPDATE = 0
const PKT_HDR_TYPE_COMP_STATUS_QUERY = 1
const PKT_HDR_TYPE_COMP_STATUS_REPLY = 2
const PKT_HDR_TYPE_COMP_SHUTDOWN_REQ = 3
const PKT_HDR_TYPE_COMP_SHUTDOWN_ACCEPT = 4
const PKT_HDR_TYPE_COMP_SPECIFIC_QUERY = 5
const PKT_HDR_TYPE_COMP_SPECIFIC_QUERY_REPLY = 6

//CXM message types
const PKT_HDR_TYPE_COMP_REGISTRATION = 10
const PKT_HDR_TYPE_COMP_PORT_STATUS_QUERY = 12
const PKT_HDR_TYPE_COMP_PORT_STATUS_REPLY = 13

const PKT_HDR_TYPE_DATA = 20

//Link states
const LINK_DOWN = 0
const LINK_UP = 1

const PKT_HDR_LEN = 8

var pktSigOffset = 4 //Offset from the end of the packet
var pktSigLen = 5    //Length in bytes

func GetScidFromPkt(msg []byte) int {
	scid := int(msg[0])
	return scid
}

func SetScidInPkt(msg []byte, scid int) {
	msg[0] = byte(scid)
}

func GetHdrTypeFromPkt(msg []byte) int {
	hdr_type := int(msg[1])
	return hdr_type
}

func GetHdrTypeString(hdr_type byte) string {
	switch hdr_type {
	case 0:
		return "PKT_HDR_TYPE_COMP_STATUS_UPDATE"
	case 1:
		return "PKT_HDR_TYPE_COMP_STATUS_QUERY"
	case 2:
		return "PKT_HDR_TYPE_COMP_STATUS_REPLY"
	case 3:
		return "PKT_HDR_TYPE_COMP_SHUTDOWN_REQ"
	case 4:
		return "PKT_HDR_TYPE_COMP_SHUTDOWN_ACCEPT"
	case 5:
		return "PKT_HDR_TYPE_COMP_SPECIFIC_QUERY"
	case 6:
		return "PKT_HDR_TYPE_COMP_SPECIFIC_QUERY_REPLY"
	case 10:
		return "PKT_HDR_TYPE_COMP_REGISTRATION"
	case 12:
		return "PKT_HDR_TYPE_COMP_PORT_STATUS_QUERY"
	case 13:
		return "PKT_HDR_TYPE_COMP_PORT_STATUS_REPLY"
	case 20:
		return "PKT_HDR_TYPE_DATA"
	default:
		return "UNKNOWN"
	}
}

func SetHdrTypeInPkt(msg []byte, hdr_type int) {
	msg[1] = byte(hdr_type)
}

func GetPortFromPkt(msg []byte) int {
	port_lsb := int(msg[2])
	port_msb := int(msg[5])
	port := port_msb<<8 | port_lsb
	return port
}

func SetPortInPkt(msg []byte, port int) {
	msg[2] = byte(port & 0xFF)
	msg[5] = byte((port >> 8) & 0xFF)
}

func GetPortStatusFromPkt(msg []byte) int {
	astatus := int(msg[3])
	return astatus
}

func GetLinkStatusString(status int) string {
	if status == LINK_UP {
		return "UP"
	} else if status == LINK_DOWN {
		return "DOWN"
	} else {
		return "INVALID"
	}
}

func SetPortStatusInPkt(msg []byte, status int) {
	msg[3] = byte(status)
}

func GetDataPacketLen(msg []byte) int {
	pkt_msb := int(msg[3])
	pkt_lsb := int(msg[4])
	pkt_len := pkt_msb<<8 | pkt_lsb
	return pkt_len
}

func GetPktSignature(msg []byte) []byte {
	//SigOffset is from the end of the packet
	//Siglen is to take it to the beginning of the signature
	sig := msg[(len(msg) - (pktSigOffset + pktSigLen)):(len(msg) - pktSigOffset)]
	return sig
}

func SetPktSignature(msg []byte, sig []byte) {
	index := len(msg) - (pktSigOffset + pktSigLen)
	if (index + len(sig)) > len(msg) {
		//Violation of packet space
		return
	}
	for ind := range sig {
		msg[index] = sig[ind]
		index++
	}
}

func SetDataPktLen(msg []byte, len int) {
	msg[4] = byte(len & 0xFF)
	msg[3] = byte((len >> 8) & 0xFF)
}

func IsAdminStatusValid(status int) bool {
	if (status == LINK_UP) || (status == LINK_DOWN) {
		return true
	}
	return false
}

func IsHdrTypeCxm(hdr_type int) bool {
	if (hdr_type >= PKT_HDR_TYPE_COMP_REGISTRATION) &&
		(hdr_type <= PKT_HDR_TYPE_COMP_PORT_STATUS_REPLY) {
		return true
	}
	return false
}

func IsHdrTypeCpm(hdr_type int) bool {
	if (hdr_type >= PKT_HDR_TYPE_COMP_STATUS_UPDATE) &&
		(hdr_type <= PKT_HDR_TYPE_COMP_SPECIFIC_QUERY_REPLY) {
		return true
	}
	return false
}

func IsHdrTypeData(hdr_type int) bool {
	if hdr_type == PKT_HDR_TYPE_DATA {
		return true
	}
	return false
}

func PrintDataPacketHeader(msg []byte) {
	scid := GetScidFromPkt(msg)
	opcode := GetHdrTypeFromPkt(msg)

	fmt.Println("Packet Header- SCID:", scid, " Header Type: ",
		GetHdrTypeString(byte(opcode)), "Length: ", GetDataPacketLen(msg))
}
