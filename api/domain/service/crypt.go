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

type Crypt interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

type AESCrypt struct {
	key    []byte
	aesGCM cipher.AEAD
}

func NewAESCrypt(key []byte) (*AESCrypt, error) {
	if len(key) != 32 {
		return nil, errors.New("invalid key length: must be 32 bytes")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("fail to create AES cipher: %w", err)
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("fail to create AES-GCM: %w", err)
	}

	return &AESCrypt{
		key:    key,
		aesGCM: aesGCM,
	}, nil
}

// Encrypt encrypts plaintext using AES-GCM
func (c *AESCrypt) Encrypt(plaintext string) (string, error) {

	nonce := make([]byte, c.aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("fail to generate nonce: %w", err)
	}

	ciphertext := c.aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func (c *AESCrypt) Decrypt(ciphertext string) (string, error) {
	decodedCiphertext, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("fail to decode base64 ciphertext: %w", err)
	}

	nonceSize := c.aesGCM.NonceSize()
	if len(decodedCiphertext) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext size")
	}

	nonce, ciphertextData := decodedCiphertext[:nonceSize], decodedCiphertext[nonceSize:]
	plaintext, err := c.aesGCM.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return "", fmt.Errorf("fail to decrypt ciphertext: %w", err)
	}

	return string(plaintext), nil
}
