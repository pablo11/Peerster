package gossip

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
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

	s.BitString = make([]byte, len(bitString))
	copy(s.BitString, bitString)

	return s
}

func (g *Gossiper) SignTx(tx *model.Transaction) {
	var sig *model.Signature

	switch {
	case tx.File != nil:
		data := tx.File.Hash()
		sig = g.Sign(data[:])
	case tx.Identity != nil:
		data := tx.Identity.Hash()
		sig = g.Sign(data[:])
	case tx.ShareTx != nil:
		data := tx.ShareTx.Hash()
		sig = g.Sign(data[:])
	case tx.VotationAnswerWrapper != nil:
		data := tx.VotationAnswerWrapper.Hash()
		sig = g.Sign(data[:])
	case tx.VotationStatement != nil:
		data := tx.VotationStatement.Hash()
		sig = g.Sign(data[:])
	}

	tx.Signature = sig
	return
}

func (g *Gossiper) SignPrivateMessage(pm *model.PrivateMessage) {
	cyptherBytes := pm.IntegrityHash()
	sig := g.Sign(cyptherBytes[:])

	sigCopy := sig.Copy()
	pm.Signature = &sigCopy
	return
}

// ===== Verification =====
func (b *Blockchain) Verify(sig *model.Signature, data []byte) bool {
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto

	hashed := sha256.Sum256(data)

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
		data := tx.File.Hash()
		verified = b.Verify(sig, data[:])
	case tx.Identity != nil:
		verified = true
	case tx.ShareTx != nil:
		data := tx.ShareTx.Hash()
		verified = b.Verify(sig, data[:])
	case tx.VotationAnswerWrapper != nil:
		data := tx.VotationAnswerWrapper.Hash()
		verified = b.Verify(sig, data[:])
	case tx.VotationStatement != nil:
		data := tx.VotationStatement.Hash()
		verified = b.Verify(sig, data[:])
	}

	return verified
}

func (b *Blockchain) VerifyPrivateMessage(pm *model.PrivateMessage) bool {
	cyptherBytes := pm.IntegrityHash()
	correctSignature := b.Verify(pm.Signature, cyptherBytes[:])
	correctSigner := pm.Origin == pm.Signature.Name
	return correctSignature && correctSigner
}
