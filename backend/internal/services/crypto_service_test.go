package services

import (
	"testing"
)

func TestCryptoService_EncryptDecrypt(t *testing.T) {
	secret := "test-secret-key-123456789012345"
	svc := NewCryptoService(secret)

	originalKey := "sk-proj-test1234567890abcdef"

	encrypted, err := svc.Encrypt(originalKey)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if encrypted == "" || encrypted == originalKey {
		t.Fatalf("Encrypted text should not be empty or equal to plaintext")
	}

	decrypted, err := svc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if decrypted != originalKey {
		t.Fatalf("Expected decrypted '%s', got '%s'", originalKey, decrypted)
	}
}

func TestCryptoService_EmptyHandling(t *testing.T) {
	svc := NewCryptoService("some-secret")

	enc, err := svc.Encrypt("")
	if err != nil || enc != "" {
		t.Fatalf("Encrypting empty string should return empty string, got: %s", enc)
	}

	dec, err := svc.Decrypt("")
	if err != nil || dec != "" {
		t.Fatalf("Decrypting empty string should return empty string, got: %s", dec)
	}
}
