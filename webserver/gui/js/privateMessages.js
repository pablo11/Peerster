$(document).ready(function() {
    loadAndDisplayOrigins()
    setupSendMessageFrom()
})

function loadAndDisplayOrigins() {
    $.get("api/origins", function(data, status) {
        const html = data.map(origin => '<a href="#" class="list-group-item" onclick="sendPrivateMessage(this)">' + origin + '</a>')
        $("#origins-list").html(html)
    })
}

function displayMsg(msg) {
    const html = '<div class="alert alert-success"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>' + msg + '</strong></div>'
    $(html).insertBefore('#origins-list')
}

function sendPrivateMessage(link) {
    const origin = $(link).html()

    $("#modalMessageTo").html(origin)
    $("#modal-msg").val("")
    $("#modal-dest").val(origin)
    $("#messageModal").modal()
}

function setupSendMessageFrom() {
    $("#sendPrivateMessage").click(function() {
        const msg = $("#modal-msg").val()
        const dest = $("#modal-dest").val()


        // Send message to API
        $.post("api/sendPrivateMessage", {
            msg: msg,
            dest: dest
        }, function(data, status) {
            console.log("Private message to " + dest + " sent: ", msg, status);
        })

        $("#messageModal").modal("hide")
        displayMsg("Message to " + dest + " sent.")

        return false
    })

    // Enable/disable send button
    $("#modal-msg").keyup(function() {
        if ($("#modal-msg").val() == "") {
            $("#sendPrivateMessage").attr("disabled", "disabled");
        } else {
            $("#sendPrivateMessage").removeAttr("disabled");
        }
    })
}
