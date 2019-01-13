package gossip

import (
	"strings"
	"strconv"
	"bytes"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
	"github.com/pablo11/Peerster/model"
	"github.com/pablo11/Peerster/util/debug"
)

type Blockchain struct {
	gossiper *Gossiper

	blocks      map[string]*model.Block
	blocksMutex sync.RWMutex

	// Store the hash of the last block of each fork with the respective blockchain length
	forks      map[string]uint64
	forksMutex sync.Mutex

	longestChain string

	txsPool      []model.Transaction
	txsPoolMutex sync.Mutex

	// Mapping of filenames in the blockchain
	filenames      map[string]*model.File
	filenamesMutex sync.Mutex

	// Mapping of identities in the blockchain [k: Name => v: Identity]
	identities      map[string]*model.Identity
	identitiesMutex sync.Mutex
	
	// Mapping of question_id a VotationStatement in the blockchain [k: question_id => v: *VotationStatement]
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
		blocks:      make(map[string]*model.Block),
		blocksMutex: sync.RWMutex{},

		forks:      make(map[string]uint64),
		forksMutex: sync.Mutex{},

		longestChain: "",

		txsPool:      make([]model.Transaction, 0),
		txsPoolMutex: sync.Mutex{},

		filenames:      make(map[string]*model.File),
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

func (b *Blockchain) GetMyAssetsJson() string {
	myAssetsStr := make([]string, 0)
	b.assetsMutex.Lock()
    for assetName, assetOwnership := range b.assets {
        amount, nonzero := assetOwnership[b.gossiper.Name]
		if nonzero {
			var totalSupply uint64 = 0
			for _, holderAmount := range assetOwnership {
				totalSupply += holderAmount
			}

			myAssetsStr = append(myAssetsStr, "\"" + assetName + "\":{\"balance\":" + strconv.Itoa(int(amount)) + ",\"totSupply\":" + strconv.Itoa(int(totalSupply)) + "}")
		}
    }
	b.assetsMutex.Unlock()

	return `{` + strings.Join(myAssetsStr, ",") + `}`
}

func (b *Blockchain) HandlePktTxPublish(gp *model.GossipPacket) {
    tp := gp.TxPublish

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
        b.addTxToPool(tp.Transaction)
    }

    // If HopLimit is > 1 decrement and broadcast
    b.broadcastTxPublishDecrementingHopLimit(tp)
}


func (b *Blockchain) isValidTx(tx *model.Transaction) (isValid bool, errorMsg string) {
    errorMsg = ""
    isValid = true

    if !b.VerifyTx(tx) {
        errorMsg = "Invalid Signature"
        isValid = false
        return
    }

    switch {
        case tx.File != nil:
            // Check if I have already seen this transactions since the last block mined
            b.filenamesMutex.Lock()
            _, filenameAlreadyClaimed := b.filenames[tx.File.Name]
            if filenameAlreadyClaimed {
                b.filenamesMutex.Unlock()
                errorMsg = "Filename already claimed"
                isValid = false
                return
            }
            b.filenamesMutex.Unlock()


        case tx.Identity != nil:
            identityName := tx.Identity.Name
            b.identitiesMutex.Lock()
            _, identityAlreadyClaimed := b.identities[identityName]
            b.identitiesMutex.Unlock()

            if identityAlreadyClaimed {
                errorMsg = "❗️ Cannot add the identity \"" + identityName + "\" because already claimed \n"
                isValid = false
                return
            }

        case tx.ShareTx != nil:
            // Check if the two identities in the share transaction are in the blockchain and that the transaction is validly signed by the sender of the transaction
            isValid, errorMsg = b.isShareTxValidlySigned(tx)

		case tx.VotationAnswerWrapper != nil:
			//To be rejected, a votation answer wrapped:
			//1. QuestionId does not exist
			//2. Replier doesn't have shares in this asset
			//3. Replier already answer this question

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
				errorMsg = "The replier "+tx.VotationAnswerWrapper.Replier+" already answered this question"
				isValid = false
				return
			}

		case tx.VotationStatement != nil:
			//To be rejected, a votation statement:
			//1. is already present with same questionID
			//2. Assetname doesn't exist
			//3. Origin has no share in this asset

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
		

    }
    return
}

