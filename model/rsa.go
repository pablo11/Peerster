package model

import (
    "crypto/rsa"
    "math/rand"
    "crypto"
)


// ========================= Digital Signature =======================
type Signature struct {
    Name string
    Signature []byte
}

// ===== Signing =====
func (g *Gossiper) Sign(data []byte) Signature {
    s := Signature{
        Name: name,
        Signature: make([]byte, 0),
    }


    rng := rand.Reader

    hashed := sha256.Sum256(data)

    signature, err := SignPSS(rng, g.PrivateKey, crypto.SHA256, hashed[:])
    if err != nil {
            fmt.Printf("Error signing: %v\n", err)
            return nil
    }

    fmt.Printf("Signature: %x\n", signature)

    // TODO DEEP COPY
    s.Signature = signature

    return s
}

func (g *Gossiper) SignFile(file *File) {
    return g.Sign(file.MetafileHash)
}

/* Useless?
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
func (b *Blockchain) Verify(sig Signature, data []byte) bool {
    hashed := sha256.Sum256(data)

    // TODO: get publicKey from Name
    b.identitiesMutex.Lock()
    pub, isIdentifiable := b.identities[sig.Name].PublicKey
    b.identitiesMutex.Unlock()
    if !isIdentifiable {
        fmt.Printf("Identity not available in the blockchain\n")
        return false
    }

    err := rsa.VerifyPSS(pub, crypto.SHA256, hashed, sig.Signature)
    if err != nil {
        fmt.Printf("Invalid signature: %v\n", err)
        return false
    }

    return true
}


func (b *Blockchain) VerifyFile(sig Signature, file *File) bool {
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



// ========================= Encryption =======================

func NewPrivateKey() *rsa.PrivateKey {
    rng := rand.Reader

    privateKey, err := rsa.GenerateKey(rng, 2048)
    if err != nil {
        fmt.Printf("Bad private key: %v\n", err)
        return nil
    }

    return privateKey
}

// ===== Encrypt =====
func (g *Gossiper) Encrypt(data []byte, publicKey *rsa.PublicKey) []byte {
    label := []byte("orders") // Optional?

    rng := rand.Reader

    encryptedData, err := EncryptOAEP(sha256.New(), rng, publicKey, data, label)
    if err != nil {
            fmt.Fprintf(os.Stderr, "Error encrypting data: %v\n", err)
            return
    }

    //fmt.Printf("EncryptedData: %x\n", encryptedData)

    return encryptedData
}

// TODO: func Encrypt for each transaction type



// ===== Decrypt =====
func (g *Gossiper) Decrypt(encryptedData []byte) []byte {
    label := []byte("orders") // Optional?

    rng := rand.Reader

    plainData, err := DecryptOAEP(sha256.New(), rng, test2048Key, encryptedData, label)
    if err != nil {
            fmt.Fprintf(os.Stderr, "Error from decryption: %s\n", err)
            return
    }

    fmt.Printf("Plain Data: %s\n", string(plainData))

    return plainData
}

// TODO: func Decrypt for each transaction type




// QUESTION: should we encrypt messages as well?
