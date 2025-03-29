package service_test

import (
	"strings"
	"testing"

	"github.com/takuoki/google-calendar-sync/api/domain/service"
)

func TestNewAESCrypt_InvalidKeyLength(t *testing.T) {
	// Test with invalid key length
	_, err := service.NewAESCrypt([]byte("shortkey"))
	if err == nil {
		t.Error("expected error for invalid key length, got nil")
	} else if !strings.HasPrefix(err.Error(), "invalid key length") {
		t.Errorf("error message does not match the expected format, got: %s", err.Error())
	}
}

func TestAESCrypt_Success(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes key
	plaintext := "Hello, World!"

	// Create AESCrypt instance
	crypt, err := service.NewAESCrypt(key)
	if err != nil {
		t.Fatalf("failed to create AESCrypt: %v", err)
	}

	// Test Encrypt
	ciphertext, err := crypt.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	// Test Decrypt
	decryptedText, err := crypt.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	// Verify decrypted text matches original plaintext
	if decryptedText != plaintext {
		t.Errorf("decrypted text does not match plaintext: got %s, want %s", decryptedText, plaintext)
	}
}

func TestAESCrypt_InvalidCiphertext(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes key

	// Create AESCrypt instance
	crypt, err := service.NewAESCrypt(key)
	if err != nil {
		t.Fatalf("failed to create AESCrypt: %v", err)
	}

	// Test Decrypt with invalid ciphertext
	_, err = crypt.Decrypt("invalidciphertext")
	if err == nil {
		t.Error("expected error for invalid ciphertext, got nil")
	} else if !strings.HasPrefix(err.Error(), "fail to decode base64 ciphertext") {
		t.Errorf("error message does not match the expected format, got: %s", err.Error())
	}
}

func TestAESCrypt_EmptyCiphertext(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32 bytes key

	// Create AESCrypt instance
	crypt, err := service.NewAESCrypt(key)
	if err != nil {
		t.Fatalf("failed to create AESCrypt: %v", err)
	}

	// Test Decrypt with empty ciphertext
	_, err = crypt.Decrypt("")
	if err == nil {
		t.Error("expected error for empty ciphertext, got nil")
	} else if !strings.HasPrefix(err.Error(), "invalid ciphertext size") {
		t.Errorf("error message does not match the expected format, got: %s", err.Error())
	}
}
