package edgetest

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const (
	KeyID    = "v1"
	ClientIP = "203.0.113.7"
	keyBytes = 32
	nonceLen = 16
)

type Keys map[string][]byte

func (k Keys) Lookup(id string) ([]byte, bool) {
	key, ok := k[id]
	return key, ok
}

func Key() []byte {
	return bytes.Repeat([]byte{7}, keyBytes)
}

func TestKeys() Keys {
	return Keys{KeyID: Key()}
}

func Sign(t *testing.T, req *http.Request, signedAt time.Time) {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	signed := security.SignedRequest{
		KeyID:     KeyID,
		Method:    req.Method,
		URI:       req.URL.RequestURI(),
		Timestamp: strconv.FormatInt(signedAt.Unix(), 10),
		Nonce:     randomNonce(t),
		ClientIP:  ClientIP,
		Body:      body,
	}
	mac := hmac.New(sha256.New, Key())
	mac.Write([]byte(signed.CanonicalString()))
	req.Header.Set("X-Edge-Key-Id", signed.KeyID)
	req.Header.Set("X-Edge-Timestamp", signed.Timestamp)
	req.Header.Set("X-Edge-Nonce", signed.Nonce)
	req.Header.Set("X-Edge-Client-Ip", signed.ClientIP)
	req.Header.Set("X-Edge-Signature", hex.EncodeToString(mac.Sum(nil)))
}

func randomNonce(t *testing.T) string {
	t.Helper()
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("nonce: %v", err)
	}
	return hex.EncodeToString(nonce)
}
