package central

import (
	"common"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"pear/rpc/centralrpc"
	"pear/rpc/serverrpc"
	"strconv"
	"time"
)

type central struct {
	// Doc -> Server Map: Used for document teammate queries
	docMap 		map[string]map[string]bool // doc -> serv list
	// Server -> Doc Map: Used for server failure lookup
	serverMap   map[string]map[string]bool // serv -> doc list
	connMap 	map[string]*rpc.Client
	clientIdCnt int
	port        int
}

func NewCentral(port int) (Central, error) {
	common.LOGV.Println("Pear Central Starting ", port)
	c := central{}
	c.port = port
	c.clientIdCnt = 0
	c.docMap = make(map[string]map[string]bool)
	c.serverMap = make(map[string]map[string]bool)
	c.connMap = make(map[string]*rpc.Client)

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
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
	return &c, nil
}

func (c *central) AddDoc(args *centralrpc.AddDocArgs, reply *centralrpc.AddDocReply) error {
	common.LOGV.Println("AddDoc: ",args)
	reply.DocId = args.DocId
	reply.Status = centralrpc.OK
	stat1 := addToMap(c.docMap, args.DocId, args.HostPort)
	stat2 := addToMap(c.serverMap, args.HostPort, args.DocId)

	if stat1 == centralrpc.DocExist || stat2 == centralrpc.DocExist {
		reply.Status = centralrpc.DocExist
	}

	// Broadcast Status to all collaborators
	teammate, ok := c.docMap[args.DocId]
	if !ok {
		common.LOGE.Println("Central Failed to Add")
		reply.Status = centralrpc.NotReady
		return nil
	}	
	err := c.broadcastAddDoc(teammate,args)		
	reply.Teammates = teammate
	return err
}

func addToMap(m map[string]map[string]bool, key1, key2 string) centralrpc.Status {
	// Update docMap
	old, ok := m[key1]
	if !ok {
		newMap := make(map[string]bool)
		newMap[key2] = true
		m[key1] = newMap
	} else {
		_, ok2 := old[key2]
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
	common.LOGV.Println("RemoveDoc: ",args)
	reply.DocId = args.DocId
	reply.Status = centralrpc.OK

	stat1 := removeMap(c.docMap, args.DocId, args.HostPort)
	stat2 := removeMap(c.serverMap, args.HostPort, args.DocId)
	if stat1 == centralrpc.DocNotExist || stat2 == centralrpc.DocNotExist {
		reply.Status = centralrpc.DocNotExist
	}

	// Broadcast new status to all collaborators
	teammate , ok := c.docMap[args.DocId]
	if ok {
		err := c.broadcastRemoveDoc(teammate,args)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeMap(m map[string]map[string]bool, key1, key2 string) centralrpc.Status {
	// Update docMap
	old, ok := m[key1]
	if !ok {
		return centralrpc.DocNotExist
	} else {
		_, ok2 := old[key2]
		if ok2 {
			// Remove current collaborator from list
			delete(old, key2)
			m[key1] = old
			if len(old) == 0 {
				// Disappear if all collaborators disappear
				delete(m,key1)
			}
			return centralrpc.OK
		} else {
			return centralrpc.DocNotExist
		}
	}
	return centralrpc.OK
}

func (c *central) AddServer(args *centralrpc.AddServerArgs, reply *centralrpc.AddServerReply) error {
	common.LOGV.Println("Server ", args.HostPort, "Added.")
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
		fmt.Fprintf(w, strconv.Itoa(c.clientIdCnt)+" localhost:9000")
		// $TODO: Race condition here
		c.clientIdCnt++
	}
}

////////////////////// Broadcast helper functions /////////////////////

func (c *central) broadcastAddDoc(teammate map[string]bool,args *centralrpc.AddDocArgs) error {
	for k, _ := range teammate {
		err := c.sendAddedDoc(k,args.HostPort, args.DocId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *central) broadcastRemoveDoc(teammate map[string]bool, args *centralrpc.RemoveDocArgs) error {
	for k, _ := range teammate {
		err := c.sendRemoveDoc(k,args.HostPort, args.DocId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *central) sendAddedDoc(dstHostPort, myHostPort, docId string) error {
	client, err := c.dialRPC(dstHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return err
	}
	
	for {
		// Make RPC Call to Master
		args := &serverrpc.AddedDocArgs{
			DocId: docId,
			HostPort: myHostPort,
		}
		var reply serverrpc.AddedDocReply
		if err := client.Call("PearServer.AddedDoc", args, &reply); err != nil {
			return err
		}
		common.LOGV.Println("Call AddedDoc[",dstHostPort,"]: ",reply)
		// Check reply from Master
		if reply.Status == serverrpc.OK {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (c *central) sendRemoveDoc(dstHostPort, myHostPort, docId string) error {
	client, err := c.dialRPC(dstHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return err
	}
	
	for {
		// Make RPC Call to Master
		args := &serverrpc.RemovedDocArgs{
			DocId: docId,
			HostPort: myHostPort,
		}
		var reply serverrpc.RemovedDocReply
		common.LOGV.Println("Call RemovedDoc[",dstHostPort,"]: ",reply)
		if err := client.Call("PearServer.RemoveDoc", args, &reply); err != nil {
			return err
		}
		// Check reply from Master
		if reply.Status == serverrpc.OK {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (c *central) dialRPC(dstHostPort string) (*rpc.Client, error) {
	// Check if old connection exists
	oldClient, ok := c.connMap[dstHostPort]
	if ok {
		return oldClient, nil
	}

	// Initialize new connection
	var client *rpc.Client
	var err error
	maxFail := 5
	for tries := 0; ; tries++ {
		client, err = rpc.DialHTTP("tcp", dstHostPort)
		if err != nil {
			if tries >= maxFail {
				return nil ,err
			}
			time.Sleep(time.Second)
		} else {	
			// Store New Conn
			c.connMap[dstHostPort] = client
			return client, nil
		}
	}
}