package model

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/binary"
)

type Transaction struct {
    File *File
    Identity *Identity
    ShareTx *ShareTx
    /*
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

        case t.ShareTx != nil:
            txContentHash = t.ShareTx.Hash()

    }

    h := sha256.New()
    h.Write(txContentHash[:])
    if t.Signature != nil {
        h.Write([]byte(t.Signature.Origin))
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
    var shareTx *ShareTx = nil

    switch {
        case t.File != nil:
            fileCopy := t.File.Copy()
            file = &fileCopy

        case t.Identity != nil:
            var publicKeyCopy []byte = make([]byte, 32)
            copy(publicKeyCopy[:], t.Identity.PublicKey[:])
            identity = &Identity{
                Name: t.Identity.Name,
                PublicKey: publicKeyCopy,
            }

        case t.ShareTx != nil:
            shareTxCopy := t.ShareTx.Copy()
            shareTx = &shareTxCopy
    }

    var signature *Signature = nil
    if t.Signature != nil {
        var signCopy []byte = make([]byte, 32)
        copy(signCopy[:], t.Signature.Signature[:])
        signature = &Signature{
            Origin: t.Signature.Origin,
            Signature: signCopy,
        }
    }

    return Transaction{
        File: file,
        Identity: identity,
        Signature: signature,
        ShareTx: shareTx,
    }
}

func (t *Transaction) String() string {
    switch {
        case t.File != nil:
            return t.File.String()

        case t.Identity != nil:
            return t.Identity.String()

        case t.ShareTx != nil:
            return t.ShareTx.String()
    }
    return ""
}

// Identity transaction ========================================================

type Identity struct {
    Name string
    PublicKey []byte
}

func (i *Identity) Hash() (out [32]byte) {
    h := sha256.New()
    binary.Write(h, binary.LittleEndian, uint32(len(i.Name)))
    h.Write([]byte(i.Name))
    h.Write(i.PublicKey)
    copy(out[:], h.Sum(nil))
    return
}

func (i *Identity) String() string {
    return "ID=" + i.Name + "(" + hex.EncodeToString(i.PublicKey) + ")"
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
