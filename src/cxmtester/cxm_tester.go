package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	pktutil "pktlib"
	"time"
)

const TESTER_OP_TEST_START = 0x01
const TESTER_OP_TEST_EXPECT = 0x02
const TESTER_OP_TEST_READER = 0x03
const TESTER_OP_TEST_READER_SUCCESS = 0x04
const TESTER_OP_TEST_READER_FAILURE = 0x05
const TESTER_OP_TEST_TIMEOUT = 0x6
const TESTER_OP_HANDSHAKE = 0xBEEF

type TesterMessage struct {
	op    int
	len   int
	port  int
	scid  int
	start time.Time
	end   time.Duration
}

var MainChannel chan TesterMessage

func CreateCxmRegMsg(msg []byte, scid int) {
	pktutil.SetScidInPkt(msg, scid)
	pktutil.SetHdrTypeInPkt(msg, pktutil.PKT_HDR_TYPE_COMP_REGISTRATION)
	//pktutil.SetPortInPkt(msg, 4)
}

func CreateCxmStatusMsg(msg []byte, scid int) {
	pktutil.SetScidInPkt(msg, scid)
	pktutil.SetHdrTypeInPkt(msg, pktutil.PKT_HDR_TYPE_COMP_STATUS_QUERY)
	pktutil.SetPortInPkt(msg, 4)
	pktutil.SetPortStatusInPkt(msg, pktutil.LINK_UP)
}

func CreateLinkRegMsg(msg []byte) {
	pktutil.SetScidInPkt(msg, 41)
	pktutil.SetHdrTypeInPkt(msg, pktutil.PKT_HDR_TYPE_LINK_REG)
	pktutil.SetPortInPkt(msg, 4)
}

func CreateLinkStatusMsg(msg []byte) {
	pktutil.SetScidInPkt(msg, 41)
	pktutil.SetHdrTypeInPkt(msg, pktutil.PKT_HDR_TYPE_LINK_STATUS)
	pktutil.SetPortInPkt(msg, 4)
	pktutil.SetPortStatusInPkt(msg, pktutil.LINK_UP)

}

func CreateCxmDataMsg(msg []byte, len int, scid int, port int) {
	sig := []byte{1, 2, 3, 4, 5}
	pktutil.SetScidInPkt(msg, scid)
	pktutil.SetPortInPkt(msg, port)
	pktutil.SetHdrTypeInPkt(msg, pktutil.PKT_HDR_TYPE_DATA)
	pktutil.SetDataPktLen(msg, len-pktutil.PKT_HDR_LEN)
	pktutil.SetPktSignature(msg, sig)
}

func ClientConnect(port int) *net.TCPConn {
	var tcpInfo net.TCPAddr

	tcpInfo.IP = net.IPv4(127, 0, 0, 1)
	tcpInfo.Port = port
	tcpInfo.Zone = ""

	conn, err := net.DialTCP("tcp4", nil, &tcpInfo)
	if err != nil {
		return nil
	}
	return conn
}

func SendMessage(scid int, msg []byte, conn *net.TCPConn) time.Time {
	_, err := conn.Write(msg)
	if err != nil {
		fmt.Println("Error sending message len ", len(msg), " to CXM server")
	}
	return time.Now()
}

func ReadMessage(scid int, conn *net.TCPConn, reader chan TesterMessage) {
	var signal TesterMessage
	msg := make([]byte, 10000)
	for {
		len, err := conn.Read(msg)
		if err != nil {
			fmt.Println("Error reading message to CXM port")
		} else {
			fmt.Println("Successfully received ", len, "bytes message from CXM ", scid)
			signal.op = TESTER_OP_TEST_READER
			signal.len = len
			signal.port = pktutil.GetPortFromPkt(msg)
			reader <- signal
		}
	}
	return
}

func RunClient(scid int, port_chan chan TesterMessage) {

	cxm_conn := ClientConnect(pktutil.SERVER_PORT_CXM)

	if cxm_conn == nil {
		fmt.Println("Error connecting to CXM server")
		return
	} else {
		fmt.Println("Connected to CXM server")
	}

	msg := make([]byte, 20)
	CreateCxmRegMsg(msg, scid)
	SendMessage(scid, msg, cxm_conn)

	var signal TesterMessage
	signal.op = TESTER_OP_HANDSHAKE
	port_chan <- signal

	read_channel := make(chan TesterMessage)
	go ReadMessage(scid, cxm_conn, read_channel)
	expected_port := 0
	expected_pkt_len := 0
	start := time.Now()
	for {
		select {
		case signal = <-port_chan:
			if signal.op == TESTER_OP_TEST_START {
				fmt.Println("SCID-", scid, ": Testing packet (len: ", signal.len,
					") to test on port ", signal.port)
				data := make([]byte, signal.len)
				CreateCxmDataMsg(data, signal.len, scid, signal.port)
				SendMessage(scid, data, cxm_conn)
			} else if signal.op == TESTER_OP_TEST_EXPECT {
				fmt.Println("SCID-", scid, ": Expecting packet (len: ", signal.len,
					") to test on port ", signal.port)
				expected_port = signal.port
				expected_pkt_len = signal.len
				start = time.Now()
			}

		case signal = <-read_channel:
			if (signal.len != expected_pkt_len) || (signal.port != expected_port) {
				fmt.Println("SCID-", scid, "Recevied unexpected length (",
					signal.len, ") packet on port ", signal.port, " while expectation ",
					expected_port, expected_pkt_len)
				signal.op = TESTER_OP_TEST_READER_FAILURE
				MainChannel <- signal
			} else {
				end := time.Since(start)
				fmt.Println("Recevied message from crossconnect on (SCID, PORT):", scid, signal.port, end)
				signal.op = TESTER_OP_TEST_READER_SUCCESS
				signal.end = end
				MainChannel <- signal
			}
			expected_port = 0
			expected_pkt_len = 0
		}
	}
}

