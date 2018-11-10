package api

import (
    "strings"
    "github.com/pablo11/Peerster/util/collections"
)

/* JsonId models the JSON response for request /api/id */
type JsonId struct {
    name string
    address string
}

func (id *JsonId) toByte() []byte {
    return []byte(`{"name":"` + id.name + `","address":"` + id.address + `"}`)
}

/* JsonPeers models the JSON response for request /api/node */
type JsonPeers struct {
    peers []string
}

func (peers *JsonPeers) toByte() []byte {
    peersStr := strings.Join(collections.Map(peers.peers, func(p string) string{
        return "\"" + p + "\""
    }), ",")
    return []byte(`[` + peersStr + `]`)
}

/* JsonOrigins models the JSON response for request /api/origins */
type JsonOrigins struct {
    origins []string
}

func (origins *JsonOrigins) toByte() []byte {
    peersStr := strings.Join(collections.Map(origins.origins, func(o string) string{
        return "\"" + o + "\""
    }), ",")
    return []byte(`[` + peersStr + `]`)
}
