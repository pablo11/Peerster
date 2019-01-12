package gossip

import (
    "fmt"
    "time"
    "sync"
    "bytes"
    "math/rand"
    "encoding/hex"
    "github.com/pablo11/Peerster/model"
    "github.com/pablo11/Peerster/util/debug"
)

type Blockchain struct {
    gossiper *Gossiper

    blocks map[string]*model.Block
    blocksMutex sync.RWMutex

    // Store the hash of the last block of each fork with the respective blockchain length
    forks map[string]uint64
    forksMutex sync.Mutex

    longestChain string

    txsPool []model.Transaction
    txsPoolMutex sync.Mutex

    // Mapping of filenames in the blockchain
    filenames map[string]*model.File
    filenamesMutex sync.Mutex

    // Mapping of identities in the blockchain [k: Name => v: Identity]
    identities map[string]*model.Identity
    identitiesMutex sync.Mutex

	
	// Mapping of assetName to array of VotationStatement in the blockchain [k: assetName => v: *VotationStatement]
    VoteStatement map[string]*model.VotationStatement
    VoteStatementMutex sync.Mutex
	
	// Mapping of votation_id to array of VotationReplyWrapped in the blockchain [votation_id: string => [holderName: string => votationAnswerWrapper: *VotationAnswerWrapper]]
	VoteAnswers map[string]map[string]*model.VotationAnswerWrapper
	VoteAnswersMutex sync.Mutex

    // Mapping of assets to users holdings: [assetName: string => [holderName: string => amount: uint64]]
    Assets map[string]map[string]uint64
    AssetsMutex sync.Mutex

}

func NewBlockchain() *Blockchain {
    return &Blockchain{
        blocks: make(map[string]*model.Block),
        blocksMutex: sync.RWMutex{},

        forks: make(map[string]uint64),
        forksMutex: sync.Mutex{},

        longestChain: "",

        txsPool: make([]model.Transaction, 0),
        txsPoolMutex: sync.Mutex{},

        filenames: make(map[string]*model.File),
        filenamesMutex: sync.Mutex{},

        identities: make(map[string]*model.Identity),
        identitiesMutex: sync.Mutex{},
		
		VoteStatement: make(map[string]*model.VotationStatement),
        VoteStatementMutex: sync.Mutex{},
		
		VoteAnswers: make(map[string]map[string]*model.VotationAnswerWrapper),
        VoteAnswersMutex: sync.Mutex{},

        Assets: make(map[string]map[string]uint64),
        AssetsMutex: sync.Mutex{},

    }
}

func (b *Blockchain) SetGossiper(g *Gossiper) {
    b.gossiper = g
}

func (b *Blockchain) HandlePktTxPublish(gp *model.GossipPacket) {
    tp := gp.TxPublish

    // Validate signature
    if tp.Transaction.Signature != nil {

        // TODO: validate signature (get the public key of the identity with name tp.Transaction.Signature.Origin and call the IsValid method of Signature)

    }

    // Discard transactions that is already in the pool
    txAlreadyInPool := false
    b.txsPoolMutex.Lock()
    for _, tx := range b.txsPool {
        if tx.HashStr() == tp.Transaction.HashStr() {
            txAlreadyInPool = true
            break
        }
    }
    b.txsPoolMutex.Unlock()

    if !txAlreadyInPool {
        // Validate transaction according to its content
        isValid, errorMsg := b.isValidTx(&tp.Transaction)
        if !isValid {
            fmt.Println("Discarding TxPublish: " + errorMsg)
            return
        }

        // If it's valid and has not yet been seen, store it in the pool of trx to be added in next block
        b.addTxToPool(&tp.Transaction)
    }

    // If HopLimit is > 1 decrement and broadcast
    b.broadcastTxPublishDecrementingHopLimit(tp)
}


