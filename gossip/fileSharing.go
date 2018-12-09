package gossip

import (
    "time"
    "math"
    "fmt"
    "os"
    "io"
    "io/ioutil"
    "sync"
    "crypto/sha256"
    "encoding/hex"
    "github.com/pablo11/Peerster/model"
)

const MAX_CHUNK_SIZE = 8192 // Chunk size in byte (8KB)
const SHARED_FILES_DIR = "_SharedFiles/"
const DOWNLOADS_DIR = "_Downloads/"
const TIMEOUT_DATA_REQUEST = 5 // Wait 5 sec before asking again the DataRequest
var CHUNKS_DIR = "_Chunks/"

type FileDownload struct {
    LocalName string
    MetaHash []byte
    NextChunkOffset int
    NextChunkHash string
    NbChunks int
    ChunksLocation []string
}

type FileSharing struct {
    gossiper *Gossiper
    // When downloading a file store it here: metaHash->file
    AvailableFiles map[string]*FileDownload
    // Mapping from hash to channel for notifying a data reply
    waitDataRequestChannels map[string]chan bool

    mutex sync.Mutex
}

func NewFileSharing() *FileSharing{
    return &FileSharing{
        AvailableFiles: make(map[string]*FileDownload),
        waitDataRequestChannels: make(map[string]chan bool),
        mutex: sync.Mutex{},
    }
}

func (fs *FileSharing) SetGossiper(g *Gossiper) {
    fs.gossiper = g

    // Make a directory for each node to simulate nodes not in the same location
    CHUNKS_DIR = CHUNKS_DIR + g.Name + "/"
    os.MkdirAll(CHUNKS_DIR, os.ModePerm);
}

func (fs *FileSharing) IndexFile(path string) {
    var err error
    var f *os.File

    // Open the file
    f, err = os.Open(SHARED_FILES_DIR + path)
    if err != nil {
        fmt.Println("ERROR: Could not open the file " + path)
        fmt.Println(err)
        return
    }
    defer f.Close()

    // Check filesize (allow up to maxNbChunks in one single metafile)
    fi, err2 := f.Stat()
    if err2 != nil {
        fmt.Println("ERROR: Could not read file length")
        fmt.Println(err)
        return
    }
    requiredNbChunks := int(math.Ceil(float64(fi.Size()) / MAX_CHUNK_SIZE))
    if requiredNbChunks > int(MAX_CHUNK_SIZE / 32) {
        fmt.Println("ERROR: The file is too large to be indexed")
        fmt.Println()
        return
    }

    var metafile []byte
    var nbChunks uint64 = 0
    // Read chunks and build up metafile
    buffer := make([]byte, MAX_CHUNK_SIZE)
    bytesread := 0
    hashBytes := make([]byte, 0)
    for {
        bytesread, err = f.Read(buffer)

        if err != nil {
            if err != io.EOF {
                fmt.Println(err)
            }
            break
        }

        // Compute hash of chunk
        hashBytes = hash(buffer[:bytesread])
        err = fs.writeBytesToFile(hex.EncodeToString(hashBytes), buffer[:bytesread])
        if (err != nil) {
            return
        }

        // Add chunk to available chunks
        metafile = append(metafile, hashBytes...)
        nbChunks += 1
    }

    metaHash := hash(metafile)

    fmt.Printf("METAHASH: %x\n", metaHash)

    fmt.Println("Number of chunks: ", nbChunks)
    fmt.Println()

    _ = fs.writeBytesToFile(hex.EncodeToString(metaHash), metafile)

    chunksLocation := make([]string, int(nbChunks))
    for i, _ := range chunksLocation {
        chunksLocation[i] = fs.gossiper.Name
    }

    fs.AvailableFiles[hex.EncodeToString(metaHash)] = &FileDownload{
        LocalName: path,
        MetaHash: metaHash,
        NextChunkOffset: int(nbChunks),
        NextChunkHash: "",
        NbChunks: int(nbChunks),
        ChunksLocation: chunksLocation,
    }
}

func (fs *FileSharing) writeBytesToFile(hash string, buffer []byte) error {
    err := ioutil.WriteFile(CHUNKS_DIR + hash, buffer, 0644)
    if (err != nil) {
        fmt.Println("ERROR: While writing metafile or chunk (hash=" + hash + ") to file")
        fmt.Println(err)
    }
    return err
}

