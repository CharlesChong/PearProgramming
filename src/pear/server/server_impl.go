
package server

type server struct {
}

func NewServer() (Server, error) {
	ps := server{}
	return &ps, nil
}