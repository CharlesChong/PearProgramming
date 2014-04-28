package server

import (
    "code.google.com/p/go.net/websocket"
    "common"
    "io"
)

type client struct {
    clientId string
    docId string
    ws    *websocket.Conn
}

func (ps *server) clientConnHandler(ws *websocket.Conn) {
    // Setup client
    var c = client{}
    var clientAck string
    c.ws = ws
    err := websocket.Message.Receive(ws, &c.clientId)
    if err == nil {
        err = websocket.Message.Receive(ws, &c.docId)
    }
    if err == nil {
        // $TODO: Get text for doc
        err = websocket.Message.Send(ws, "setDoc    This is the text for " + c.docId)
    }
    if err == nil {
        err = websocket.Message.Receive(c.ws, &clientAck)
    }
    if err != nil {
        common.LOGE.Println("Websocket error during setup: " + err.Error())
        return
    }
    if (clientAck != "setDoc    ok") {
        common.LOGE.Println("Did not get setup ack from client");
        return
    }
    // Store information
    ps.clients[c.clientId] = &c
    clientList, ok := ps.documents[c.docId]
    if ok {
        // Doc exists on server
        _, ok2 := clientList[c.clientId]
        if !ok2 {
            clientList[c.clientId] = true
            ps.documents[c.docId] = clientList    
        } else {
            common.LOGE.Println("Client ID already exists")
            return
        }
    } else {
        // New doc on server
        newClientList := make(map[string]bool)
        newClientList[c.clientId] = true
        ps.documents[c.docId] = newClientList
        err = sendAddDoc(ps,ps.myHostPort,c.docId)
        if err != nil {
            common.LOGE.Println("Error adding doc: " + err.Error())
        }
    }

    ps.clientReadHandler(&c)
}

func (ps *server) clientReadHandler(c *client) {
    // Read handler
    for {
        var msg string
        err := websocket.Message.Receive(c.ws, &msg)
        if err != nil {
            if err != io.EOF {
                common.LOGV.Println("Websocket error: " + err.Error())
            }
            ps.closeClient(c)
            return
        }
        if len(msg) < 10 {
            common.LOGE.Println("Received invalid command from client " + c.clientId + ": " + msg)
        } else {
            command := msg[0:10]
            args := msg[10:len(msg)]
            common.LOGV.Println(command+":"+args)
            switch command {
            case "getDoc    ":
            case "vote      ":
            case "complete  ":
            case "requestTxn":
            default:
                common.LOGE.Println("Received unrecognized command from client " + c.clientId + ": " + msg)
            }
        }
    }

}

func (ps *server) closeClient (c *client) {
    clientList, ok := ps.documents[c.docId]
    if ok {
        delete(clientList, c.clientId)
        if len(clientList) == 0 {
            delete(ps.documents, c.docId)
            err := sendRemoveDoc(ps,ps.myHostPort,c.docId)
            if err != nil {
                common.LOGE.Println("Error removing doc: " + err.Error())
            }
        }
    } else {
        common.LOGE.Println("Unrecorded client has closed")
    }
    delete(ps.clients, c.clientId)
}