func (b *Blockchain) isShareTxValidlySigned(tx *model.Transaction) (isValid bool, errorMsg string) {
    errorMsg = ""
    isValid = true
    isValidSignature := true

    // Make sure that the sender (From) and destinatary (To) identities are in the blockchain
	if tx.ShareTx.From != "" {
		b.identitiesMutex.Lock()
	    fromIdentity, isFromRegistered := b.identities[tx.ShareTx.From]
	    b.identitiesMutex.Unlock()
		if !isFromRegistered {
			errorMsg = "No identities found for the sender of the share transaction"
	        isValid = false
	        return
		}

        isValidSignature = tx.Signature.Name == fromIdentity.Name
	}

	b.identitiesMutex.Lock()
    toIdentity, isToRegistered := b.identities[tx.ShareTx.To]
    b.identitiesMutex.Unlock()
    if !isToRegistered {
        errorMsg = "No identities found for the destinatary of the share transaction"
        isValid = false
        return
    }

    if tx.ShareTx.From == "" {
        isValidSignature = tx.Signature.Name == toIdentity.Name
    }


    if !isValidSignature {
        errorMsg = "Invalid signature of share transaction"
        isValid = false
        return
    }

    return
}

// This function assumes that the ShareTxs's signature contained in the list of transactions are already validated
// To simplify the implementation, the order of thransactions in the block must be valid to be accepted
func (b *Blockchain) validateBlockShareTxs(txs []model.Transaction) bool {
    tmpAssets := make(map[string]map[string]uint64)

    for i, tx := range txs {
        if tx.ShareTx != nil {
            b.AssetsMutex.Lock()
            asset, assetExists := b.Assets[tx.ShareTx.Asset]
            b.AssetsMutex.Unlock()

            if assetExists {
                // Check if sender has sufficeint balance
                _, tmpAssetPresent := tmpAssets[tx.ShareTx.Asset]
                if !tmpAssetPresent {
                    // The asset doesn't exist in our temporary version of the assets, get the user balance from
                    fromBalance, fromBalanceExists := asset[tx.ShareTx.From]
                    if !fromBalanceExists {
                        debug.Debug("Discarding block: invalid asset transaction " + strconv.Itoa(i) + " - " + tx.ShareTx.From)
                        return false
                    }
                    tmpAssets[tx.ShareTx.Asset] = make(map[string]uint64)
                    tmpAssets[tx.ShareTx.Asset][tx.ShareTx.From] = fromBalance

                    toBalance, toBalanceExists := asset[tx.ShareTx.To]
                    if !toBalanceExists {
                        toBalance = 0
                    }
                    tmpAssets[tx.ShareTx.Asset][tx.ShareTx.To] = toBalance
                }

                fromAmount, fromAmountNonzero := tmpAssets[tx.ShareTx.Asset][tx.ShareTx.From]
                if !fromAmountNonzero || fromAmount < tx.ShareTx.Amount {
                    debug.Debug("Discarding block: invalid asset transaction2")
                    return false
                } else {
                    tmpAssets[tx.ShareTx.Asset][tx.ShareTx.From] -= tx.ShareTx.Amount
                    if tmpAssets[tx.ShareTx.Asset][tx.ShareTx.From] == 0 {
                        delete(tmpAssets[tx.ShareTx.Asset], tx.ShareTx.From)
                    }

                    _, toAmountExists := tmpAssets[tx.ShareTx.Asset][tx.ShareTx.To]
                    if toAmountExists {
                        tmpAssets[tx.ShareTx.Asset][tx.ShareTx.To] += tx.ShareTx.Amount
                    } else {
                        tmpAssets[tx.ShareTx.Asset][tx.ShareTx.To] = tx.ShareTx.Amount
                    }
                }
            } else {
                // Create the new asset
                tmpAssets[tx.ShareTx.Asset] = make(map[string]uint64)
                tmpAssets[tx.ShareTx.Asset][tx.ShareTx.To] = tx.ShareTx.Amount
            }
        }
    }
    return true
}

func (b *Blockchain) validateBlockIdentities(txs []model.Transaction) bool {
    tmpIds := make(map[string]bool)

    for _, tx := range txs {
        if tx.Identity != nil {
            _, isThere := tmpIds[tx.Identity.Name]
            if isThere {
                return false
            } else {
                tmpIds[tx.Identity.Name] = true
            }
        }
    }
    return true
}


