var ws;
var clientId;
var editor;

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
    $("#openSidePanelButton").click()
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
    if (msg.length < 10) {
        console.log("Received an improper command :" + msg)
        return
    }
    var command = msg.substr(0, 10);
    var args = msg.substring(10, msg.length);
    console.log(command + ":" + args);
    switch(command) {
    case "setDoc    ":
        editor.setValue(args);
        break;
    case "getDoc    ":
        break;
    case "vote      ":
        break;
    case "comple    ":
        break;
    case "requestTxn":
        break;
    default:
        console.log("Received unrecognized command")
    }
}

function editorChange(e) {
    //ws.send("requestTxn" + e.data.text);
    ws.send("")
}