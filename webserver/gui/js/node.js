$(document).ready(function() {
    loadAndDisplayPeers()
    setAddPeerForm()
})

function setAddPeerForm() {
    $("#add-peer-form").submit(function() {
        // Disable submit btn
        $("#add-peer-form-submit-btn").prop("disabled", true)

        const newPeer = $("#new-peer-input").val()
        if (newPeer == null || newPeer == "") {
            // Enable submit btn
            $("#add-peer-form-submit-btn").prop("disabled", false)
            return false
        }

        const jsonParameters = {
            peer: newPeer
        }

        $.post("api/node", {
            peer: newPeer
        }, function(data, status) {
            console.log("Peer added: ", newPeer);

            $("#peers-list").append('<a href="#" class="list-group-item">' + newPeer + '</a>')

            // Clean input
            $("#new-peer-input").val("")

            // Enable submit btn
            $("#add-peer-form-submit-btn").prop("disabled", false)
            return false
        }).fail(function() {
            console.log("Peer NOT added");

            // Clean input
            $("#new-peer-input").val("")

            // Enable submit btn
            $("#add-peer-form-submit-btn").prop("disabled", false)

            displayErrorMsg("Invalid address")

            return false
        })
        return false
    })
}

function loadAndDisplayPeers() {
    $.get("api/nodes", function(data, status) {
        const html = data.map(peer => '<a href="#" class="list-group-item">' + peer + '</a>')
        $("#peers-list").html(html)
    })
}

function displayErrorMsg(msg) {
    const html = '<div class="alert alert-danger"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>' + msg + '</strong></div>'
    $(html).insertBefore('#add-peer-form')
}