// If dest is "", the file is downloaded from multiple sources. Sources are
// retreived from the corresponding FullMatches entry
func (fs *FileSharing) RequestFile(filename, dest, metahash string) {
    byteHash, err := hex.DecodeString(metahash)
    if err != nil {
        fmt.Println("ERROR: The provided request is not an hash")
        return
    }

    // Check if the file is locally available


    // Prepare chunk locations
    chunksLocation := make([]string, MAX_CHUNK_SIZE / 32)
    if dest == "" {
        // Download from multiple sources
        for _, fullMatch := range fs.gossiper.FullMatches {
            if hex.EncodeToString(fullMatch.MetaHash) == metahash {
                chunksLocation = fullMatch.ChunksLocation
                dest = fullMatch.ChunksLocation[0]
                filename = fullMatch.Filename
            }
        }

        // Check if the FullMatch was found
        if dest == "" {
            fmt.Println("ERROR: Could not download file from multiple sources, the FullMatch is missing")
            return
        }
    } else {
        // Download from single origin dest
        for i, _ := range chunksLocation {
            chunksLocation[i] = dest
        }
    }

    // Add this file to the AvailableFiles map
    fs.AvailableFiles[metahash] = &FileDownload{
        LocalName: filename,
        MetaHash: nil,
        NextChunkOffset: 0,
        NextChunkHash: "",
        NbChunks: 0,
        ChunksLocation: chunksLocation,
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
        fmt.Println("ERROR: Invalid packet. Dropped")
        return
    }

    // If this node is not the destinatary forward the packet
    if dr.Destination != fs.gossiper.Name {
        fmt.Println("Forwarding DataReply packet to " + dr.Destination)
        if dr.HopLimit > 1 {
            dr.HopLimit -= 1
            fs.sendDataReply(dr)
        }
        return
    }

    // Notify packet received
    fs.notifyChannelForHash(hex.EncodeToString(dr.HashValue))

    file, isAvailableFiles := fs.AvailableFiles[hex.EncodeToString(dr.HashValue)]
    if isAvailableFiles && file.MetaHash == nil {
        fmt.Println("DOWNLOADING metafile of " + file.LocalName + " from " + dr.Origin)

        nbChunks := len(dr.Data) / 32

        // Store the metafile
        err := fs.writeBytesToFile(hex.EncodeToString(dr.HashValue), dr.Data)
        if (err != nil) {
            return
        }

        // Ask for next Chunk: send DataRequest packet with HashValue equal to the first hash present in the Metafile
        firstChunkHash := fs.getChunkHashFromMetafile(hex.EncodeToString(dr.HashValue), 0)
        if firstChunkHash == nil {
            fmt.Println("ERROR: If we get here, the metafile is empty")
            return
        }

        fs.AvailableFiles[hex.EncodeToString(dr.HashValue)].NextChunkOffset = 0
        fs.AvailableFiles[hex.EncodeToString(dr.HashValue)].MetaHash = dr.HashValue
        fs.AvailableFiles[hex.EncodeToString(dr.HashValue)].NextChunkHash = hex.EncodeToString(firstChunkHash)
        fs.AvailableFiles[hex.EncodeToString(dr.HashValue)].NbChunks = nbChunks
        fs.requestData(fs.AvailableFiles[hex.EncodeToString(dr.HashValue)].ChunksLocation[0], firstChunkHash)
    } else {
        // Store the chunk
        err := fs.writeBytesToFile(hex.EncodeToString(dr.HashValue), dr.Data)
        if (err != nil) {
            return
        }

        // Find metafile requesting this chunk
        for metahash, file := range fs.AvailableFiles {
            if file.NextChunkHash == hex.EncodeToString(dr.HashValue) {
                fs.AvailableFiles[metahash].NextChunkOffset += 1
                fmt.Println("DOWNLOADING " + file.LocalName + " chunk", fs.AvailableFiles[metahash].NextChunkOffset, "from " + dr.Origin)
                nextChunkHash := fs.getChunkHashFromMetafile(metahash, file.NextChunkOffset)
                if nextChunkHash == nil {
                    // The download is complete. Reconstruct the file and save it with the local name
                    fs.reconstructFile(metahash, file.LocalName)
                } else {
                    // Request next chunk
                    fs.AvailableFiles[metahash].NextChunkHash = hex.EncodeToString(nextChunkHash)
                    fs.requestData(fs.AvailableFiles[metahash].ChunksLocation[file.NextChunkOffset], nextChunkHash)
                }

                return
            }
        }
    }
}

func (fs *FileSharing) HandleDataRequest(dr *model.DataRequest) {
    bytesToSend := fs.readChunkFile(hex.EncodeToString(dr.HashValue))
    if (bytesToSend != nil) {
        //fmt.Println("🧰 I have it!", hex.EncodeToString(dr.HashValue))

        dReply := model.DataReply{
            Origin: fs.gossiper.Name,
            Destination: dr.Origin,
            HopLimit: 10,
            HashValue: hash(bytesToSend),
            Data: bytesToSend,
        }

        fs.sendDataReply(&dReply)
        return
    } else {
        //fmt.Println("🙁 I don't have it.", hex.EncodeToString(dr.HashValue))
    }

    if dr.Destination != fs.gossiper.Name {
        // If I don't have the metafile/chunk, forward the request to the destination node
        fmt.Println("Forwarding DataRequest packet to " + dr.Destination)
        if dr.HopLimit > 1 {
            dr.HopLimit -= 1
            fs.sendDataRequest(dr)
        }
    }
}

