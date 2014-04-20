
package serverrpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK           Status = iota + 1 // The RPC was a success.
	NotReady                       // The storage servers are still getting ready.
)

type Message string

type Node struct {
	HostPort string // The host:port address of the storage server node.
	NodeID   uint32 // The ID identifying this storage server node.
}

type RegisterArgs struct {
	ServerInfo Node
}

type RegisterReply struct {
	Status  Status
	Servers []Node
}

type VoteArgs struct {
	Msg Message
}

type VoteReply struct {
	Vote bool
	Msg Message
}

type CompleteArgs struct {
	Rollback bool
	Msg Message
}

type CompleteReply struct {
	Msg Message
}