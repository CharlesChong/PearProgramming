package main

import (
    "log"
    "net/http"
    "fmt"
    "html"
)

const defaultPort = "8080";

func main() {
    http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
    });
    http.Handle("/", http.FileServer(http.Dir("client")))

    fmt.Println("Launching central server on " + defaultPort + "...")
    log.Fatal(http.ListenAndServe(":" + defaultPort, nil))
}