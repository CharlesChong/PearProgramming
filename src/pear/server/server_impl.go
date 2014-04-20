
package server

type server struct {
}

func NewServer(masterServerHostPort string, numNodes, port int, nodeID uint32) (Server, error) {
	ps := server{}
	return &ps, nil
}
