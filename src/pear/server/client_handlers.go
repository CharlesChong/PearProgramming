package server

import (
    "code.google.com/p/go.net/websocket"
    "common"
    "io"
    "time"
)

type client struct {
    clientId string
    docId string
    ws    *websocket.Conn
}

func (ps *server) clientConnHandler(ws *websocket.Conn) {
    var c = client{}
    c.ws = ws
    err := websocket.Message.Receive(ws, &c.clientId)
    if err == nil {
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

    ps.clients[c.clientId] = &c
    go ps.clientReadHandler(c.clientId)
}

func (ps *server) clientReadHandler (clientId string) {
    c := ps.clients[clientId]
    for {
        //c.ws.SetDeadline(time.Now().Add(time.Minute))
        var msg string
        err := websocket.Message.Receive(c.ws, &msg)
        if err != nil {
            if err != io.EOF {
                common.LOGV.Println("Websocket error: " + err.Error())
            }
            ps.closeClient(clientId)
            return
        }
        if len(msg) < 10 {
            common.LOGE.Println("Received invalid command from client " + clientId + ": " + msg)
        } else {
            command := msg[0:10]
            args := msg[10:len(msg)]
            common.LOGV.Println(command + ":" + args) // $
            switch command {
            case "setDoc    ":
            case "getDoc    ":
            case "vote      ":
            case "comple    ":
            case "requestTxn":
            default:
                common.LOGE.Println("Received unrecognized command from client " + clientId + ": " + msg)
            }
        }
    }
}

func (ps *server) closeClient (clientId string) {
}