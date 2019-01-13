package api

import (
    "net/http"
    "strconv"
)

func (a *ApiHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
    myAssetsJson := a.gossiper.Blockchain.GetMyAssetsJson()
    sendJSON(w, []byte(myAssetsJson))
}

func (a *ApiHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    assetName, isAssetNamePresent := r.PostForm["assetName"]
    totSupply, isTotSupplyPresent := r.PostForm["totSupply"]
    if !isAssetNamePresent || len(assetName) != 1 || !isTotSupplyPresent || len(totSupply) != 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    intTotSupply, err := strconv.Atoi(totSupply[0])
    if err != nil {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    go a.gossiper.Blockchain.SendShareTx(assetName[0], a.gossiper.Name, uint64(intTotSupply))

    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) SendShares(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    amount, isAmountPresent := r.PostForm["amount"]
    dest, isDestPresent := r.PostForm["dest"]
    assetName, isAssetNamePresent := r.PostForm["assetName"]
    if !isAmountPresent || len(amount) != 1 || !isDestPresent || len(dest) != 1 || !isAssetNamePresent || len(assetName) != 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    intAmount, err := strconv.Atoi(amount[0])
    if err != nil {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    go a.gossiper.Blockchain.SendShareTx(assetName[0], dest[0], uint64(intAmount))

    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) GetAssetVotes(w http.ResponseWriter, r *http.Request) {
    asset, isPresent := r.URL.Query()["asset"]
    if !isPresent || len(asset[0]) < 1 {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    assetVotesJson := a.gossiper.Blockchain.GetAssetVotesJson(asset[0])

    sendJSON(w, []byte(assetVotesJson))
}

func (a *ApiHandler) CreateAssetVote(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    question, isQuestionPresent := r.PostForm["question"]
    asset, isAssetPresent := r.PostForm["asset"]
    if !isQuestionPresent || len(question) < 1 || question[0] == "" || !isAssetPresent || len(asset) < 1 || asset[0] == "" {
        w.Header().Set("Server", "Cryptop GO server")
        w.WriteHeader(400)
        return
    }

    go a.gossiper.LaunchVotation(question[0], asset[0])

    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}

func (a *ApiHandler) VoteOnAssetVote(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
    question, isQuestionPresent := r.PostForm["question"]
    asset, isAssetPresent := r.PostForm["asset"]
	origin, isOriginPresent := r.PostForm["origin"]
	answer, isAnswerPresent := r.PostForm["answer"]
    if !isQuestionPresent || len(question) < 1 || question[0] == "" || !isAssetPresent || len(asset) < 1 || asset[0] == "" {
        w.Header().Set("Server", "Cryptop GO server1")
        w.WriteHeader(400)
        return
    }

	if !isOriginPresent || len(origin) < 1 || origin[0] == "" || !isAnswerPresent || len(answer) < 1 || answer[0] == "" {
        w.Header().Set("Server", "Cryptop GO server2")
        w.WriteHeader(400)
        return
    }

    answerBool := true
    if answer[0] != "true" && answer[0] != "false" {
        w.Header().Set("Server", "Cryptop GO server3")
        w.WriteHeader(400)
        return
    }

    if answer[0] == "false" {
        answerBool = false
    }

    go a.gossiper.AnswerVotation(question[0], asset[0], origin[0], answerBool)

    w.Header().Set("Server", "Cryptop GO server")
    w.WriteHeader(200)
}
