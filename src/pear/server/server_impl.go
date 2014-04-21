
package server

import (
	"pear/rpc/serverrpc"
	"net/rpc"
	"net/http"
	"net"
	"common"
	"time"
	"fmt"
)

type server struct {
	servers              map[serverrpc.Node]bool
	serverList           []serverrpc.Node
	masterServerHostPort string
	numNodes             int
	port                 int
	nodeID               uint32 // Upper bound hash range
}

func NewServer(masterServerHostPort string, numNodes, port int, nodeID uint32) (Server, error) {
	ps := server{}
	ps.servers = make(map[serverrpc.Node]bool)
	ps.serverList = []serverrpc.Node{}
	ps.masterServerHostPort = masterServerHostPort
	ps.numNodes = numNodes
	ps.port = port
	ps.nodeID = nodeID
	common.LOGV.Println("Pear Server startin")
	myHostPort := fmt.Sprintf("localhost:%d", port)

	// Create the server socket that will listen for incoming RPCs.
	listener, err := net.Listen("tcp", myHostPort)
	if err != nil {
		return nil, err
	}

	// Wrap the tribServer before registering it for RPC.
	err = rpc.RegisterName("PearServer", serverrpc.Wrap(ps))
	if err != nil {
		return nil, err
	}

	// Setup the HTTP handler that will server incoming RPCs and
	// serve requests in a background goroutine.
	rpc.HandleHTTP()
	go http.Serve(listener, nil)

	// Start with 1 coordinator and register all machines
	// After complete respond with all servers
	if len(masterServerHostPort) == 0 {
		err = coordinatorInit(&ps, myHostPort)
	} else {
		err = participantInit(&ps, myHostPort)
	}
	if err != nil {
		common.LOGE.Println(err)
		return nil, err
	}

	return &ps, nil
}

func participantInit(ps *server, myHostPort string) error {
	common.LOGV.Println("participant Init:", ps.nodeID, ps.port)
	var client *rpc.Client
	var err error

	maxFail := 5
	for tries := 0; tries < maxFail; tries++ {
		client, err = rpc.DialHTTP("tcp", ps.masterServerHostPort)
		if err != nil {
			common.LOGE.Printf("dialing:", err)
			if tries >= maxFail {
				return err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	for {
		// Make RPC Call to Master
		srvInfo := serverrpc.Node{HostPort: myHostPort,
			NodeID: ps.nodeID}
		args := &serverrpc.RegisterArgs{ServerInfo: srvInfo}
		var reply serverrpc.RegisterReply
		if err := client.Call("PearServer.RegisterParticipant", args, &reply); err != nil {
			return err
		}

		// Check reply from Master
		if reply.Status == serverrpc.OK {
			// findHashRange(ss)
			return nil
		}

		time.Sleep(time.Second)
	}
}

func coordinatorInit(ps *server, myHostPort string) error {

	// Register Coordinator into Servers
	srvInfo := serverrpc.Node{HostPort: myHostPort, NodeID: ps.nodeID}
	ps.servers[srvInfo] = true
	ps.serverList = append(ps.serverList, srvInfo)

	// Sleep until all servers are done
	for len(ps.servers) < ps.numNodes {
		time.Sleep(time.Second)
	}

	common.LOGV.Println("Ring Initialization Complete")

	// findHashRange(ss)

	return nil

}

func (ps server) RegisterParticipant(args *serverrpc.RegisterArgs,reply *serverrpc.RegisterReply) error {
	common.LOGV.Println("Register Server ", args.ServerInfo.NodeID)

	ps.servers[args.ServerInfo] = true

	if len(ps.servers) == ps.numNodes {
		// All Servers complete, compile list
		reply.Status = serverrpc.OK
		for k, _ := range ps.servers {
			if k.NodeID != ps.nodeID {
				ps.serverList = append(ps.serverList, k)
			}
		}
		reply.Servers = ps.serverList
		common.LOGV.Println("READY BABE")

	} else {
		reply.Status = serverrpc.NotReady
		reply.Servers = []serverrpc.Node{}
	}

	return nil
}

func (ps server) VotePhase(args *serverrpc.VoteArgs, reply *serverrpc.VoteReply) error {
	fmt.Println("Vote: ",args.Msg)
	reply.Vote = true
	reply.Msg = args.Msg
	fmt.Println("Vote Rsp: ",reply.Msg, " ",reply.Vote)
	return nil
}

func (ps server) CompletePhase(args *serverrpc.CompleteArgs, reply *serverrpc.CompleteReply) error {
	fmt.Println("Cmp: ",args.Msg," rollback?",args.Rollback)
	reply.Msg = args.Msg
	fmt.Println("Cmp Rsp ",reply.Msg)
	return nil
}
