package serverrpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK           Status = iota + 1 // The RPC was a success.
	NotReady                       // The storage servers are still getting ready.
)

type DocId string
type Doc string
type Message string

type AddedDocArgs struct {
	DocId		DocId
	HostPort	string
}

type AddedDocReply struct {
	DocId   DocId
	Status  Status
}

type RemovedDocArgs struct {
	DocId   	DocId
	HostPort	string
}

type RemovedDocReply struct {
	DocId   DocId
	Status  Status
}

type GetDocArgs struct {
	DocId   DocId
}

type GetDocReply struct {
	Doc     Doc
	DocId   DocId
	Status  Status
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