package server

import (
    "code.google.com/p/go.net/websocket"
    "common"
)

type client struct {
    DocId int
    ws    *websocket.Conn
}

func (ps *server) ClientHandler(ws *websocket.Conn) {
    var clientId string
    err := websocket.Message.Receive(ws, &clientId)
    common.LOGV.Println(clientId)
    /*if err == io.EOF {
        return
    } else */if err != nil {
        common.LOGE.Println("Websocket error: " + err.Error())
        return
    }
    var docId string
    err = websocket.Message.Receive(ws, &docId)
    if err != nil {
        common.LOGE.Println("Websocket error: " + err.Error())
        return
    }
    common.LOGV.Println(docId)
}