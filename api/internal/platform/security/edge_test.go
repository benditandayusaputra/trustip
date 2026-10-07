package security_test

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

type edgeVectors struct {
	KeyID  string       `json:"key_id"`
	KeyHex string       `json:"key_hex"`
	Cases  []edgeVector `json:"cases"`
}

type edgeVector struct {
	Name      string `json:"name"`
	Method    string `json:"method"`
	URI       string `json:"uri"`
	Timestamp int64  `json:"timestamp"`
	Nonce     string `json:"nonce"`
	ClientIP  string `json:"client_ip"`
	Body      string `json:"body"`
	Canonical string `json:"canonical"`
	Signature string `json:"signature"`
}

type keyring map[string][]byte

func (k keyring) Lookup(id string) ([]byte, bool) {
	key, ok := k[id]
	return key, ok
}

func loadVectors(t *testing.T) edgeVectors {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "vectors", "edge_signature.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}
	var vectors edgeVectors
	if err := json.Unmarshal(raw, &vectors); err != nil {
		t.Fatalf("decode vectors: %v", err)
	}
	return vectors
}

func (v edgeVector) request(keyID string) security.SignedRequest {
	return security.SignedRequest{
		KeyID:     keyID,
		Method:    v.Method,
		URI:       v.URI,
		Timestamp: strconv.FormatInt(v.Timestamp, 10),
		Nonce:     v.Nonce,
		ClientIP:  v.ClientIP,
		Body:      []byte(v.Body),
		Signature: v.Signature,
	}
}

func newVerifier(t *testing.T, vectors edgeVectors, now time.Time) (*security.EdgeVerifier, *clock.Manual) {
	t.Helper()
	key, err := hex.DecodeString(vectors.KeyHex)
	if err != nil {
		t.Fatalf("decode key: %v", err)
	}
	manual := clock.NewManual(now)
	return security.NewEdgeVerifier(keyring{vectors.KeyID: key}, manual), manual
}

func TestCanonicalString_MatchesSharedVectors(t *testing.T) {
	vectors := loadVectors(t)
	for _, vector := range vectors.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			if got := vector.request(vectors.KeyID).CanonicalString(); got != vector.Canonical {
				t.Fatalf("canonical = %q, want %q", got, vector.Canonical)
			}
		})
	}
}

func TestVerify_SharedVectors_AreAccepted(t *testing.T) {
	vectors := loadVectors(t)
	for _, vector := range vectors.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			verifier, _ := newVerifier(t, vectors, time.Unix(vector.Timestamp, 0))
			if err := verifier.Verify(vector.request(vectors.KeyID)); err != nil {
				t.Fatalf("Verify: %v", err)
			}
		})
	}
}

func TestVerify_Rejections(t *testing.T) {
	vectors := loadVectors(t)
	vector := vectors.Cases[2]
	cases := []struct {
		name   string
		change func(*security.SignedRequest)
		want   error
	}{
		{
			name:   "missing signature",
			change: func(r *security.SignedRequest) { r.Signature = "" },
			want:   security.ErrEdgeMissingHeaders,
		},
		{
			name:   "unknown key id",
			change: func(r *security.SignedRequest) { r.KeyID = "v9" },
			want:   security.ErrEdgeUnknownKey,
		},
		{
			name:   "tampered body",
			change: func(r *security.SignedRequest) { r.Body = []byte(`{"qty":200}`) },
			want:   security.ErrEdgeBadSignature,
		},
		{
			name:   "spoofed client ip",
			change: func(r *security.SignedRequest) { r.ClientIP = "198.51.100.1" },
			want:   security.ErrEdgeBadSignature,
		},
		{
			name:   "malformed nonce",
			change: func(r *security.SignedRequest) { r.Nonce = "short" },
			want:   security.ErrEdgeMissingHeaders,
		},
		{
			name:   "malformed timestamp",
			change: func(r *security.SignedRequest) { r.Timestamp = "soon" },
			want:   security.ErrEdgeMissingHeaders,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verifier, _ := newVerifier(t, vectors, time.Unix(vector.Timestamp, 0))
			request := vector.request(vectors.KeyID)
			tc.change(&request)
			if err := verifier.Verify(request); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestVerify_TimestampWindow(t *testing.T) {
	vectors := loadVectors(t)
	vector := vectors.Cases[0]
	cases := []struct {
		name   string
		offset time.Duration
		want   error
	}{
		{"exactly 60 seconds late is accepted", 60 * time.Second, nil},
		{"61 seconds late is rejected", 61 * time.Second, security.ErrEdgeStale},
		{"exactly 60 seconds early is accepted", -60 * time.Second, nil},
		{"61 seconds early is rejected", -61 * time.Second, security.ErrEdgeStale},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verifier, _ := newVerifier(t, vectors, time.Unix(vector.Timestamp, 0).Add(tc.offset))
			if err := verifier.Verify(vector.request(vectors.KeyID)); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestVerify_ReusedNonceWithin120Seconds_IsRejected(t *testing.T) {
	vectors := loadVectors(t)
	vector := vectors.Cases[0]
	verifier, manual := newVerifier(t, vectors, time.Unix(vector.Timestamp, 0))
	if err := verifier.Verify(vector.request(vectors.KeyID)); err != nil {
		t.Fatalf("first use: %v", err)
	}

	manual.Advance(30 * time.Second)

	if err := verifier.Verify(vector.request(vectors.KeyID)); !errors.Is(err, security.ErrEdgeReplayed) {
		t.Fatalf("err = %v, want replay rejection", err)
	}
}

func TestVerify_BadSignature_DoesNotBurnNonce(t *testing.T) {
	vectors := loadVectors(t)
	vector := vectors.Cases[0]
	verifier, _ := newVerifier(t, vectors, time.Unix(vector.Timestamp, 0))
	forged := vector.request(vectors.KeyID)
	forged.Signature = vector.Signature[:len(vector.Signature)-2] + "00"

	if err := verifier.Verify(forged); !errors.Is(err, security.ErrEdgeBadSignature) {
		t.Fatalf("forged err = %v", err)
	}

	if err := verifier.Verify(vector.request(vectors.KeyID)); err != nil {
		t.Fatalf("genuine request rejected after forged attempt: %v", err)
	}
}
