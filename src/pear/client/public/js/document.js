$(function(){   
    var editor = ace.edit("editor");
    editor.setTheme("ace/theme/cobalt");
    editor.getSession().setMode("ace/mode/javascript");
    editor.getSession().setUseWrapMode(true);

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
                console.log("RAWR")
            }
        });
    });
    $("#openSidePanelButton").click()
});