package gossip

import (
    "fmt"
    "bytes"
    //"encoding/hex"
    "time"
    "github.com/pablo11/Peerster/model"
)



func (g *Gossiper) HandlePktTxPublish(gp *model.GossipPacket) {
    tp := gp.TxPublish

    // Check if I have already seen this transactions since the last block mined
    g.filesForNextBlockMutex.Lock()
    for _, pendingFile := range g.filesForNextBlock {
        if bytes.Equal(pendingFile.File.MetafileHash, tp.File.MetafileHash) && pendingFile.File.Name == tp.File.Name && pendingFile.File.Size == tp.File.Size {
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
    g.addTxPublishToPool(tp)

    // If HopLimit is > 1 decrement and broadcast
    g.broadcastTxPublishDecrementingHopLimit(tp)
}

func (g *Gossiper) HandlePktBlockPublish(gp *model.GossipPacket) {
    bp := gp.BlockPublish

    // Validate PoW
    if !bp.Block.IsValid() {
        fmt.Println("Discarding BlockPublish since the PoW is invalid")
        return
    }

    // Check if I already have it in the blockchain


    // Do I have the block in the blockchain or the parent block in the blockchain?
    parentInBlockchain := false
    blockInBlockchain := false
    g.blockchainMutex.Lock()
    for i := len(g.blockchain) - 1; i >= 0 && !parentInBlockchain && !blockInBlockchain; i-- {
        // Check if block is already in the blockchain
        if bytes.Equal(g.blockchain[i].PrevHash[:], bp.Block.PrevHash[:]) {
            blockInBlockchain = true
        }

        // Check if parent is in the blockchain
        blockHash := g.blockchain[i].Hash()
        if bytes.Equal(blockHash[:], bp.Block.PrevHash[:]) {
            parentInBlockchain = true
        }
    }
    g.blockchainMutex.Unlock()

    if blockInBlockchain {
        // Forward the blockPublish since someone could not have it
        g.broadcastBlockPublishDecrementingHopLimit(bp)
        fmt.Println("Discarding BlockPublish since block is already in the blockchain")
        return
    }

    isGenesisBlock := false
    if !parentInBlockchain {
        // Check if it's the genesis block
        isGenesisBlock = bytes.Equal(bp.Block.PrevHash[:], make([]byte, 32))

        /*
        isGenesisBlock = true
        for _, b := range bp.Block.PrevHash {
            isGenesisBlock = isGenesisBlock && b == byte(0)
        }
        */
    }

    if !parentInBlockchain && !isGenesisBlock {
        fmt.Println("Discarding BlockPublish since parent block isn't in the blockchain")
        return
    }

    // Append block to blockchain
    g.blockchainMutex.Lock()
    g.blockchain = append(g.blockchain, &bp.Block)
    g.blockchainMutex.Unlock()

    g.printBlockchain()

    // Integrate transactions in the filesName mapping
    g.filesNameMutex.Lock()
    for _, trx := range bp.Block.Transactions {
        g.filesName[trx.File.Name] = &trx.File
    }
    g.filesNameMutex.Unlock()

    // If HopLimit is > 1 decrement and broadcast
    g.broadcastBlockPublishDecrementingHopLimit(bp)
}

func (g *Gossiper) printBlockchain() {

    // TODO: store a variable containing the string, add a new block to string when it's added to the blockchain
    // this results in less  calculations of the hash of blocks

    chainStr := ""
    for i := len(g.blockchain) - 1; i >= 0; i-- {
        chainStr += " " + g.blockchain[i].String()
    }

    fmt.Println("CHAIN" + chainStr)
}

func (g *Gossiper) broadcastTxPublish(tp *model.TxPublish) {
    gp := model.GossipPacket{TxPublish: tp}
    go g.sendGossipPacket(&gp, g.peers)
}

func (g *Gossiper) broadcastTxPublishDecrementingHopLimit(tp *model.TxPublish) {
    // If HopLimit is > 1 decrement and broadcast
    if tp.HopLimit > 1 {
        tp.HopLimit -= 1
        g.broadcastTxPublish(tp)
    }
}

func (g *Gossiper) broadcastBlockPublish(bp *model.BlockPublish) {
    gp := model.GossipPacket{BlockPublish: bp}
    go g.sendGossipPacket(&gp, g.peers)
}

func (g *Gossiper) broadcastBlockPublishDecrementingHopLimit(bp *model.BlockPublish) {
    // If HopLimit is > 1 decrement and broadcast
    if bp.HopLimit > 1 {
        bp.HopLimit -= 1
        g.broadcastBlockPublish(bp)
    }
}

func (g *Gossiper) SendTxPublish(file *model.File) {
    tp := model.TxPublish{
        File: *file,
        HopLimit: 10,
    }

    // Add the transaction to the pool of transactions to be added in the next block
    g.addTxPublishToPool(&tp)

    g.broadcastTxPublish(&tp)
}

func (g *Gossiper) addTxPublishToPool(tp *model.TxPublish) {
    // Add only if not already there
    g.filesForNextBlockMutex.Lock()
    alreadyPresnet := false
    for _, trx := range g.filesForNextBlock {
        if trx.File.Name == tp.File.Name && trx.File.Size == tp.File.Size && bytes.Equal(trx.File.MetafileHash, tp.File.MetafileHash) {
            alreadyPresnet = true
        }
    }

    if !alreadyPresnet {
        g.filesForNextBlock = append(g.filesForNextBlock, tp)
    }
    g.filesForNextBlockMutex.Unlock()
}

func (g *Gossiper) startMining() {
    time.Sleep(GENESIS_BLOCK_WAIT_TIME * time.Second)

    fmt.Println("START MINING " + time.Now().String())

    var zeroBytes [32]byte = [32]byte{}

    for {
        if len(g.filesForNextBlock) > 0 {

            // Create the block from filesForNextBlock
            g.blockchainMutex.Lock()
            var prevHash [32]byte = [32]byte{} //zeroBytes[0:32]
            if len(g.blockchain) > 0 {
                prevHash = g.blockchain[len(g.blockchain) - 1].Hash()
            }
            g.blockchainMutex.Unlock()

            g.filesForNextBlockMutex.Lock()
            transactions := make([]model.TxPublish, len(g.filesForNextBlock))
            for i, trx := range g.filesForNextBlock {
                transactions[i] = *trx
            }
            g.filesForNextBlockMutex.Unlock()

            block := model.Block{
                PrevHash: prevHash,
                Nonce: zeroBytes,
                Transactions: transactions,
            }

            // Mine block
            _ = block.Mine()

            // Remove transactions mined from the filesForNextBlock (assume that there are no new TxPhublish added in the middle of the list)
            g.filesForNextBlockMutex.Lock()
            g.filesForNextBlock = g.filesForNextBlock[len(transactions):len(g.filesForNextBlock)]
            g.filesForNextBlockMutex.Unlock()

            // Add block to blockchain
            g.blockchainMutex.Lock()
            g.blockchain = append(g.blockchain, &block)
            g.blockchainMutex.Unlock()

            g.broadcastBlockPublish(&model.BlockPublish{
                Block: block,
                HopLimit: 20,
            })
        } else {
            // Wait a bit before checking again
            time.Sleep(1 * time.Second)
        }
    }

}
