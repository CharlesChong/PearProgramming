package server

import (
    "code.google.com/p/go.net/websocket"
	"common"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"pear/rpc/centralrpc"
	"pear/rpc/serverrpc"
	"strconv"
	"time"
)

type rpcFn func(*rpc.Client,string, string) error 

type server struct {
	centralHostPort string
	myHostPort		string
	port            int
	// clientID -> client struct with docs
	// client struct in clientHandlers.go file
	clients         map[string]*client 
	// doc -> clientList (clientID)
	documents  		map[string]map[string]bool
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

	// Test Code here!
	// TODO Remove test code
	// Test 1: Add doc
	// err = ps.sendAddDoc("Hello")
	// if ps.myHostPort == "localhost:9001" {
	// 	common.LOGV.Println("Testing Remove")
	// 	err = ps.sendRemoveDoc("Hello")
	// }

	// Test 2: Get Doc
	// var doc string
	// err = ps.sendAddDoc("Hello")
	// if ps.myHostPort == "localhost:9001" {
	// 	doc , err = ps.ClientGetDoc("Hello")
	// 	if err != nil {
	// 		common.LOGE.Println(err)
	// 		return nil, err
	// 	}
	// 	common.LOGV.Println("Result: ",doc)
	// }

	// Test 3: 2PC begin
	// err = ps.sendAddDoc("Hello")
	// time.Sleep(time.Second*5)
	// if ps.myHostPort == "localhost:9001" {
	// 	msg := serverrpc.Message {
	// 		TId : "Yo dis my Tid",
	// 		Body : "HEY THAR",
	// 	}
	// 	_, err  = ps.ClientRequestTxn(&msg,"Hello")
	// }


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

	clientList, ok := ps.documents[args.DocId]
	if ok {
		for client, _ := range clientList {
			doc, err := ps.clients[client].sendRequest(getDocCmd, "")
			if err == nil {
				reply.Doc = doc
				reply.Status = serverrpc.OK
				return nil
			} 
			break
		}
	} 
	reply.Doc = "ERROR"
	reply.Status = serverrpc.NotReady
	return nil
}

func (ps *server) VotePhase(args *serverrpc.VoteArgs, reply *serverrpc.VoteReply) error {
	common.LOGV.Println("Vote: ", args.Msg)
	reply.Msg = args.Msg
	var vote bool
	var finalVote bool
	clientList, ok := ps.documents[args.DocId]
	if ok {
		for client, _ := range clientList {
			common.LOGV.Println("Before",client,":",args.Msg.ToString())
			rsp, err := ps.clients[client].sendRequest(voteCmd, args.Msg.ToString())
			common.LOGV.Println(rsp)
			if err == nil {
				vote ,err = strconv.ParseBool(rsp)
				if err != nil {
					common.LOGE.Println("Invalid Vote",rsp)
					reply.Status = serverrpc.NotReady
					reply.Vote = false
					return nil
				} else {
					if !vote {
						finalVote = false
					}
				}
			} else {
				common.LOGE.Println(err)
				reply.Vote = false
				reply.Status = serverrpc.NotReady
				return nil
			}
		}
		reply.Vote = finalVote
		reply.Status = serverrpc.OK
		return nil
	} 
	reply.Vote = false
	reply.Status = serverrpc.NotReady
	return nil
}

func (ps *server) CompletePhase(args *serverrpc.CompleteArgs, reply *serverrpc.CompleteReply) error {
	common.LOGV.Println("Complete: ", args.Msg, " commit?", args.Commit)
	reply.Msg = args.Msg
	clientList, ok := ps.documents[args.DocId]
	if ok {
		for client, _ := range clientList {
			args.Msg.Body = strconv.FormatBool(args.Commit)
			rsp, err := ps.clients[client].sendRequest(completeCmd, args.Msg.ToString())
			if rsp != "ok" {
				reply.Status = serverrpc.NotReady
				return nil
			}
			if err != nil {
				common.LOGE.Println(err)
				reply.Status = serverrpc.NotReady
				return nil
			}
		}
		reply.Status = serverrpc.OK
	} else {
		reply.Status = serverrpc.NotReady
	}
	return nil
}

///////////////////// Client Handler Calls ///////////////////
/////////////////////// Sending RPC Calls //////////////////////

func (ps *server) sendAddDoc(docId string) error {
	err := ps.makeRPCCall(ps.RPCAddDoc,ps.centralHostPort,docId)
	return err
}

func (ps *server) sendRemoveDoc(docId string) error {
	err := ps.makeRPCCall(RPCRemoveDoc,ps.centralHostPort,docId)
	return err
}

// Client Request Txn:
// Begin 2PC
// Returns true when will Commit
func (ps *server) ClientRequestTxn(msg *serverrpc.Message,docId string) (bool,error) {
	willCommit := true
	serverList, ok := ps.docToServerMap[docId]
	rpcCh := make(chan bool, len(serverList))
	errorCh := make(chan error)
	if ok {
		// 1. Broadcast Vote to all collaborators
		for s , _ := range serverList {
			go ps.makeRPCCall(RPCVote(rpcCh,errorCh,msg),s,docId)
		}

		for _ , _ = range serverList {
			select {
			case res := <- rpcCh:
				if res == false {
					willCommit = false
				}
			case err := <- errorCh:
				close(rpcCh)
				return false, err
			}
		}
		// 2. Broadcast Complete Phase to all collaborators
		for s , _ := range serverList {
			go ps.makeRPCCall(RPCComplete(willCommit,msg),s,docId)
		}
		// 3. Notify own clients about Complete phase
		return willCommit, nil
	} else {
		return false, errors.New("Invalid doc to Server Map")
	}
}

