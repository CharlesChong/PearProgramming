var ws;
$(function(){
    $.get("http://" + centralHostPort, {docId: docId})
        .done(function(data) {
            setupServer(data);
        }).fail(function(data) {
            alert("Failed to retrieve server information");
            console.log(data);
        });
    setupGUI();
});

function setupGUI() {
    var editor = ace.edit("editor");
    editor.setTheme("ace/theme/cobalt");
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
    ws = new WebSocket("ws://" + serverHostPort);
    ws.onmessage = serverHandler;
}

function serverHandler(e) {
    console.log(event.data)
    console.log("Received: " + e.data);
    console.log(e)
}

function editorChange(e) {
    ws.send(e.data.text);
}