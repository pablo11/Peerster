package model

import (
    "crypto/sha256"
    "encoding/hex"
    //"github.com/pablo11/Peerster/util/debug"
)

type Transaction struct {
    File *File
    Identity *Identity
    ShareTx *ShareTx
	VotationStatement *VotationStatement
    VotationAnswerWrapper *VotationAnswerWrapper
    Signature *Signature
}

func (t *Transaction) Hash() (out [32]byte) {
    var txContentHash [32]byte
    switch {
        case t.File != nil:
            txContentHash = t.File.Hash()

        case t.Identity != nil:
            txContentHash = t.Identity.Hash()

        case t.ShareTx != nil:
            txContentHash = t.ShareTx.Hash()

        case t.VotationStatement != nil:
            txContentHash = t.VotationStatement.Hash()

        case t.VotationAnswerWrapper != nil:
            txContentHash = t.VotationAnswerWrapper.Hash()
    }

    h := sha256.New()
    h.Write(txContentHash[:])
    if t.Signature != nil {
        h.Write([]byte(t.Signature.Name))
        h.Write(t.Signature.BitString)
    }
    copy(out[:], h.Sum(nil))
    return
}

func (t *Transaction) HashStr() string {
    hash := t.Hash()
    return hex.EncodeToString(hash[:])
}

func (t *Transaction) Copy() Transaction {
    txCopy := Transaction{
        File: nil,
        Identity: nil,
        ShareTx: nil,
		VotationStatement: nil,
		VotationAnswerWrapper: nil,
        Signature: nil,
    }

    switch {
        case t.File != nil:
            fileCopy := t.File.Copy()
            txCopy.File = &fileCopy

        case t.Identity != nil:
            identityCopy := t.Identity.Copy()
            txCopy.Identity = &identityCopy

        case t.ShareTx != nil:
            shareTxCopy := t.ShareTx.Copy()
            txCopy.ShareTx = &shareTxCopy

		case t.VotationStatement != nil:
			votationStatementCopy := t.VotationStatement.Copy()
			txCopy.VotationStatement = &votationStatementCopy

		case t.VotationAnswerWrapper != nil:
			votationAnswerWrapperCopy := t.VotationAnswerWrapper.Copy()
			txCopy.VotationAnswerWrapper = &votationAnswerWrapperCopy
    }

    if t.Signature != nil {
        signatureCopy := t.Signature.Copy()
        txCopy.Signature = &signatureCopy
    }

    return txCopy
}

func (t *Transaction) String() string {
    switch {
        case t.File != nil:
            return t.File.String()

        case t.Identity != nil:
            return t.Identity.String()

        case t.ShareTx != nil:
            return t.ShareTx.String()

        case t.VotationStatement != nil:
            return t.VotationStatement.String()

        case t.VotationAnswerWrapper != nil:
            return t.VotationAnswerWrapper.String()
    }
    return ""
}
