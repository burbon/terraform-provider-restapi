package restapi

import (
	"flag"
	"os"
	"testing"
)

var debug_flag = flag.Bool("debug", false, "logger in debug mode")
var debug bool

func TestMain(m *testing.M) {
	flag.Parse()
	debug = *debug_flag
	os.Exit(m.Run())
}
