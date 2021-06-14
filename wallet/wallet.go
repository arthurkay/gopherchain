package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

// Wallet defines the structure
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// ValidateAddress validates the characters used the address name
// making sure they conform to the base58 encoding scheme
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualCheckSum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetCheckSum := Checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Equal(actualCheckSum, targetCheckSum)
}

// Checksum creates a sha256 cryptographic cipher
// from the provided payload
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]
}

// PublicKeyHash creates a ripemd160 byte key
// from the sha256 hash of the public key
func PublicKeyHash(pubkey []byte) []byte {
	pubHash := sha256.Sum256(pubkey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil)
	return publicRipMD
}

// NewJKeyPai returns an ecdsa private key and a
// public byte key
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pub
}

// Address returns a byte of the base58
// Base58 eliminates 0 O 1 l + / from the keyspace
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)
	return address
}

// MakeWallet creates a wallet that holds the private
// public key pairs
func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}
