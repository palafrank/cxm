/*
TODO:
 - Validate component fields
 - Check for duplicate components in the comp list
*/

package main

import "fmt"

import "strings"

import "net"

type Port struct {
	PortNum  int
	LinkType string
	PhyType  string
}

type Component struct {
	Name        string              `cdf:"NAME"`
	Remote      bool                `cdf:"REMOTE"`
	LinkAdmin   bool                `cdf:"LINK_ADMIN"`
	Type        string              `cdf:"TYPE"`
	Scid        int                 `cdf:"SCID"`
	MinPort     int                 `cdf:"MINPORT"`
	MaxPort     int                 `cdf:"MAXPORT"`
	Ports       []*Port             ``
	Connections map[int]*Connection ``
	Channel     chan *[]byte        ``
	Sock        *net.TCPConn        ``
}

const g_max_port_per_comp = 10

type ParserInterface struct {
	comp_array       []*Component
	conn_array       []*Connection
	scid_mapped_comp map[int]*Component
}

func (c *ParserInterface) CompAlloc() interface{} {
	return new(Component)
}

func (c *ParserInterface) CompFinish(comp_int interface{}) {
	comp := comp_int.(*Component)
	comp.Connections = make(map[int]*Connection, (comp.MaxPort - comp.MinPort))
	c.comp_array = append(c.comp_array, comp)
	comp.Channel = make(chan *[]byte, 10)
}

func (c *ParserInterface) ConnAlloc() interface{} {
	conn := new(Connection)
	conn.Ports = make(map[string]*ConnectionPorts, g_max_port_per_comp)
	return conn
}

func (c *ParserInterface) ConnAddPort(comp_name string, port int, conn_int interface{}) {
	conn := conn_int.(*Connection)
	connPort := new(ConnectionPorts)
	connPort.Comp = comp_name
	connPort.Port = port
	comp := GetCompFromName(comp_name)
	connPort.Scid = comp.Scid
	key := conn.GenerateKeyForConnPort(comp.Scid, connPort.Port)
	conn.Ports[key] = connPort
	comp.AddConnectionToComponent(connPort, conn)
}

func (c *ParserInterface) ConnFinish(z interface{}) {
}

func (c *ParserInterface) ParseComplete() {
	AddCompsToScidMap()
}

var g_components ParserInterface

func IsValidScid(scid int) bool {
	if g_components.scid_mapped_comp[scid] == nil {
		return false
	}
	return true
}

func GetNumComps() int {
	return len(g_components.comp_array)
}

func GetCompFromName(name string) *Component {
	for i := 0; i < len(g_components.comp_array); i++ {
		if strings.Compare(name, g_components.comp_array[i].Name) == 0 {
			return g_components.comp_array[i]
		}
	}
	return nil
}

func AddCompsToScidMap() {
	g_components.scid_mapped_comp = make(map[int]*Component, len(g_components.comp_array))

	for i := 0; i < len(g_components.comp_array); i++ {
		g_components.scid_mapped_comp[g_components.comp_array[i].Scid] = g_components.comp_array[i]
		//fmt.Println("Adding scid ", g_comp_array[i].Scid)
		//fmt.Println("Added scid ", g_scid_mapped_comp[g_comp_array[i].Scid] )
	}
}

func GetConnFromScidAndPort(scid int, port int) *Connection {
	//comp := g_scid_mapped_comp[scid]
	//fmt.Println("Get the scid ", comp)
	return g_components.scid_mapped_comp[scid].Connections[port]
}

func GetCompFromScid(scid int) *Component {
	return g_components.scid_mapped_comp[scid]
}

func (c *Component) InitComponent(socket *net.TCPConn) {
	c.Sock = socket

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
