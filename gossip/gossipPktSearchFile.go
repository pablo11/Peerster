package gossip

import (
    "fmt"
    "time"
    "strings"
    "math/rand"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/collections"
)

type ActiveSearch struct {
    Keywords []string
    LastBudget uint64
    NotifyChannel chan bool
    // FileName -> FileChunksMatches
    Matches map[string]*FileMatches
}

type FileMatches struct {
    MetaHash []byte
    NbChunks uint64
    // Map: chunck nb -> node having it
    Chunks map[int]string
}

func (g *Gossiper) HandlePktSearchRequest(gp *model.GossipPacket) {
    sr := gp.SearchRequest
    // Discard SearchRequest if it's a duplicate
    if g.checkDuplicateSearchRequests(sr) {
        return
    }

    // Process request locally (if I have files matching the SearchRequest)
    searchResults := make([]*model.SearchResult, 0)
    for _, file := range g.FileSharing.AvailableFiles {
        filename := file.LocalName
        for i := 0; i < len(sr.Keywords); i++ {
            if strings.Contains(filename, sr.Keywords[i]) {
                // Following my implementation, if I have chunk nb n, I also have all chunks nb smaller than n
                chunkMap := make([]uint64, 0)
                for k := 0; k < file.NextChunkOffset; k++ {
                    chunkMap = append(chunkMap, uint64(k + 1))
                }

                searchResults = append(searchResults, &model.SearchResult{
                    FileName: filename,
                    MetafileHash: file.MetaHash,
                    ChunkMap: chunkMap,
                    ChunkCount: uint64(file.NbChunks),
                })
            }
        }
    }

    // Send a SearchReply if at least one file matches the SearchRequest
    if len(searchResults) > 0 {
        go g.sendSearchReplyFor(sr.Origin, searchResults)
    }

    // Subtract 1 from the request's budget
    sr.Budget -= 1

    // If request's budget is greater than zero redistribute requests to neighbors
    if sr.Budget <= 0 {
        return
    }

    // Propagate the SearchRequest subdividing the budget
    g.budgetPropagation(sr.Budget, sr.Origin, sr.Keywords)
}

func (g *Gossiper) budgetPropagation(budget uint64, origin string, keywords []string) {
    // Propagate the SearchRequest subdividing the budget
    peersBudget := g.subdivideBudget(budget)
    if len(peersBudget) > 0 {
        for peerAddr, peerBudget := range peersBudget {
            go g.sendSearchRequest(origin, peerBudget, keywords, peerAddr, false)
        }
    }
}

// Returns a boolean indicating if the request is a duplicate
func (g *Gossiper) checkDuplicateSearchRequests(sr *model.SearchRequest) bool {
    g.mutex.Lock()

    // Search for duplicate
    searchRequestUid := sr.Origin + strings.Join(sr.Keywords, ",")
    _, isDuplicate := g.ProcessingSearchRequests[searchRequestUid]
    if isDuplicate {
        return true
    }

    // If it isn't a duplicate, put it in the ProcessingSearchRequests array and
    // set a timer to remove it after SEARCH_REQUEST_DUPLICATE_PERIOD seconds
    g.ProcessingSearchRequests[searchRequestUid] = true

    g.mutex.Unlock()

    go func() {
        g.mutex.Lock()
        defer g.mutex.Unlock()

        time.Sleep(SEARCH_REQUEST_DUPLICATE_PERIOD * time.Millisecond)
        delete(g.ProcessingSearchRequests, searchRequestUid)
    }()
    return false
}

func (g *Gossiper) sendSearchReplyFor(dest string, results []*model.SearchResult) {
    sr := model.SearchReply{
        Origin: g.Name,
        Destination: dest,
        HopLimit: 10,
        Results: results,
    }

    g.sendSearchReply(&sr)
}

func (g *Gossiper) sendSearchReply(sr *model.SearchReply) {
    destPeer := g.GetNextHopForDest(sr.Destination)
    if destPeer == "" {
        return
    }
    gp := model.GossipPacket{SearchReply: sr}
    go g.sendGossipPacket(&gp, []string{destPeer})
}

func (g *Gossiper) sendSearchRequest(origin string, budget uint64, keywords []string, destPeer string, randomPeer bool) {
    peer := destPeer
    if randomPeer {
        peer = g.getNRandomPeers(1)[0]
    }

    sr := model.SearchRequest{
        Origin: origin,
        Budget: budget,
        Keywords: keywords,
    }

    gp := model.GossipPacket{SearchRequest: &sr}
    go g.sendGossipPacket(&gp, []string{peer})
}

