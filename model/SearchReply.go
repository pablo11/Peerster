package model

type SearchReply struct {
     Origin string
     Destination string
     HopLimit uint32
     Results []*SearchResult
}
