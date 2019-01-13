package api

import (
    "fmt"
    "os"
    "io"
    "io/ioutil"
    "net/http"
    "strings"
    "strconv"
	"log"
    "encoding/base64"
	"encoding/hex"
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
    go a.gossiper.SendPublicMessage(msg, true)

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

    go a.gossiper.FileSharing.IndexFile(handler.Filename)

    // Respond to request with ok
    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) RequestFile(w http.ResponseWriter, r *http.Request) {
    // Parse POST "hash"
    r.ParseForm()
    filename, isFilenamePresent := r.PostForm["filename"]
    dest, isDestPresent := r.PostForm["dest"]
    hash, isHashPresent := r.PostForm["hash"]
    if !isHashPresent || len(hash) != 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    hashStr := hash[0]

    if isFilenamePresent && len(filename) == 1 && isDestPresent && len(dest) == 1 && dest[0] != "0" {
        go a.gossiper.FileSharing.RequestFile(filename[0], dest[0], hashStr)
    } else if !isFilenamePresent && !isDestPresent {
        go a.gossiper.FileSharing.RequestFile("", "", hashStr)
    } else {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    // Respond to request with ok
    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
    /*
    filesShared, err := ioutil.ReadDir(SHARED_FILES_DIR)
    if err != nil {
        fmt.Println(err)
    }
    */

    filesDownloaded, err := ioutil.ReadDir(DOWNLOADED_FILES_DIR)
    if err != nil {
        fmt.Println(err)
    }

    filesJson := make([]string, 0)
    /*
    for _, f := range filesShared {
        if !strings.HasPrefix(f.Name(), ".") {
            filesJson = append(filesJson, "{\"path\": \"" + SHARED_FILES_DIR + f.Name() + "\", \"name\": \"" + f.Name() + "\"}")
        }
    }
    */
    for _, f := range filesDownloaded {
        if !strings.HasPrefix(f.Name(), ".") {
            filesJson = append(filesJson, "{\"path\": \"" + DOWNLOADED_FILES_DIR + f.Name() + "\", \"name\": \"" + f.Name() + "\"}")
        }
    }

    sendJSON(w, []byte(`[` + strings.Join(filesJson, ",") + `]`))
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

func (a *ApiHandler) SearchFiles(w http.ResponseWriter, r *http.Request) {
    // Get search query from request
    r.ParseForm()
    query, isQueryPresent := r.PostForm["query"]
    budget, isBudgetPresent := r.PostForm["budget"]
    if !isQueryPresent || len(query) < 1 || query[0] == "" || !isBudgetPresent || len(budget) < 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    keywords := strings.Split(query[0], ",")
    budgetVal, err := strconv.Atoi(budget[0])
    if err != nil {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    if budgetVal == 0 {
        go a.gossiper.StartSearchRequest(2, keywords, true)
    } else {
        go a.gossiper.StartSearchRequest(uint64(budgetVal), keywords, false)
    }

    // Respond to request with ok
    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) SearchResults(w http.ResponseWriter, r *http.Request) {
    fileMatches := a.gossiper.GetFullMatches()
    jsonFiles := make([]string, len(fileMatches))
    for i, f := range fileMatches {
        jsonFiles[i] = "{\"filename\":\"" + f.Filename + "\", \"metahash\":\"" + f.MetaHash + "\"}"
    }

    sendJSON(w, []byte(`[` + strings.Join(jsonFiles, ",") + `]`))
}

func (a *ApiHandler) VotationCreate(w http.ResponseWriter, r *http.Request) {
	//Create a votation
	r.ParseForm()
    question, isQuestionPresent := r.PostForm["question"]
    asset, isAssetPresent := r.PostForm["asset"]
    if !isQuestionPresent || len(question) < 1 || question[0] == "" || !isAssetPresent || len(asset) < 1 || asset[0] == "" {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

	//Check for correctness will be done in votation
	 go a.gossiper.LaunchVotation(question[0],asset[0])
}

func (a *ApiHandler) Votations(w http.ResponseWriter, r *http.Request) {

	a.gossiper.Blockchain.VoteStatementMutex.Lock()
	voteStatements := a.gossiper.Blockchain.VoteStatement //Is lock here enought because after we iterate...
	a.gossiper.Blockchain.VoteStatementMutex.Unlock()
	a.gossiper.Blockchain.AssetsMutex.Lock()
	assetsMap := a.gossiper.Blockchain.Assets //Same here
	a.gossiper.Blockchain.AssetsMutex.Unlock()


	var jsonFiles []string

	for _, vs := range voteStatements {
		haveTheAsset := false
		//Only send question for assets you have
		for shareholders,shares := range assetsMap[vs.AssetName] {
			if shareholders == a.gossiper.Name && shares > 0{
				haveTheAsset = true
			}
		}

		if haveTheAsset {
			jsonFiles = append(jsonFiles,"{\"question\":\"" + vs.Question + "\", \"origin\":\"" + vs.Origin + "\", \"asset\":\"" + vs.AssetName +"\"}")
		}
	}

	sendJSON(w, []byte(`[` + strings.Join(jsonFiles, ",") + `]`))
}

func (a *ApiHandler) VotationReply(w http.ResponseWriter, r *http.Request) {
	//Create a votation reply
	r.ParseForm()
    question, isQuestionPresent := r.PostForm["question"]
    asset, isAssetPresent := r.PostForm["asset"]
	origin, isOriginPresent := r.PostForm["origin"]
	answer, isAnswerPresent := r.PostForm["answer"]
    if !isQuestionPresent || len(question) < 1 || question[0] == "" || !isAssetPresent || len(asset) < 1 || asset[0] == ""{
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

	if !isOriginPresent || len(origin) < 1 || origin[0] == "" || !isAnswerPresent || len(answer) < 1 || answer[0] == ""{
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

	answer_bool, err := strconv.ParseBool(answer[0])
	if err != nil {
		w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
	}

	//Check for correctness will be done in votation
	 go a.gossiper.AnswerVotation(question[0], asset[0], origin[0], answer_bool)
}

func (a *ApiHandler) VotationResult(w http.ResponseWriter, r *http.Request) {

	//reply only answers for votating I'm in

	a.gossiper.Blockchain.VoteStatementMutex.Lock()
	voteStatements := a.gossiper.Blockchain.VoteStatement //Is lock here enought because after we iterate...
	a.gossiper.Blockchain.VoteStatementMutex.Unlock()
	a.gossiper.Blockchain.AssetsMutex.Lock()
	assetsMap := a.gossiper.Blockchain.Assets //Same here
	a.gossiper.Blockchain.AssetsMutex.Unlock()
	a.gossiper.Blockchain.VoteAnswersMutex.Lock()
	voteAnswers := a.gossiper.Blockchain.VoteAnswers //Same here
	a.gossiper.Blockchain.VoteAnswersMutex.Unlock()

	var jsonFiles []string

	for question_id, vs := range voteStatements {
		haveTheAsset := false
		//Only send question for assets you have
		for shareholders,shares := range assetsMap[vs.AssetName] {
			if shareholders == a.gossiper.Name && shares > 0{
				haveTheAsset = true
			}
		}

		if haveTheAsset {
			holderNames, votationAnswerExists := voteAnswers[question_id]
			a.gossiper.QuestionKeyMutex.Lock()
			key, keyExists := a.gossiper.QuestionKey[question_id]
			a.gossiper.QuestionKeyMutex.Unlock()
			if votationAnswerExists && keyExists{
				for holderName, answer := range holderNames{
					key_byte, err := hex.DecodeString(key)
					if err != nil{
						log.Fatal("Error occured when decoding key")
						break
					}
					ans_decrypted, err := answer.Decrypt(key_byte)
					if err != nil{
						log.Fatal("Error occured when decrypting answer")
						break
					}
					bool_str := strconv.FormatBool(ans_decrypted.Answer)
					jsonFiles = append(jsonFiles,"{\"question\":\"" + vs.Question + "\", \"origin\":\"" + vs.Origin + "\", \"asset\":\"" + vs.AssetName +"\", \"replier\":\"" + holderName +"\", \"answer\":\"" + bool_str +"\"}")
				}
			}
		}
	}

	sendJSON(w, []byte(`[` + strings.Join(jsonFiles, ",") + `]`))
}
