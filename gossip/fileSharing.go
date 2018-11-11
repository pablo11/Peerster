package gossip

import (
    "fmt"
    "os"
    "io"
    "crypto/sha256"
    "github.com/pablo11/Peerster/model"
)

const MAX_CHUNK_SIZE = 8000 // Chunk size in byte (8KB)
const SHARED_FILES_DIR = "_SharedFiles/"
const DOWNLOADS_DIR = "_Downloads/"

type File struct {
    LocalName string
    MetaHash []byte
    NextChunkOffset int
    NextChunkHash string
}

type FileSharing struct {
    gossiper *Gossiper
    // Store all chunks and metafiles present on this node in a map hash->bytes
    metafiles map[string][]byte
    chunks map[string][]byte

    // When downloading a file store it here: metaHash->file
    downloading map[string]*File
}

func NewFileSharing() *FileSharing{
    return &FileSharing{
        metafiles: make(map[string][]byte),
        chunks: make(map[string][]byte),
        downloading: make(map[string]*File),
    }
}

func (fs *FileSharing) SetGossiper(g *Gossiper) {
    fs.gossiper = g
}

func (fs *FileSharing) IndexFile(path string) {
    // Open the file
    f, err := os.Open(SHARED_FILES_DIR + path)
    if err != nil {
        fmt.Println("ERROR: Could not open the file " + path)
        fmt.Println(err)
        return
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

        // Add chunk to available chunks
        fs.chunks[string(hashBytes)] = buffer[:bytesread]

        metafile = append(metafile, hashBytes...)

        fmt.Println("BYTES READ:", bytesread)
        fmt.Printf("HASH: %x", hashBytes)
        fmt.Println()
        //fmt.Println("\nCHUNK:", string(buffer[:bytesread]))
    }

    metaHash := hash(metafile)

    fs.metafiles[string(metaHash)] = metafile
}

func (fs *FileSharing) RequestFile(filename, dest, metahash string) {
    _, isAlreadyPresent := fs.downloading[metahash]
    if isAlreadyPresent {
        fmt.Println("ALREADy DOWNLOADING THIS PIECE OF DATA")
        return
    }

    // Add this file to the downloading map
    fs.downloading[metahash] = &File{
        LocalName: filename,
        MetaHash: nil,
        NextChunkOffset: 0,
        NextChunkHash: "",
    }

    // Prepare and send the request
    dr := model.DataRequest{
        Origin: fs.gossiper.Name,
        Destination: dest,
        HopLimit: 10,
        HashValue: []byte(metahash),
    }

    fs.sendDataRequest(&dr)
}

func (fs *FileSharing) HandleDataReply(dr *model.DataReply) {
    // Check validity of packet: HashValue must be equal to hash(dr.Data)
    if !dr.IsValid() {
        fmt.Println("⚠️ ERROR: invalid packet. Dropped")
        return
    }

    // If this node is not the destinatary forward the packet
    if dr.Destination != fs.gossiper.Name {
        fmt.Println("🧠 Forwarding DataReply packet to " + dr.Destination)
        if dr.HopLimit > 1 {
            dr.HopLimit -= 1
            fs.sendDataReply(dr)
        }
        return
    }

    file, isDownloading := fs.downloading[string(dr.HashValue)]
    if !isDownloading {
        fmt.Println("⁉️ NOT WAITING FOR THIS FILE")
        return
    }

    if file.MetaHash == nil {
        fmt.Println("DOWNLOADING metafile of " + file.LocalName + " from " + dr.Origin)

        // Store the metafile
        fs.metafiles[string(dr.HashValue)] = dr.Data

        // Ask for next Chunk: send DataRequest packet with HashValue equal to the first hash present in the Metafile
        firstChunkHash := fs.getChunkHashFromMetafile(string(dr.HashValue), 0)
        if firstChunkHash == nil {
            fmt.Println("⚠️ ERROR: If we get here, the metafile is empty")
            return
        }

        fs.requestData(dr.Origin, firstChunkHash)
        fs.downloading[string(dr.HashValue)].NextChunkOffset = 0
        fs.downloading[string(dr.HashValue)].MetaHash = dr.HashValue
        fs.downloading[string(dr.HashValue)].NextChunkHash = string(firstChunkHash)
    } else {
        // Store the chunk
        fs.chunks[string(dr.HashValue)] = dr.Data

        // Find metafile requesting this chunk
        for metahash, file := range fs.downloading {
            if file.NextChunkHash == string(dr.HashValue) {
                fmt.Println("DOWNLOADING " + file.LocalName + " chunk " + string(fs.downloading[metahash].NextChunkOffset) + " from " + dr.Origin)
                fs.downloading[metahash].NextChunkOffset += 1
                nextChunkHash := fs.getChunkHashFromMetafile(string(dr.HashValue), file.NextChunkOffset)
                if nextChunkHash == nil {
                    // The download is complete. Reconstruct the file and save it with the local name
                    fs.reconstructFile(metahash, file.LocalName)
                } else {
                    // Request next chunk
                    fs.requestData(dr.Origin, nextChunkHash)
                    fs.downloading[metahash].NextChunkHash = string(nextChunkHash)
                }

                return
            }
        }
    }
}

