
package centralrpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK           Status = iota + 1 // The RPC was a success.
	NotReady                       // The storage servers are still getting ready.
)

type ServerId string
type Message string

type Node struct {
	HostPort string // The host:port address of the storage server node.
}

type AddServerArgs struct {
	HostPort string
}

type AddServerReply struct {
	Status Status
}

type AddDocArgs struct {
	DocId string
	HostPort string
}

type AddDocReply struct {
	Teammates []string
}

type RemoveDocArgs struct {
	DocId string
	HostPort string
}

type RemoveDocReply struct {
	Status Status
}
