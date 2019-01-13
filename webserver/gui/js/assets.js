$(document).ready(function() {
    setupCreateNewAsset()
    setupAssetDetailsModal()
    setupSendShares()


    fetchAndDisplayMyAssets()
    setInterval(function() {
        fetchAndDisplayMyAssets()
    }, 2000)
})

function setupCreateNewAsset() {
    $('#create-asset-btn').click(function() {
        $.post("api/asset/create", {
            assetName: $("#createAssetName").val(),
            totSupply: $("#createAssetTotSuppl").val(),
        }, function(data, status) {
            //console.log("File requested");
            window.alert("Your asset creation was submitted to the blockchain. It'll be visible as soon as it's included in a block.")
        })

        $("#newAssetModal").modal("hide")
    })
}

function fetchAndDisplayMyAssets() {
    $.get("api/assets/list", function(assets, status) {
        //console.log(assets, status);

        var html = ""
        for (assetName in assets) {
            var a = assets[assetName]
            html += '<tr class="assetRow" data-toggle="modal" data-target="#assetModal"><td>' + assetName + '</td><td>' + a.balance + '</td><td>' + a.totSupply + '</td></tr>'
        }

        $('#listAssetsRows').html(html)
    })
}

function setupAssetDetailsModal() {
    $('.assetRow').click(function() {
        console.log("hello");
    })
}

function setupSendShares() {
    $('#send-shares-btn').click(function() {
        $.post("api/asset/send", {
            amount: $("#sendAssetAmount").val(),
            dest: $("#sendAssetDest").val(),
            assetName: $('#modalAssetName').html(),
        }, function(data, status) {
            window.alert("Your transaction was submitted to the blockchain. It'll be executed if valid.")
        })

        $("#assetModal").modal("hide")
    })
}
