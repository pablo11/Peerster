package model

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/binary"
)

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

func (f *File) Hash() (out [32]byte) {
    h := sha256.New()
    binary.Write(h, binary.LittleEndian, uint32(len(f.Name)))
    h.Write([]byte(f.Name))
    h.Write(f.MetafileHash)
    copy(out[:], h.Sum(nil))
    return
}

func (f *File) HashStr() string {
    hash := f.Hash()
    return hex.EncodeToString(hash[:])
}

func (f *File) Copy() File {
    var metafileHashCopy []byte
    copy(metafileHashCopy[:], f.MetafileHash[:])

    return File{
        Name: f.Name,
        Size: f.Size,
        MetafileHash: metafileHashCopy,
    }
}
