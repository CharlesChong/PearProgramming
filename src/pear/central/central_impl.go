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
	// Doc -> Server Map: Used for document teammate queries
	docMap 		map[string] map[string]bool // doc -> serv list
	// Server -> Doc Map: Used for server failure lookup
	serverMap 	map[string] map[string]bool // serv -> doc list
	clientIdCnt int
	port 		int
}

func NewCentral(port int) (Central, error) {
	common.LOGV.Println("Pear Central Starting ",port)
	c := central{}
	c.port = port
	c.clientIdCnt = 0
	c.docMap = make(map[string]map[string]bool)
	c.serverMap = make(map[string]map[string]bool)

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

func (c *central) AddDoc(args *centralrpc.AddDocArgs, reply *centralrpc.AddDocReply) error {
	reply.DocId = args.DocId
	reply.Status = centralrpc.OK

	stat1 := addToMap(c.docMap,args.DocId,args.HostPort)
	stat2 := addToMap(c.serverMap,args.HostPort,args.DocId)

	if (stat1 == centralrpc.DocExist || stat2 ==centralrpc.DocExist) {
		reply.Status = centralrpc.DocExist
	}

	reply.Teammates = c.docMap[args.DocId]
	c.broadcastAddDoc(args)
	return nil
}

func addToMap(m map[string]map[string]bool,key1, key2 string) centralrpc.Status{
	// Update docMap
	old,ok := m[key1]
	if !ok {
		newMap := make(map[string]bool)
		newMap[key2] = true
		m[key1] = newMap
	} else {
		_,ok2 := old[key2]
		if ok2 {
			return centralrpc.DocExist
		} else {
			old[key2] = true
			m[key1] = old		
		}
	}
	return centralrpc.OK
}

func (c *central) RemoveDoc(args *centralrpc.RemoveDocArgs, reply *centralrpc.RemoveDocReply) error {
	reply.DocId = args.DocId
	reply.Status = centralrpc.OK
	stat1 := removeMap(c.docMap,args.DocId,args.HostPort)
	stat2 := removeMap(c.serverMap,args.HostPort,args.DocId)

	if (stat1 == centralrpc.DocNotExist || stat2 ==centralrpc.DocNotExist) {
		reply.Status = centralrpc.DocNotExist
	}

	c.broadcastRemoveDoc(args)
	return nil
}
func removeMap(m map[string]map[string]bool,key1, key2 string) centralrpc.Status{
	// Update docMap
	old,ok := m[key1]
	if !ok {
		return centralrpc.DocNotExist
	} else {
		_,ok2 := old[key2]
		if ok2 {
			delete(old,key2)
			m[key1] = old
			return centralrpc.OK
		} else {
			return centralrpc.DocNotExist	
		}
	}
	return centralrpc.OK
}

func (c *central) broadcastAddDoc(args *centralrpc.AddDocArgs) {
	common.LOGV.Println("TODO")
}

func (c *central) broadcastRemoveDoc(args *centralrpc.RemoveDocArgs) {
	common.LOGV.Println("TODO")
}

func (c *central) AddServer(args *centralrpc.AddServerArgs, reply *centralrpc.AddServerReply) error {
	common.LOGV.Println("Server ",args.HostPort, "Added.")
	_, ok := c.serverMap[args.HostPort]
	if !ok {
		c.serverMap[args.HostPort] = make(map[string]bool)
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