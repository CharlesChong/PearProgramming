package server

import (
    "code.google.com/p/go.net/websocket"
	"common"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"pear/rpc/centralrpc"
	"pear/rpc/serverrpc"
	"strconv"
	"time"
)

type server struct {
	centralHostPort string
	myHostPort		string
	port            int
	// clientID -> client struct with docs
	// client struct in clientHandlers.go file
	clients         map[string]*client 
	// doc -> clientList (clientID)
	documents  map[string]map[string]bool
	// connection map remembers old clients
	connMap         map[string]*rpc.Client
	// doc -> serverID (hostport) -> exists bool
	docToServerMap  map[string]map[string]bool
}

func NewServer(centralHostPort string, port int) (Server, error) {
	ps := server{}
	ps.centralHostPort = centralHostPort
	ps.myHostPort = fmt.Sprintf("localhost:%d", port)
	ps.port = port
	ps.clients = make(map[string]*client)
	ps.documents = make(map[string]map[string]bool)
	ps.connMap = make(map[string]*rpc.Client)
	ps.docToServerMap = make(map[string]map[string]bool)	

	// Create the server socket that will listen for incoming RPCs.
	listener, err := net.Listen("tcp", ps.myHostPort)
	if err != nil {
		return nil, err
	}

	// Wrap the tribServer before registering it for RPC.
	err = rpc.RegisterName("PearServer", serverrpc.Wrap(&ps))
	if err != nil {
		return nil, err
	}

	// Setup the HTTP handler that will server incoming RPCs and
	// serve requests in a background goroutine.
	rpc.HandleHTTP()
	go http.Serve(listener, nil)

	// err = coordinatorInit(&ps, ps.myHostPort)
	err = participantInit(&ps, ps.myHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return nil, err
	}

	// Test Code here! TODO: Remove
	err = sendAddDoc(&ps,ps.myHostPort,"Hello")
	if ps.myHostPort == "localhost:9001" {
		common.LOGV.Println("Testing Remove")
		err = sendRemoveDoc(&ps,ps.myHostPort, "Hello")
	}

	http.Handle("/", websocket.Handler(ps.clientConnHandler))
	go http.ListenAndServe(":" + strconv.Itoa(port), nil)

	return &ps, nil
}

func participantInit(ps *server, myHostPort string) error {
	client, err := ps.dialRPC(ps.centralHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return err
	}
	for {
		// Make RPC Call to Master
		args := &centralrpc.AddServerArgs{
			HostPort: myHostPort,
		}
		var reply centralrpc.AddServerReply
		if err := client.Call("PearCentral.AddServer", args, &reply); err != nil {
			return err
		}
		// Check reply from Master
		if reply.Status == centralrpc.OK {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func (ps *server) AddedDoc(args *serverrpc.AddedDocArgs, reply *serverrpc.AddedDocReply) error {
	common.LOGV.Println("Added Doc",args)
	reply.DocId = args.DocId
	reply.Teammates = make(map[string]bool)
	reply.Status = serverrpc.OK

	serverMap, ok := ps.docToServerMap[args.DocId]
	if !ok {
		newMap := make(map[string]bool)
		newMap[args.HostPort] = true
		ps.docToServerMap[args.DocId] = newMap
		reply.Teammates = newMap
	} else {
		_, ok2 := serverMap[args.HostPort]
		if !ok2 {
			serverMap[args.HostPort] = true
			ps.docToServerMap[args.DocId] = serverMap
		} else {
			reply.Status = serverrpc.DocExist
		}
	}
	return nil
}

func (ps *server) RemovedDoc(args *serverrpc.RemovedDocArgs, reply *serverrpc.RemovedDocReply) error {
	common.LOGV.Println("Removed Doc: ", args)
	reply.DocId = args.DocId
	reply.Status = serverrpc.OK

	serverMap, ok := ps.docToServerMap[args.DocId]
	if ok {
		_, ok2 := serverMap[args.HostPort]
		if ok2 {
			delete(serverMap, args.HostPort)
			ps.docToServerMap[args.DocId] = serverMap
			return nil
		}
	}
	reply.Status = serverrpc.DocNotExist

	return nil
}

func (ps *server) GetDoc(args *serverrpc.GetDocArgs, reply *serverrpc.GetDocReply) error {
	common.LOGV.Println("GetDoc: ", args)
	reply.DocId = args.DocId
	
	reply.Doc = "Fake Doc"
	reply.Status = serverrpc.OK
	return nil
}

func (ps *server) VotePhase(args *serverrpc.VoteArgs, reply *serverrpc.VoteReply) error {
	common.LOGV.Println("Vote: ", args.Msg)
	reply.Vote = true
	reply.Msg = args.Msg
	return nil
}

func (ps *server) CompletePhase(args *serverrpc.CompleteArgs, reply *serverrpc.CompleteReply) error {
	common.LOGV.Println("Cmp: ", args.Msg, " rollback?", args.Rollback)
	reply.Msg = args.Msg
	return nil
}

func sendAddDoc(ps *server, myHostPort,docId string) error {
	client, err := ps.dialRPC(ps.centralHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return err
	}

	for {
		// Make RPC Call to Master
		args := &centralrpc.AddDocArgs{
			DocId: docId,
			HostPort: myHostPort,
		}
		var reply centralrpc.AddDocReply
		if err := client.Call("PearCentral.AddDoc", args, &reply); err != nil {
			return err
		}
		common.LOGV.Println("Call: AddDoc[",ps.centralHostPort,"]: ",reply)
		// Check reply from Master
		if reply.Status == centralrpc.OK {
			return nil
		}

		time.Sleep(time.Second)
	}
}

func sendRemoveDoc(ps *server, myHostPort, docId string) error {
	client, err := ps.dialRPC(ps.centralHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return err
	}

	for {
		// Make RPC Call to Master
		args := &centralrpc.RemoveDocArgs{
			DocId: docId,
			HostPort: myHostPort,
		}
		var reply centralrpc.RemoveDocReply
		if err := client.Call("PearCentral.RemoveDoc", args, &reply); err != nil {
			return err
		}

		// Check reply from Master
		if reply.Status == centralrpc.OK {
			common.LOGV.Println(reply)
			return nil
		}

		time.Sleep(time.Second)
	}
}

func (ps *server) dialRPC(dstHostPort string) (*rpc.Client, error) {
	// Check if old connection exists
	oldClient, ok := ps.connMap[dstHostPort]
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
			// Store in conn Map
			ps.connMap[dstHostPort] = client
			return client, nil
		}
	}
}

func (ps *server) AddClient(clientId, docId string) serverrpc.Status {
	clientList, ok := ps.documents[docId]
	if ok {
		_, ok2 := clientList[clientId]
		if !ok2 {
			clientList[clientId] = true
			ps.documents[docId] = clientList	
		} else {
			return serverrpc.ClientExist
		}
		
	} else {
		newClientList := make(map[string]bool)
		newClientList[clientId] = true
		ps.documents[docId] = newClientList
	}
	return serverrpc.OK
}
