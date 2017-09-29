package main

import "fmt"
import "bufio"
import "strings"
import "strconv"
import "pktlib"

type ConnectionPorts struct {
    Comp string
    Scid int
    Port int
    admin_status int
    oper_status int
}

type Connection struct {
  State string
  ConnType string
  Ports map[string]*ConnectionPorts
}


var g_connections []*Connection
var g_max_port_per_comp = 10

func (c Connection) SwitchDataPacket (msg []byte, pktlen int) {
  SrcPort := pktutil.GetPortFromPkt(msg)
  SrcScid := pktutil.GetScidFromPkt(msg)
  for _, port := range c.Ports {
    //Skip the source of the traffic
    //fmt.Println("TRYING: Switching packet from (SCID, PORT: ", SrcScid, SrcPort,
    //  ") -> (SCID, PORT: ", port.Scid, port.Port, ") len ", len(msg))
    if((port.Scid == SrcScid) && (port.Port == SrcPort)) {
      continue
    }
    if(port.oper_status == pktutil.LINK_UP) {
      pktutil.SetScidInPkt(msg, port.Scid)
      pktutil.SetPortInPkt(msg, port.Port)
      //Handover packet to SCID channel
      comp := GetCompFromScid(port.Scid)
      fmt.Println("Switching packet from (SCID, PORT: ", SrcScid, SrcPort,
        ") -> (SCID, PORT: ", port.Scid, port.Port, ") len ", len(msg))
      comp.Channel <- msg
      //Send the packet to appropriate SCID channel
    }
  }
}

func (c Connection) GenerateKeyForConnPort (scid int, port int) string {
  return (strconv.Itoa(scid) + strconv.Itoa(port))
}

func (c Connection) PrintConnection (index int) {
  fmt.Printf("Connection %d::", index)
  fmt.Printf("%s", c.ConnType)
  fmt.Printf("::%s::", c.State)
  for _, ports := range c.Ports {
    fmt.Printf(" %s,%d,%d ", ports.Comp, ports.Scid, ports.Port)
  }
  fmt.Println("::")
}

func parse_conn(scanner *bufio.Scanner) {

  conn := new(Connection)
  conn.Ports = make(map[string]*ConnectionPorts, g_max_port_per_comp)
  splitStrings := strings.Split(scanner.Text(), " ")

  for i:=0; i<len(splitStrings); i++ {
    if strings.HasPrefix(splitStrings[i], "STATE") {
      conn.State = strings.TrimPrefix(splitStrings[i], "STATE=")
    } else if strings.HasPrefix(splitStrings[i], "TYPE") {
      conn.ConnType = strings.TrimPrefix(splitStrings[i], "TYPE=")
    } else {
      connString := strings.Split(splitStrings[i], ",")
      if len(connString) == 2 {
        connPort := new(ConnectionPorts)
        connPort.Comp = connString[0]
        connPort.Port, _ = strconv.Atoi(connString[1])
        comp := GetCompFromName(connPort.Comp)
        //fmt.Println("Going to add to ", &comp)
        connPort.Scid = comp.Scid
        //conn.Ports = append(conn.Ports, connPort)
        key := conn.GenerateKeyForConnPort(connPort.Scid, connPort.Port)
        conn.Ports[key] = connPort
        //fmt.Println("Added connection to ", connPort.Scid, connPort.Port, key)
        comp.AddConnectionToComponent(connPort, conn)
      }
    }
  }
  g_connections = append(g_connections, conn)
}

func print_connections() {
  for conn := range g_connections {
    g_connections[conn].PrintConnection(conn)
  }
}
