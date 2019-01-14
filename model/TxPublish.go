package model

import (
	"encoding/hex"
)

type TxPublish struct {
	Transaction Transaction
	HopLimit    uint32
}

func (t *TxPublish) Hash() (out [32]byte) {
	return t.Transaction.Hash()
}

func (t *TxPublish) HashStr() string {
	hash := t.Hash()
	return hex.EncodeToString(hash[:])
}

func (t *TxPublish) Copy() TxPublish {
	tp := TxPublish{
		Transaction: t.Transaction.Copy(),
		HopLimit:    t.HopLimit,
	}

	return tp
}
