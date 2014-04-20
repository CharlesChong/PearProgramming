package main

import (
    "log"
    "net/http"
    "fmt"
    "html"
)

const defaultPort = "8080";

func main() {
    http.HandleFunc("/getServers", getServers)

    fmt.Println("Launching central server on " + defaultPort + "...")
    log.Fatal(http.ListenAndServe(":" + defaultPort, nil))
}

func getServers(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
    r.ParseForm()
    fmt.Println(r.Form)
}