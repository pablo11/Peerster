package model

import (
    "crypto/sha256"
)

type DataReply struct {
    Origin string
    Destination string
    HopLimit uint32
    HashValue []byte
    Data []byte
}

func (dr *DataReply) IsValid() bool {
    h := sha256.New()
    h.Write(dr.Data)
    hash := h.Sum(nil)
    return string(dr.HashValue) == string(hash)
}

func (dr *DataReply) String(isMetafile bool, filename string, chunkNb int) string {
    if isMetafile {
        return "DOWNLOADING metafiler of " + filename + " from " + dr.Origin
    } else {
        return "DOWNLOADING " + filename + " chunk " + string(chunkNb) + " from " + dr.Origin
    }
}
