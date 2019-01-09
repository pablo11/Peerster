package model

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/binary"
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
    h.Write([]byte(t.Signature.Origin))
    h.Write(t.Signature.Signature)
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
            var metafileHashCopy []byte = make([]byte, 32)
            copy(metafileHashCopy[:], t.File.MetafileHash[:])
            file = &File{
                Name: t.File.Name,
                Size: t.File.Size,
                MetafileHash: metafileHashCopy,
            }

        case t.Identity != nil:
            var publicKeyCopy []byte = make([]byte, 32)
            copy(publicKeyCopy[:], t.Identity.PublicKey[:])
            identity = &Identity{
                Name: t.Identity.Name,
                PublicKey: publicKeyCopy,
            }
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
    }
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
