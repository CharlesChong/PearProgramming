package server

import (
    "code.google.com/p/go.net/websocket"
    "common"
)

type client struct {
    clientId string
    docId string
    ws    *websocket.Conn
}

func (ps *server) ClientHandler(ws *websocket.Conn) {
    var c = client{}
    c.ws = ws
    err := websocket.Message.Receive(ws, &c.clientId)
    /*if err == io.EOF {
        return
    } else */if err == nil {
        err = websocket.Message.Receive(ws, &c.docId)
    }
    if err == nil {
        // $TODO: Get text for doc
        err = websocket.Message.Send(ws, "setDoc    This is the text for " + c.docId)
    }    
    if err != nil {
        common.LOGE.Println("Websocket error during setup: " + err.Error())
        return
    }

    //io.Copy(ws, ws)
}