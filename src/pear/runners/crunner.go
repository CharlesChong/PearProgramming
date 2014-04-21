package main

import (
    "fmt"
    "log"
    "net/http"
    "pear/central"
)

const defaultPort = "3000";

func main() {
    http.HandleFunc("/", central.GetServers)

    fmt.Println("Launching central server on " + defaultPort + "...")
    log.Fatal(http.ListenAndServe(":" + defaultPort, nil))
}