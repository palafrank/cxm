package main

import "bufio"
import "os"
import "strings"

//import "fmt"

func ParseConfig(path string) (int, int, int) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return -1, 0, 0
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		if strings.Compare(scanner.Text(), comp_start) == 0 {
			parse_comp(scanner)
		} else if strings.HasPrefix(scanner.Text(), "conn") == true {
			parse_conn(scanner)
		}
	}
	AddCompsToScidMap()

	for i := 0; i < len(g_comp_array); i++ {
		//g_comp_array[i].PrintComponent()
		//fmt.Println(g_comp_array[i])
	}

	//conn := GetConnFromScidAndPort(41, 4)
	//fmt.Println(conn)

	return 0, len(g_comp_array), len(g_connections)
}

func ParseExmp() int {
	return 36
}
