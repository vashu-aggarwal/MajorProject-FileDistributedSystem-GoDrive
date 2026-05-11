package pipeline

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Encrypt takes a 32-byte AES-256 key and any plaintext byte slice.
// It returns the encrypted bytes in the format: [12-byte nonce] + [ciphertext + GCM auth tag].
// AES-GCM provides both confidentiality and built-in integrity checking.
func Encrypt(key []byte, plaintext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("encryption: key must be exactly 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate a cryptographically secure random nonce.
	// A new nonce is generated for every encryption call to ensure
	// that the same plaintext never produces the same ciphertext.
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal encrypts and appends the GCM authentication tag.
	// We prepend the nonce to the ciphertext so it can be extracted during decryption.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt takes a 32-byte AES-256 key and the ciphertext produced by Encrypt.
// It returns the original plaintext bytes.
// If the ciphertext has been tampered with, decryption will fail with an error.
func Decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("encryption: key must be exactly 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("encryption: ciphertext is too short to be valid")
	}

	// Split the nonce from the actual ciphertext
	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Open decrypts and verifies the GCM authentication tag.
	// If any byte was tampered with, this will return an error.
	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, errors.New("encryption: decryption failed - data may be corrupted or tampered with")
	}

	return plaintext, nil
}
