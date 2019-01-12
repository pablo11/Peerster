package gossip

import (
    "fmt"
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha256"
    "crypto"
    "github.com/pablo11/Peerster/model"
)


// ===== Signing =====
func (g *Gossiper) Sign(data []byte) *model.Signature {
    var opts rsa.PSSOptions
    opts.SaltLength = rsa.PSSSaltLengthAuto

    s := &model.Signature{
        Name: g.Name,
    }

    rng := rand.Reader

    hashed := sha256.Sum256(data)

    bitString, err := rsa.SignPSS(rng, g.PrivateKey, crypto.SHA256, hashed[:], &opts)
    if err != nil {
            fmt.Printf("Error signing: %v\n", err)
            return nil
    }

    fmt.Printf("Signature bitstring: %x\n", bitString)

    s.BitString = make([]byte, len(bitString))
    copy(s.BitString, bitString)

    return s
}

func (g *Gossiper) SignFile(file *model.File) *model.Signature {
    return g.Sign(file.MetafileHash)
}

/*
func (g *Gossiper) SignIdentity(identity *Identity) Signature {
    return g.Sign([]byte(identity.Name))
}

func (g *Gossiper) SignShareTx(shareTx *ShareTx) Signature {
    return g.Sign(shareTx.Hash())
}

func (g *Gossiper) SignPublishShareTx(publishShareTx *PublishShareTx) Signature {
    return g.Sign(publishShareTx.Hash())
}

func (g *Gossiper) SignVotingStatement(votingStatement *VotingStatement) Signature {
    return g.Sign(votingStatement.Hash())
}

func (g *Gossiper) SignVotingReply(votingReply *VotingReply) Signature {
    return g.Sign(votingReply.Hash())
}
*/


// ===== Validation =====
func (b *Blockchain) Verify(sig model.Signature, data []byte) bool {
    var opts rsa.PSSOptions
    opts.SaltLength = rsa.PSSSaltLengthAuto

    hashed := sha256.Sum256(data)

    // TODO: get publicKey from Name
    b.identitiesMutex.Lock()
    identity, isIdentifiable := b.identities[sig.Name]
    b.identitiesMutex.Unlock()
    if !isIdentifiable {
        fmt.Printf("Identity not available in the blockchain\n")
        return false
    }
    pub := identity.PublicKeyObj()

    err := rsa.VerifyPSS(pub, crypto.SHA256, hashed[:], sig.BitString, &opts)
    if err != nil {
        fmt.Printf("Invalid signature: %v\n", err)
        return false
    }

    return true
}


func (b *Blockchain) VerifyFile(sig model.Signature, file *model.File) bool {
    return b.Verify(sig, file.MetafileHash)
}

/*
func (b *Blockchain) VerifyIdentity(sig Signature, identity *Identity) bool {
    return b.Verify(sig, []byte(identity.Name))
}

func (b *Blockchain) VerifyShareTx(sig Signature, shareTx *ShareTx) bool {
    return b.Verify(sig, shareTx.Hash())
}

func (b *Blockchain) VerifyPublishShareTx(sig Signature, publishShareTx *PublishShareTx) bool {
    return b.Verify(sig, publishShareTx.Hash())
}

func (b *Blockchain) VerifyVotingStatement(sig Signature, votingStatement *VotingStatement) bool {
    return b.Verify(sig, votingStatement.Hash())
}

func (b *Blockchain) VerifyVotingReply(sig Signature, votingReply *VotingReply) bool {
    return b.Verify(sig, votingReply.Hash())
}
*/
