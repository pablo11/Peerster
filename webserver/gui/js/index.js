$(document).ready(function() {
    setSendMsgForm()
})

function setSendMsgForm() {
    $("#send-msg-form").submit(function() {
        // Disable submit btn
        $("#send-msg-form-submit-btn").prop("disabled", true)

        const msg = $("#msg-input").val()
        if (msg == null || msg == "") {
            // Enable submit btn
            $("#send-msg-form-submit-btn").prop("disabled", false)
            return false
        }

        // Send message to API
        $.post("api/sendPublicMessage", {
            msg: msg
        }, function(data, status) {
            console.log("Message sent: ", msg, status);

            // Clean input
            $("#msg-input").val("")

            // Enable submit btn
            $("#send-msg-form-submit-btn").prop("disabled", false)
        })

        return false
    })
}
