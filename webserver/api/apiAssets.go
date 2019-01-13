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

func (a *ApiHandler) VoteOnAssetVote(w http.ResponseWriter, r *http.Request) {

}
