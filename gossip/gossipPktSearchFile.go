package gossip

import (
    "fmt"
    "time"
    "strings"
    "strconv"
    "math/rand"
    "encoding/hex"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/collections"
)

type ActiveSearch struct {
    Keywords []string
    LastBudget uint64
    NotifyChannel chan bool
    // Metahash -> FileMatch
    Matches map[string]*FileMatch
}

type FileMatch struct {
    Filename string
    MetaHash []byte
    NbChunks uint64
    // Map: chunck nb -> node having it
    ChunksLocation []string
}

func (g *Gossiper) HandlePktSearchRequest(gp *model.GossipPacket) {
    sr := gp.SearchRequest
    // Discard SearchRequest if it's a duplicate
    if g.checkDuplicateSearchRequests(sr.Origin, sr.Keywords) {
        return
    }

    // Process request locally (if I have files matching the SearchRequest)
    g.searchFileLocally(sr.Keywords, sr.Origin)

    // Subtract 1 from the request's budget
    sr.Budget -= 1

    // If request's budget is greater than zero redistribute requests to neighbors
    if sr.Budget <= 0 {
        return
    }

    // Propagate the SearchRequest subdividing the budget
    g.budgetPropagation(sr.Budget, sr.Origin, sr.Keywords)
}

func (g *Gossiper) searchFileLocally(keywords []string, origin string) {
    fmt.Println("Searching files locally")

    searchResults := make([]*model.SearchResult, 0)
    for _, file := range g.FileSharing.AvailableFiles {
        filename := file.LocalName
        for i := 0; i < len(keywords); i++ {
            if strings.Contains(filename, keywords[i]) {
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
        go g.sendSearchReplyFor(origin, searchResults)
    }
}

func (g *Gossiper) budgetPropagation(budget uint64, origin string, keywords []string) {
    // Propagate the SearchRequest subdividing the budget
    peersBudget := g.subdivideBudget(budget)
    if len(peersBudget) > 0 {
        for peerAddr, peerBudget := range peersBudget {
            go g.sendSearchRequest(origin, peerBudget, keywords, peerAddr)
        }
    }
}

// Returns a boolean indicating if the request is a duplicate
func (g *Gossiper) checkDuplicateSearchRequests(origin string, keywords []string) bool {
    g.processingSearchRequestsMutex.Lock()

    // Search for duplicate
    searchRequestUid := origin + strings.Join(keywords, ",")
    _, isDuplicate := g.processingSearchRequests[searchRequestUid]
    if isDuplicate {
        g.processingSearchRequestsMutex.Unlock()
        return true
    }

    // If it isn't a duplicate, put it in the ProcessingSearchRequests array and
    // set a timer to remove it after SEARCH_REQUEST_DUPLICATE_PERIOD seconds
    g.processingSearchRequests[searchRequestUid] = true

    g.processingSearchRequestsMutex.Unlock()

    go func() {
        time.Sleep(SEARCH_REQUEST_DUPLICATE_PERIOD * time.Millisecond)
        g.processingSearchRequestsMutex.Lock()
        delete(g.processingSearchRequests, searchRequestUid)
        defer g.processingSearchRequestsMutex.Unlock()
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

    if dest == g.Name {
        // Handle search request from client
        g.HandlePktSearchReply(&model.GossipPacket{SearchReply: &sr})
    } else {
        g.sendSearchReply(&sr)
    }
}

func (g *Gossiper) sendSearchReply(sr *model.SearchReply) {
    destPeer := g.GetNextHopForDest(sr.Destination)
    if destPeer == "" {
        return
    }
    gp := model.GossipPacket{SearchReply: sr}
    go g.sendGossipPacket(&gp, []string{destPeer})
}

func (g *Gossiper) sendSearchRequest(origin string, budget uint64, keywords []string, destPeer string) {
    sr := model.SearchRequest{
        Origin: origin,
        Budget: budget,
        Keywords: keywords,
    }
    gp := model.GossipPacket{SearchRequest: &sr}
    go g.sendGossipPacket(&gp, []string{destPeer})
}

func (g *Gossiper) subdivideBudget(budget uint64) map[string]uint64 {
    peersBudget := make(map[string]uint64)
    nbPeers := uint64(len(g.peers))
    if nbPeers < 1 {
        return peersBudget
    }

    if budget >= nbPeers {
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

func (g *Gossiper) StartSearchRequest(budget uint64, keywords []string, startExpandingRing bool) {
    fmt.Println("💡 SEARCH STARTED for keywords=" + strings.Join(keywords, ","))

    // Discard SearchRequest if it's a duplicate
    if g.checkDuplicateSearchRequests(g.Name, keywords) {
        return
    }

    g.activeSearchRequestMutex.Lock()

    doLocalSearch := g.activeSearchRequest == nil

    if g.activeSearchRequest != nil && g.activeSearchRequest.LastBudget >= budget {
        fmt.Println("WARNING: A SearchRequest is already beeing searched")
        return
    }

    g.activeSearchRequest = &ActiveSearch{
        Keywords: keywords,
        LastBudget: budget,
        NotifyChannel: make(chan bool),
        Matches: make(map[string]*FileMatch),
    }

    g.activeSearchRequestMutex.Unlock()

    if doLocalSearch {
        go g.searchFileLocally(keywords, g.Name)
    }

    // Propagate the SearchRequest subdividing the budget
    go g.budgetPropagation(budget, g.Name, keywords)

    // Keep record of SearchRequests sent until 2 mathces are got, in the meantime
    // every 1 second double the budget and sent a new request (up to a threshold of 32)

    ticker := time.NewTicker(SEARCH_REQUEST_BUDGET_DOUBLING_PERIOD * time.Second)
    defer ticker.Stop()

    select {
    case <-g.activeSearchRequest.NotifyChannel:
        // Match threshold reached, print
        fmt.Println("✅ 2 MATHCHES")
        if startExpandingRing {
            ticker.Stop()
        }
        g.activeSearchRequestMutex.Lock()
        g.activeSearchRequest = nil
        g.activeSearchRequestMutex.Unlock()

        fmt.Println("✅ 2 MATHCHES 1")

        // TODO: download the match files



        return

    case <-ticker.C:
        // Send a new SearchRequest doubling budget if smaller than MAX_SEARCH_BUDGET
        ticker.Stop()

        if !startExpandingRing {
            return
        }

        if budget >= MAX_SEARCH_BUDGET {
            fmt.Println("⛔️ MAX BUDGET REACHED")
            return
        }

        go g.StartSearchRequest(budget * 2, keywords, startExpandingRing)
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
        chunkMapStr := make([]string, len(result.ChunkMap))
        for i := 0; i < len(result.ChunkMap); i++ {
            chunkMapStr[i] = strconv.Itoa(int(result.ChunkMap[i]))
        }

        hexMetahash := hex.EncodeToString(result.MetafileHash)
        fmt.Println("FOUND match " + result.FileName + " at " + sr.Origin + " metafile=" + hexMetahash + " chunks=" + strings.Join(chunkMapStr, ","))

        g.activeSearchRequestMutex.Lock()

        _, exists := g.activeSearchRequest.Matches[hexMetahash]
        if !exists {
            g.activeSearchRequest.Matches[hexMetahash] = &FileMatch{
                Filename: result.FileName,
                MetaHash: result.MetafileHash,
                NbChunks: result.ChunkCount,
                ChunksLocation: make([]string, result.ChunkCount),
            }
        }

        // Store location of each chunk
        for _, chunkNb := range result.ChunkMap {
            g.activeSearchRequest.Matches[hexMetahash].ChunksLocation[int(chunkNb) - 1] = sr.Origin
        }

        g.activeSearchRequestMutex.Unlock()

        // Check if it's a full match
        if len(g.activeSearchRequest.Matches[hexMetahash].ChunksLocation) == int(g.activeSearchRequest.Matches[hexMetahash].NbChunks) {
            isDuplicate := false
            g.FullMatchesMutex.Lock()
            for _, fullMatch := range g.FullMatches {
                if hex.EncodeToString(fullMatch.MetaHash) == hexMetahash {
                    isDuplicate = true
                }
            }

            if !isDuplicate {
                g.FullMatches = append(g.FullMatches, g.activeSearchRequest.Matches[hexMetahash])
                if len(g.FullMatches) >= SEARCH_REQUEST_MATCH_THRESHOLD {
                    fmt.Println("SEARCH FINISHED")
                    g.activeSearchRequest.NotifyChannel <- true
                }
            }
            g.FullMatchesMutex.Unlock()
        }
    }
}
