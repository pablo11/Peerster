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

/* JsonMessage and JsonMessages model the JSON response fro request /api/message */
type JsonMessage struct {
    from string
    msg string
}

func (m *JsonMessage) toString() string {
    return `{"from":"` + m.from + `","msg":"` + m.msg + `"}`
}

type JsonMessages struct {
    messages []JsonMessage
}

func (m *JsonMessages) toByte() []byte {
    messagesStr := strings.Join(mapJsonMsg(m.messages, func(m JsonMessage) string{
        return m.toString()
    }), ",")
    return []byte(`[` + messagesStr + `]`)
}


func mapJsonMsg(vs []JsonMessage, f func(JsonMessage) string) []string {
    vsm := make([]string, len(vs))
    for i, v := range vs {
        vsm[i] = f(v)
    }
    return vsm
}
