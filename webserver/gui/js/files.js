$(document).ready(function() {
    listFiles()
    setupFileUpload()
    setupRequestFile()
})

function listFiles() {
    $.get("api/listFiles", function(data, status) {
        // Order by name
        data.sort(function(a, b) {
            return a.name > b.name
        })

        console.log(data);
        displayListedFiles(data)
    })
}

function displayListedFiles(files) {
    var html = ""
    for (var f of files) {
        const b64pth = btoa(f.path)
        html += '<a href="/api/downloadFile?path=' + b64pth + '" target="_blank" class="list-group-item">' + f.name + '</a>'
    }

    $("#listFiles").html(html)
}

function setupFileUpload() {
    $("#file-upload-progress-bar").hide()
    $("#upload-file-btn").prop("disabled", true)


    $("#file").on("change", function() {
        $("#upload-file-btn").prop("disabled", false)
    })

    $("#upload-file-btn").click(function() {
        $("#file-upload-progress-bar").show()
        $.ajax({
            url: '/api/uploadFile',
            type: 'POST',
            data: new FormData($('form')[0]),
            cache: false,
            contentType: false,
            processData: false,
            xhr: function() {
                var myXhr = $.ajaxSettings.xhr();
                if (myXhr.upload) {
                    // Handling the upload progress
                    myXhr.upload.addEventListener('progress', function(e) {
                        if (e.lengthComputable) {
                            $('progress').attr({
                                value: e.loaded,
                                max: e.total,
                            });
                        }
                    }, false);
                }
                return myXhr;
            }
        }).done(function(data) {
            console.log("finished");
            uploadCompleted("Upload successful")
        }).fail(function(data) {
            console.log(data);
            uploadCompleted("There was an error uploading your file")
        });

        return false;
    })
}

function uploadCompleted(message) {
    window.alert(message)
    $("#file").val("")
    $('progress').attr({
        value: 0,
        max: 100,
    });
    $("#upload-file-btn").prop("disabled", true)
    $("#file-upload-progress-bar").hide()
}

function setupRequestFile() {
    // Load nodes
    $.get("api/origins", function(data, status) {
        const html = data.map(origin => '<option value="' + origin + '">' + origin + '</option>')
        $("#requestFromNode").html('<option value="0" selected="selected" disabled="disabled">Select a node</option>' + html)


        $("#request-file-btn").click(function() {
            // Send message to API
            $.post("api/requestFile", {
                filename: $("#filename").val(),
                dest: $("#requestFromNode").val() || "0",
                hash: $("#fileHash").val()
            }, function(data, status) {
                console.log("File requested");
                window.alert("The file was requested, reload the page in a couple of seconds and find it in the available files")
            })

            return false;
        })
    })
}
