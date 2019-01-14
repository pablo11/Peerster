var assets = {}
var nodeName = ""
var currentAsset = ""
var currentVotes = {}
var identityRegistered = false

$(document).ready(function() {
    loadNodeId()
    checkIfIdentityIsRegistered(true)
    //setupIdentityRegistrationRequirement()

    setupCreateNewAsset()
    setupSendShares()
    setupAskQuestion()

    fetchAndDisplayMyAssets()
    setInterval(function() {
        if (identityRegistered) {
            fetchAndDisplayMyAssets()
        }
    }, 2000)

    //  Update the asset's votations every 2 seconds
    setInterval(function() {
        showAsset(currentAsset)
    }, 3000)

    $("#assetModal").on('hide.bs.modal', function() {
        currentAsset = ""
        currentVotes = {}
    });
})

function loadNodeId() {
    $.get("api/id", function(data, status) {
        nodeName = data.name
    })
}

function setupIdentityRegistrationRequirement() {
    if (!identityRegistered) {
        $(".claimIdentityBtnLoader").hide()
        $("#registerIdentityModal").modal("show")

        $("#claimIdentityBtn").click(function() {
            $("#claimIdentityBtn").prop("disabled", true)
            $(".claimIdentityBtnTitle").hide()
            $(".claimIdentityBtnLoader").show()

            $.get("api/identity/register", function(data, status) {
                waitForIdentityRegistration()
            })
        })
    }
}

function checkIfIdentityIsRegistered(isFirstTime = false) {
    $.get("api/identity/check", function(data, status) {
        identityRegistered = data.identityRegistered
        if (identityRegistered) {
            $("#registerIdentityModal").modal("hide")
        } else {
            if (isFirstTime) {
                setupIdentityRegistrationRequirement()
            } else {
                waitForIdentityRegistration()
            }
        }
    })
}

function waitForIdentityRegistration() {
    setTimeout(function() {
        checkIfIdentityIsRegistered()
    }, 2000)
}

function setupCreateNewAsset() {
    $('#create-asset-btn').click(function() {
        $.post("api/asset/create", {
            assetName: $("#createAssetName").val(),
            totSupply: $("#createAssetTotSuppl").val()
        }, function(data, status) {
            //console.log("File requested");
            window.alert("Your asset creation was submitted to the blockchain. It'll be visible as soon as it's included in a block.")
        })

        $("#newAssetModal").modal("hide")
    })
}

function fetchAndDisplayMyAssets() {
    $.get("api/assets/list", function(assetsReturned, status) {
        //console.log(assets, status);
        assets = assetsReturned
        var htmlRows = []
        for (assetName in assets) {
            var a = assets[assetName]
            //data-toggle="modal" data-target="#assetModal"
            htmlRows.push('<tr onclick="showAsset(\'' + assetName + '\')"><td>' + assetName + '</td><td>' + a.balance + '</td><td>' + a.totSupply + '</td></tr>')
        }

        htmlRows = htmlRows.sort((a, b) => {
            return a.toLowerCase() > b.toLowerCase()
        })

        $('#listAssetsRows').html(htmlRows.join(""))
    })
}

function showAsset(assetName) {
    if (assetName == "") {
        return
    }

    var asset = assets[assetName]
    currentAsset = assetName

    $('#modalAssetName').html(assetName)
    $('#modalAssetBalance').html(asset.balance)
    $('#modalAssetTotSupply').html(asset.totSupply)

    $("#assetModal").modal("show")
    $.get("api/asset/votes?asset=" + assetName, function(votes, status) {
        // Check if the new votes are differeent from the ones saved
        var newVotesDiffer = Object.keys(votes).length == 0
        for (vote in votes) {
            if (!currentVotes.hasOwnProperty(vote)) {
                // There's a new vote statement
                newVotesDiffer = true
            } else {
                if (Object.keys(currentVotes[vote].answers).length != Object.keys(votes[vote].answers).length) {
                    // There are new votes to the question
                    console.log("votes differ 2", currentVotes[vote].answers, votes[vote].answers);
                    newVotesDiffer = true
                } else {
                    for (holder in votes[vote].answers) {
                        if (!currentVotes[vote].answers.hasOwnProperty(holder) || currentVotes[vote].answers[holder].reply != votes[vote].answers[holder].reply || currentVotes[vote].answers[holder].balance != votes[vote].answers[holder].balance) {
                            newVotesDiffer = true
                        }
                    }
                }
            }
        }
        if (!newVotesDiffer) {
            return
        }

        currentVotes = votes

        var htmlRows = []
        for (vote in votes) {
            v = votes[vote]

            var rowElements = prepareRowElements(v.answers)

            htmlRows.push('<tr><td>' + v.question  + '</td><td>' + rowElements.decision + '</td><td>' + rowElements.nbAnswers + '</td><td>' + v.origin + '</td><td>' + rowElements.htmlVote + '</td></tr>')
        }

        htmlRows = htmlRows.sort((a, b) => {
            return a.toLowerCase() > b.toLowerCase()
        })

        $('#assetModalListVotes').html('')
        $('#assetModalListVotes').html(htmlRows.join(""))
    })

}