// This function assumes that the transaction and it's content is already validated (identities existence, valid signature, prevent doublespending)
func (b *Blockchain) applyShareTxs(txs []model.Transaction) {
    for _, tx := range txs {
        if tx.ShareTx != nil {
            b.AssetsMutex.Lock()
            asset, assetExists := b.Assets[tx.ShareTx.Asset]
            b.AssetsMutex.Unlock()
            if assetExists {
                // The asset exists, we need to do a transaction from the sender to the destinatary (is the sender has sufficient balance)
                if asset[tx.ShareTx.From] < tx.ShareTx.Amount {
                    // The holder doesn't have enough to asset to perform the transaction
                    fmt.Println("ShareTx discarded since the sender doesn't have a sufficient balance")
                } else {
                    b.AssetsMutex.Lock()
                    asset[tx.ShareTx.From] -= tx.ShareTx.Amount
                    _, destinataryHashShares := asset[tx.ShareTx.To]
                    if destinataryHashShares {
                        asset[tx.ShareTx.To] += tx.ShareTx.Amount
                    } else {
                        asset[tx.ShareTx.To] = tx.ShareTx.Amount
                    }
                    b.AssetsMutex.Unlock()
                }
            } else {
                // The asset doesn't exist yet, we need to create it and assign all the amount to the initiator of the transaction
                b.AssetsMutex.Lock()
                b.Assets[tx.ShareTx.Asset] = make(map[string]uint64)
                b.Assets[tx.ShareTx.Asset][tx.ShareTx.To] = tx.ShareTx.Amount
                b.AssetsMutex.Unlock()
            }
            //*/

		}
	}
}


func (b *Blockchain) addTxToPool(t model.Transaction) {
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

    // Validate transactions of type ShareTx
    validTxShares := b.validateBlockShareTxs(bp.Block.Transactions)
    if !validTxShares {
        fmt.Println("Invalid asset transactions")
        return
    }

    // Validate transactions of type Identity
    validIdentities := b.validateBlockIdentities(bp.Block.Transactions)
    if !validIdentities {
        fmt.Println("Invalid identity transactions")
        return
    }

    fmt.Printf("🔗 NEW BLOCK \n\n")

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

	// 🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨 TODO: MOVE THAT UP, not always 🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨

    // Integrate transactions
    b.integrateValidTxs(&bp.Block)

	b.printAssetsOwnership()

	b.printVotings()

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
				identityCopy := tx.Identity.Copy()
				b.identitiesMutex.Lock()
				b.identities[tx.Identity.Name] = &identityCopy
				b.identitiesMutex.Unlock()


			case tx.VotationAnswerWrapper != nil:
				vawCopy := tx.VotationAnswerWrapper.Copy()
				questionId := vawCopy.GetVotationId()
				b.VoteAnswersMutex.Lock()
				answers, answersExist := b.VoteAnswers[questionId]
				if !answersExist {
					answers = make(map[string]*model.VotationAnswerWrapper)
					answers[vawCopy.Replier] = &vawCopy
					b.VoteAnswers[questionId] = answers
				} else {
					answers[vawCopy.Replier] = &vawCopy
				}


				b.VoteAnswersMutex.Unlock()

			case tx.VotationStatement != nil:
				vsCopy := tx.VotationStatement.Copy()
				questionId := vsCopy.GetId()
				b.VoteStatementMutex.Lock()
				b.VoteStatement[questionId] = &vsCopy
				b.VoteStatementMutex.Unlock()
        }
    }

    b.applyShareTxs(block.Transactions)
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

		chainStr += "\n" + block.String()
		currentHash = block.PrevHashStr()
	}
	b.blocksMutex.RUnlock()

	fmt.Println("⛓ CHAIN" + chainStr + "\n")
}

func (b *Blockchain) printAssetsOwnership() {
	toPrint := ""
	b.AssetsMutex.Lock()
	for assetName, asset := range b.Assets {
		toPrint += assetName + ":"
		for ownerName, amount := range asset {
			toPrint += "\n---" + ownerName + ": " + strconv.Itoa(int(amount))
		}
		toPrint += "\n"
	}
	b.AssetsMutex.Unlock()

	fmt.Println("ASSET OWNERSHIP:\n" + toPrint)
}

