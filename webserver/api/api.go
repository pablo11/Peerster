package api

import (
    //"io/ioutil" // For thesting only
    //"fmt"
    "net/http"
    "strings"
    "github.com/pablo11/Peerster/gossip"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/validator"
)

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

func (a *ApiHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
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
    go a.gossiper.SendMessage(msg)

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
/*
func sendFakeJson(file string) []byte {
    b, err := ioutil.ReadFile(file) // just pass the file name
    if err != nil {
        fmt.Print(err)
    }
    return b
}
*/