function prepareRowElements(answers) {
    var totReplies = 0
    var positiveAnswers = 0
    var negativeAnswers = 0
    var thisNodeReply = ""
    for (holderName in answers) {
        if (holderName == nodeName) {
            console.log("This node reply found");
            thisNodeReply = answers[holderName].reply
        }

        totReplies += answers[holderName].balance
        if (answers[holderName].reply == "yes") {
            positiveAnswers += answers[holderName].balance
        } else {
            negativeAnswers += answers[holderName].balance
        }
    }

    var positiveAnswersPercentage = (totReplies > 0) ? (parseFloat(positiveAnswers) / parseFloat(totReplies) * 100.0).toFixed(1) : 0
    var decision = ""
    if (positiveAnswersPercentage > 50) {
        decision = "<b style=\"color:green;\">Yes</b> with " + positiveAnswersPercentage + "%"
    } else {
        decision = "<b style=\"color:red;\">No</b> with " + (100 - positiveAnswersPercentage) + "%"
    }

    var nbAnswers = Object.keys(answers).length || 0
    if (nbAnswers == 0) {
        decision = "-"
    }

    var htmlVote = (thisNodeReply != "") ? thisNodeReply : `<div class="btn-group">
        <button type="button" class="btn btn-xs btn-success" onclick="voteOnAsset(this, '` + v.question + `',true, '` + v.origin + `')">yes</button>
        <button type="button" class="btn btn-xs btn-danger" onclick="voteOnAsset(this, '` + v.question + `',false, '` + v.origin + `')">no</button>
    </div>`

    return {decision: decision, nbAnswers: nbAnswers, thisNodeReply: thisNodeReply, htmlVote: htmlVote}
}

function setupSendShares() {
    $('#send-shares-btn').click(function() {
        $("#send-shares-btn").prop("disabled", true)
        $.post("api/asset/send", {
            amount: $("#sendAssetAmount").val(),
            dest: $("#sendAssetDest").val(),
            assetName: $('#modalAssetName').html()
        }, function(data, status) {
            $("#sendAssetAmount").val('')
            $("#sendAssetDest").val('')
            $("#send-shares-btn").prop("disabled", false)
            window.alert("Your transaction was submitted to the blockchain. It'll be executed if valid.")
        })

        $("#assetModal").modal("hide")
    })
}

function setupAskQuestion() {
    $('#propose-vote-btn').click(function() {
        $("#propose-vote-btn").prop("disabled", true)
        $.post("api/asset/newVote", {
            question: $("#proposeVoteQuestion").val(),
            asset: $('#modalAssetName').html()
        }, function(data, status) {
            $("#proposeVoteQuestion").val('')
            $("#propose-vote-btn").prop("disabled", false)
            window.alert("Your question was submitted to the blockchain. Asset holders will shortly be able to vote on it.")
        })
    })
}

function voteOnAsset(button, question, answer, origin) {
    var answerStr = answer ? "yes" : "no"
    var answerBoolStr = answer ? "true" : "false"

    //$(button).parent().parent().html(answerStr)

    // Find the right question and answer in the currentVotes and add the vote
    for (vote in currentVotes) {
        if (currentVotes[vote].question == question) {
            currentVotes[vote].answers[origin] = {reply: answerStr, balance: parseInt($('#modalAssetBalance').html())}

            var rowElements = prepareRowElements(currentVotes[vote].answers)

            var htmlRowContent = '<td>' + question  + '</td><td>' + rowElements.decision + '</td><td>' + rowElements.nbAnswers + '</td><td>' + currentVotes[vote].origin + '</td><td>' + rowElements.htmlVote + '</td>'
            $(button).parent().parent().parent().html(htmlRowContent)
        }
    }

    $.post("api/asset/vote", {
        question: question,
        asset: $('#modalAssetName').html(),
        origin: origin,
        answer: answerBoolStr
    }, function(data, status) {
        window.alert("Your answer was submitted to the blockchain.")
    })
}
