package main

import (
	"fmt"
	"net"
	pktutil "pktlib"
)

var g_pkttracker_handle interface{}

func ValidateCxmPkt(msg []byte, len int) bool {
	return true
}

func ProcessCxmPacket(msg []byte, socket *net.TCPConn) {
	hdr_type := pktutil.GetHdrTypeFromPkt(msg)
	scid := pktutil.GetScidFromPkt(msg)
	port := pktutil.GetPortFromPkt(msg)
	oper_status := pktutil.GetPortStatusFromPkt(msg)

	switch hdr_type {
	case pktutil.PKT_HDR_TYPE_COMP_REGISTRATION:
		fmt.Println("Recevied a Component Registration message for (scid,port): ", scid, port)
		comp := GetCompFromScid(scid)
		if comp == nil {
			fmt.Println("Received Registration from unknown component : ", scid)
			return
		}
		fmt.Println("Registering Component: ", scid, " Socket: ", socket)
		comp.InitComponent(socket)
	case pktutil.PKT_HDR_TYPE_COMP_PORT_STATUS_QUERY:
		fallthrough
	case pktutil.PKT_HDR_TYPE_COMP_PORT_STATUS_REPLY:
		conn := GetConnFromScidAndPort(scid, port)
		if conn == nil {
			fmt.Println("Invalid connection ", scid, port)
			return
		}
		key := conn.GenerateKeyForConnPort(scid, port)
		if conn.Ports[key].admin_status == pktutil.LINK_UP {
			conn.Ports[key].oper_status = oper_status
		} else {
			conn.Ports[key].oper_status = pktutil.LINK_DOWN
		}
		fmt.Println("Recevied a Component Staus message for (scid,port): ", scid,
			port, pktutil.GetLinkStatusString(oper_status))
	default:
		fmt.Println("Unrecoganized CXM message header type: ", hdr_type)
	}

}

func ProcessCpmPacket(msg []byte) {
	fmt.Println("Handling CPM messages")
}

func ProcessDataPacket(msg []byte, len int) {
	scid := pktutil.GetScidFromPkt(msg)
	port := pktutil.GetPortFromPkt(msg)
	payload_len := pktutil.GetDataPacketLen(msg)

	if (len - pktutil.PKT_HDR_LEN) != payload_len {
		fmt.Printf("Mismatch in calculated payload length (%d) and packet length: (%d)\n",
			payload_len, len)
		return
	}
	fmt.Println("Data packet received")

	conn := GetConnFromScidAndPort(scid, port)
	if conn != nil {
		conn.SwitchDataPacket(msg, len)
	} else {
		fmt.Println("Cannot find the connection to switch data packet.")
	}
}

func ProcessCxmMsg(conn *net.TCPConn, msg []byte, len int) {

	if !ValidateCxmPkt(msg, len) {
		fmt.Println("Invalid message received on Link connection")
		return
	}

	hdr_type := pktutil.GetHdrTypeFromPkt(msg)

	if pktutil.IsHdrTypeCxm(hdr_type) {
		ProcessCxmPacket(msg, conn)
	} else if pktutil.IsHdrTypeCpm(hdr_type) {
		ProcessCpmPacket(msg)
	} else if pktutil.IsHdrTypeData(hdr_type) {
		ProcessDataPacket(msg, len)
	} else {
		fmt.Println("Unrecoganized header type on cxm port: ", hdr_type)
	}
}

/*
  This routine listens on the SCID specific channel and writes out
  the data to the socket when data is available. The originator SCID comp
  would have already modified the SCID/Port in the packet before sending it
  to the channel.
*/

func WriteCxmPackets(sig chan int, scid int, conn *net.TCPConn) {
	defer conn.Close()
	comp := GetCompFromScid(scid)
	fmt.Println("Getting comp for scid: ", scid)
	if comp == nil {
		fmt.Println("Did not find Comp ", scid, ". Closing socket.")
		return
	}
	down := 0
	for {
		select {
		case data := <-comp.Channel:
			fmt.Println("Received message to send to SCID: ", scid, " len ", len(*data))
			len, err := conn.Write(*data)
			if err != nil {
				fmt.Println("Error transmitting packet")
			} else {
				fmt.Println("Sent message of len ", len)
			}
		case down = <-sig:
			fmt.Println("Write function for scid ", scid, " shutting down.", down)
			break
		}
		if down == 1 {
			break
		}
	}
}

/*
  Reads all packets coming on the CXM connection for a specific SCID
  The first time packet comes, the SCID is determined and a write socket
  is initiated.
  The channel between the read and write socket determines that there
  is an error on the socket and we need to get out of these read/write routines.
*/

func ReadCxmPackets(conn *net.TCPConn) {
	scid := 0
	StartWriteThread := false
	read_msg := make([]byte, 9000)
	sig := make(chan int)
	defer close(sig)
	for {
		pkt_len, err := conn.Read(read_msg)
		if err != nil {
			fmt.Println("Error (", err, ") reading from the socket. Closing Read for SCID ", scid)
			sig <- 1
			break
		} else {
			if pkt_len == 0 {
				fmt.Println("Got a packet of len 0")
			}
			msg := read_msg[:pkt_len]
			if StartWriteThread == false {
				go WriteCxmPackets(sig, pktutil.GetScidFromPkt(msg), conn)
				StartWriteThread = true
			}
			fmt.Println("Recevid cxm message of length ", pkt_len, "Header type: ",
				pktutil.GetHdrTypeString(byte(pktutil.GetHdrTypeFromPkt(msg))),
				"SCID: ", pktutil.GetScidFromPkt(msg))
			scid = pktutil.GetScidFromPkt(msg)
			ProcessCxmMsg(conn, msg, pkt_len)
		}
	}
}

/*
  Purpose: Go routine to start up the Read routine that will read from the
  CXM socket. This routine will then exit
*/
func HandleCxmConnection(MainChannel chan []byte, conn *net.TCPConn) int {
	fmt.Println("Got a CXM connection")
	go ReadCxmPackets(conn)
	return 0
}