func (fs *FileSharing) HandleDataRequest(dr *model.DataRequest) {
    // Check if I have the piece of data
    var bytesToSend []byte = nil
    data, isMetafileAvailable := fs.metafiles[string(dr.HashValue)]
    if isMetafileAvailable {
        bytesToSend = data
    } else {
        data, isChunkAvailable := fs.chunks[string(dr.HashValue)]
        if isChunkAvailable {
            bytesToSend = data
        }
    }

    if bytesToSend != nil {
        dReply := &model.DataReply{
            Origin: fs.gossiper.Name,
            Destination: dr.Origin,
            HopLimit: 10,
            HashValue: dr.HashValue,
            Data: bytesToSend,
        }

        fs.sendDataReply(dReply)
    }

    // If I don't have the metafile/chunk, forward the request to the destination node
    fmt.Println("🧠 Forwarding DataRequest packet to " + dr.Destination)
    if dr.HopLimit > 1 {
        dr.HopLimit -= 1
        fs.sendDataRequest(dr)
    }
}

func (fs *FileSharing) reconstructFile(metahash, filename string) {
    f, err := os.Create(DOWNLOADS_DIR + filename)
    if err != nil {
        fmt.Println("⚠️ ERROR: While creating the file")
        fmt.Println(err)
        return
    }
    defer f.Close()

    metafileByteOffset := 0
    for {
        nextChunkHash := fs.getChunkHashFromMetafile(metahash, metafileByteOffset)
        if nextChunkHash == nil {
            break
        }

        chunkToWrite := fs.chunks[string(nextChunkHash)]
        _, err := f.Write(chunkToWrite)
        if err != nil {
            fmt.Println("⚠️ ERROR: While writing the file")
            fmt.Println(err)
            return
        }
        metafileByteOffset += 32
    }

    f.Sync()

    // Remove it from downloading
    delete(fs.downloading, metahash)

    fmt.Println("RECONSTRUCTED file " + filename)
    fmt.Println()
}

func (fs *FileSharing) getChunkHashFromMetafile(metahash string, offset int) []byte {
    metafile, isPresent := fs.metafiles[metahash]
    if !isPresent {
        fmt.Println("⚠️ ERROR: If we get here, the metafile isn't available for some reason")
        return nil
    }

    byteOffset := offset * 32
    if byteOffset > len(metafile) {
        return nil
    }

    endByteOffset := byteOffset + 32
    if endByteOffset > len(metafile) {
        endByteOffset = endByteOffset
    }

    chunkHash := metafile[byteOffset:endByteOffset]
    if len(chunkHash) == 32 {
        fmt.Println()
        fmt.Println("⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️⛔️")
        fmt.Println()
    }
    return chunkHash
}


func (fs *FileSharing) requestData(dest string, hashValue []byte) {
    // Prepare and send DataRequest packet
    dr := model.DataRequest{
        Origin: fs.gossiper.Name,
        Destination: dest,
        HopLimit: 10,
        HashValue: hashValue,
    }

    fs.sendDataRequest(&dr)
}

func (fs *FileSharing) sendDataRequest(dr *model.DataRequest) {
    // Get hop-peer if existing
    destPeer := fs.gossiper.GetNextHopForDest(dr.Destination)
    if destPeer == "" {
        return
    }

    gp := model.GossipPacket{DataRequest: dr}
    fs.gossiper.sendGossipPacket(&gp, []string{destPeer})
}

func (fs *FileSharing) sendDataReply(dr *model.DataReply) {
    destPeer := fs.gossiper.GetNextHopForDest(dr.Destination)
    if destPeer == "" {
        return
    }

    gp := model.GossipPacket{DataReply: dr}
    fs.gossiper.sendGossipPacket(&gp, []string{destPeer})
}

func hash(toHash []byte) []byte {
    h := sha256.New()
    h.Write(toHash)
    return h.Sum(nil)
}
