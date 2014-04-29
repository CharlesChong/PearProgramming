package main

import (
	"flag"
	"fmt"
	"log"
	"pear/server"
)

const defaultPort = 9000
const defaultCentralPort = "localhost:3000"
// const defaultHost = "localhost"
var (
	// myHost			= flag.String("host", defaultHost, "host")
	myPort          = flag.Int("port", defaultPort, "port number to listen on")
	centralHostPort = flag.String("central", defaultCentralPort, "central storage server host port (if non-empty then this storage server is a slave)")
	nodeID          = flag.Uint("id", 0, "a 32-bit unsigned node ID to use for consistent hashing")
)

func main() {
	fmt.Println("PearServer Running....")
	flag.Parse()

	// Create and start the StorageServer.
	_, err := server.NewServer(*centralHostPort, *myPort) //, *myHost)
	if err != nil {
		log.Fatalln("Failed to create storage server:", err)
	}

	// Run the storage server forever.
	select {}

}
