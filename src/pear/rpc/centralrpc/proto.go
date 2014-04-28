package centralrpc

// Status represents the status of a RPC's reply.
type Status int

const (
	OK       Status = iota + 1 // The RPC was a success.
	NotReady                   // The storage servers are still getting ready.
	DocExist
	DocNotExist
)

type AddDocArgs struct {
	DocId    string
	HostPort string
}

type AddDocReply struct {
	DocId     string
	Teammates map[string]bool
	Status    Status
}

type RemoveDocArgs struct {
	DocId    string
	HostPort string
}

type RemoveDocReply struct {
	DocId  string
	Status Status
}

type AddServerArgs struct {
	HostPort string
}

type AddServerReply struct {
	Status Status
}
