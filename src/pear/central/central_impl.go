package central

import (
	"pear/rpc/centralrpc"
	"net/rpc"
	"net/http"
	"net"
	"common"
	// "time"
	// "errors"
	"strconv"
	"log"
	"fmt"
)

type central struct {
	docMap 		map[string] []string // doc -> serv list
	serverMap 	map[string] []string // serv -> doc list
	clientIdCnt int
	port 		int
}

func NewCentral(port int) (Central, error) {
	common.LOGV.Println("Pear Central Starting ",port)
	c := central{}
	c.port = port
	c.clientIdCnt = 0
	c.docMap = make(map[string] []string)
	c.serverMap = make(map[string] []string)

	myHostPort := fmt.Sprintf("localhost:%d", port)

	// Create the server socket that will listen for incoming RPCs.
	listener, err := net.Listen("tcp", myHostPort)
	if err != nil {
		return nil, err
	}

	// Wrap the tribServer before registering it for RPC.
	err = rpc.RegisterName("PearCentral", centralrpc.Wrap(&c))
	if err != nil {
		return nil, err
	}

	// Setup the HTTP handler that will server incoming RPCs and
	// serve requests in a background goroutine.
	rpc.HandleHTTP()
	go http.Serve(listener, nil)

    http.HandleFunc("/", c.NewClient)
    log.Fatal(http.ListenAndServe(":" + strconv.Itoa(port), nil))

	return &c, nil
}

func (c *central) AddDoc(args *centralrpc.AddDocArgs,reply *centralrpc.AddDocReply) error {

	return nil
}

func (c *central) RemoveDoc(args *centralrpc.RemoveDocArgs,reply *centralrpc.RemoveDocReply) error {
	
	return nil
}

func (c *central) AddServer(args *centralrpc.AddServerArgs,reply *centralrpc.AddServerReply) error {
	common.LOGV.Println("Server ",args.HostPort, "Added.")
	_, ok := c.serverMap[args.HostPort]
	if !ok {
		c.serverMap[args.HostPort] = []string{}
		reply.Status = centralrpc.OK
	} else {
		reply.Status = centralrpc.NotReady
	}

	return nil
}

func (c *central) NewClient(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    if (r.Form["docId"] != nil) && (len(r.Form["docId"]) > 0) {
        // var docId = r.Form["docId"][0]
        // $TODO: Actually return something meaningful
        fmt.Fprintf(w, strconv.Itoa(c.clientIdCnt) +" localhost:9000")
        // $TODO: Race condition here
        c.clientIdCnt++
    }
}