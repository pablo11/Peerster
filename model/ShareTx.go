package model

import (
    "strconv"
    "crypto/sha256"
    "encoding/hex"
    "encoding/binary"
)

type ShareTx struct {
    Asset string
    Amount uint64
    From string // The name of the sender of the transaction
    To string // The name of the destinatary taken from the identities in the blockchain
}

func (st *ShareTx) Hash() (out [32]byte) {
    h := sha256.New()
    binary.Write(h, binary.LittleEndian, st.Amount)
    h.Write([]byte(st.Asset))
    h.Write([]byte(st.From))
    h.Write([]byte(st.To))
    copy(out[:], h.Sum(nil))
    return
}

func (st *ShareTx) HashStr() string {
    hash := st.Hash()
    return hex.EncodeToString(hash[:])
}

func (st *ShareTx) String() string {
    return "SHARE_TX=" + st.From + "->" + st.To + "(" + strconv.Itoa(int(st.Amount)) + " " + st.Asset +")"
}

func (st *ShareTx) Copy() ShareTx {
    return ShareTx{
        Asset: st.Asset,
        Amount: st.Amount,
        From: st.From,
        To: st.To,
    }
}
