package serverrpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK       Status = iota + 1 // The RPC was a success.
	NotReady                   // The storage servers are still getting ready.
	DocExist
	DocNotExist
	InvalidServer
)

type Message string

type AddedDocArgs struct {
	DocId    string
	HostPort string
}

type AddedDocReply struct {
	DocId     string
	Teammates map[string]bool
	Status    Status
}

type RemovedDocArgs struct {
	DocId    string
	HostPort string
}

type RemovedDocReply struct {
	DocId  	string
	Status 	Status
}

type GetDocArgs struct {
	DocId 	string
}

type GetDocReply struct {
	Doc    string
	DocId  string
	Status Status
}

type VoteArgs struct {
	Msg Message
}

type VoteReply struct {
	Vote bool
	Msg  Message
}

type CompleteArgs struct {
	Rollback bool
	Msg      Message
}

type CompleteReply struct {
	Msg Message
}
