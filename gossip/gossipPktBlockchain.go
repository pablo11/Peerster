package gossip

import (
    "fmt"
    "bytes"
    //"encoding/hex"
    "github.com/pablo11/Peerster/model"
)



func (g *Gossiper) HandlePktTxPublish(gp *model.GossipPacket) {
    tp := gp.TxPublish

    // Check if I have already seen this transactions since the last block mined
    g.filesForNextBlockMutex.Lock()
    for _, pendingFile := range g.filesForNextBlock {
        if bytes.Equal(pendingFile.MetafileHash, tp.File.MetafileHash) && pendingFile.Name == tp.File.Name && pendingFile.Size == tp.File.Size {
            g.filesForNextBlockMutex.Unlock()
            fmt.Println("Discarding TxPublish since already received")
            return
        }
    }
    g.filesForNextBlockMutex.Unlock()

    // Check that the TxPublish is valid, i.e. nobody has already claimed the name
    g.filesNameMutex.Lock()
    _, filenameAlreadyCaimed := g.filesName[tp.File.Name]
    if filenameAlreadyCaimed {
        g.filesNameMutex.Unlock()
        fmt.Println("Discarding TxPublish since name already claimed")
        return
    }
    g.filesNameMutex.Unlock()

    // If it's valid and has not yet been seen, store it in the pool of trx to be added in next block
    g.filesForNextBlockMutex.Lock()
    g.filesForNextBlock = append(g.filesForNextBlock, &tp.File)
    g.filesForNextBlockMutex.Unlock()

    // If HopLimit is > 1 decrement and broadcast
    if tp.HopLimit > 1 {
        tp.HopLimit -= 1
        g.broadcastTxPublish(tp)
    }
}

func (g *Gossiper) HandlePktBlockPublish(gp *model.GossipPacket) {
    bp := gp.BlockPublish

    // Do I have the parent block in the blockchain?
    parentInBlockchain := false
    g.blockchainMutex.Lock()
    for i := len(g.blockchain) - 1; i >= 0; i-- {
        blockHash := g.blockchain[i].Hash()
        if bytes.Equal(blockHash[:], bp.Block.PrevHash[:]) {
            parentInBlockchain = true
        }
    }
    g.blockchainMutex.Unlock()

    // Check if it's the genesis block
    isGenesisBlock := true
    for _, b := range bp.Block.PrevHash {
        isGenesisBlock = isGenesisBlock && b == byte(0)
    }

    if !parentInBlockchain && !isGenesisBlock {
        fmt.Println("Discarding BlockPublish since parent block isn't in the blockchain")
        return
    }

    // Validate PoW
    if !bp.Block.IsValid() {
        fmt.Println("Discarding BlockPublish since the PoW is invalid")
        return
    }

    // Append block to blockchain
    g.blockchainMutex.Lock()
    g.blockchain = append(g.blockchain, &bp.Block)
    g.blockchainMutex.Unlock()

    // Integrate transactions in the filesName mapping
    g.filesNameMutex.Lock()
    for _, trx := range bp.Block.Transactions {
        g.filesName[trx.File.Name] = &trx.File
    }
    g.filesNameMutex.Unlock()

    // If HopLimit is > 1 decrement and broadcast
    if bp.HopLimit > 1 {
        bp.HopLimit -= 1
        g.broadcastBlockPublish(bp)
    }
}

func (g *Gossiper) broadcastTxPublish(tp *model.TxPublish) {
    gp := model.GossipPacket{TxPublish: tp}
    go g.sendGossipPacket(&gp, g.peers)
}

func (g *Gossiper) broadcastBlockPublish(bp *model.BlockPublish) {
    gp := model.GossipPacket{BlockPublish: bp}
    go g.sendGossipPacket(&gp, g.peers)
}

func (g *Gossiper) SendTxPublish(file *model.File) {
    tp := model.TxPublish{
        File: *file,
        HopLimit: 10,
    }

    g.broadcastTxPublish(&tp)
}