func RunTestTimer() {
	var signal TesterMessage
	time.Sleep(time.Millisecond * 5000)
	signal.op = TESTER_OP_TEST_TIMEOUT
	MainChannel <- signal
}

func main() {

	MainChannel = make(chan TesterMessage)

	PortMap := [][]int{
		0:  {29, 25, 11, 1, 0, 0},
		1:  {29, 33, 11, 5, 0, 0},
		2:  {29, 1, 2, 1, 0, 0},
		3:  {29, 2, 2, 2, 0, 0},
		4:  {32, 25, 11, 17, 0, 0},
		5:  {32, 33, 11, 21, 0, 0},
		6:  {32, 1, 2, 17, 0, 0},
		7:  {32, 2, 2, 18, 0, 0},
		8:  {32, 14, 41, 1, 0, 0},
		9:  {32, 14, 11, 17, 1, 0}, //FAIL case
		10: {29, 25, 2, 1, 1, 0},   //FAIL case
	}

	ChannelMap := make(map[int]chan TesterMessage, len(PortMap)*2)
	//Create a channel for every SCID being tested
	num_scid := 0
	for _, val := range PortMap {
		if ChannelMap[val[0]] == nil {
			Channel := make(chan TesterMessage)
			ChannelMap[val[0]] = Channel
			num_scid++
		}
		if ChannelMap[val[2]] == nil {
			Channel := make(chan TesterMessage)
			ChannelMap[val[2]] = Channel
			num_scid++
		}
	}

	//Start a client per SCID
	for scid, channel := range ChannelMap {
		fmt.Println("Starting Client for SCID ", scid)
		go RunClient(scid, channel)
	}

	//Finish handshaking with all the Clients
	done := 0
	for key, _ := range ChannelMap {
		select {
		case signal := <-ChannelMap[key]:
			if signal.op == TESTER_OP_HANDSHAKE {
				fmt.Println("Client SCID ", key, " has started")
				done++
			}
		}
	}

	if done != num_scid {
		log.Fatalln("Not all SCIDs completed handshake")
	}

	time.Sleep(time.Millisecond * 1000)

	rand.Seed(42)
	total_tests := 0
	var total_time time.Duration
	var signal TesterMessage
	//for {
	for key, val := range PortMap {
		fmt.Println("Testing connection ", key, ": (", val[0], ",", val[1], ") <--> (",
			val[2], ",", val[3], ")")
		signal.op = TESTER_OP_TEST_START
		signal.len = rand.Intn(1024)
		signal.port = val[1]
		ChannelMap[val[0]] <- signal
		signal.op = TESTER_OP_TEST_EXPECT
		signal.port = val[3]
		signal.start = time.Now()
		ChannelMap[val[2]] <- signal

		go RunTestTimer()
		expected := 0
		test_done := 0

		for test_done == 0 {
			select {
			case data := <-MainChannel:
				if data.op == TESTER_OP_TEST_READER_FAILURE {
					if val[4] == 0 {
						log.Fatalln("Testcase failed")
					}
					val[5] = 1
				} else if data.op == TESTER_OP_TEST_READER_SUCCESS {
					expected = 1
					total_time += data.end
					total_tests += 1
				} else if data.op == TESTER_OP_TEST_TIMEOUT {
					if expected == 1 {
						fmt.Println("Testcase for connection ", key, " passed")
						val[5] = 0
					} else {
						//log.Fatalln("Testcase failed. Did not receive expected packet")
						val[5] = 1
					}
					test_done = 1
				}
			}
		}

		//}
		fmt.Println("Average time: ", time.Duration(int(total_time)/total_tests))
	}

	fmt.Println("Test Results:")
	for key, val := range PortMap {
		fmt.Println("Connection ", key, ": (", val[0], ",", val[1], ") <--> (",
			val[2], ",", val[3], ") Expected: ", val[4], "Result: ", val[5])
	}

}
