package cxmparser

//Component is a device with many ports
//Connection is a link between two ports on two different components

type CxmWriter interface {
	ConnAlloc() interface{}
	ConnAddPort(comp string, port int, conn interface{})
	ConnFinish(interface{})
	CompAlloc() interface{}
	CompFinish(interface{})
	ParseComplete()
}

type ParserError int

const (
	ParserErrorCode_OK   ParserError = 0
	ParserErrorCode_FAIL ParserError = 1
)

var ParseErrorString = map[ParserError]string{
	0: "Parsing was successful",
	1: "Parsing failed",
}

func (p ParserError) Error() string {
	return ParseErrorString[p]
}

func (p ParserError) Fail() bool {
	if p != ParserErrorCode_OK {
		return true
	}
	return false
}

func ParseConfig(path string, writer CxmWriter) ParserError {
	return parseCdfConfig(path, writer)
}
