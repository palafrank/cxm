
/*
TODO:
 - Validate component fields
 - Check for duplicate components in the comp list
*/

package main

import "fmt"
import "bufio"
import "strings"
import "strconv"
import "net"

type Port struct {
  PortNum int
  LinkType string
  PhyType string
}

type Component struct {
  Name string
  Remote bool
  LinkAdmin bool
  Type string
  Scid int
  MinPort int
  MaxPort int
  Ports []*Port
  Connections map[int]*Connection
  Channel chan []byte
  Sock *net.TCPConn
}

var comp_start = "comp {"
var comp_end   = "}"
var g_comp_array []*Component
var g_scid_mapped_comp map[int]*Component

func IsValidScid(scid int) bool {
  if g_scid_mapped_comp[scid] == nil {
    return false
  }
  return true
}

func GetNumComps() int {
  return len(g_comp_array)
}

func GetCompFromName(name string) *Component {
  for i:=0; i<len(g_comp_array); i++ {
    if strings.Compare(name, g_comp_array[i].Name) == 0 {
      return g_comp_array[i]
    }
  }
  return nil
}

func AddCompsToScidMap() {
  g_scid_mapped_comp = make(map[int]*Component, len(g_comp_array))

  for i:=0; i<len(g_comp_array); i++ {
      g_scid_mapped_comp[g_comp_array[i].Scid] = g_comp_array[i]
      //fmt.Println("Adding scid ", g_comp_array[i].Scid)
      //fmt.Println("Added scid ", g_scid_mapped_comp[g_comp_array[i].Scid] )
  }
}


func GetConnFromScidAndPort(scid int, port int) *Connection {
    //comp := g_scid_mapped_comp[scid]
    //fmt.Println("Get the scid ", comp)
    return g_scid_mapped_comp[scid].Connections[port]
}

func GetCompFromScid(scid int) *Component {
  return g_scid_mapped_comp[scid]
}

func (c *Component) InitComponent(socket *net.TCPConn) {
  c.Sock = socket;

  for port, conn := range c.Connections {
    key := conn.GenerateKeyForConnPort(c.Scid, port)
    fmt.Println("Scid ", c.Scid, " Port ", port, "LINK_UP")
    conn.Ports[key].admin_status = 1
    conn.Ports[key].oper_status = 1
  }
}

func (c *Component) AddConnectionToComponent(port *ConnectionPorts, conn *Connection) {
  c.Connections[port.Port] = conn
  //fmt.Println("Adding Commp/Conn map ", c.Scid, port.Port, conn)
  //fmt.Println("The data: ", c)
}


func (c Component) PrintComponent() {
  fmt.Println("Component: ", c.Name)
  fmt.Println(" Remote: ", c.Remote)
  fmt.Println(" LinkAdmin: ", c.LinkAdmin)
  fmt.Println(" Type: ", c.Type)
  fmt.Println(" SCID: ", c.Scid)
  fmt.Println(" Ports: ", c.MinPort, "..", c.MaxPort)
  fmt.Println(" NumPortConn: ", len(c.Connections))
}


func (c Component) ValidateFields() bool {
  return true
}

func (c *Component) ParsePort(ports []string) {
  portNums := strings.Split(ports[1], "..")
  minPort, _ := strconv.Atoi(portNums[0])
  maxPort, _ := strconv.Atoi(portNums[1])
  c.MinPort = minPort
  c.MaxPort = maxPort
  c.Connections = make(map[int]*Connection, (maxPort - minPort))
  for i:=minPort; i<=maxPort; i++ {
    port := new(Port)
    port.PortNum = i
    if len(ports) >= 4 {
      port.LinkType = strings.TrimPrefix(ports[2], "LINK=")
      port.PhyType = strings.TrimPrefix(ports[3], "PHY=")
    }
    if c.ValidateFields() == true {
      c.Ports = append(c.Ports, port)
    } else {
      //Error handling in parser
    }
  }
}

func parse_comp(scanner *bufio.Scanner) {
  comp := new(Component)
  for ;strings.Compare(scanner.Text(), comp_end) != 0 && scanner.Scan(); {
    trimmedString := strings.TrimSpace(scanner.Text())
    splitString := strings.Split(trimmedString, " ")
    switch(splitString[0]) {
      case "NAME" :
        comp.Name = splitString[1]
      case "REMOTE" :
        if strings.Compare(splitString[1], "yes") == 0 {
          comp.Remote = true
        } else {
          comp.Remote = false
        }
      case "LINK_ADMIN" :
        if strings.Compare(splitString[1], "yes") == 0 {
          comp.LinkAdmin = true
        } else {
          comp.LinkAdmin = false
        }
      case "TYPE" :
        comp.Type = splitString[1]
      case "SCID":
        comp.Scid, _ = strconv.Atoi(splitString[1])
      case "PORT":
        comp.ParsePort(splitString)
      case "}":
      default:
        fmt.Println("Unkown component attribute ", splitString[0])
    }
  }
  comp.Channel = make(chan []byte, 10)
  g_comp_array = append(g_comp_array, comp)
}
