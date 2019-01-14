package model

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
)

const MINING_DIFFICULTY int = 6 // Each unit of mining difficulty corresponds to 4 bits

type BlockPublish struct {
	Block    Block
	HopLimit uint32
}

type Block struct {
	PrevHash     [32]byte
	Nonce        [32]byte
	Transactions []Transaction
}

func (b *Block) HashStr() string {
	hash := b.Hash()
	return hex.EncodeToString(hash[:])
}

func (b *Block) PrevHashStr() string {
	return hex.EncodeToString(b.PrevHash[:])
}

func (b *Block) Hash() (out [32]byte) {
	h := sha256.New()
	h.Write(b.PrevHash[:])

	h.Write(b.Nonce[:])

	binary.Write(h, binary.LittleEndian, uint32(len(b.Transactions)))

	for _, t := range b.Transactions {
		th := t.Hash()
		h.Write(th[:])
	}

	copy(out[:], h.Sum(nil))

	return
}

func (b *Block) Mine() {
	var nonce [32]byte
	for {
		rand.Read(nonce[:])
		b.Nonce = nonce
		if b.IsValid() {
			hash := b.Hash()
			fmt.Println("FOUND-BLOCK " + hex.EncodeToString(hash[:]))
			return
		}
	}
}

func (b *Block) IsValid() bool {
	blockHash := b.Hash()
	blockHashStr := hex.EncodeToString(blockHash[:])

	return blockHashStr[0:MINING_DIFFICULTY] == strings.Repeat("0", MINING_DIFFICULTY)
}

func (b *Block) Copy() Block {
	var prevHashCopy [32]byte = [32]byte{}
	copy(prevHashCopy[:], b.PrevHash[:])
	var nonceCopy [32]byte = [32]byte{}
	copy(nonceCopy[:], b.Nonce[:])

	transactionsCopy := make([]Transaction, len(b.Transactions))
	for idx, tx := range b.Transactions {
		transactionsCopy[idx] = tx.Copy()
	}

	return Block{
		PrevHash:     prevHashCopy,
		Nonce:        nonceCopy,
		Transactions: transactionsCopy,
	}
}

func (b *Block) IsGenesis() bool {
	return bytes.Equal(b.PrevHash[:], make([]byte, 32))
}

func (b *Block) String() string {
	transactionsStr := make([]string, len(b.Transactions))
	for i, tx := range b.Transactions {
		transactionsStr[i] = tx.String()
	}

	return b.HashStr() + ":" + b.PrevHashStr() + "\nTransactions: \n" + strings.Join(transactionsStr, "\n")
}
