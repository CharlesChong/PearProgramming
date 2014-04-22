package serverrpc

type RemoteServer interface {
	VotePhase(*VoteArgs, *VoteReply) error
	CompletePhase(*CompleteArgs, *CompleteReply) error
	AddedDoc(*AddedDocArgs, *AddedDocReply) error
	RemovedDoc(*RemovedDocArgs, *RemovedDocReply) error
	GetDoc(*GetDocArgs, *GetDocReply) error
}

type Server struct {
	// Embed all methods into the struct. See the Effective Go section about
	// embedding for more details: golang.org/doc/effective_go.html#embedding
	RemoteServer
}

// Wrap wraps s in a type-safe wrapper struct to ensure that only the desired
// StorageServer methods are exported to receive RPCs.
func Wrap(s RemoteServer) RemoteServer {
	return &Server{s}
}
