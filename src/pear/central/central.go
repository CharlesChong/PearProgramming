package central

import (
    "net/http"
    "fmt"
)

func GetServers(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    if (r.Form["docId"] != nil) && (len(r.Form["docId"]) > 0) {
        // var docId = r.Form["docId"][0]
        // $TODO: Actually return something meaningful
        fmt.Fprintf(w, "localhost:9000")
    }
}