package gossip

import (
    "fmt"
    "time"
    "strings"
    "math/rand"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/collections"
)

func (g *Gossiper) HandlePktSearchRequest(gp *model.GossipPacket) {
    sr := gp.SearchRequest
    // Discard SearchRequest if it's a duplicate
    if g.checkDuplicateSearchRequests(sr) {
        return
    }

    // Process request locally (if I have files matching the SearchRequest)
    searchResults := make([]*model.SearchResult, 0)
    for filename, file := range g.FileSharing.AvailableFiles {
        for i := 0; i < len(sr.Keywords); i++ {
            if strings.Contains(filename, sr.Keywords[i]) {
                searchResults = append(searchResults, &model.SearchResult{
                    FileName: filename,
                    MetafileHash: file.MetaHash,
                    ChunkMap: make([]uint64, 0),
                    ChunkCount: file.NbChunks,
                })
            }
        }
    }

    // Send a SearchReply if at least one file matches the SearchRequest
    if len(searchResults) > 0 {
        go g.sendSearchReply(sr.Origin, searchResults)
    }

    // Subtract 1 from the request's budget
    sr.Budget -= 1

    // If request's budget is greater than zero redistribute requests to neighbors
    if sr.Budget <= 0 {
        return
    }

    // Propagate the SearchRequest subdividing the budget
    peersBudget := g.subdivideBudget(sr.Budget)
    if len(peersBudget) > 0 {
        for peerAddr, peerBudget := range peersBudget {
            go g.sendSearchRequest(peerAddr, sr.Origin, peerBudget, sr.Keywords)
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

func (g *Gossiper) sendSearchReply(dest string, results []*model.SearchResult) {
    destPeer := g.GetNextHopForDest(dest)
    if destPeer == "" {
        return
    }

    sr := model.SearchReply{
        Origin: g.Name,
        Destination: dest,
        HopLimit: 10,
        Results: results,
    }

    gp := model.GossipPacket{SearchReply: &sr}
    go g.sendGossipPacket(&gp, []string{destPeer})
}

func (g *Gossiper) sendSearchRequest(destPeer, origin string, budget uint64, keywords []string) {
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

func (g *Gossiper) HandlePktSearchReply(gp *model.GossipPacket) {

}
