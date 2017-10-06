package cxmparser

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var comp_start = "comp {"
var comp_end = "}"

//Function to map the CDF tags to field index
func parseMapCdfFields(t reflect.Type) map[string]int {
	mapper := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		val, exists := sf.Tag.Lookup("cdf")
		if exists {
			mapper[val] = i
		}
	}
	return mapper
}

//Parse the connection lines in the file
func parseCdfConn(scanner *bufio.Scanner, writer CxmWriter) {

	conn := writer.ConnAlloc()
	v := reflect.ValueOf(conn).Elem()
	t := v.Type()
	if v.Kind() != reflect.Struct {
		panic("Did not get a struct for connection")
	}
	mapper := parseMapCdfFields(t)

	splitStrings := strings.Split(scanner.Text(), " ")

	for i := 0; i < len(splitStrings); i++ {
		if strings.HasPrefix(splitStrings[i], "STATE") {
			index, ok := mapper["STATE"]
			if ok && v.Field(index).Kind() == reflect.String {
				state := strings.TrimPrefix(splitStrings[i], "STATE=")
				v.Field(index).Set(reflect.ValueOf(state))
			}
		} else if strings.HasPrefix(splitStrings[i], "TYPE") {
			index, ok := mapper["TYPE"]
			if ok && v.Field(index).Kind() == reflect.String {
				state := strings.TrimPrefix(splitStrings[i], "TYPE=")
				v.Field(index).Set(reflect.ValueOf(state))
			}
		} else {
			connString := strings.Split(splitStrings[i], ",")
			if len(connString) == 2 {
				val, _ := strconv.Atoi(connString[1])
				writer.ConnAddPort(connString[0], val, conn)
			}
		}
	}
	writer.ConnFinish(conn)
}

//Parse the Component lines in the file
func parseCdfComp(scanner *bufio.Scanner, writer CxmWriter) {
	comp := writer.CompAlloc()

	v := reflect.ValueOf(comp).Elem()
	if v.Kind() != reflect.Struct {
		panic("Did not get a struct for comp")
	}
	t := v.Type()
	mapper := parseMapCdfFields(t)

	for strings.Compare(scanner.Text(), comp_end) != 0 && scanner.Scan() {
		trimmedString := strings.TrimSpace(scanner.Text())
		splitString := strings.Split(trimmedString, " ")

		switch splitString[0] {

		case "NAME":
			i, ok := mapper[splitString[0]]
			if ok && v.Field(i).Kind() == reflect.String {
				v.Field(i).Set(reflect.ValueOf(splitString[1]))
			}
		case "REMOTE":
		case "LINK_ADMIN":
			i, ok := mapper[splitString[0]]
			if ok && v.Field(i).Kind() == reflect.Bool {
				if strings.Compare(splitString[1], "yes") == 0 {
					v.Field(i).Set(reflect.ValueOf(true))
				} else {
					v.Field(i).Set(reflect.ValueOf(false))
				}
			}
		case "TYPE":
			i, ok := mapper[splitString[0]]
			if ok && v.Field(i).Kind() == reflect.String {
				v.Field(i).Set(reflect.ValueOf(splitString[1]))
			}
		case "SCID":
			i, ok := mapper[splitString[0]]
			if ok && v.Field(i).Kind() == reflect.Int {
				val, _ := strconv.Atoi(splitString[1])
				v.Field(i).Set(reflect.ValueOf(val))
			}
		case "PORT":
			portNums := strings.Split(splitString[1], "..")
			i, ok := mapper["MINPORT"]
			if ok && v.Field(i).Kind() == reflect.Int {
				val, _ := strconv.Atoi(portNums[0])
				v.Field(i).Set(reflect.ValueOf(val))
			}
			i, ok = mapper["MAXPORT"]
			if ok && v.Field(i).Kind() == reflect.Int {
				val, _ := strconv.Atoi(portNums[1])
				v.Field(i).Set(reflect.ValueOf(val))
			}
		case comp_end:
		default:
			fmt.Println("Unkown component attribute ", splitString[0])
		}
	}
	writer.CompFinish(comp)
}

//Parse a CDF file
func parseCdfConfig(path string, writer CxmWriter) ParserError {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return ParserErrorCode_FAIL
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if strings.Compare(scanner.Text(), comp_start) == 0 {
			parseCdfComp(scanner, writer)
		} else if strings.HasPrefix(scanner.Text(), "conn") == true {
			parseCdfConn(scanner, writer)
		}
	}
	writer.ParseComplete()
	return ParserErrorCode_OK
}
