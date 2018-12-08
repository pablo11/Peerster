package gossip

import (
    "time"
    "strings"
    "github.com/pablo11/Peerster/model"
)

func (g *Gossiper) HandlePktSearchRequest(gp *model.GossipPacket) {
    // Discard SearchRequest if it's a duplicate
    if g.checkDuplicateSearchRequests(gp.SearchRequest) {
        return
    }

    // Process request locally (if I have chunks of the file in question)


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

func (g *Gossiper) HandlePktSearchReply(gp *model.GossipPacket) {

}
