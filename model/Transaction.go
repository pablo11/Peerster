package model

import (
    "crypto/sha256"
    "encoding/hex"
)

type Transaction struct {
    File *File
    Identity *Identity
    /*
    ShareTx *ShareTx
    VotingRequest *VotingRequest
    VotingReply *VotingReply
    ...
    */
    Signature *Signature
}

func (t *Transaction) Hash() (out [32]byte) {
    var txContentHash [32]byte
    switch {
        case t.File != nil:
            txContentHash = t.File.Hash()

        case t.Identity != nil:
            txContentHash = t.Identity.Hash()

    }

    h := sha256.New()
    h.Write(txContentHash[:])
    if t.Signature != nil {
        h.Write([]byte(t.Signature.Name))
        h.Write(t.Signature.Signature)
    }
    copy(out[:], h.Sum(nil))
    return
}

func (t *Transaction) HashStr() string {
    hash := t.Hash()
    return hex.EncodeToString(hash[:])
}

func (t *Transaction) Copy() Transaction {
    var file *File = nil
    var identity *Identity = nil

    switch {
        case t.File != nil:
            fileCopy := t.File.Copy()
            file = &fileCopy

        case t.Identity != nil:
            identityCopy := t.Identity.Copy()
            identity = &identityCopy
    }

    var signature *Signature = nil
    if t.Signature != nil {
        signatureCopy := t.Signature.Copy()
        signature = &signatureCopy
    }

    return Transaction{
        File: file,
        Identity: identity,
        Signature: signature,
    }
}

func (t *Transaction) String() string {
    switch {
        case t.File != nil:
            return t.File.String()

        case t.Identity != nil:
            return t.Identity.String()
    }
    return ""
}



// ShareTx transaction =========================================================
/*
type ShareTx struct {
    Asset string
    Amount uint64
    To []byte // Public key of the destination account
}

func (st *ShareTx) Hash() (out [32]byte) {
    h := sha256.New()
    binary.Write(h, binary.LittleEndian, st.Amount))
    h.Write([]byte(st.Asset))
    h.Write(i.To)
    copy(out[:], h.Sum(nil))
    return
}
*/
