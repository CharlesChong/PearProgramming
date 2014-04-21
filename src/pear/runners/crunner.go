package main

import (
    "fmt"
    "flag"
    "log"
    "pear/central"
)

const defaultPort = 3000;

var (
	myPort = flag.Int("port", defaultPort, "port number to listen on")
)

func main() {
	fmt.Println("PearCentral Running....")
	flag.Parse()

	// Create and start the StorageServer.
	_, err := central.NewCentral(*myPort)
	if err != nil {
		log.Fatalln("Failed to create storage server:", err)
	}
}