func (ps *server) ClientGetDoc(docId string) (string ,error) {
	// Check if server has other clients with document
	resCh := make(chan string)
	clientList, ok := ps.documents[docId]
	if ok {
		for client, _ := range clientList {
			if client != ps.myHostPort {
				doc, err := ps.clients[client].sendRequest(getDocCmd, "")
				if err != nil {
					return "", err
				} else {
					return doc, nil
				}
			}
		}
	} 
	// Ask Another server for document
	serverList, ok2 := ps.docToServerMap[docId]
	if ok2 {
		for server, _ := range serverList {
			if server != ps.myHostPort {
				go ps.makeRPCCall(RPCGetDoc(resCh),server,docId)
				doc := <- resCh
				return doc, nil
			}
		}
		// No Other Pear Servers has Document -> New Document Created
		return "NEW DOCUMENT: " + docId, nil
	} else {
		// New Document
		return "NEW DOCUMENT FROM ERROR: " + docId, errors.New("Doc Not Registered")
	}
	
}

func (ps *server) makeRPCCall(rpcCall rpcFn,dstHostPort, docId string) error {
	rpcClient, err := ps.dialRPC(dstHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return err
	}
	err = rpcCall(rpcClient,docId,ps.myHostPort)
	return err
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

func (ps *server) RPCAddDoc(client *rpc.Client, docId, myHostPort string) error  {
	// Pear Server -> Pear Central: Requesting Add Fresh New Document
	for {
		// Make RPC Call to Master
		args := &centralrpc.AddDocArgs{
			DocId: docId,
			HostPort: myHostPort,
		};
		var reply centralrpc.AddDocReply
		if err := client.Call("PearCentral.AddDoc", args, &reply); err != nil {
			return err
		}
		common.LOGV.Println("Call:",":",reply)
		// Check reply from Master
		if reply.Status == centralrpc.OK {
			_, ok := ps.docToServerMap[docId]
			if !ok {
				ps.docToServerMap[docId] = make(map[string]bool)
			}
			for k,_ := range reply.Teammates {
				ps.docToServerMap[docId][k] = true
			}
			return nil
		}
		time.Sleep(time.Second)
	}
}

func RPCRemoveDoc(client *rpc.Client, docId, myHostPort string) error  {
	// Pear Server -> Pear Central: Removing Existing Client/Hostport combo
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

func RPCGetDoc (resCh chan string) rpcFn {
	return func (client *rpc.Client, docId, myHostPort string) error  {
		// Pear Server -> Pear Server: Get Existing doc from another srv
		for {
			// Make RPC Call to Master
			args := &serverrpc.GetDocArgs{
				DocId: docId,
				HostPort: myHostPort,
			}
			var reply serverrpc.GetDocReply
			if err := client.Call("PearServer.GetDoc", args, &reply); err != nil {
				return err
			}
			// Check reply from Master
			if reply.Status == serverrpc.OK {
				resCh <- reply.Doc
				return nil
			}
			time.Sleep(time.Second)
		}
	}
}

func RPCVote(rspChan chan bool,errorCh chan error, msg *serverrpc.Message) rpcFn {
	// Pear Server -> Pear Server: Get Vote for people for txn
	return func (client *rpc.Client, docId, myHostPort string) error  {
		for {
			// Make RPC Call to Master
			args := &serverrpc.VoteArgs{
				DocId: docId,
				HostPort: myHostPort,
				Msg: msg,
			}
			var reply serverrpc.VoteReply
			if err := client.Call("PearServer.VotePhase", args, &reply); err != nil {
				errorCh <- err
				return err
			}
			common.LOGV.Println("Call: VotePhase ",reply)
			// Check reply from Master
			if reply.Status == serverrpc.OK {
				rspChan <- reply.Vote
				return nil
			}
			common.LOGV.Println(reply.Status)
			time.Sleep(time.Second)
		}
	}
}

func RPCComplete(commit bool,msg *serverrpc.Message) rpcFn {
	// Pear Server -> Pear Server: Get Complete for people for txn
	return func (client *rpc.Client, docId, myHostPort string) error  {
		for {
			// Make RPC Call to Master
			args := &serverrpc.CompleteArgs{
				Commit: commit,
				DocId: docId,
				HostPort: myHostPort,
				Msg: msg,
			}
			var reply serverrpc.CompleteReply
			if err := client.Call("PearServer.CompletePhase", args, &reply); err != nil {
				return err
			}
			common.LOGV.Println("Call: CompletePhase ",reply)
			// Check reply from Master
			if reply.Status == serverrpc.OK {
				return nil
			}
			common.LOGV.Println(reply.Status)
			time.Sleep(time.Second)
		}
	}
}
