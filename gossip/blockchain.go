package gossip

import (
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

	// Mapping of assets to users holdings: [assetName: string => [holderName: string => amount: uint64]]
	assets      map[string]map[string]uint64
	assetsMutex sync.Mutex
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

		identities:      make(map[string]*model.Identity),
		identitiesMutex: sync.Mutex{},

		assets:      make(map[string]map[string]uint64),
		assetsMutex: sync.Mutex{},
	}
}

func (b *Blockchain) SetGossiper(g *Gossiper) {
	b.gossiper = g
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
            // Check if the two identities in the share transaction are in the blockchain and that the transaction is validly signed by the sender of the transaction
            isValid, errorMsg = b.isShareTxValidlySigned(tx.ShareTx)
    }
    return
}

func (b *Blockchain) isShareTxValidlySigned(st *model.ShareTx) (isValid bool, errorMsg string) {
    errorMsg = ""
    isValid = true

    // Make sure that the sender (From) and destinatary (To) identities are in the blockchain
    b.identitiesMutex.Lock()
    _ /*fromIdentity*/, isFromRegistered := b.identities[st.From]
    _ /*toIdentity*/, isToRegistered := b.identities[st.To]
    b.identitiesMutex.Unlock()
    if !isFromRegistered || !isToRegistered {
        errorMsg = "No identities found for the share transaction"
        isValid = false
        return
    }

    // Validate signature with the identity of the sender (From)

    isValidSignature := true // TODO: need to check the signature against the sender identity

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

    for _, tx := range txs {
        if tx.ShareTx != nil {
            b.assetsMutex.Lock()
            asset, assetExists := b.assets[tx.ShareTx.Asset]
            b.assetsMutex.Unlock()

            if assetExists {
                // Check if sender has sufficeint balance
                _, tmpAssetPresent := tmpAssets[tx.ShareTx.Asset]
                if !tmpAssetPresent {
                    // The asset doesn't exist in our temporary version of the assets, get the user balance from
                    fromBalance, fromBalanceExists := asset[tx.ShareTx.From]
                    if !fromBalanceExists {
                        debug.Debug("Discarding block: invalid asset transaction")
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
                    debug.Debug("Discarding block: invalid asset transaction")
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

// This function assumes that the transaction and it's content is already validated (identities existence, valid signature, prevent doublespending)
func (b *Blockchain) applyShareTxs(txs []model.Transaction) {
    for _, tx := range txs {
        if tx.ShareTx != nil {
            b.assetsMutex.Lock()
            asset, assetExists := b.assets[tx.ShareTx.Asset]
            b.assetsMutex.Unlock()
            if assetExists {
                // The asset exists, we need to do a transaction from the sender to the destinatary (is the sender has sufficient balance)
                if asset[tx.ShareTx.From] < tx.ShareTx.Amount {
                    // The holder doesn't have enough to asset to perform the transaction
                    fmt.Println("ShareTx discarded since the sender doesn't have a sufficient balance")
                } else {
                    b.assetsMutex.Lock()
                    asset[tx.ShareTx.From] -= tx.ShareTx.Amount
                    _, destinataryHashShares := asset[tx.ShareTx.To]
                    if destinataryHashShares {
                        asset[tx.ShareTx.To] += tx.ShareTx.Amount
                    } else {
                        asset[tx.ShareTx.To] = tx.ShareTx.Amount
                    }
                    b.assetsMutex.Unlock()
                }
            } else {
                // The asset doesn't exist yet, we need to create it and assign all the amount to the initiator of the transaction
                b.assetsMutex.Lock()
                b.assets[tx.ShareTx.Asset] = make(map[string]uint64)
                b.assets[tx.ShareTx.Asset][tx.ShareTx.From] = tx.ShareTx.Amount
                b.assetsMutex.Unlock()
            }
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

	// 🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨 TODO: MOVE THAT UP, not always 🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨

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
				identityCopy := tx.Identity.Copy()
				b.identitiesMutex.Lock()
				b.identities[tx.Identity.Name] = &identityCopy
				b.identitiesMutex.Unlock()
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
    b.identitiesMutex.Lock()
    _, isThere := b.identities[identityName]
    b.identitiesMutex.Unlock()
    if isThere {
        fmt.Printf("❗️ Cannot add the identity \"%v\" because already claimed \n\n", identityName)
        return
    }


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

	fmt.Printf("👤 New Identity - Name: %v \n", identityName)
	//fmt.Printf("PrivateKey: Private Exponent=%v\n Prime factors=%v \n", privateKey.D, privateKey.Primes)
	//fmt.Printf("PublicKey: Modulus=%v\n Public Exponent=%v \n", newIdentity.PublicKey.N, newIdentity.PublicKey.E)

    // the first 25 chars are always the same
    fmt.Printf("PrivateKey: %v \n", model.PrivateKeyString(privateKey))

    // the first 19 chars are always the same
    fmt.Printf("PublicKey:  %v\n\n", model.PublicKeyString(newIdentity.PublicKeyObj()))


	tx := model.Transaction{
		Identity: newIdentity,
	}
	b.SendTxPublish(&tx)
}
