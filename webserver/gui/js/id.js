$(document).ready(function() {
    loadAndDisplayId()
})

function loadAndDisplayId() {
    $.get("api/id", function(data, status) {
        $("#id-node-name").html(data.name)
        $("#id-public-address").html(data.address)
    })
}
