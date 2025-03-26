package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type Crypt struct {
	key []byte
}

func NewCrypt(key []byte) (*Crypt, error) {
	if len(key) != 32 {
		return nil, errors.New("invalid key length: must be 32 bytes")
	}

	return &Crypt{
		key: key,
	}, nil
}

// Encrypt encrypts plaintext using AES-GCM
func (c *Crypt) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher([]byte(c.key))
	if err != nil {
		return "", fmt.Errorf("fail to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("fail to create AES-GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("fail to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (c *Crypt) Decrypt(ciphertext string) (string, error) {
	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("fail to decode base64 ciphertext: %w", err)
	}

	block, err := aes.NewCipher([]byte(c.key))
	if err != nil {
		return "", fmt.Errorf("fail to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("fail to create AES-GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(decodedCiphertext) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext size")
	}

	nonce, ciphertextData := decodedCiphertext[:nonceSize], decodedCiphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return "", fmt.Errorf("fail to decrypt ciphertext: %w", err)
	}

	return string(plaintext), nil
}
