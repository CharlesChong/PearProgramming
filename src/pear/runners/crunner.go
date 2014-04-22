package main

import (
	"flag"
	// "fmt"
	"log"
	"pear/central"
)

const defaultPort = 3000

var (
	myPort = flag.Int("port", defaultPort, "port number to listen on")
)

func main() {
	flag.Parse()

	// Create and start the StorageServer.
	_, err := central.NewCentral(*myPort)
	if err != nil {
		log.Fatalln("Failed to create storage server:", err)
	}
	select {}
}
