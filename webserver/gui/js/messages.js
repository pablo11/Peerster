const SCROLL_BOTTOM_AUTO_ON_NEW_MESSAGES = true
const REFRESH_MESSAGES_PERIOD = 2000

var myName = ""
var messages = []

$(document).ready(function() {
    // Hide scroll down btn
    $('.scroll-to-bottom').hide()

    // Scroll to bottom
    window.scrollTo(0,document.body.scrollHeight);

    // Set the visibility of scroll bottom btn
    setScrollToBottomBtn()
    $(document).scroll(function() {
        toggleScrollToBottomBtn()
    })

    // Load identity and messages
    getIdentity(function() {
        loadMessages()
    })

    // Refresh messages every REFRESH_MESSAGES_PERIOD milliseconds
    setInterval(function() {
        loadMessages()
    }, REFRESH_MESSAGES_PERIOD)
})

function setScrollToBottomBtn() {
    $('.scroll-to-bottom').click(function() {
        $("html, body").animate({ scrollTop: $(document).height() }, 1000);
    })
}

function toggleScrollToBottomBtn() {
    const scroll = $(window).scrollTop()
    const documentHeight = $(document).height()
    const windowHeight = $(window).height()

    if (documentHeight - scroll - windowHeight) {
        $('.scroll-to-bottom').show()
    } else {
        $('.scroll-to-bottom').hide()
    }
}

function addMessage(senderName, message, fromMe = false, animated = true) {
    const additionalClass = fromMe ? ' message-from-me' : ''
    const html = '<div class="well message' + additionalClass + '"><b>' + senderName + '</b><p class="mb-0">' + message + '</p></div>'
    $('#messages').append(html)

    const animationDuration = animated ? 500 : 0
    if (SCROLL_BOTTOM_AUTO_ON_NEW_MESSAGES) {
        $("html, body").animate({ scrollTop: $(document).height() }, animationDuration);
    } else {
        toggleScrollToBottomBtn()
    }
}

function getIdentity(onReturn) {
    $.get("api/id", function(data, status) {
        myName = data.name
        onReturn()
    })
}

function loadMessages() {
    $.get("api/message", function(data, status) {
        console.log("Refreshing messages");
        if (messages.length != data.length) {
            for (var i = messages.length; i < data.length; i++) {
                messages.push(data[i])

                addMessage(data[i].from, data[i].msg, data[i].from == myName, false)
            }
        }
    })
}
