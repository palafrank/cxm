package main

import (
	"cxmparser"
	"os"
	"testing"
)

func TestParsing(t *testing.T) {

	err := cxmparser.ParseConfig("../../testdata/isim.cdf", &g_components)
	if err.Fail() {
		t.Error("Failed to parse the configuration file")
		return
	}
	t.Log("Successfully parsed the file")
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
