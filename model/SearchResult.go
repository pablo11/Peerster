package model

type SearchResult struct {
     FileName string
     MetafileHash []byte
     ChunkMap []uint64
     ChunkCount uint64
}
