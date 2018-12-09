package model

type FileDownload struct {
    LocalName string
    MetaHash []byte
    NextChunkOffset int
    NextChunkHash string
    NbChunks int
    ChunksLocation []string
}

type ActiveSearch struct {
    Keywords []string
    LastBudget uint64
    NotifyChannel chan bool
    // Metahash -> FileMatch
    Matches map[string]*FileMatch
}

type FileMatch struct {
    Filename string
    MetaHash string
    NbChunks uint64
    // Map: chunck nb -> node having it
    ChunksLocation []string
}

// For blockchain filename to hash mapping
type File struct {
    Name string
    Size int64
    MetafileHash []byte
}
