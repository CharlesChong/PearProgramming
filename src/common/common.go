package common

import (
	"pear/rpc/serverrpc"
	// "io/ioutil"
	"log"
	"os"
)

var LOGE = log.New(os.Stderr, "ERROR ", log.Lshortfile)
// var LOGV = log.New(ioutil.Discard, "VERBOSE ", log.Lshortfile)

var LOGV = log.New(os.Stdout, "VERBOSE ", log.Lshortfile)

var serverStatusName = [...]string{
	"N/A", "OK", "NotReady",
}

func ToString(d serverrpc.Status) string {
	return "$ " + serverStatusName[d]
}
