package model

import (
    "fmt"
    "os"
    "io"
    "crypto/sha256"
)

const MAX_CHUNK_SIZE = 8000 // Chunk size in byte (8KB)

type File struct {
    LocalName string
    Size int
    Metafile []byte
    MetaHash []byte
}

func NewFile(path string) *File {

    metafileHash, metafile := createMetaFile(path)

    return &File{
        LocalName: "test",
        Size: 123,
        Metafile: metafile,
        MetaHash: metafileHash,
    }
}

func createMetaFile(path string) ([]byte, []byte) {
    // Open the file
    f, err := os.Open(path)
    if err != nil {
        fmt.Println("Could not open the file " + path)
        fmt.Println(err)
        return nil, nil
    }
    defer f.Close()

    var metafile []byte
    buffer := make([]byte, MAX_CHUNK_SIZE)

    // Read chunks and build up metafile
    for {
        bytesread, err := f.Read(buffer)

        if err != nil {
            if err != io.EOF {
                fmt.Println(err)
            }

            break
        }

        // Compute hash of chunk
        hashBytes := hash(buffer[:bytesread])
        metafile = append(metafile, hashBytes...)

        fmt.Println("BYTES READ:", bytesread)
        fmt.Printf("HASH: %x", hashBytes)
        fmt.Println()
        //fmt.Println("\nCHUNK:", string(buffer[:bytesread]))
    }

    hash := hash(metafile)

    return hash, metafile
}

func hash(toHash []byte) []byte {
    h := sha256.New()
    h.Write(toHash)
    return h.Sum(nil)
}
