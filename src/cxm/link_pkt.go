package main

import (
	"fmt"
	"net"
	pktutil "pktlib"
	"time"
)

func ValidateLinkPkt(msg []byte, len int) bool {
	scid := pktutil.GetScidFromPkt(msg)
	hdr_type := pktutil.GetHdrTypeFromPkt(msg)
	if IsValidScid(scid) {
		if (hdr_type == pktutil.PKT_HDR_TYPE_LINK_REG) || (hdr_type == pktutil.PKT_HDR_TYPE_LINK_STATUS) {
			return true
		}
	}
	return false
}

func ProcessLinkStatusPkt(msg []byte) {
	port := pktutil.GetPortFromPkt(msg)
	admin_status := pktutil.GetPortStatusFromPkt(msg)
	scid := pktutil.GetScidFromPkt(msg)

	if !pktutil.IsAdminStatusValid(admin_status) {
		fmt.Println("Invalid admin status received")
		return
	}

	conn := GetConnFromScidAndPort(scid, port)
	if conn == nil {
		fmt.Println("No connection found for (SCID, Port) ", scid, port)
	}
	key := conn.GenerateKeyForConnPort(scid, port)
	conn.Ports[key].admin_status = admin_status
	fmt.Println("Setting admin status for (Scid, Port) ", scid, port,
		pktutil.GetLinkStatusString(admin_status))

}

func ProcessLinkPkt(conn *net.TCPConn, msg []byte, len int) {

	if !ValidateLinkPkt(msg, len) {
		fmt.Println("Invalid message received on Link connection")
		return
	}

	hdr_type := pktutil.GetHdrTypeFromPkt(msg)
	scid := pktutil.GetScidFromPkt(msg)
	port := pktutil.GetPortFromPkt(msg)

	switch hdr_type {
	case pktutil.PKT_HDR_TYPE_LINK_REG:
		fmt.Println("Received a LINK registration message from (SCID, Port) ", scid, port)
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(time.Duration(3) * time.Second)
	case pktutil.PKT_HDR_TYPE_LINK_STATUS:
		fmt.Println("Received a LINK status message from (SCID, Port) ", scid, port)
		ProcessLinkStatusPkt(msg)
	default:
		fmt.Println("Unrecoganized LINK message. Ignoring.")
	}
}

func HandleLinkConnection(conn *net.TCPConn) int {
	defer conn.Close()
	msg := make([]byte, 9000)
	fmt.Println("Got a LINK connection")
	for {
		pkt_len, err := conn.Read(msg)
		if err != nil {
			fmt.Println("Error in Link socket. Bailing out. err:", err)
			break
		} else {
			//fmt.Println(string(msg))
			fmt.Println("Recevid link message of length ", pkt_len)
			fmt.Println(msg[1])
			ProcessLinkPkt(conn, msg, pkt_len)
		}
	}
	return 0
}
