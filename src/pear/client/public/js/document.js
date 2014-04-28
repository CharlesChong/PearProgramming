var ws;
var clientId;
var editor;
var settingDoc = false;
var transactionNum = 0;
var committed = null;
var committing = null;
var currTransactionId = null;

$(function(){
    $.get("http://" + centralHostPort, {docId: docId})
        .done(function(data) {
            var reply = data.split(" ", 2);
            clientId = (reply[0]);
            setupServer(reply[1]);
        }).fail(function(data) {
            alert("Failed to retrieve server information");
            console.log(data);
        });
    setupGUI();
});

function setupGUI() {
    editor = ace.edit("editor");
    editor.setTheme("ace/theme/solarized_dark");
    editor.getSession().setMode("ace/mode/javascript");
    editor.getSession().setUseWrapMode(true);
    editor.getSession().on('change', editorChange);

    $("#openSidePanelButton").click(function(){
        $("#editor").animate({right:"300px"}, {
            duration: 500,
            start: function(){
                $.pageslide({ direction: "left", speed:500, modal: true , href: "#sidePanel"});
            },
            progress: function(){
                editor.resize();
            }
        });
    });
    $("#closeSidePanelButton").click(function(){
        $("#editor").animate({right:"0px"}, {
            duration: 500,
            start: function(){
                $.pageslide.close();
            },
            progress: function(){
                editor.resize();
            }
        });
    });
    $("#openSidePanelButton").click();
    editor.focus();
}

function setupServer(serverHostPort) {
    ws = new WebSocket("ws://" + serverHostPort, ["Message"]);
    ws.onopen = function () {
        ws.send(clientId+"");
        ws.send(docId);
        ws.onmessage = serverHandler;
    }
}

function serverHandler(e) {
    var msg = e.data
    if (msg.length < 12) {
        console.log("Received an improper command :" + msg);
        return;
    }
    var command = msg.substr(0, 10);
    var body = msg.substring(10, msg.length);
    var msgIdArr = body.split(" ", 1);
    if (msgIdArr.length == 0) {
        console.log("Received a command without an ID");
        return;
    }
    var msgId = msgIdArr[0];
    var args = msg.substr(11 + msgId.length, msg.length)
    console.log(msgId + ". " + command + ":" + args);
    switch(command) {
    case "setDoc    ":
        settingDoc = true;
        editor.setValue(args);
        editor.gotoLine(0);
        settingDoc = false;
        committed = args;
        ws.send("setDoc    ok");
        break;
    case "getDoc    ":
        ws.send("getDoc    " + msgId + " " + editor.getValue());
        break;
    case "vote      ":
        // get transactionId
        var transactionIdArr = args.split(" ", 1);
        if (transactionIdArr.length == 0) {
            console.log("Received a vote request without a transactionId");
            return;
        }
        var transactionId = transactionIdArr[0];
        var body = args.substr(1 + transactionId.length, args.length);
        if (committing && transactionId !== currTransactionId) {
            ws.send("vote      " + msgId + " " + "false")
        } else {
            currTransactionId = transactionId
            committing = body
            ws.send("vote      " + msgId + " " + "true")
        }
        break;
    case "complete  ":
        var transactionIdArr = args.split(" ", 1);
        if (transactionIdArr.length == 0) {
            console.log("Received a complete request without a transactionId");
            return;
        }
        transactionId = transactionIdArr[0];
        if (transactionId === currTransactionId){
            if (args.substr(1 + transactionId.length, args.length) === "true") {
                committed = committing;
                committing = null;
                currTransactionId == null;
                editor.setValue(committed);
                editor.gotoLine(0);
            } else {
                committing = null;
                currTransactionId == null;
                editor.setValue(committed);
                editor.gotoLine(0);
            }
        }
        ws.send("complete  " + msgId + " " + "ok")
        break;
    case "requestTxn":
        break;
    default:
        console.log("Received unrecognized command")
    }
}

function editorChange(e) {
    
}

function requestTxn() {
    if (!settingDoc && !committing) {
        currTransactionId = clientId + ":" + transactionNum;
        transactionNum++;
        committing = editor.getValue();
        ws.send("requestTxn" + currTransactionId + " " + committing);
    }
}