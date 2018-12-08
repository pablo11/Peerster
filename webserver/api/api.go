package api

import (
    "fmt"
    "os"
    "io"
    "io/ioutil"
    "net/http"
    "strings"
    "encoding/base64"
    "github.com/pablo11/Peerster/gossip"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/validator"
)

const SHARED_FILES_DIR = "_SharedFiles/"
const DOWNLOADED_FILES_DIR = "_Downloads/"

type ApiHandler struct {
    gossiper *gossip.Gossiper
}

func NewApiHandler(gossiper *gossip.Gossiper) *ApiHandler {
    return &ApiHandler{
        gossiper: gossiper,
    }
}

func (a *ApiHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
    messages := a.gossiper.GetAllMessages()
    json := strings.Join(mapRumorMessages(messages, func(m *model.RumorMessage) string {
        return m.ToJSON()
    }), ",")

    sendJSON(w, []byte(`[` + json + `]`))
}

func mapRumorMessages(vs []*model.RumorMessage, f func(*model.RumorMessage) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}

func (a *ApiHandler) SendPublicMessage(w http.ResponseWriter, r *http.Request) {
    // Parse POST "msg"
    r.ParseForm()
    postedMsg, isPresent := r.PostForm["msg"]
    if !isPresent || len(postedMsg) != 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    msg := postedMsg[0]

    // Send message to gossiper
    go a.gossiper.SendPublicMessage(msg)

    // Respond to request with ok
    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) GetOrigins(w http.ResponseWriter, r *http.Request) {
    jsonOrigins := JsonOrigins{
        origins: a.gossiper.GetOrigins(),
    }

    sendJSON(w, jsonOrigins.toByte())
}

func (a *ApiHandler) SendPrivateMessage(w http.ResponseWriter, r *http.Request) {
    // Parse POST "msg" and "dest"
    r.ParseForm()
    postedMsg, msgIsPresent := r.PostForm["msg"]
    postedDest, destIsPresent := r.PostForm["dest"]
    if !msgIsPresent || !destIsPresent || len(postedMsg) != 1 || len(postedDest) != 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    msg := postedMsg[0]
    dest := postedDest[0]

    // Send private message
    pm := model.NewPrivateMessage(a.gossiper.Name, msg, dest)
    a.gossiper.SendPrivateMessage(pm)

    // Respond to request with ok
    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
    jsonPeers := JsonPeers{
        peers: a.gossiper.GetPeers(),
    }

    sendJSON(w, jsonPeers.toByte())
}

func (a *ApiHandler) AddNode(w http.ResponseWriter, r *http.Request) {
    // Parse POST "peer"
    r.ParseForm()
    postedNewPeer, isPresent := r.PostForm["peer"]
    if !isPresent || len(postedNewPeer) != 1 || !validator.IsGossipAddr(postedNewPeer[0]) {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    peer := postedNewPeer[0]

    // Add peer to gossiper
    a.gossiper.AddPeer(peer)

    // Respond to request with ok
    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) GetId(w http.ResponseWriter, r *http.Request) {
    jsonId := JsonId{
        name: a.gossiper.Name,
        address: a.gossiper.GetAddress(),
    }

    sendJSON(w, jsonId.toByte())
}

func sendJSON(w http.ResponseWriter, json []byte) {
    w.Header().Set("Server", "Cryptop GO server")
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(200)
    w.Write(json)
}

func sendError(w http.ResponseWriter, errorMsg string) {
    w.Header().Set("Server", "Cryptop GO server")
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(200)
    w.Write([]byte("{error:" + errorMsg + "}"))
}

func (a *ApiHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
    r.ParseMultipartForm(0)
    file, handler, err := r.FormFile("file")
    if err != nil {
        sendError(w, err.Error())
        fmt.Println(err)
        return
    }
    defer file.Close()

    f, err := os.OpenFile(SHARED_FILES_DIR + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
    if err != nil {
        sendError(w, err.Error())
        fmt.Println(err)
        return
    }
    defer f.Close()
    io.Copy(f, file)

    // Respond to request with ok
    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
    filesShared, err := ioutil.ReadDir(SHARED_FILES_DIR)
    if err != nil {
        fmt.Println(err)
    }

    filesDownloaded, err := ioutil.ReadDir(DOWNLOADED_FILES_DIR)
    if err != nil {
        fmt.Println(err)
    }

    filesJson := make([]string, 0)
    for _, f := range filesShared {
        if !strings.HasPrefix(f.Name(), ".") {
            filesJson = append(filesJson, "{\"path\": \"" + SHARED_FILES_DIR + f.Name() + "\", \"name\": \"" + f.Name() + "\"}")
        }
    }
    for _, f := range filesDownloaded {
        if !strings.HasPrefix(f.Name(), ".") {
            filesJson = append(filesJson, "{\"path\": \"" + DOWNLOADED_FILES_DIR + f.Name() + "\", \"name\": \"" + f.Name() + "\"}")
        }
    }

    json := strings.Join(filesJson, ",")

    sendJSON(w, []byte(`[` + json + `]`))
}

func (a *ApiHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
    // Get file path from request
    path, ok := r.URL.Query()["path"]
    if !ok || len(path[0]) < 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    decodedPath, err := base64.StdEncoding.DecodeString(path[0])
    if err != nil {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(500)
        return
	}

    w.Header().Set("Server", "Cryptop GO server")
    w.Header().Add("Content-Disposition", "Attachment")
    http.ServeFile(w, r, string(decodedPath))
}
