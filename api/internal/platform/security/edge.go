package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	canonicalVersion = "v1"
	timestampWindow  = 60 * time.Second
	nonceLifetime    = 120 * time.Second
)

var (
	ErrEdgeMissingHeaders = errors.New("edge signature headers are missing or malformed")
	ErrEdgeUnknownKey     = errors.New("edge signature key id is unknown")
	ErrEdgeStale          = errors.New("edge signature timestamp is outside the allowed window")
	ErrEdgeBadSignature   = errors.New("edge signature does not match")
	ErrEdgeReplayed       = errors.New("edge signature nonce was already used")
)

var (
	nonceFormat     = regexp.MustCompile(`^[0-9a-f]{32}$`)
	signatureFormat = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type KeyLookup interface {
	Lookup(id string) ([]byte, bool)
}

type Clock interface {
	Now() time.Time
}

type SignedRequest struct {
	KeyID     string
	Method    string
	URI       string
	Timestamp string
	Nonce     string
	ClientIP  string
	Body      []byte
	Signature string
}

func (r SignedRequest) CanonicalString() string {
	bodyHash := sha256.Sum256(r.Body)
	return strings.Join([]string{
		canonicalVersion, r.Method, r.URI, r.Timestamp, r.Nonce, r.ClientIP, hex.EncodeToString(bodyHash[:]),
	}, "\n")
}

type EdgeVerifier struct {
	keys   KeyLookup
	clock  Clock
	nonces *nonceCache
}

func NewEdgeVerifier(keys KeyLookup, clock Clock) *EdgeVerifier {
	return &EdgeVerifier{keys: keys, clock: clock, nonces: newNonceCache(nonceLifetime)}
}

func (v *EdgeVerifier) Verify(request SignedRequest) error {
	signedAt, err := parseHeaders(request)
	if err != nil {
		return err
	}
	key, ok := v.keys.Lookup(request.KeyID)
	if !ok {
		return ErrEdgeUnknownKey
	}
	now := v.clock.Now()
	if now.Sub(signedAt).Abs() > timestampWindow {
		return ErrEdgeStale
	}
	if !hasValidSignature(key, request) {
		return ErrEdgeBadSignature
	}
	if !v.nonces.claim(request.Nonce, now) {
		return ErrEdgeReplayed
	}
	return nil
}

func parseHeaders(request SignedRequest) (time.Time, error) {
	if !nonceFormat.MatchString(request.Nonce) || !signatureFormat.MatchString(request.Signature) {
		return time.Time{}, ErrEdgeMissingHeaders
	}
	if request.KeyID == "" || request.ClientIP == "" {
		return time.Time{}, ErrEdgeMissingHeaders
	}
	seconds, err := strconv.ParseInt(request.Timestamp, 10, 64)
	if err != nil {
		return time.Time{}, ErrEdgeMissingHeaders
	}
	return time.Unix(seconds, 0), nil
}

func hasValidSignature(key []byte, request SignedRequest) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(request.CanonicalString()))
	expected := mac.Sum(nil)
	provided, err := hex.DecodeString(request.Signature)
	return err == nil && hmac.Equal(expected, provided)
}
