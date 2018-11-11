package gossip

import (
    "fmt"
    "os"
    "io"
    "crypto/sha256"
    "encoding/hex"
    "github.com/pablo11/Peerster/model"
)

const MAX_CHUNK_SIZE = 8192 // Chunk size in byte (8KB)
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

    // Read chunks and build up metafile
    for {
        buffer := make([]byte, MAX_CHUNK_SIZE)
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
        fs.chunks[hex.EncodeToString(hashBytes)] = buffer[:bytesread]
        metafile = append(metafile, hashBytes...)
    }

    metaHash := hash(metafile)

    fmt.Printf("METAHASH: %x", metaHash)
    fmt.Println()

    fs.metafiles[hex.EncodeToString(metaHash)] = metafile
}

func (fs *FileSharing) RequestFile(filename, dest, metahash string) {
    // Add this file to the downloading map
    fs.downloading[metahash] = &File{
        LocalName: filename,
        MetaHash: nil,
        NextChunkOffset: 0,
        NextChunkHash: "",
    }

    byteHash, err := hex.DecodeString(metahash)
    if err != nil {
        fmt.Println("⚠️ ERROR: The provided request is not an hash")
        return
    }

    // Prepare and send the request
    dr := model.DataRequest{
        Origin: fs.gossiper.Name,
        Destination: dest,
        HopLimit: 10,
        HashValue: byteHash,
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

    file, isDownloading := fs.downloading[hex.EncodeToString(dr.HashValue)]
    if isDownloading && file.MetaHash == nil {
        fmt.Println("DOWNLOADING metafile of " + file.LocalName + " from " + dr.Origin)

        // Store the metafile
        fs.metafiles[hex.EncodeToString(dr.HashValue)] = dr.Data

        // Ask for next Chunk: send DataRequest packet with HashValue equal to the first hash present in the Metafile
        firstChunkHash := fs.getChunkHashFromMetafile(hex.EncodeToString(dr.HashValue), 0)
        if firstChunkHash == nil {
            fmt.Println("⚠️ ERROR: If we get here, the metafile is empty")
            return
        }

        fs.requestData(dr.Origin, firstChunkHash)
        fs.downloading[hex.EncodeToString(dr.HashValue)].NextChunkOffset = 0
        fs.downloading[hex.EncodeToString(dr.HashValue)].MetaHash = dr.HashValue
        fs.downloading[hex.EncodeToString(dr.HashValue)].NextChunkHash = hex.EncodeToString(firstChunkHash)
    } else {
        // Store the chunk
        fs.chunks[hex.EncodeToString(dr.HashValue)] = dr.Data

        // Find metafile requesting this chunk
        for metahash, file := range fs.downloading {
            if file.NextChunkHash == hex.EncodeToString(dr.HashValue) {
                fs.downloading[metahash].NextChunkOffset += 1
                fmt.Println("DOWNLOADING " + file.LocalName + " chunk", fs.downloading[metahash].NextChunkOffset, "from " + dr.Origin)
                nextChunkHash := fs.getChunkHashFromMetafile(metahash, file.NextChunkOffset)
                if nextChunkHash == nil {
                    // The download is complete. Reconstruct the file and save it with the local name
                    fs.reconstructFile(metahash, file.LocalName)
                } else {
                    // Request next chunk
                    fs.requestData(dr.Origin, nextChunkHash)
                    fs.downloading[metahash].NextChunkHash = hex.EncodeToString(nextChunkHash)
                }

                return
            }
        }
    }
}

func (fs *FileSharing) HandleDataRequest(dr *model.DataRequest) {
    // Check if I have the piece of data
    var bytesToSend []byte = nil
    data, isMetafileAvailable := fs.metafiles[hex.EncodeToString(dr.HashValue)]
    if isMetafileAvailable {
        bytesToSend = data
    } else {
        data, isChunkAvailable := fs.chunks[hex.EncodeToString(dr.HashValue)]
        if isChunkAvailable {
            bytesToSend = data
        }
    }

    if bytesToSend != nil {
        fmt.Println("🧰 I have it!", hex.EncodeToString(dr.HashValue))

        dReply := &model.DataReply{
            Origin: fs.gossiper.Name,
            Destination: dr.Origin,
            HopLimit: 10,
            HashValue: hash(bytesToSend),
            Data: bytesToSend,
        }

        fs.sendDataReply(dReply)
        return
    }

    if dr.Destination != fs.gossiper.Name {
        // If I don't have the metafile/chunk, forward the request to the destination node
        fmt.Println("🧠 Forwarding DataRequest packet to " + dr.Destination)
        if dr.HopLimit > 1 {
            dr.HopLimit -= 1
            fs.sendDataRequest(dr)
        }
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

        chunkToWrite := fs.chunks[hex.EncodeToString(nextChunkHash)]
        _, err := f.Write(chunkToWrite)
        if err != nil {
            fmt.Println("⚠️ ERROR: While writing the file")
            fmt.Println(err)
            return
        }
        metafileByteOffset += 1
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
    if byteOffset >= len(metafile) {
        return nil
    }

    endByteOffset := byteOffset + 32
    if endByteOffset > len(metafile) {
        endByteOffset = len(metafile)
    }

    return metafile[byteOffset:endByteOffset]
}


func (fs *FileSharing) requestData(dest string, hashValue []byte) {
    // Prepare and send DataRequest packet
    dr := model.DataRequest{
        Origin: fs.gossiper.Name,
        Destination: dest,
        HopLimit: 10,
        HashValue: hashValue,
    }

    fmt.Println("REQUESTING DATA " + hex.EncodeToString(hashValue))

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