func (b *Blockchain) isValidTx(tx *model.Transaction) (isValid bool, errorMsg string) {
    errorMsg = ""
    isValid = true

    switch {
        case tx.File != nil:
            // Check if I have already seen this transactions since the last block mined
            b.filenamesMutex.Lock()
            _, filenameAlreadyCaimed := b.filenames[tx.File.Name]
            if filenameAlreadyCaimed {
                b.filenamesMutex.Unlock()
                errorMsg = "Filename already claimed"
                isValid = false
                return
            }
            b.filenamesMutex.Unlock()


        case tx.Identity != nil:

            // TODO: implement

        case tx.ShareTx != nil:


            // TODO: make sure that the destinatary of the transaction is in the blockchain identities.
            // Also make sure that the transaction is signed by the private key corresponding to the public key claimed by the from name


            // Check if the asset already exists
            b.AssetsMutex.Lock()
            asset, assetExists := b.Assets[tx.ShareTx.Asset]
            b.AssetsMutex.Unlock()

            if assetExists && asset[tx.ShareTx.From] < tx.ShareTx.Amount {
                errorMsg = tx.ShareTx.From + " doesn't have enough " + tx.ShareTx.Asset
                isValid = false
                return
            }

            /*
            if assetExists {

                // TODO: make sure Fom and To are present in the asset shares mapping

                // The asset exists, we need to do a transaction from one holder to another
                if asset[tx.ShareTx.From] < tx.ShareTx.Amount {
                    // The holder doesn't have enough to asset to perform the transaction


                } else {
                    b.assetsMutex.Lock()
                    asset[tx.ShareTx.From] -= tx.ShareTx.Amount
                    asset[tx.ShareTx.To] += tx.ShareTx.Amount
                    b.assetsMutex.Unlock()
                }
            } else {
                // The asset doesn't exist yet, we need to create it and assign all the amount to the initiator of the transaction
                b.assetsMutex.Lock()
                b.assets[tx.ShareTx.Asset] = make(map[string]uint64)
                b.assets[tx.ShareTx.Asset][tx.ShareTx.From] = tx.ShareTx.Amount
                b.assetsMutex.Unlock()
            }
            //*/
			
		case tx.VotationAnswerWrapper != nil:
			//To be rejected, a votation answer wrapped:
			//1. QuestionId does not exist
			//2. Replier doesn't have shares in this asset
			//3. Replier already answer this question
			debug.Debug("Checking votation answer transaction correctness")
			//1.
			questionId := tx.VotationAnswerWrapper.GetVotationId()
			
			b.VoteStatementMutex.Lock()
			_, votationExist := b.VoteStatement[questionId]
			b.VoteStatementMutex.Unlock()
			
			if !votationExist{
				errorMsg = "The votation "+questionId+" does not exists"
				isValid = false
				return
			}
			
			//2.
			b.AssetsMutex.Lock()
			asset, assetExists := b.Assets[tx.VotationAnswerWrapper.AssetName]
			b.AssetsMutex.Unlock()
			
			if !assetExists{
				errorMsg = "The asset "+ tx.VotationAnswerWrapper.AssetName +" doesn't exist"
				isValid = false
				return
			}
			
			share, shareExists := asset[tx.VotationAnswerWrapper.Replier]
			if !shareExists || share <= 0 {
				errorMsg = "The replier "+tx.VotationAnswerWrapper.Replier+" does not have shares in asset "+ tx.VotationAnswerWrapper.AssetName
				isValid = false
				return
			}
			
			//3.
			b.VoteAnswersMutex.Lock()
			voteAnswer, voteAnswerExists := b.VoteAnswers[questionId]
			var replierAlreadyAnswer bool
			if voteAnswerExists {
				_,replierAlreadyAnswer = voteAnswer[tx.VotationAnswerWrapper.Replier]
			}
			b.VoteAnswersMutex.Unlock()
			
			if replierAlreadyAnswer {
				errorMsg = "The replier "+tx.VotationAnswerWrapper.Replier+" already answer this question"
				isValid = false
				return
			}
			debug.Debug("Checking votation answer transaction correctness -> OK")
		
		case tx.VotationStatement != nil:
			//To be rejected, a votation statement:
			//1. is already present with same questionID
			//2. Assetname doesn't exist
			//3. Origin has no share in this asset
			
			debug.Debug("Checking votation statement transaction correctness")
			
			//1.
			questionId := tx.VotationStatement.GetId()
			
			b.VoteStatementMutex.Lock()
			_, votationExist := b.VoteStatement[questionId]
			b.VoteStatementMutex.Unlock()
			
			if votationExist{
				errorMsg = "The votation "+questionId+" already exists"
				isValid = false
				return
			}
			
			//2.
			b.AssetsMutex.Lock()
			asset, assetExists := b.Assets[tx.VotationStatement.AssetName]
			b.AssetsMutex.Unlock()
			
			if !assetExists{
				errorMsg = "The asset "+ tx.VotationStatement.AssetName +"doesn't exist"
				isValid = false
				return
			} 
			
			//3.
			share, shareExists := asset[tx.VotationStatement.Origin]
			if !shareExists || share <= 0 {
				errorMsg = "The origin "+tx.VotationStatement.Origin+"does not have shares in asset "+ tx.VotationStatement.AssetName
				isValid = false
				return
			}
		
			debug.Debug("Checking votation statement transaction correctness -> OK")
		
    }
    return
}

func (b *Blockchain) addTxToPool(t *model.Transaction) {
    // Add only if not already there
    b.txsPoolMutex.Lock()
    for _, tx := range b.txsPool {
        if tx.HashStr() == t.HashStr() {
            b.txsPoolMutex.Unlock()
            return
        }
    }

    // Append a copy of the transaction to the pool of transactions for newxt block
    b.txsPool = append(b.txsPool, t.Copy())
    b.txsPoolMutex.Unlock()
}

