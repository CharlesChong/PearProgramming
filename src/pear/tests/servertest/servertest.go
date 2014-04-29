package main 

import (
	"flag"
	"fmt"
	"log"
	"pear/server"
)

const defaultPort = 9000
const defaultCentralPort = "localhost:3000"

var (
	myPort          = flag.Int("port", defaultPort, "port number to listen on")
	centralHostPort = flag.String("central", defaultCentralPort, "central storage server host port (if non-empty then this storage server is a slave)")
	nodeID          = flag.Uint("id", 0, "a 32-bit unsigned node ID to use for consistent hashing")
)

func main() {
	fmt.Println("Running Pear Server Test....")
	flag.Parse()

	// Create and start the StorageServer.
	ps , err := server.NewServer(*centralHostPort, *myPort)
	if err != nil {
		log.Fatalln("Failed to create storage server:", err)
	}

	
	// Test Code here!
	// TODO Remove test code
	// Test 1: Add doc
	err = ps.sendAddDoc("Hello")
	if ps.myHostPort == "localhost:9001" {
		common.LOGV.Println("Testing Remove")
		err = ps.sendRemoveDoc("Hello")
	}

	// Test 2: Get Doc
	// var doc string
	// err = ps.sendAddDoc("Hello")
	// if ps.myHostPort == "localhost:9001" {
	// 	doc , err = ps.ClientGetDoc("Hello")
	// 	if err != nil {
	// 		common.LOGE.Println(err)
	// 		return nil, err
	// 	}
	// 	common.LOGV.Println("Result: ",doc)
	// }

	// Test 3: 2PC begin
	// err = ps.sendAddDoc("Hello")
	// time.Sleep(time.Second*5)
	// if ps.myHostPort == "localhost:9001" {
	// 	msg := serverrpc.Message {
	// 		TId : "Yo dis my Tid",
	// 		Body : "HEY THAR",
	// 	}
	// 	_, err  = ps.ClientRequestTxn(&msg,"Hello")
	// }

	// Test 4: Dead Pear Server
	// err = ps.sendAddDoc("Hello")
	// time.Sleep(time.Second )

	// if ps.myHostPort == "localhost:9001" {
		// common.LOGV.Println("Testing Add")
		// for {
		// 	common.LOGV.Println("Sending..")
		// 	err = ps.sendAddDoc("Hello")
		// 	time.Sleep(time.Second * 5)
		// }
	// }
	// if ps.myHostPort != "localhost:9001" {
	// 	ps.handleDead("localhost:9001")
	// }

}
