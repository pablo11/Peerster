package api

import (
    "io/ioutil" // For thesting only
    "fmt"
    "net/http"
    //"encoding/json"
    "github.com/pablo11/Peerster/gossip"
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
    sendJSON(w, sendFakeJson("fake/messages.json"))

    // TODO
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

    fmt.Println("Sending message: " + msg)

    // Send message to gossiper
    go a.gossiper.SendSimpleMessage(msg)

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

    fmt.Println("Adding peer: " + peer)

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

func sendFakeJson(file string) []byte {
    b, err := ioutil.ReadFile(file) // just pass the file name
    if err != nil {
        fmt.Print(err)
    }
    return b
}

/*
func (a *ApiHandler) sendMessageToGossiper(msg string) bool {
    conn, e := net.Dial("udp", a.gossiperAddr)
	defer conn.Close()
	if e != nil {
		fmt.Println(e)
        return false
	}
	conn.Write([]byte(msg))
    return true
}
*/
