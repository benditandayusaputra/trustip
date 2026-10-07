package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	tokenAlgorithm    = "HS256"
	tokenType         = "JWT"
	tokenPartsCount   = 3
	refreshTokenBytes = 32
)

var (
	ErrTokenMalformed  = errors.New("access token is malformed")
	ErrTokenAlgorithm  = errors.New("access token algorithm is not accepted")
	ErrTokenUnknownKey = errors.New("access token key id is unknown")
	ErrTokenSignature  = errors.New("access token signature does not match")
	ErrTokenExpired    = errors.New("access token has expired")
)

type AccessClaims struct {
	Subject          string `json:"sub"`
	SessionID        string `json:"sid"`
	Organization     string `json:"org,omitempty"`
	OrganizationType string `json:"org_type,omitempty"`
	Role             string `json:"role,omitempty"`
	IssuedAt         int64  `json:"iat"`
	ExpiresAt        int64  `json:"exp"`
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
	Type      string `json:"typ"`
}

type TokenSigner struct {
	keys SecretRing
}

func NewTokenSigner(keys SecretRing) *TokenSigner {
	return &TokenSigner{keys: keys}
}

func (s *TokenSigner) Sign(claims AccessClaims) (string, error) {
	keyID, key := s.keys.Active()
	header, err := encodeSegment(tokenHeader{Algorithm: tokenAlgorithm, KeyID: keyID, Type: tokenType})
	if err != nil {
		return "", err
	}
	payload, err := encodeSegment(claims)
	if err != nil {
		return "", err
	}
	signingInput := header + "." + payload
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sign(key, signingInput)), nil
}

func (s *TokenSigner) Verify(token string, now time.Time) (AccessClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != tokenPartsCount {
		return AccessClaims{}, ErrTokenMalformed
	}
	key, err := s.keyFor(parts[0])
	if err != nil {
		return AccessClaims{}, err
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, sign(key, parts[0]+"."+parts[1])) {
		return AccessClaims{}, ErrTokenSignature
	}
	var claims AccessClaims
	if err := decodeSegment(parts[1], &claims); err != nil {
		return AccessClaims{}, err
	}
	if now.Unix() >= claims.ExpiresAt {
		return AccessClaims{}, ErrTokenExpired
	}
	return claims, nil
}

func (s *TokenSigner) keyFor(encodedHeader string) ([]byte, error) {
	var header tokenHeader
	if err := decodeSegment(encodedHeader, &header); err != nil {
		return nil, err
	}
	if header.Algorithm != tokenAlgorithm || header.Type != tokenType {
		return nil, ErrTokenAlgorithm
	}
	key, ok := s.keys.Lookup(header.KeyID)
	if !ok {
		return nil, ErrTokenUnknownKey
	}
	return key, nil
}

type RefreshToken struct {
	Plain string
	Hash  []byte
}

func NewRefreshToken() (RefreshToken, error) {
	random := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(random); err != nil {
		return RefreshToken{}, fmt.Errorf("read refresh token: %w", err)
	}
	plain := base64.RawURLEncoding.EncodeToString(random)
	return RefreshToken{Plain: plain, Hash: HashRefreshToken(plain)}, nil
}

func HashRefreshToken(plain string) []byte {
	digest := sha256.Sum256([]byte(plain))
	return digest[:]
}

func sign(key []byte, signingInput string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(signingInput))
	return mac.Sum(nil)
}

func encodeSegment(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode token segment: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(encoded), nil
}

func decodeSegment(segment string, target any) error {
	decoded, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return ErrTokenMalformed
	}
	if err := json.Unmarshal(decoded, target); err != nil {
		return ErrTokenMalformed
	}
	return nil
}