func (b *Blockchain) HandlePktBlockPublish(gp *model.GossipPacket) {
    bp := gp.BlockPublish

    // Validate PoW
    if !bp.Block.IsValid() {
        fmt.Println("Discarding BlockPublish since the PoW is invalid")
        return
    }

    blockHash := bp.Block.Hash()
    blockHashStr := hex.EncodeToString(blockHash[:])

    // Check if we already have the block
    b.blocksMutex.Lock()
    _, isPresent := b.blocks[blockHashStr]
    b.blocksMutex.Unlock()
    if isPresent {
        // Forward the blockPublish since someone could not have it
        b.broadcastBlockPublishDecrementingHopLimit(bp)
        //fmt.Println("Discarding BlockPublish since block is already in the blockchain")
        return
    }

    // Validate all transactions in the block before integrating it into the blockchain
    for _, tx := range bp.Block.Transactions {
        isValid, errorMsg := b.isValidTx(&tx)
        if !isValid {
            fmt.Println("Invalid transaction: " + errorMsg)
            return
        }
    }

    //fmt.Printf("🧩 NEW BLOCK %+v\n\n", bp)

    // Store block
    newBlock := bp.Block.Copy()
    b.blocksMutex.Lock()
    b.blocks[blockHashStr] = &newBlock
    b.blocksMutex.Unlock()

    // Check if this block is the continuation of a fork
    isNewFork := true
    b.forksMutex.Lock()
    for lastHash, blockchainLength := range b.forks {
        if lastHash == bp.Block.PrevHashStr() {
            delete(b.forks, lastHash)
            b.forks[blockHashStr] = blockchainLength + 1
            isNewFork = false

            // Check if we modified the longest chain
            if b.longestChain == lastHash {
                b.updateLongestChain(blockHashStr, &bp.Block)
            } else {
                // Check if this fork becomes the longest chain
                if blockchainLength + 1 > b.forks[b.longestChain] {
                    // We are switching to a new longest chain
                    fmt.Printf("FORK-LONGER rewind %d blocks\n", b.computeNbBlocksRewind(blockHashStr, b.longestChain))
                    b.updateLongestChain(blockHashStr, nil)
                } else {
                    fmt.Println("BLOCK ADDED TO A SHORTER FORK")
                }
            }
            break
        }
    }
    b.forksMutex.Unlock()

    if isNewFork {
        // Count length of blockchain
        blockchainLength := b.forkLength(blockHashStr)

        b.forksMutex.Lock()
        b.forks[blockHashStr] = blockchainLength
        b.forksMutex.Unlock()

        if b.longestChain == "" && bp.Block.IsGenesis() {
            // It's the first genesis block
            b.updateLongestChain(blockHashStr, &bp.Block)
        } else {
            fmt.Println("FORK-SHORTER " + hex.EncodeToString(bp.Block.PrevHash[:]))
        }
    }

    // Integrate transactions
    b.integrateValidTxs(&bp.Block)

    // If HopLimit is > 1 decrement and broadcast
    b.broadcastBlockPublishDecrementingHopLimit(bp)
}

func (b *Blockchain) integrateValidTxs(block *model.Block) {
    for _, tx := range block.Transactions {
        switch {
            case tx.File != nil:
                fileCopy := tx.File.Copy()
                b.filenamesMutex.Lock()
                b.filenames[tx.File.Name] = &fileCopy
                b.filenamesMutex.Unlock()

            case tx.Identity != nil:

                // TODO
				
			case tx.VotationAnswerWrapper != nil:
				vawCopy := tx.VotationAnswerWrapper.Copy()
				questionId := vawCopy.GetVotationId()
				b.VoteAnswersMutex.Lock()
				answers, answersExist := b.VoteAnswers[questionId]
				if !answersExist {
					answers = make(map[string]*model.VotationAnswerWrapper)
				}
				answers[vawCopy.Replier] = &vawCopy
				//I could decrypt here if you want.
				b.VoteAnswersMutex.Unlock()
				
			case tx.VotationStatement != nil:
				vsCopy := tx.VotationStatement.Copy()
				questionId := vsCopy.GetId()
				b.VoteStatementMutex.Lock()
				b.VoteStatement[questionId] = &vsCopy
				b.VoteStatementMutex.Unlock()
        }
    }
}

func (b *Blockchain) updateLongestChain(blockHashStr string, block *model.Block) {
    b.longestChain = blockHashStr

    // Remove transactions from pool of transactions for next block to mine
    b.txsPoolMutex.Lock()
    txsForNextBlock := make([]model.Transaction, 0)
    if block != nil {
        for _, f := range b.txsPool {
            toAdd := true
            for _, tx := range block.Transactions {
                if tx.HashStr() == f.HashStr() {
                    toAdd = false
                }
            }
            if toAdd {
                txsForNextBlock = append(txsForNextBlock, f)
            }
        }
    }
    b.txsPool = txsForNextBlock
    b.txsPoolMutex.Unlock()

    b.printBlockchain(blockHashStr)
}

