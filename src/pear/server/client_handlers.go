package server

import (
    "code.google.com/p/go.net/websocket"
    "io"
)

func ClientHandler(ws *websocket.Conn) {
    io.Copy(ws, ws)
}