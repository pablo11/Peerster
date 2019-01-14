package model

type SearchRequest struct {
	Origin   string
	Budget   uint64
	Keywords []string
}
