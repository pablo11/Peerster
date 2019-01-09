package gossip
/*
import (
    "fmt"
    "time"
    "bytes"
    "math/rand"
    "encoding/hex"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/debug"
)

func (g *Gossiper) HandlePktTxPublish(gp *model.GossipPacket) {
    tp := gp.TxPublish

    // Validate signature
    if tp.Transaction.Signature != nil {

        // TODO: validate signature (get the public key of the identity with name tp.Transaction.Signature.Origin and call the IsValid method of Signature)

    }

    // Validate transaction according to its content
    isValid, errorMsg := g.isValidTx(tp.Transaction)
    if !isValid {
        fmt.Println("Discarding TxPublish: " + errorMsg)
        return
    }

    // If it's valid and has not yet been seen, store it in the pool of trx to be added in next block
    g.addTxPublishToPool(*tp)

    // If HopLimit is > 1 decrement and broadcast
    g.broadcastTxPublishDecrementingHopLimit(tp)
}

func (g *Gossiper) isValidTx(tx *model.Transaction) (isValid bool, errorMsg string) {
    errorMsg = nil
    isValid = nil

    switch {
        case t.File != nil:
            // Check if I have already seen this transactions since the last block mined
            g.filesNameMutex.Lock()
            _, filenameAlreadyCaimed := g.filesName[tx.File.Name]
            if filenameAlreadyCaimed {
                g.filesNameMutex.Unlock()
                errorMsg = "Filename already claimed"
                isValid = false
                return
            }
            g.filesNameMutex.Unlock()


        case t.Identity != nil:

            // TODO: implement

    }
}

func (g *Gossiper) addTxToPool(t model.Transaction) {
    // Add only if not already there
    g.txsPoolMutex.Lock()
    alreadyPresnet := false
    for _, tx := range g.txsPool {
        if tx.File.Name == t.File.Name {
            alreadyPresnet = true
        }
    }

    if !alreadyPresnet {
        var metafileHashCopy []byte = make([]byte, 32)
        copy(metafileHashCopy[:], tp.File.MetafileHash[:])
        tx := model.TxPublish{
            File: model.File{
                Name: tp.File.Name,
                Size: tp.File.Size,
                MetafileHash: metafileHashCopy,
            },
            HopLimit: tp.HopLimit,
        }

        g.txsForNextBlock = append(g.txsForNextBlock, tx)
    }

    g.txsPoolMutex.Unlock()
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
        //fmt.Println("Discarding BlockPublish since block is already in the blockchain")
        return
    }

    //fmt.Printf("🧩 NEW BLOCK %+v\n\n", bp)

    // Store block
    newBlock := g.deepCopyBlock(bp.Block)

    g.blocksMutex.Lock()
    g.blocks[blockHashStr] = &newBlock
    g.blocksMutex.Unlock()

    // Check if this block is the continuation of a fork
    isNewFork := true
    g.forksMutex.Lock()
    for lastHash, blockchainLength := range g.forks {
        if lastHash == bp.Block.PrevHashStr() {
            delete(g.forks, lastHash)
            g.forks[blockHashStr] = blockchainLength + 1
            isNewFork = false

            // Check if we modified the longest chain
            if g.longestChain == lastHash {
                g.updateLongestChain(blockHashStr, &bp.Block)
            } else {
                // Check if this fork becomes the longest chain
                if blockchainLength + 1 > g.forks[g.longestChain] {
                    // We are switching to a new longest chain
                    fmt.Printf("FORK-LONGER rewind %d blocks\n", g.computeNbBlocksRewind(blockHashStr, g.longestChain))
                    g.updateLongestChain(blockHashStr, nil)
                } else {
                    fmt.Println("BLOCK ADDED TO A SHORTER FORK")
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
        g.forks[blockHashStr] = blockchainLength
        g.forksMutex.Unlock()

        if g.longestChain == "" && bytes.Equal(bp.Block.PrevHash[:], make([]byte, 32)) {
            // It's the first genesis block
            g.updateLongestChain(blockHashStr, &bp.Block)
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

func (g *Gossiper) deepCopyBlock(b model.Block) model.Block {
    var newPrevHash [32]byte
    var newNonce [32]byte
    copy(newPrevHash[:], b.PrevHash[:])
    copy(newNonce[:], b.Nonce[:])

    var newTransactions []model.TxPublish = make([]model.TxPublish, len(b.Transactions))
    for i, tx := range b.Transactions {
        var metafileHashCopy []byte = make([]byte, 32)
        copy(metafileHashCopy[:], tx.File.MetafileHash[:])
        newTransactions[i] = model.TxPublish{
            File: model.File{
                Name: tx.File.Name,
                Size: tx.File.Size,
                MetafileHash: metafileHashCopy,
            },
            HopLimit: tx.HopLimit,
        }
    }

    return model.Block{
        PrevHash: newPrevHash,
        Nonce: newNonce,
        Transactions: newTransactions,
    }
}

func (g *Gossiper) updateLongestChain(blockHashStr string, block *model.Block) {
    g.longestChain = blockHashStr

    // Remove transactions from pool of transactions for next block to mine
    g.txsForNextBlockMutex.Lock()
    txsForNextBlockTmp := make([]model.TxPublish, 0)
    if block != nil {
        for _, f := range g.txsForNextBlock {
            toAdd := true
            for _, tx := range block.Transactions {
                if tx.HashStr() == f.HashStr() {
                    toAdd = false
                }
            }
            if toAdd {
                txsForNextBlockTmp = append(txsForNextBlockTmp, f)
            }
        }
    }
    g.txsForNextBlock = txsForNextBlockTmp
    g.txsForNextBlockMutex.Unlock()

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
        currentHash = block.PrevHashStr()
    }

    g.blocksMutex.Unlock()
    return length
}

func (g *Gossiper) computeNbBlocksRewind(newHeadHash, oldHeadHash string) int {
    rewind := 0

    newHeadHashCurrent := newHeadHash
    oldHeadHashCurrent := oldHeadHash

    for {
        g.blocksMutex.Lock()
        new, _ := g.blocks[newHeadHashCurrent]
        old, isNotGenesis := g.blocks[oldHeadHashCurrent]
        g.blocksMutex.Unlock()
        oldHash := old.Hash()
        if !isNotGenesis || bytes.Equal(oldHash[:], new.PrevHash[:]) {
            break
        }
        rewind += 1

        newHeadHashCurrent = new.PrevHashStr()
        oldHeadHashCurrent = old.PrevHashStr()
    }
    return rewind
}

func (g *Gossiper) printBlockchain(headHash string) {
    currentHash := headHash
    chainStr := ""

    g.blocksMutex.RLock()
    for {
        block, isPresent := g.blocks[currentHash]
        if !isPresent {
            break
        }

        chainStr += " " + block.String()
        currentHash = block.PrevHashStr()
    }
    g.blocksMutex.RUnlock()

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
    g.addTxPublishToPool(tp)

    g.broadcastTxPublish(&tp)
}

func (g *Gossiper) addTxPublishToPool(tp model.TxPublish) {
    // Add only if not already there
    g.txsForNextBlockMutex.Lock()
    alreadyPresnet := false
    for _, trx := range g.txsForNextBlock {
        if trx.File.Name == tp.File.Name {
            alreadyPresnet = true
        }
    }

    if !alreadyPresnet {
        var metafileHashCopy []byte = make([]byte, 32)
        copy(metafileHashCopy[:], tp.File.MetafileHash[:])
        tx := model.TxPublish{
            File: model.File{
                Name: tp.File.Name,
                Size: tp.File.Size,
                MetafileHash: metafileHashCopy,
            },
            HopLimit: tp.HopLimit,
        }

        g.txsForNextBlock = append(g.txsForNextBlock, tx)
    }

    g.txsForNextBlockMutex.Unlock()
}

func (g *Gossiper) startMining() {
    time.Sleep(GENESIS_BLOCK_WAIT_TIME * time.Second)

    for {
        g.txsForNextBlockMutex.Lock()
        txsForNextBlockLength := len(g.txsForNextBlock)
        g.txsForNextBlockMutex.Unlock()

        if txsForNextBlockLength > 0 {
            //debug.Debug("START MINING " + time.Now().String())

            block := g.createBlockAndMine()
            if block != nil {
                bp := &model.BlockPublish{
                    Block: *block,
                    HopLimit: 20,
                }

                g.HandlePktBlockPublish(&model.GossipPacket{BlockPublish: bp})
            }
        } else {
            // Wait a bit before checking again
            time.Sleep(1 * time.Second)
        }
    }

}

func (g *Gossiper) createBlockAndMine() *model.Block {
    var nonce [32]byte
    for {
        var prevHash [32]byte = [32]byte{}

        g.forksMutex.Lock()
        longestChainLength := g.forks[g.longestChain]
        g.forksMutex.Unlock()

        if longestChainLength > 0 {
            data, _ := hex.DecodeString(g.longestChain)
            copy(prevHash[:], data[:32])
        }

        // Prepare transactions to insert in the block
        g.txsForNextBlockMutex.Lock()
        // Stop mining if there are no more transaction  to be added to the blockchain
        if len(g.txsForNextBlock) == 0 {
            g.txsForNextBlockMutex.Unlock()
            return nil
        }

        transactions := make([]model.TxPublish, len(g.txsForNextBlock))
        copy(transactions[:], g.txsForNextBlock)
        g.txsForNextBlockMutex.Unlock()

        // Create the block with a random nonce
        rand.Read(nonce[:])
        block := model.Block{
            PrevHash: prevHash,
            Nonce: nonce,
            Transactions: transactions,
        }

        if block.IsValid() {
            fmt.Println("FOUND-BLOCK " + block.HashStr())

            // Remove transactions mined from the txsForNextBlock (assume that there are no new TxPhublish added in the middle of the list)
            g.txsForNextBlockMutex.Lock()
            g.txsForNextBlock = g.txsForNextBlock[len(transactions):len(g.txsForNextBlock)]
            g.txsForNextBlockMutex.Unlock()

            return &block
        }
    }
}
*/