func (b *Blockchain) printVotings() {
	toPrint := ""
	question_prints := make(map[string]string)
	b.VoteStatementMutex.Lock()
	for question_id, vs := range b.VoteStatement{
		question_prints[question_id] = question_id + ":  " + vs.Question+ " from " +vs.Origin+" on asset "+ vs.AssetName
	}
	b.VoteStatementMutex.Unlock()

	question_keys_copy := make(map[string]string)
	b.gossiper.QuestionKeyMutex.Lock()
	for question_id, _ := range question_prints{
		key, keyExists := b.gossiper.QuestionKey[question_id]
		if keyExists {
			question_keys_copy[question_id] = key
		}
	}
	b.gossiper.QuestionKeyMutex.Unlock()

	b.VoteAnswersMutex.Lock()
	for question_id, question_print := range question_prints {
		toPrint += question_print
		for voteReplier, vote := range b.VoteAnswers[question_id]{

			//TODO: I HAVE TO LOCK HERE
			key, keyExists := question_keys_copy[question_id]
			var bool_str string
			if keyExists{
				key_byte, err := hex.DecodeString(key)

				ans_decrypted, err := vote.Decrypt(key_byte)
				if err != nil{
					fmt.Println("failled to decrypt answer")
					return
				}

				bool_str = strconv.FormatBool(ans_decrypted.Answer)
			}
			toPrint += "\n---" + voteReplier +" "+ bool_str
		}
		toPrint += "\n"
	}

	b.VoteAnswersMutex.Unlock()

	fmt.Println("VOTATIONS:\n" + toPrint)
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

func (b *Blockchain) SendShareTx(asset, to string, amount uint64) {
	shareTx := model.ShareTx{
		Asset: asset,
		Amount: amount,
		From: "",
		To: to,
	}

	if b.gossiper.Name != to {
		shareTx.From = b.gossiper.Name
	}

	shareTx.GenerateNonce()

	tx := model.Transaction{
		ShareTx: &shareTx,
	}
	b.SendTxPublish(&tx)
}

func (b *Blockchain) SendTxPublish(tx *model.Transaction) {
    if tx.Signature == nil {
        b.gossiper.SignTx(tx)
    }

    tp := model.TxPublish{
		Transaction: *tx,
		HopLimit:    10,
	}

	// Add the transaction to the pool of transactions to be added in the next block

    b.addTxToPool(*tx)

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
					Block:    *block,
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
			PrevHash:     prevHash,
			Nonce:        nonce,
			Transactions: transactions,
		}

		if block.IsValid() {
			fmt.Println("FOUND-BLOCK " + block.HashStr() + "\n")

			// Remove transactions mined from the txsForNextBlock (assume that there are no new TxPhublish added in the middle of the list)
			b.txsPoolMutex.Lock()
			b.txsPool = b.txsPool[len(transactions):len(b.txsPool)]
			b.txsPoolMutex.Unlock()

			return &block
		}
	}
}

func (b *Blockchain) SendIdentityTx(identityName string) {
    newIdentity := &model.Identity{
		Name: identityName,
	}

	var privateKey *rsa.PrivateKey


	// Identity for THIS peer
	if identityName == b.gossiper.Name {
		privateKey = b.gossiper.PrivateKey
		newIdentity.SetPublicKey(&privateKey.PublicKey)
	} else {
        // Identity for ANOTHER peer
		privateKey = NewPrivateKey()
		newIdentity.SetPublicKey(&privateKey.PublicKey)
	}


	tx := &model.Transaction{
		Identity: newIdentity,
	}

    b.gossiper.SignTx(tx)

    isValid, err := b.isValidTx(tx)
    if isValid {
        alreadyPending := b.isAlreadyPendingIdentity(newIdentity)

        if !alreadyPending {
            fmt.Printf("👤 New Identity - Name: %v \n", identityName)
            fmt.Printf("PrivateKey: %v \n", model.PrivateKeyString(privateKey))
            fmt.Printf("PublicKey:  %v\n\n", model.PublicKeyString(newIdentity.PublicKeyObj()))
            //fmt.Printf("Hash: %v \n", newIdentity.HashStr())
            b.SendTxPublish(tx)
        } else {
            fmt.Println("❗️ Cannot add the identity \"" + identityName + "\" because already in the pending pool\n")
        }
    } else {
        fmt.Println(err)
    }
}


func (b *Blockchain) isAlreadyPendingIdentity(newIdentity *model.Identity) bool {
    for _, tx := range b.txsPool {
        if tx.Identity != nil {
            if tx.Identity.Name == newIdentity.Name {
                return true
            }
        }
    }
    return false
}
