package api

import (
    "fmt"
    "io/ioutil"
    "mime"
    "os"
    "path/filepath"
    "crypto/rand"

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
    go a.gossiper.SendMessage(msg)

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

func (a *ApiHandler) UploadFile() http.HandlerFunc {
    const maxUploadSize = 2 * 1024 // 2 mb
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// validate file size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
            fmt.Println("FILE_TOO_BIG")
			return
		}

		// parse and validate file and post parameters
		file, _, err := r.FormFile("file")
		if err != nil {
            fmt.Println("INVALID_FILE")
            fmt.Print(err)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
            fmt.Println("INVALID_FILE2")
			return
		}

		// check file type, detectcontenttype only needs the first 512 bytes
		filetype := http.DetectContentType(fileBytes)
		switch filetype {
		case "image/jpeg", "image/jpg":
		case "image/gif", "image/png":
		case "application/pdf":
			break
		default:
            fmt.Println("INVALID_FILE_TYPE")
			return
		}
		fileName := randToken(12)
		newPath := filepath.Join("_SharedFiles/", fileName+fileEndings[0])
		fmt.Printf("FileType: %s, File: %s\n", fileType, newPath)

		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
            fmt.Println("CANT_WRITE_FILE")
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
            fmt.Println("CANT_WRITE_FILE")
			return
		}
		w.Write([]byte("SUCCESS"))
	})
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
