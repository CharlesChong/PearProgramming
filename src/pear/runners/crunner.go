package main

import (
    "fmt"
    "flag"
    "log"
    "strconv"
    "net/http"
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
    http.HandleFunc("/", central.NewClient)

    fmt.Println("Launching central server on ", defaultPort , "...")
    defaultPortStr := ":" + strconv.Itoa(defaultPort)
    log.Fatal(http.ListenAndServe(defaultPortStr, nil))
}