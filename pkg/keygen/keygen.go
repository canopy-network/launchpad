package keygen

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"

	"github.com/canopy-network/canopy/lib/crypto"
	"golang.org/x/crypto/argon2"
)

// KeyPair represents a blockchain key pair with private key, public key, and address
type KeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
	Address    string `json:"address"`
}

// EncryptedKeyPair represents an encrypted form of a private key
type EncryptedKeyPair struct {
	PublicKey            string `json:"publicKey"`
	EncryptedPrivateKey  string `json:"encryptedPrivateKey"`
	Salt                 string `json:"salt"`
	Address              string `json:"address"`
}

// GenerateKeyPair creates a new BLS12-381 key pair for blockchain use
func GenerateKeyPair() (*KeyPair, error) {
	blsKey, err := crypto.NewBLS12381PrivateKey()
	if err != nil {
		return nil, err
	}

	blsPub := blsKey.PublicKey()

	return &KeyPair{
		PrivateKey: blsKey.String(),
		PublicKey:  blsPub.String(),
		Address:    blsPub.Address().String(),
	}, nil
}

// GenerateEncryptedKeyPair creates a new BLS12-381 key pair and encrypts the private key with a password
func GenerateEncryptedKeyPair(password string) (*KeyPair, *EncryptedKeyPair, error) {
	blsKey, err := crypto.NewBLS12381PrivateKey()
	if err != nil {
		return nil, nil, err
	}

	blsPub := blsKey.PublicKey()
	address := blsPub.Address().String()

	keyPair := &KeyPair{
		PrivateKey: blsKey.String(),
		PublicKey:  blsPub.String(),
		Address:    address,
	}

	encryptedKeyPair, err := EncryptPrivateKey(blsPub.Bytes(), blsKey.Bytes(), []byte(password), address)
	if err != nil {
		return nil, nil, err
	}

	return keyPair, encryptedKeyPair, nil
}

// EncryptPrivateKey encrypts a private key using AES-GCM with Argon2 key derivation
func EncryptPrivateKey(publicKey, privateKey, password []byte, address string) (*EncryptedKeyPair, error) {
	// generate random 16 bytes salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// derive an AES-GCM encryption key and nonce using the password and salt
	gcm, nonce, err := kdf(password, salt)
	if err != nil {
		return nil, err
	}

	// encrypt the private key with AES-GCM using the derived key and nonce
	return &EncryptedKeyPair{
		PublicKey:           hex.EncodeToString(publicKey),
		EncryptedPrivateKey: hex.EncodeToString(gcm.Seal(nil, nonce, privateKey, nil)),
		Salt:                hex.EncodeToString(salt),
		Address:             address,
	}, nil
}

// DecryptPrivateKey decrypts an encrypted private key using the password
func DecryptPrivateKey(epk *EncryptedKeyPair, password []byte) ([]byte, error) {
	salt, err := hex.DecodeString(epk.Salt)
	if err != nil {
		return nil, err
	}

	encrypted, err := hex.DecodeString(epk.EncryptedPrivateKey)
	if err != nil {
		return nil, err
	}

	gcm, nonce, err := kdf(password, salt)
	if err != nil {
		return nil, err
	}

	plainText, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// kdf derives an AES-GCM encryption key and nonce from a password and salt using Argon2 key derivation
func kdf(password, salt []byte) (gcm cipher.AEAD, nonce []byte, err error) {
	// use Argon2 to derive a 32 byte key from the password and salt
	key := argon2.Key(password, salt, 3, 32*1024, 4, 32)

	// init AES block cipher with the derived key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// init AES-GCM mode with the AES cipher block
	if gcm, err = cipher.NewGCM(block); err != nil {
		return
	}

	// return the gcm and the 12 byte nonce
	return gcm, key[:12], nil
}
