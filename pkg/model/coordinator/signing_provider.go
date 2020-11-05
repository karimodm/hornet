package coordinator

import (
	"crypto/ed25519"

	iotago "github.com/iotaledger/iota.go"

	"github.com/gohornet/hornet/pkg/keymanager"
	"github.com/gohornet/hornet/pkg/model/milestone"
)

// MilestoneSignerProvider provides milestone signers.
type MilestoneSignerProvider interface {
	// MilestoneIndexSigner returns a new signer for the milestone index.
	MilestoneIndexSigner(milestone.Index) MilestoneIndexSigner
	// PublicKeysCount returns the amount of public keys in a milestone.
	PublicKeysCount() int
}

// MilestoneIndexSigner is a signer for a particular milestone.
type MilestoneIndexSigner interface {
	// PublicKeys returns a slice of the used public keys.
	PublicKeys() []iotago.MilestonePublicKey
	// PublicKeysSet returns a map of the used public keys.
	PublicKeysSet() iotago.MilestonePublicKeySet
	// SigningFunc returns a function to sign the particular milestone.
	SigningFunc() iotago.MilestoneSigningFunc
}

// InMemoryEd25519MilestoneSignerProvider provides InMemoryEd25519MilestoneIndexSigner.
type InMemoryEd25519MilestoneSignerProvider struct {
	privateKeys     []ed25519.PrivateKey
	keyManger       *keymanager.KeyManager
	publicKeysCount int
}

// NewInMemoryEd25519MilestoneSignerProvider create a new InMemoryEd25519MilestoneSignerProvider.
func NewInMemoryEd25519MilestoneSignerProvider(privateKeys []ed25519.PrivateKey, keyManager *keymanager.KeyManager, publicKeysCount int) *InMemoryEd25519MilestoneSignerProvider {

	return &InMemoryEd25519MilestoneSignerProvider{
		privateKeys:     privateKeys,
		keyManger:       keyManager,
		publicKeysCount: publicKeysCount,
	}
}

// MilestoneIndexSigner returns a new signer for the milestone index.
func (p *InMemoryEd25519MilestoneSignerProvider) MilestoneIndexSigner(index milestone.Index) MilestoneIndexSigner {

	pubKeySet := p.keyManger.GetPublicKeysSetForMilestoneIndex(index)

	keyPairs := p.keyManger.GetKeyPairsForMilestoneIndex(index, p.privateKeys, p.PublicKeysCount())
	pubKeys := []iotago.MilestonePublicKey{}
	for pubKey := range keyPairs {
		pubKeys = append(pubKeys, pubKey)
	}

	milestoneSignFunc := iotago.InMemoryEd25519MilestoneSigner(keyPairs)

	return &InMemoryEd25519MilestoneIndexSigner{
		pubKeys:     pubKeys,
		pubKeySet:   pubKeySet,
		signingFunc: milestoneSignFunc,
	}
}

// PublicKeysCount returns the amount of public keys in a milestone.
func (p *InMemoryEd25519MilestoneSignerProvider) PublicKeysCount() int {
	return p.publicKeysCount
}

// InMemoryEd25519MilestoneIndexSigner is an in memory signer for a particular milestone.
type InMemoryEd25519MilestoneIndexSigner struct {
	pubKeys     []iotago.MilestonePublicKey
	pubKeySet   iotago.MilestonePublicKeySet
	signingFunc iotago.MilestoneSigningFunc
}

// PublicKeys returns a slice of the used public keys.
func (s *InMemoryEd25519MilestoneIndexSigner) PublicKeys() []iotago.MilestonePublicKey {
	return s.pubKeys
}

// PublicKeysSet returns a map of the used public keys.
func (s *InMemoryEd25519MilestoneIndexSigner) PublicKeysSet() iotago.MilestonePublicKeySet {
	return s.pubKeySet
}

// SigningFunc returns a function to sign the particular milestone.
func (s *InMemoryEd25519MilestoneIndexSigner) SigningFunc() iotago.MilestoneSigningFunc {
	return s.signingFunc
}

// RemoteEd25519MilestoneSignerProvider provides RemoteEd25519MilestoneIndexSigner.
type RemoteEd25519MilestoneSignerProvider struct {
	remoteEndpoint  string
	keyManger       *keymanager.KeyManager
	publicKeysCount int
}

// NewRemoteEd25519MilestoneSignerProvider create a new RemoteEd25519MilestoneSignerProvider.
func NewRemoteEd25519MilestoneSignerProvider(remoteEndpoint string, keyManager *keymanager.KeyManager, publicKeysCount int) *RemoteEd25519MilestoneSignerProvider {
	return &RemoteEd25519MilestoneSignerProvider{
		remoteEndpoint:  remoteEndpoint,
		keyManger:       keyManager,
		publicKeysCount: publicKeysCount,
	}
}

// MilestoneIndexSigner returns a new signer for the milestone index.
func (p *RemoteEd25519MilestoneSignerProvider) MilestoneIndexSigner(index milestone.Index) MilestoneIndexSigner {

	pubKeySet := p.keyManger.GetPublicKeysSetForMilestoneIndex(index)

	pubKeys := []iotago.MilestonePublicKey{}
	for pubKey := range pubKeySet {
		pubKeys = append(pubKeys, pubKey)
	}

	milestoneSignFunc := iotago.RemoteEd25519MilestoneSigner(p.remoteEndpoint, pubKeys)

	return &RemoteEd25519MilestoneIndexSigner{
		pubKeys:     pubKeys,
		pubKeySet:   pubKeySet,
		signingFunc: milestoneSignFunc,
	}
}

// PublicKeysCount returns the amount of public keys in a milestone.
func (p *RemoteEd25519MilestoneSignerProvider) PublicKeysCount() int {
	return p.publicKeysCount
}

// RemoteEd25519MilestoneIndexSigner is an in memory signer for a particular milestone.
type RemoteEd25519MilestoneIndexSigner struct {
	pubKeys     []iotago.MilestonePublicKey
	pubKeySet   iotago.MilestonePublicKeySet
	signingFunc iotago.MilestoneSigningFunc
}

// PublicKeys returns a slice of the used public keys.
func (s *RemoteEd25519MilestoneIndexSigner) PublicKeys() []iotago.MilestonePublicKey {
	return s.pubKeys
}

// PublicKeysSet returns a map of the used public keys.
func (s *RemoteEd25519MilestoneIndexSigner) PublicKeysSet() iotago.MilestonePublicKeySet {
	return s.pubKeySet
}

// SigningFunc returns a function to sign the particular milestone.
func (s *RemoteEd25519MilestoneIndexSigner) SigningFunc() iotago.MilestoneSigningFunc {
	return s.signingFunc
}
