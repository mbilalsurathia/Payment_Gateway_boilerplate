package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	// encryptionKey is used for encrypting sensitive data
	// In a real system, this should be securely stored and accessed
	encryptionKey []byte
)

func init() {
	// Load encryption key from environment variable
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		// For development only - use a hardcoded key
		// In production, this should fail if no key is provided
		keyStr = "1234567890abcdef1234567890abcdef" // 32 bytes = 256 bits
	}

	var err error
	encryptionKey, err = hex.DecodeString(keyStr)
	if err != nil {
		// Log error and use a default key for development
		// In production, this should fail
		encryptionKey = []byte("1234567890abcdef1234567890abcdef")
	}
}

// MaskData masks data using base64 encoding (non-encrypted, for logging)
func MaskData(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Encrypt encrypts data using AES-GCM
func Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Never use more than 2^32 random nonces with a given key
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Encrypt and authenticate
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)

	return result, nil
}

// Decrypt decrypts data using AES-GCM
func Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 12 {
		return nil, errors.New("ciphertext too short")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Extract nonce from ciphertext
	nonce := ciphertext[:12]
	actualCiphertext := ciphertext[12:]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Decrypt and verify
	plaintext, err := aesgcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns a base64-encoded result
func EncryptString(plaintext string) (string, error) {
	encrypted, err := Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptString decrypts a base64-encoded string
func DecryptString(encryptedBase64 string) (string, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return "", err
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
