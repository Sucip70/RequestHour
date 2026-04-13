package gamecrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// AudioPayload is encrypted into the audioToken returned by POST /game/question.
type AudioPayload struct {
	Link    string `json:"link"`
	Session string `json:"session"`
	SongID  int    `json:"songId"`
}

var ErrDecrypt = errors.New("decrypt audio token")

// Key derives a 32-byte AES-256 key from a secret string.
func Key(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// EncryptPayload marshals p as JSON and returns base64(nonce|ciphertext|tag).
func EncryptPayload(key []byte, p AudioPayload) (string, error) {
	plain, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plain, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptPayload reverses EncryptPayload.
func DecryptPayload(key []byte, tokenB64 string) (AudioPayload, error) {
	var zero AudioPayload
	raw, err := base64.StdEncoding.DecodeString(tokenB64)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrDecrypt, err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return zero, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return zero, err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return zero, ErrDecrypt
	}
	nonce, ct := raw[:ns], raw[ns:]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrDecrypt, err)
	}
	var p AudioPayload
	if err := json.Unmarshal(plain, &p); err != nil {
		return zero, fmt.Errorf("%w: %w", ErrDecrypt, err)
	}
	return p, nil
}
