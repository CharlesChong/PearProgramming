package centralrpc

type RemoteCentral interface {
	AddDoc(*AddDocArgs,*AddDocReply) error
	RemoveDoc(*RemoveDocArgs,*RemoveDocReply) error
	AddServer(*AddServerArgs,*AddServerReply) error
}

type Central struct {
	// Embed all methods into the struct. See the Effective Go section about
	// embedding for more details: golang.org/doc/effective_go.html#embedding
	RemoteCentral
}

// Wrap wraps s in a type-safe wrapper struct to ensure that only the desired
// StorageCentral methods are exported to receive RPCs.
func Wrap(c RemoteCentral) RemoteCentral {
	return &Central{c}
}
