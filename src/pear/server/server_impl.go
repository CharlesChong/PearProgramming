package server

import (
    "code.google.com/p/go.net/websocket"
	"pear/rpc/serverrpc"
	"pear/rpc/centralrpc"
	"net/rpc"
	"net/http"
	"log"
	"strconv"
	"net"
	"common"
	"time"
	"fmt"
)

type server struct {
	centralHostPort 	 string
	port                 int
	// map[client] document
	// map[document] client
	// map[server] conn
	// map[document] server_list
}

func NewServer(centralHostPort string,port int) (Server, error) {
	common.LOGV.Println("Pear Server starting")
	ps := server{}
	ps.centralHostPort = centralHostPort
	ps.port = port
	
	myHostPort := fmt.Sprintf("localhost:%d", port)

	// Create the server socket that will listen for incoming RPCs.
	listener, err := net.Listen("tcp", myHostPort)
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

	// err = coordinatorInit(&ps, myHostPort)
	err = participantInit(&ps, myHostPort)
	if err != nil {
		common.LOGE.Println(err)
		return nil, err
	}

	http.Handle("/", websocket.Handler(ps.ClientHandler))
	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(port), nil))

	return &ps, nil
}

func participantInit(ps *server, myHostPort string) error {
	common.LOGV.Println("Participant Init:", ps.port)
	var client *rpc.Client
	var err error

	maxFail := 5
	for tries := 0; tries < maxFail; tries++ {
		client, err = rpc.DialHTTP("tcp", ps.centralHostPort)
		if err != nil {
			common.LOGE.Printf("Dialing:", err)
			if tries >= maxFail {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	common.LOGV.Println("Connection Made")

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
			common.LOGV.Println("Registered")
			return nil
		}

		time.Sleep(time.Second)
	}
}

func (ps *server) AddedDoc(args *serverrpc.AddedDocArgs,reply *serverrpc.AddedDocReply) error {
	
	return nil
}

func (ps *server) RemovedDoc(args *serverrpc.RemovedDocArgs,reply *serverrpc.RemovedDocReply) error {

	return nil
}

func (ps *server) GetDoc(args *serverrpc.GetDocArgs,reply *serverrpc.GetDocReply) error {
	reply.DocId = args.DocId
	reply.Doc = "Fake Doc"
	reply.Status = serverrpc.OK
	return nil
}

func (ps *server) VotePhase(args *serverrpc.VoteArgs, reply *serverrpc.VoteReply) error {
	fmt.Println("Vote: ",args.Msg)
	reply.Vote = true
	reply.Msg = args.Msg
	fmt.Println("Vote Rsp: ",reply.Msg, " ",reply.Vote)
	return nil
}

func (ps *server) CompletePhase(args *serverrpc.CompleteArgs, reply *serverrpc.CompleteReply) error {
	fmt.Println("Cmp: ",args.Msg," rollback?",args.Rollback)
	reply.Msg = args.Msg
	fmt.Println("Cmp Rsp ",reply.Msg)
	return nil
}
