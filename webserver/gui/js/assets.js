$(document).ready(function() {
    setupCreateNewAsset()


    fetchAndDisplayMyAssets()
    setInterval(function() {
        fetchAndDisplayMyAssets()
    }, 2000)
})

function prepareModal(asset) {



}

function setupCreateNewAsset() {
    $('#create-asset-btn').click(function() {
        $.post("api/assets/create", {
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
        console.log(assets, status);

        var html = ""
        for (assetName in assets) {
            var a = assets[assetName]
            html += '<tr data-toggle="modal" data-target="#assetModal"><td>' + assetName + '</td><td>' + a.balance + '</td><td>' + a.totSupply + '</td></tr>'
        }

        $('#listAssetsRows').html(html)
    })
}
