package gossip

import (
    "fmt"
    "bytes"
    "encoding/hex"
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

    blockHash := bp.Block.Hash()
    blockHashStr := hex.EncodeToString(blockHash[:])

    // Check if we already have the block
    g.blocksMutex.Lock()
    _, isPresent := g.blocks[blockHashStr]
    g.blocksMutex.Unlock()
    if isPresent {
        // Forward the blockPublish since someone could not have it
        g.broadcastBlockPublishDecrementingHopLimit(bp)
        fmt.Println("Discarding BlockPublish since block is already in the blockchain")
        return
    }

    // Store block
    g.blocksMutex.Lock()
    g.blocks[blockHashStr] = &bp.Block
    g.blocksMutex.Unlock()

    // Check if this block is the continuation of a fork
    isNewFork := true
    g.forksMutex.Lock()
    for lastHash, blockchainLength := range g.forks {
        if lastHash == hex.EncodeToString(bp.Block.PrevHash[:]) {
            delete(g.forks, lastHash)
            g.forks[blockHashStr] = blockchainLength + 1
            isNewFork = false

            // Check if we modified the longest chain
            if g.longestChain == lastHash {
                g.updateLongestChain(blockHashStr)
            } else {
                // Check if this fork becomes the longest chain
                if blockchainLength + 1 > g.forks[g.longestChain] {
                    // We are switching to a new longest chain
                    fmt.Printf("FORK-LONGER rewind %d blocks\n", g.computeNbBlocksRewind(blockHashStr, g.longestChain))
                    g.updateLongestChain(blockHashStr)
                }
            }
            break
        }
    }
    g.forksMutex.Unlock()

    if isNewFork {
        // Count length of blockchain
        blockchainLength := g.forkLength(blockHashStr)

        g.forksMutex.Lock()
        g.forks[blockHashStr] = uint64(blockchainLength + 1)
        g.forksMutex.Unlock()

        if bytes.Equal(bp.Block.PrevHash[:], make([]byte, 32)) {
            // It's the genesis block
            g.updateLongestChain(blockHashStr)
        } else {
            fmt.Println("FORK-SHORTER " + hex.EncodeToString(bp.Block.PrevHash[:]))
        }
    }

    // Integrate transactions in the filesName mapping
    g.filesNameMutex.Lock()
    for _, trx := range bp.Block.Transactions {
        g.filesName[trx.File.Name] = &trx.File
    }
    g.filesNameMutex.Unlock()

    // If HopLimit is > 1 decrement and broadcast
    g.broadcastBlockPublishDecrementingHopLimit(bp)
}

func (g *Gossiper) updateLongestChain(blockHashStr string) {
    g.longestChain = blockHashStr
    g.printBlockchain(blockHashStr)
}

func (g *Gossiper) forkLength(blockHash string) uint64 {
    currentHash := blockHash
    var length uint64 = 0
    g.blocksMutex.Lock()

    for {
        block, isPresent := g.blocks[currentHash]
        if !isPresent {
            break
        }

        length += 1
        currentHash = hex.EncodeToString(block.PrevHash[:])
    }

    g.blocksMutex.Unlock()
    return length
}

func (g *Gossiper) computeNbBlocksRewind(newHeadHash, oldHeadHash string) int {
    rewind := 0

    newHeadHashCurrent := newHeadHash
    oldHeadHashCurrent := oldHeadHash

    for {
        new, _ := g.blocks[newHeadHashCurrent]
        old, isNotGenesis := g.blocks[oldHeadHashCurrent]
        oldHash := old.Hash()
        if !isNotGenesis || bytes.Equal(oldHash[:], new.PrevHash[:]) {
            break
        }
        rewind += 1

        newHeadHashCurrent = hex.EncodeToString(new.PrevHash[:])
        oldHeadHashCurrent = hex.EncodeToString(old.PrevHash[:])
    }
    return rewind
}

func (g *Gossiper) printBlockchain(headHash string) {
    currentHash := headHash
    chainStr := ""
    g.blocksMutex.Lock()
    for {
        block, isPresent := g.blocks[currentHash]
        if !isPresent {
            break
        }

        chainStr += " " + block.String()
    }
    g.blocksMutex.Unlock()

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
            var prevHash [32]byte = [32]byte{} //zeroBytes[0:32]
            g.forksMutex.Lock()
            longestChainLength := g.forks[g.longestChain]
            g.forksMutex.Unlock()
            if longestChainLength > 0 {
                g.blocksMutex.Lock()
                prevHash = g.blocks[g.longestChain].PrevHash
                g.blocksMutex.Unlock()
            }


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
            block.Mine()

            // Remove transactions mined from the filesForNextBlock (assume that there are no new TxPhublish added in the middle of the list)
            g.filesForNextBlockMutex.Lock()
            g.filesForNextBlock = g.filesForNextBlock[len(transactions):len(g.filesForNextBlock)]
            g.filesForNextBlockMutex.Unlock()

            bp := &model.BlockPublish{
                Block: block,
                HopLimit: 20,
            }

            g.HandlePktBlockPublish(&model.GossipPacket{BlockPublish: bp})

            g.broadcastBlockPublish(bp)
        } else {
            // Wait a bit before checking again
            time.Sleep(1 * time.Second)
        }
    }

}