func (fs *FileSharing) reconstructFile(metahash, filename string) {
    f, err := os.Create(DOWNLOADS_DIR + filename)
    if err != nil {
        fmt.Println("ERROR: While creating the file")
        fmt.Println(err)
        return
    }
    defer f.Close()

    metafileByteOffset := 0
    nextChunkHash := make([]byte, 0)
    chunkToWrite := make([]byte, 0)
    var nbChunks uint64 = 0
    for {
        nextChunkHash = nextChunkHash[:0]
        nextChunkHash = fs.getChunkHashFromMetafile(metahash, metafileByteOffset)
        if nextChunkHash == nil {
            break
        }

        chunkToWrite = fs.readChunkFile(hex.EncodeToString(nextChunkHash))
        if (chunkToWrite == nil) {
            return
        }

        _, err = f.Write(chunkToWrite)
        if err != nil {
            fmt.Println("ERROR: While writing the file")
            fmt.Println(err)
            return
        }
        metafileByteOffset += 1
        nbChunks += 1
    }

    f.Sync()

    fmt.Println("RECONSTRUCTED file " + filename)
    fmt.Println()
}

func (fs *FileSharing) readChunkFile(hash string) []byte {
    data, err := ioutil.ReadFile(CHUNKS_DIR + hash)
    if (err != nil) {
        fmt.Println("ERROR: While reading metafile or chunk (hash=" + hash + ") from file")
        fmt.Println(err)
        return nil
    }
    return data
}

func (fs *FileSharing) getChunkHashFromMetafile(metahash string, offset int) []byte {
    metafile := fs.readChunkFile(metahash)
    if (metafile == nil) {
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

    //fmt.Println("REQUESTING DATA " + hex.EncodeToString(hashValue))

    fs.sendDataRequest(&dr)
}

func (fs *FileSharing) sendDataRequest(dr *model.DataRequest) {
    if dr.Destination == fs.gossiper.Name {
        fs.HandleDataRequest(dr)
        return
    }

    // Get hop-peer if existing
    destPeer := fs.gossiper.GetNextHopForDest(dr.Destination)
    if destPeer == "" {
        return
    }

    gp := model.GossipPacket{DataRequest: dr}
    go fs.gossiper.sendGossipPacket(&gp, []string{destPeer})

    go fs.waitDataReply(dr)
}

func (fs *FileSharing) getChannelForHash(datahash string) chan bool {
    fs.mutex.Lock()
    defer fs.mutex.Unlock()
    _, channelExists := fs.waitDataRequestChannels[datahash]
    if !channelExists {
        fs.waitDataRequestChannels[datahash] = make(chan bool)
    }
    return fs.waitDataRequestChannels[datahash]
}

func (fs *FileSharing) removeChannelForHash(datahash string) {
    fs.mutex.Lock()
    defer fs.mutex.Unlock()
    _, channelExists := fs.waitDataRequestChannels[datahash]
    if channelExists {
        fs.waitDataRequestChannels[datahash] = nil
        delete(fs.waitDataRequestChannels, datahash)
    }
}

func (fs *FileSharing) notifyChannelForHash(datahash string) {
    fs.mutex.Lock()
    defer fs.mutex.Unlock()
    _, channelExists := fs.waitDataRequestChannels[datahash]
    if channelExists {
        fs.waitDataRequestChannels[datahash] <- true
    }
}

func (fs *FileSharing) waitDataReply(dr *model.DataRequest) {
    ticker := time.NewTicker(TIMEOUT_DATA_REQUEST * time.Second)
    defer ticker.Stop()

    datahash := hex.EncodeToString(dr.HashValue)
    channel := fs.getChannelForHash(datahash)

    select {
    case <-channel:
        ticker.Stop()
        fs.removeChannelForHash(datahash)
        // If we get here it's because the DataReply was received

    case <-ticker.C:
        ticker.Stop()
        fs.removeChannelForHash(datahash)

        fmt.Println("DataRequest timed out, resending request")

        // If we get ther, it means that the DataReply was not received
        fs.sendDataRequest(dr)
    }
}

func (fs *FileSharing) sendDataReply(dr *model.DataReply) {
    if dr.Destination == fs.gossiper.Name {
        fs.HandleDataReply(dr)
        return
    }

    destPeer := fs.gossiper.GetNextHopForDest(dr.Destination)
    if destPeer == "" {
        return
    }

    gp := model.GossipPacket{DataReply: dr}
    go fs.gossiper.sendGossipPacket(&gp, []string{destPeer})
}

func hash(toHash []byte) []byte {
    h := sha256.New()
    h.Write(toHash)
    return h.Sum(nil)
}
