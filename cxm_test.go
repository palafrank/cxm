package main

import (
  "testing"
  "net"
  "fmt"
  "time"
)

func TestCxmParser1(t *testing.T) {
  num_comps, num_conn := ParseConfig("./testdata/isim.cdf")
  //num_comps := ParseExmp()
  //num_comps := GetNumComps()
  if(( num_comps != 36) || (num_conn !=36)) {
    t.Error("Error in Component parsing. Expected: 36, 36 , Returned: ", num_comps, num_conn)
  } else {
    fmt.Println("PASSED: TestCxmParser1")
  }
}

func ClientConnect(port int) *net.TCPConn {
  var tcpInfo net.TCPAddr;


  tcpInfo.IP = net.IPv4(127, 0, 0, 1)
  tcpInfo.Port = port
  tcpInfo.Zone = ""

  conn, err := net.DialTCP("tcp4", nil, &tcpInfo)
  if(err != nil) {
    return nil
  }
  return conn
}


func CreateLinkRegMsg(msg []byte) {
  SetScidInPkt(msg, 1)
  SetHdrTypeInPkt(msg, PKT_HDR_TYPE_LINK_REG)
  SetPortInPkt(msg, 5)
}

func CreateLinkStatusMsg(msg []byte) {
  SetScidInPkt(msg, 1)
  SetHdrTypeInPkt(msg, PKT_HDR_TYPE_LINK_STATUS)
  SetPortInPkt(msg, 5)
  SetPortStatusInPkt(msg, LINK_UP)
}

func CreateCxmRegMsg(msg []byte) {
  SetScidInPkt(msg, 41)
  SetHdrTypeInPkt(msg, PKT_HDR_TYPE_COMP_REGISTRATION)
  SetPortInPkt(msg, 4)
}

func CreateCxmStatusMsg(msg []byte) {
  SetScidInPkt(msg, 1)
  SetHdrTypeInPkt(msg, PKT_HDR_TYPE_COMP_STATUS_QUERY)
  SetPortInPkt(msg, 5)
  SetPortStatusInPkt(msg, LINK_UP)
}

func CreateCxmDataMsg(msg []byte) {
  SetScidInPkt(msg, 1)
  SetHdrTypeInPkt(msg, PKT_HDR_TYPE_COMP_REGISTRATION)
  SetDataPktLen(msg, 16)
}

func TestCxmServer(t *testing.T) {

  ParseConfig("./testdata/isim.cdf")

  msg := make([]byte, 20)
  StartAllServers()
  time.Sleep(1000 * 1000 * 1000 * 5)

  conn := ClientConnect(SERVER_PORT_CXM)

  if(conn == nil) {
    t.Error("Error connecting to CXM server")
  } else {
    fmt.Println("Connected to CXM server")
  }
  CreateCxmRegMsg(msg)
  _, err := conn.Write(msg)
  if(err != nil) {
    t.Error("Error sending message to CXM server")
  } else {
    fmt.Println("Successfully sent message to CXM server")
  }

  time.Sleep(1000 * 1000 * 1000 * 5)

  err = conn.Close()
  if(err != nil) {
    t.Error("Error closing the CXM connection")
  } else {
    fmt.Println("Successfully closed the CXM connection")
  }

/*
  if(!ClientConnect(SERVER_PORT_LINK)) {
    t.Error("Error connecting to LINK server")
  } else {
    fmt.Println("PASSED: TestLinkServer1")
  }

  if(!ClientConnect(SERVER_PORT_PROBE)) {
    t.Error("Error connecting to CXM server")
  } else {
    fmt.Println("PASSED: TestProbeServer1")
  }

  if(!ClientConnect(SERVER_PORT_WIRE)) {
    t.Error("Error connecting to Wire server")
  } else {
    fmt.Println("PASSED: TestWireServer1")
  }
*/

}
