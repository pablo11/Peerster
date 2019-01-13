var assets = {}
var nodeName = ""
var currentAsset = ""
var currentVotes = {}
var identityRegistered = false

$(document).ready(function() {
    loadNodeId()
    setupIdentityRegistrationRequirement()

    if (identityRegistered) {
        setupCreateNewAsset()
        setupSendShares()
        setupAskQuestion()

        fetchAndDisplayMyAssets()
        setInterval(function() {
            fetchAndDisplayMyAssets()
        }, 2000)

        //  Update the asset's votations every 3 seconds

        setInterval(function() {
            showAsset(currentAsset)
        }, 2000)
    }

    $("#assetModal").on('hide.bs.modal', function() {
        currentAsset = ""
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

function checkIfIdentityIsRegistered() {
    $.get("api/identity/check", function(data, status) {
        identityRegistered = data.identityRegistered
        if (identityRegistered) {
            $("#registerIdentityModal").modal("hide")
        } else {
            waitForIdentityRegistration()
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
        var html = ""
        for (assetName in assets) {
            var a = assets[assetName]
            //data-toggle="modal" data-target="#assetModal"
            html += '<tr onclick="showAsset(\'' + assetName + '\')"><td>' + assetName + '</td><td>' + a.balance + '</td><td>' + a.totSupply + '</td></tr>'
        }

        $('#listAssetsRows').html(html)
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
                newVotesDiffer = true
            } else {
                if (Object.keys(currentVotes[vote].answers).length != Object.keys(votes[vote].answers).length) {
                    newVotesDiffer = true
                } else {
                    for (holder in votes[vote].answers) {
                        if (!currentVotes[vote].answers.hasOwnProperty(holder) || currentVotes[vote].answers[holder] != votes[vote].answers[holder]) {
                            newVotesDiffer = true
                        }
                    }
                }
            }
        }
        if (!newVotesDiffer) {
            console.log("compared but equals");
            return
        }

        $('#assetModalListVotes').html('')

        console.log("compare", currentVotes, votes);

        currentVotes = votes

        console.log(votes);
        var html = ""
        for (vote in votes) {
            v = votes[vote]
            var positiveAnswers = 0
            var negativeAnswers = 0
            var thisNodeReply = ""
            for (holderName in v.answers) {
                if (holderName == nodeName) {
                    thisNodeReply = v.answers[nodeName]
                }

                if (v.answers[holderName] == "yes") {
                    positiveAnswers += 1
                } else {
                    negativeAnswers += 1
                }
            }

            htmlVote = (thisNodeReply != "") ? thisNodeReply : `<div class="btn-group">
                <button type="button" class="btn btn-xs btn-success" onclick="voteOnAsset(this, '` + v.question + `',true, '` + v.origin + `')">yes</button>
                <button type="button" class="btn btn-xs btn-danger" onclick="voteOnAsset(this, '` + v.question + `',false, '` + v.origin + `')">no</button>
            </div>`
            html += '<tr><td>' + v.question  + '</td><td>' + positiveAnswers + '/' + negativeAnswers + '</td><td>' + v.origin + '</td><td>' + htmlVote + '</td></tr>'
        }
        $('#assetModalListVotes').html(html)
    })

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
    $(button).parent().parent().html(answerStr)

    $.post("api/asset/vote", {
        question: question,
        asset: $('#modalAssetName').html(),
        origin: origin,
        answer: answerBoolStr
    }, function(data, status) {
        // Find the right question and answer in the currentVotes and add the vote
        for (vote in currentVotes) {
            if (currentVotes[vote].question == question) {
                currentVotes[vote][origin] = answerStr
            }
        }

        window.alert("Your answer was submitted to the blockchain.")
    })
}