func (g *Gossiper) subdivideBudget(budget uint64) map[string]uint64 {
    peersBudget := make(map[string]uint64)
    nbPeers := uint64(len(g.peers))
    if nbPeers < 1 {
        return peersBudget
    }

    if budget > nbPeers {
        // Divide the budget among all peers
        minBudgetPerPeer := budget / nbPeers
        for _, p := range g.peers {
            peersBudget[p] = minBudgetPerPeer
        }

        for _, p := range g.getNRandomPeers(budget - minBudgetPerPeer * nbPeers) {
            peersBudget[p] += 1
        }
    } else {
        // Select budget peers at random and send them 1 unit of budget
        for _, p := range g.getNRandomPeers(budget) {
            peersBudget[p] = 1
        }
    }
    return peersBudget
}

func (g *Gossiper) getNRandomPeers(n uint64) []string {
    if len(g.peers) < int(n) {
        fmt.Println("ERROR: not enough peers to select " + string(n) + " at random")
        return make([]string, 0)
    }
    tmpPeers := g.peers
    randomPeers := make([]string, int(n))

    for i := 0; i < int(n); i++ {
        randomPeer := tmpPeers[rand.Intn(len(tmpPeers))]
        randomPeers = append(randomPeers, randomPeer)

        tmpPeers = collections.Filter(tmpPeers, func(p string) bool{
            return p != randomPeer
        })
    }

    return randomPeers
}

func (g *Gossiper) startSearchRequest(budget uint64, keywords []string) {
    if g.ActiveSearchRequest != nil && g.ActiveSearchRequest.LastBudget >= budget {
        fmt.Println("WARNING: The SearchRequest is already beeing searched")
        return
    }

    g.ActiveSearchRequest = &ActiveSearch{
        Keywords: keywords,
        LastBudget: budget,
        NotifyChannel: make(chan bool),
        Matches: make(map[string]*FileMatches)
    }

    go g.sendSearchRequest(g.Name, budget, keywords, "", true)


    // TODO: Propagate or not?

/*
    // Propagate the SearchRequest subdividing the budget
    g.budgetPropagation(sr.Budget, sr.Origin, sr.Keywords)
*/

    // Keep record of SearchRequests sent until 2 mathces are got, in the meantime
    // every 1 second double the budget and sent a new request (up to a threshold of 32)

    ticker := time.NewTicker(SEARCH_REQUEST_BUDGET_DOUBLING_PERIOD * time.Second)
    defer ticker.Stop()

    select {
    case <-g.ActiveSearchRequest.NotifyChannel:
        // Match threshold reached, print
        ticker.Stop()
        g.ActiveSearchRequests = nil

        fmt.Println("✅ 2 MATHCHES")

        // TODO: download the match files



        return

    case <-ticker.C:
        // Send a new SearchRequest doubling budget if smaller than MAX_SEARCH_BUDGET
        ticker.Stop()

        if budget >= MAX_SEARCH_BUDGET {
            fmt.Println("⛔️ MAX BUDGET REACHED")
            return
        }

        go g.startSearchRequest(budget * 2, keywords)
    }
}

func (g *Gossiper) HandlePktSearchReply(gp *model.GossipPacket) {
    sr := gp.SearchReply
    // Forward pkts not for me
    if sr.Destination != g.Name {
        fmt.Println("Forwarding DataRequest packet to " + sr.Destination)
        if sr.HopLimit > 1 {
            sr.HopLimit -= 1
            g.sendSearchReply(sr)
        }
        return
    }

    // Don't care about files that I already have
    for _, result := range sr.Results {
        fmt.Println("FOUND match " + result.FileName + " at " + sr.Origin + " metafile=" + hex.EncodeToString(result.MetafileHash) + " chunks=" + strings.Join(result.ChunkMap, ","))

        _, exists := g.ActiveSearchRequest.Match[result.FileName]
        if !exists {
            g.ActiveSearchRequest.Match[result.FileName] = &FileMatches{
                MetaHash: result.MetafileHash,
                NbChunks: result.ChunkCount,
                Chunks: make(map[int]string),
            }
        }

        // Store location of each chunk
        for _, chunkNb := range result.ChunkMap {
            g.ActiveSearchRequest.Match[result.FileName].Chunks[chunkNb] = sr.Origin
        }
    }

    // Check if there are full matches (files for which we have all chunks)
    nbOfFullMatches := 0
    for filename, fileMatches := range g.ActiveSearchRequest.Match {
        if fileMatches.NbChunks == len(fileMatches.Chunks) {
            nbOfFullMatches += 1
        }
    }

    // Check if I have two matches, in that case send the signal to trigger the end of SearchRequest
    if nbOfFullMatches >= SEARCH_REQUEST_MATCH_THRESHOLD {
        fmt.Println("SEARCH FINISHED")
        g.ActiveSearchRequest.NotifyChannel <- true


        // TODO: send send back the full matchig files, since the ActiveSearchRequest
        // will be overwritten at next SearchRequest

    }

}
