package api

import (
    "strings"
    "github.com/pablo11/Peerster/util/collections"
)

/* JsonId models the JSON response fro request /api/id */
type JsonId struct {
    name string
    address string
}

func (id *JsonId) toByte() []byte {
    return []byte(`{"name":"` + id.name + `","address":"` + id.address + `"}`)
}

/* JsonPeers models the JSON response fro request /api/node */
type JsonPeers struct {
    peers []string
}

func (peers *JsonPeers) toByte() []byte {
    peersStr := strings.Join(collections.Map(peers.peers, func(p string) string{
        return "\"" + p + "\""
    }), ",")
    return []byte(`[` + peersStr + `]`)
}
