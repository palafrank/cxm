package main

import "fmt"

import "strconv"
import pktutil "pktlib"

type ConnectionPorts struct {
	Comp         string
	Scid         int
	Port         int
	admin_status int
	oper_status  int
}

type Connection struct {
	State    string `cdf:"STATE"`
	ConnType string `cdf:"TYPE"`
	Ports    map[string]*ConnectionPorts
}

var g_connections []*Connection

func (c Connection) SwitchDataPacket(msg []byte, pktlen int) {
	SrcPort := pktutil.GetPortFromPkt(msg)
	SrcScid := pktutil.GetScidFromPkt(msg)
	for _, port := range c.Ports {
		//Skip the source of the traffic
		//fmt.Println("TRYING: Switching packet from (SCID, PORT: ", SrcScid, SrcPort,
		//  ") -> (SCID, PORT: ", port.Scid, port.Port, ") len ", len(msg))
		if (port.Scid == SrcScid) && (port.Port == SrcPort) {
			continue
		}
		if port.oper_status == pktutil.LINK_UP {
			pktutil.SetScidInPkt(msg, port.Scid)
			pktutil.SetPortInPkt(msg, port.Port)
			//Handover packet to SCID channel
			comp := GetCompFromScid(port.Scid)
			fmt.Println("Switching packet from (SCID, PORT: ", SrcScid, SrcPort,
				") -> (SCID, PORT: ", port.Scid, port.Port, ") len ", len(msg))
			comp.Channel <- &msg
			//Send the packet to appropriate SCID channel
		}
	}
}

func (c Connection) GenerateKeyForConnPort(scid int, port int) string {
	return (strconv.Itoa(scid) + strconv.Itoa(port))
}

func (c Connection) PrintConnection(index int) {
	fmt.Printf("Connection %d::", index)
	fmt.Printf("%s", c.ConnType)
	fmt.Printf("::%s::", c.State)
	for _, ports := range c.Ports {
		fmt.Printf(" %s,%d,%d ", ports.Comp, ports.Scid, ports.Port)
	}
	fmt.Println("::")
}

func print_connections() {
	for conn := range g_connections {
		g_connections[conn].PrintConnection(conn)
	}
}
