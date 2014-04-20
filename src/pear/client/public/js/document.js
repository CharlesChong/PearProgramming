$(function(){
    $("#openSidePanelButton").pageslide({ direction: "left", modal: true });

    var editor = ace.edit("editor");
    editor.setTheme("ace/theme/cobalt");
    editor.getSession().setMode("ace/mode/javascript");
    editor.getSession().setUseWrapMode(true);
});