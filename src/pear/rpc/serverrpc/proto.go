package serverrpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK       Status = iota + 1 // The RPC was a success.
	NotReady                   // The storage servers are still getting ready.
	DocExist
	DocNotExist
	InvalidServer
	ClientExist
	ClientNotExist
)


type Message struct {
	TId		string
	Body 	string
}

func (msg *Message) ToString () string {
	return msg.TId + " " + msg.Body
}

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
	DocId 		string
	HostPort 	string
}

type GetDocReply struct {
	Doc    string
	DocId  string
	Status Status
}

type VoteArgs struct {
	Msg 		*Message
	DocId 		string
	HostPort 	string
}

type VoteReply struct {
	Vote 		bool
	Msg  		*Message
	Status 		Status
}

type CompleteArgs struct {
	Commit	 	bool
	DocId 	 	string
	HostPort 	string
	Msg      	*Message
}

type CompleteReply struct {
	Msg 	*Message
	Status 	Status
}