func (b *Blockchain) forkLength(blockHash string) uint64 {
    currentHash := blockHash
    var length uint64 = 0

    b.blocksMutex.Lock()
    for {
        block, isPresent := b.blocks[currentHash]
        if !isPresent {
            break
        }

        length += 1
        currentHash = block.PrevHashStr()
    }

    b.blocksMutex.Unlock()
    return length
}

func (b *Blockchain) computeNbBlocksRewind(newHeadHash, oldHeadHash string) int {
    rewind := 0

    newHeadHashCurrent := newHeadHash
    oldHeadHashCurrent := oldHeadHash

    for {
        b.blocksMutex.Lock()
        new, _ := b.blocks[newHeadHashCurrent]
        old, isNotGenesis := b.blocks[oldHeadHashCurrent]
        b.blocksMutex.Unlock()
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

func (b *Blockchain) printBlockchain(headHash string) {
    currentHash := headHash
    chainStr := ""

    b.blocksMutex.RLock()
    for {
        block, isPresent := b.blocks[currentHash]
        if !isPresent {
            break
        }

        chainStr += " " + block.String()
        currentHash = block.PrevHashStr()
    }
    b.blocksMutex.RUnlock()

    fmt.Println("CHAIN" + chainStr)
}

func (b *Blockchain) broadcastTxPublish(tp *model.TxPublish) {
    gp := model.GossipPacket{TxPublish: tp}
    go b.gossiper.sendGossipPacket(&gp, b.gossiper.peers)
}

func (b *Blockchain) broadcastTxPublishDecrementingHopLimit(tp *model.TxPublish) {
    // If HopLimit is > 1 decrement and broadcast
    if tp.HopLimit > 1 {
        tp.HopLimit -= 1
        b.broadcastTxPublish(tp)
    }
}

func (b *Blockchain) broadcastBlockPublish(bp *model.BlockPublish) {
    gp := model.GossipPacket{BlockPublish: bp}
    go b.gossiper.sendGossipPacket(&gp, b.gossiper.peers)
}

func (b *Blockchain) broadcastBlockPublishDecrementingHopLimit(bp *model.BlockPublish) {
    // If HopLimit is > 1 decrement and broadcast
    if bp.HopLimit > 1 {
        bp.HopLimit -= 1
        b.broadcastBlockPublish(bp)
    }
}

func (b *Blockchain) SendFileTx(file *model.File) {
    tx := model.Transaction{
        File: file,
    }
    b.SendTxPublish(&tx)
}

func (b *Blockchain) SendTxPublish(tx *model.Transaction) {
    tp := model.TxPublish{
        Transaction: *tx,
        HopLimit: 10,
    }

    // Add the transaction to the pool of transactions to be added in the next block
    b.addTxToPool(tx)

    b.broadcastTxPublish(&tp)
}

func (b *Blockchain) StartMining() {
    time.Sleep(GENESIS_BLOCK_WAIT_TIME * time.Second)

    for {
        b.txsPoolMutex.Lock()
        txsForNextBlockLength := len(b.txsPool)
        b.txsPoolMutex.Unlock()

        if txsForNextBlockLength > 0 {
            //debug.Debug("START MINING " + time.Now().String())

            block := b.createBlockAndMine()
            if block != nil {
                bp := &model.BlockPublish{
                    Block: *block,
                    HopLimit: 20,
                }

                b.HandlePktBlockPublish(&model.GossipPacket{BlockPublish: bp})
            }
        } else {
            // Wait a bit before checking again
            time.Sleep(1 * time.Second)
        }
    }

}

func (b *Blockchain) createBlockAndMine() *model.Block {
    var nonce [32]byte
    for {
        var prevHash [32]byte = [32]byte{}

        b.forksMutex.Lock()
        longestChainLength := b.forks[b.longestChain]
        b.forksMutex.Unlock()

        if longestChainLength > 0 {
            data, _ := hex.DecodeString(b.longestChain)
            copy(prevHash[:], data[:32])
        }

        // Prepare transactions to insert in the block
        b.txsPoolMutex.Lock()
        // Stop mining if there are no more transaction  to be added to the blockchain
        if len(b.txsPool) == 0 {
            b.txsPoolMutex.Unlock()
            return nil
        }

        transactions := make([]model.Transaction, len(b.txsPool))
        copy(transactions[:], b.txsPool)
        b.txsPoolMutex.Unlock()

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
            b.txsPoolMutex.Lock()
            b.txsPool = b.txsPool[len(transactions):len(b.txsPool)]
            b.txsPoolMutex.Unlock()

            return &block
        }
    }
}
