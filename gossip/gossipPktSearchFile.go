package gossip

import (
    "time"
    "strings"
    "github.com/pablo11/Peerster/model"
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


    // If request's budget is greater than zero redistribute requests to neighbors

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

func (g *Gossiper) HandlePktSearchReply(gp *model.GossipPacket) {

}
