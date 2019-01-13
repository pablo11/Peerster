package gossip

import (
    "fmt"
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha256"
    "crypto"
    //"encoding/hex"
    "github.com/pablo11/Peerster/model"
    //"github.com/pablo11/Peerster/util/debug"
)


// ===== Signing =====
func (g *Gossiper) Sign(data [32]byte) *model.Signature {
    var opts rsa.PSSOptions
    opts.SaltLength = rsa.PSSSaltLengthAuto

    s := &model.Signature{
        Name: g.Name,
    }

    rng := rand.Reader

    hashed := sha256.Sum256(data[:])

    bitString, err := rsa.SignPSS(rng, g.PrivateKey, crypto.SHA256, hashed[:], &opts)
    if err != nil {
            fmt.Printf("Error signing: %v\n", err)
            return nil
    }

    s.BitString = make([]byte, len(bitString))
    copy(s.BitString, bitString)

    //s.PrintSignature()
    return s
}

func (g *Gossiper) SignTx(tx *model.Transaction) {
    var sig *model.Signature

    switch {
        case tx.File != nil:
            sig = g.Sign(tx.File.Hash())
        case tx.Identity != nil:
            sig = g.Sign(tx.Identity.Hash())
        case tx.ShareTx != nil:
            sig = g.Sign(tx.ShareTx.Hash())
        case tx.VotationAnswerWrapper != nil:
            sig = g.Sign(tx.VotationAnswerWrapper.Hash())
        case tx.VotationStatement != nil:
            sig = g.Sign(tx.VotationStatement.Hash())
    }

    tx.Signature = sig
    return
}


// ===== Verification =====
func (b *Blockchain) Verify(sig *model.Signature, data [32]byte) bool {
    var opts rsa.PSSOptions
    opts.SaltLength = rsa.PSSSaltLengthAuto

    hashed := sha256.Sum256(data[:])

    // TODO: get publicKey from Name
    b.identitiesMutex.Lock()
    identity, isIdentifiable := b.identities[sig.Name]
    b.identitiesMutex.Unlock()
    if !isIdentifiable {
        fmt.Printf("❓👤 Identity not available in the blockchain\n")
        return false
    }
    pub := identity.PublicKeyObj()

    err := rsa.VerifyPSS(pub, crypto.SHA256, hashed[:], sig.BitString, &opts)
    if err != nil {
        fmt.Printf("❌🔏 Invalid signature: %v\n", err)
        return false
    }

    fmt.Printf("✅🔏 Valid signature\n")
    return true
}

func (b *Blockchain) VerifyTx(tx *model.Transaction) bool {

    sig := tx.Signature
    verified := false
    switch {
        case tx.File != nil:
            verified = b.Verify(sig, tx.File.Hash())
        case tx.Identity != nil:
            verified = true
            //verified = b.Verify(sig, tx.Identity.Hash())
        case tx.ShareTx != nil:
            verified = b.Verify(sig, tx.ShareTx.Hash())
        case tx.VotationAnswerWrapper != nil:
            verified = b.Verify(sig, tx.VotationAnswerWrapper.Hash())
        case tx.VotationStatement != nil:
            /*debug.Debug("Votation tx hash: " + tx.HashStr())
            bytevote := tx.VotationStatement.Hash()
            debug.Debug("Votation question hash: " + hex.EncodeToString(bytevote[:]))
            debug.Debug("Votation signature check: ")
            tx.Signature.PrintSignature()*/

            verified = b.Verify(sig, tx.VotationStatement.Hash())
    }

    return verified
}
