package main

import "fmt"
import "net"
import pktutil "pktlib"
import "os"
import "cxmparser"
import "pkttracker"

var parseFile = "./testdata/isim.cdf"
var trackFile = "./adm_tracker.log"

func HandleWireConnection(conn *net.TCPConn) int {
	defer conn.Close()
	msg := make([]byte, 9000)
	fmt.Println("Got a WIRE connection")
	for {
		_, err := conn.Read(msg)
		if err != nil {
			fmt.Println("Error in Wire socket. Bailing out. err:", err)
			break
		} else {
			fmt.Println(string(msg))
		}
	}
	return 0
}

func HandleProbeConnection(conn *net.TCPConn) int {
	msg := make([]byte, 9000)
	fmt.Println("Got a PROBE connection")
	for {
		_, err := conn.Read(msg)
		if err != nil {
			fmt.Println("Error in Probe socket. Bailing out. err:", err)
			break
		} else {
			fmt.Println(string(msg))
		}
	}
	return 0
}

func StartServer(MainChannel chan []byte, ip []byte, port int) {
	//fmt.Printf("Starting server on port %d \n", port)
	var tcpInfo net.TCPAddr

	tcpInfo.IP = net.IPv4(ip[0], ip[1], ip[2], ip[3])
	tcpInfo.Port = port
	tcpInfo.Zone = ""

	ln, err := net.ListenTCP("tcp4", &tcpInfo)
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			// handle error
		}
		switch port {
		case pktutil.SERVER_PORT_CXM:
			go HandleCxmConnection(MainChannel, conn)
		case pktutil.SERVER_PORT_LINK:
			go HandleLinkConnection(conn)
		case pktutil.SERVER_PORT_WIRE:
			go HandleWireConnection(conn)
		case pktutil.SERVER_PORT_PROBE:
			go HandleProbeConnection(conn)
		}
	}
}

func StartAllServers(MainChannel chan []byte) {
	ip := []byte{0, 0, 0, 0}

	go StartServer(MainChannel, ip, pktutil.SERVER_PORT_CXM)
	go StartServer(MainChannel, ip, pktutil.SERVER_PORT_LINK)
	go StartServer(MainChannel, ip, pktutil.SERVER_PORT_WIRE)
	go StartServer(MainChannel, ip, pktutil.SERVER_PORT_PROBE)

}

func admProgHelp() {
	fmt.Println("ADM commandline help: ")
	fmt.Println("adm_switch [-f <config_path_filename>] [-t <track_filename]")
	fmt.Println("config_path_filename: Full path and file name of the adm switch configuration file")
	fmt.Println("track_filename: Full path and file name of the packet tracking file")
}

func admArgParse(args []string) bool {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f":
			if i+1 < len(args) {
				i = i + 1
				parseFile = args[i]
			} else {
				return false
			}
		case "-t":
			if i+1 < len(args) {
				i = i + 1
				trackFile = args[i]
			} else {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func main() {

	MainChannel := make(chan []byte)

	args := os.Args
	if (len(args) > 1) && !admArgParse(args[1:]) {
		fmt.Println("Error in command line arguments")
		admProgHelp()
		return
	}

	err := cxmparser.ParseConfig(parseFile, &g_components)
	if err.Fail() {
		fmt.Println("Parsing configuration file failed", parseFile)
		return
	}

	g_pkttracker_handle = pkttracker.PktTrackerInit()

	fmt.Println("Done with parsing ", parseFile, ". Start servers....")
	StartAllServers(MainChannel)

	for {
		select {
		case data := <-MainChannel:
			//Some SCID is trying to switch data packet to another SCID
			scid := pktutil.GetScidFromPkt(data)
			fmt.Println("Message of length ", len(data),
				"received on the main channel For SCID: ", scid)
			comp := GetCompFromScid(scid)
			comp.Channel <- &data
		}
	}

}
