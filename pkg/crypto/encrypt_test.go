package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "hello world"},
		{"chinese text", "你好世界"},
		{"special characters", "!@#$%^&*()_+-=[]{}|;:,.<>?"},
		{"long text", "This is a longer text message for testing encryption and decryption functionality with various characters and symbols."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := EncryptString(tc.plaintext)
			if err != nil {
				t.Errorf("EncryptString failed: %v", err)
			}

			if ciphertext == tc.plaintext {
				t.Error("Ciphertext should not be the same as plaintext")
			}

			decrypted, err := DecryptString(ciphertext)
			if err != nil {
				t.Errorf("DecryptString failed: %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("Decrypted text mismatch: got %q, want %q", decrypted, tc.plaintext)
			}
		})
	}
}

func TestEncryptDecryptBytes(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"empty bytes", []byte{}},
		{"simple bytes", []byte("hello world")},
		{"binary data", []byte{0x00, 0x01, 0x02, 0x03, 0xFF}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := Encrypt(tc.plaintext)
			if err != nil {
				t.Errorf("Encrypt failed: %v", err)
			}

			if len(tc.plaintext) > 0 && string(ciphertext) == string(tc.plaintext) {
				t.Error("Ciphertext should not be the same as plaintext")
			}

			decrypted, err := Decrypt(ciphertext)
			if err != nil {
				t.Errorf("Decrypt failed: %v", err)
			}

			if string(decrypted) != string(tc.plaintext) {
				t.Errorf("Decrypted bytes mismatch")
			}
		})
	}
}