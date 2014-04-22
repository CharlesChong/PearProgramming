package server

import (
	"code.google.com/p/go.net/websocket"
	"io"
)

type client struct {
	DocId int
	ws    *websocket.Conn
}

func ClientHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}
