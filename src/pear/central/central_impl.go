package central

import (
	"common"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"pear/rpc/centralrpc"
	"pear/rpc/serverrpc"
	"strconv"
	"time"
)

type central struct {
	// Representations:
	// Server - hostport
	// Doc - document name (docId)

	// Doc -> Server Map: Used for document teammate queries
	docMap 		map[string]map[string]bool // doc -> serv list
	// Server -> Doc Map: Used for server failure lookup
	serverMap   map[string]map[string]bool // serv -> doc list
	// Remembers initalized connections
	connMap 	map[string]*rpc.Client
	// Counter for creating clientIds
	clientIdCnt int
	port        int
	myHostPort  string
}

func NewCentral(port int) (Central, error) {
	common.LOGV.Println("$PearCentral on ",port)
	c := central{}
	c.port = port
	c.clientIdCnt = 0
	c.docMap = make(map[string]map[string]bool)
	c.serverMap = make(map[string]map[string]bool)
	c.connMap = make(map[string]*rpc.Client)

	c.myHostPort = fmt.Sprintf("localhost:%d", port)

	// Create the server socket that will listen for incoming RPCs.
	listener, err := net.Listen("tcp", c.myHostPort)
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
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
	return &c, nil
}

func (c *central) AddDoc(args *centralrpc.AddDocArgs, reply *centralrpc.AddDocReply) error {
	common.LOGV.Println("$Request AddDoc: ",args)
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
	common.LOGV.Println("$Request RemoveDoc: ",args)
	reply.DocId = args.DocId
	reply.Status = centralrpc.OK
	
	// Broadcast new status to all collaborators
	teammate , ok := c.docMap[args.DocId]
	if ok {
		common.LOGV.Println(teammate)
		err := c.broadcastRemoveDoc(teammate,args)
		if err != nil {
			return err
		}
	}

	stat1 := removeMap(c.docMap, args.DocId, args.HostPort,true)
	stat2 := removeMap(c.serverMap, args.HostPort, args.DocId,false)
	if stat1 == centralrpc.DocNotExist || stat2 == centralrpc.DocNotExist {
		reply.Status = centralrpc.DocNotExist
	}
	
	return nil
}

func removeMap(m map[string]map[string]bool, key1, key2 string, isDocMap bool) centralrpc.Status {
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
			if isDocMap && len(old) == 0 {
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
	common.LOGV.Println("$Server ", args.HostPort, "Added.")
	_, ok := c.serverMap[args.HostPort]
	if !ok {
		c.serverMap[args.HostPort] = make(map[string]bool)
	}
	reply.Status = centralrpc.OK
	return nil
}

func (c *central) RemoveServer(args *centralrpc.RemoveServerArgs, reply *centralrpc.RemoveServerReply) error {
	common.LOGV.Println("$Server ", args.HostPort, "Removed.")
	c.handleDead(args.HostPort)
	reply.Status = centralrpc.OK
	return nil
}

func (c *central) NewClient(w http.ResponseWriter, r *http.Request) {
	if len(c.serverMap) == 0 {
		fmt.Fprintf(w, "No available pear servers")
	} else {
		serverIdx := rand.Intn(len(c.serverMap))
		for hostPort, _ := range c.serverMap {
			if serverIdx == 0 {
				err := c.sendCheckAlive(hostPort)
				if err == nil {
					fmt.Fprintf(w, strconv.Itoa(c.clientIdCnt)+" "+ hostPort)
					// $TODO: Race condition here
					c.clientIdCnt++
				} else {
					c.handleDead(hostPort)
					c.NewClient(w,r)
				}
				break
			} else {
				serverIdx--
			}
		}
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
			c.handleDead(dstHostPort)
			return err
		}
		common.LOGV.Println("$Broadcast AddedDoc[",dstHostPort,"]: ",reply)
		// Check reply from Master
		if reply.Status == serverrpc.OK {
			return nil
		} else if reply.Status == serverrpc.DocExist {
			common.LOGE.Println("Doc ",docId," Already Exist")
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
		if err := client.Call("PearServer.RemovedDoc", args, &reply); err != nil {
			c.handleDead(dstHostPort)
			return err
		}
		common.LOGV.Println("$Broadcast RemovedDoc[",dstHostPort,"]: ",reply)
		// Check reply from Master
		if reply.Status == serverrpc.OK {
			return nil
		} else if reply.Status == serverrpc.DocNotExist {
			common.LOGE.Println("Doc ",docId," does not exist.")
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (c *central) sendCheckAlive(dstHostPort string) error {
	client, err := c.dialRPC(dstHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return err
	}
	
	for {
		// Make RPC Call to Master
		args := &serverrpc.CheckAliveArgs{
			HostPort: c.myHostPort,
		}
		var reply serverrpc.CheckAliveReply
		if err := client.Call("PearServer.CheckAlive", args, &reply); err != nil {
			c.handleDead(dstHostPort)
			return err
		}
		common.LOGV.Println("$CheckAlive[",dstHostPort,"]: ",reply)
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
				c.handleDead(dstHostPort)
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

func (c *central) handleDead(serverId string) {
	common.LOGV.Println("Handling Dead ",serverId)
	docList, ok := c.serverMap[serverId]
	if ok {
		// Update Everyone: Remove all docs from dead server
		for docId ,_ := range docList {
			teammate, ok2 := c.docMap[docId]
			if ok2 {
				// Update all collaborators for relevant doc
				for k , _ := range teammate {
					if k != serverId {
						err := c.sendRemoveDoc(k,serverId, docId)
						if err != nil {
							common.LOGE.Println("Gave up broadcast Dead server")
							return
						}
					}
				}
				// Remove dead server from docMap
				delete(teammate, serverId)
				c.docMap[docId] = teammate
			}
		}
		// Remove dead server from serverMap
		delete(c.serverMap,serverId)
	} else {
		common.LOGE.Println("Server ",serverId," does not exist")
	}
}