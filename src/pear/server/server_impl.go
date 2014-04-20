
package server

import (
	"pear/rpc/serverrpc"
	"net/rpc"
	"net/http"
	"net"
	"fmt"
)

type server struct {

}

func NewServer(masterServerHostPort string, numNodes, port int, nodeID uint32) (Server, error) {
	ps := server{}
	fmt.Println("Pear Server startin")
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

	return &ps, nil
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
