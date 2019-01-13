package gossip

import (
    "fmt"
    "os"
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha256"
    "github.com/pablo11/Peerster/model"
)

// ========================= Encryption =======================

func NewPrivateKey() *rsa.PrivateKey {
    rng := rand.Reader

    privateKey, err := rsa.GenerateKey(rng, 2048)
    if err != nil {
        fmt.Printf("❌ Bad private key: %v\n", err)
        return nil
    }

    return privateKey
}

// ===== Encrypt =====
func (g *Gossiper) Encrypt(data []byte, publicKey *rsa.PublicKey) []byte {
    label := []byte("orders") // Optional?

    rng := rand.Reader

    encryptedData, err := rsa.EncryptOAEP(sha256.New(), rng, publicKey, data, label)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌🔒 Error encrypting data: %v\n", err)
        return nil
    }

    fmt.Printf("🔒EncryptedData: %x\n", encryptedData)

    return encryptedData
}


func (g *Gossiper) NewEncryptedPrivateMessage(origin, text, dest string) *model.PrivateMessage {
    g.Blockchain.identitiesMutex.Lock()
    toIdentity, isIdentifiable := g.Blockchain.identities[dest]
    g.Blockchain.identitiesMutex.Unlock()
    if !isIdentifiable {
        fmt.Printf("❓👤 Identity not available in the blockchain\n")
        return nil
    }

    cypherBytes := g.Encrypt([]byte(text), toIdentity.PublicKeyObj())
    cypherText := string(cypherBytes[:])

    return &model.PrivateMessage{
        Origin: origin,
        ID: 0,
        Text: cypherText,
        Destination: dest,
        HopLimit: 10,
        IsEncrypted: true,
    }
}

/*
func (g *Gossiper) NewRSAPrivateMessage(origin, text, dest string, isEncrypted, isSigned bool) *model.PrivateMessage {
    g.Blockchain.identitiesMutex.Lock()
    toIdentity, isToIdentifiable := g.Blockchain.identities[dest]
    fromIdentity, isFromIdentifiable := g.Blockchain.identities[origin]
    g.Blockchain.identitiesMutex.Unlock()
    if !isToIdentifiable  {
        fmt.Printf("❓👤 Destination Identity not available in the blockchain\n")
        return nil
    }

    if !isFromIdentifiable  {
        fmt.Printf("❓👤 Origin Identity not available in the blockchain\n")
        return nil
    }

    cypherBytes := g.Encrypt([]byte(text), toIdentity.PublicKeyObj())
    cypherText := string(cypherBytes[:])

    Sign(data [32]byte)

    return &model.PrivateMessage{
        Origin: origin,
        ID: 0,
        Text: cypherText,
        Destination: dest,
        HopLimit: 10,
        IsEncrypted: isEncrypted,
        IsSigned: isSigned,
    }
}*/


// ===== Decrypt =====
func (g *Gossiper) Decrypt(encryptedData []byte) []byte {
    label := []byte("orders") // Optional?

    rng := rand.Reader

    plainBytes, err := rsa.DecryptOAEP(sha256.New(), rng, g.PrivateKey, encryptedData, label)
    if err != nil {
        fmt.Fprintf(os.Stderr, "❌🔓 Error from decryption: %s\n", err)
        return nil
    }

    fmt.Printf("🔓 Plain Bytes: %s\n", string(plainBytes))

    return plainBytes
}

func (g *Gossiper) DecryptPrivateMessage(pm *model.PrivateMessage) {
    cypherBytes := []byte(pm.Text)

    plainBytes := g.Decrypt(cypherBytes)
    pm.Text = string(plainBytes)
}
