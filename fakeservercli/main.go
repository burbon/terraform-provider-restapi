package main

import (
	"flag"
	"fmt"
	"github.com/Mastercard/terraform-provider-restapi/fakeserver"
	"github.com/Mastercard/terraform-provider-restapi/log"
	"os"
)

func main() {
	api_server_objects := make(map[string]map[string]interface{})

	port := flag.Int("port", 8080, "The port fakeserver will listen on")
	debug := flag.Bool("debug", false, "Enable debug output of the server")
	static_dir := flag.String("static_dir", "", "Serve static content from this directory")

	flag.Parse()

	logger := log.New(*debug)
	svrOpts := &fakeserver.FakeServerOpts{
		Port:    *port,
		Objects: api_server_objects,
		Start:   false,
		Debug:   *debug,
		Logger:  logger,
		Dir:     *static_dir,
	}
	svr := fakeserver.NewFakeServer(svrOpts)

	fmt.Printf("Starting server on port %d...\n", *port)
	fmt.Println("Objects are at /api/objects/{id}")

	internal_server := svr.GetServer()
	err := internal_server.ListenAndServe()
	if nil != err {
		fmt.Printf("Error with the internal TCP server: %s", err)
		os.Exit(1)
	}
}
