package security_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const referenceJWT = "eyJhbGciOiJIUzI1NiIsImtpZCI6InYxIiwidHlwIjoiSldUIn0." +
	"eyJzdWIiOiIwMTkyZjBjNC03YjFlLTdjM2EtOWQyZS01ZjZhN2I4YzlkMGUiLCJzaWQiOiIwMTkyZjBjNC03YjFlLTdjM2EtOWQyZS0w" +
	"MDAwMDAwMDAwMDEiLCJpYXQiOjE3OTIzOTY4MDAsImV4cCI6MTc5MjM5NzcwMH0." +
	"mmgd78k8VgofsENNHlIyfhi8NqEZnqrK6QnxUJqGsgo"

func signingKeys() pepperRing {
	return pepperRing{activeID: "v1", peppers: map[string][]byte{"v1": bytes.Repeat([]byte{5}, 32)}}
}

func issuedAt() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func sampleClaims() security.AccessClaims {
	return security.AccessClaims{
		Subject:   "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e",
		SessionID: "0192f0c4-7b1e-7c3a-9d2e-000000000001",
		IssuedAt:  issuedAt().Unix(),
		ExpiresAt: issuedAt().Add(15 * time.Minute).Unix(),
	}
}

func TestSign_RSEC16_MatchesAnIndependentHS256Implementation(t *testing.T) {
	token, err := security.NewTokenSigner(signingKeys()).Sign(sampleClaims())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if token != referenceJWT {
		t.Fatalf("token = %s", token)
	}
}

func TestVerify_RSEC16_ValidTokenGivesBackItsClaims(t *testing.T) {
	claims, err := security.NewTokenSigner(signingKeys()).Verify(referenceJWT, issuedAt())
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if claims != sampleClaims() {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestVerify_RSEC16_Rejections(t *testing.T) {
	parts := strings.Split(referenceJWT, ".")
	noneHeader := "eyJhbGciOiJub25lIiwia2lkIjoidjEiLCJ0eXAiOiJKV1QifQ"
	otherKid := "eyJhbGciOiJIUzI1NiIsImtpZCI6InY5IiwidHlwIjoiSldUIn0"
	cases := []struct {
		name  string
		token string
		at    time.Time
		want  error
	}{
		{"expired", referenceJWT, issuedAt().Add(15 * time.Minute), security.ErrTokenExpired},
		{"alg none", noneHeader + "." + parts[1] + ".", issuedAt(), security.ErrTokenAlgorithm},
		{"unknown kid", otherKid + "." + parts[1] + "." + parts[2], issuedAt(), security.ErrTokenUnknownKey},
		{"tampered payload", parts[0] + "." + parts[1] + "x." + parts[2], issuedAt(), security.ErrTokenSignature},
		{"two parts", parts[0] + "." + parts[1], issuedAt(), security.ErrTokenMalformed},
		{"garbage", "not-a-token", issuedAt(), security.ErrTokenMalformed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := security.NewTokenSigner(signingKeys()).Verify(tc.token, tc.at)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestVerify_RotatedKey_StillAcceptsTokensSignedWithTheOldKey(t *testing.T) {
	rotated := signingKeys()
	rotated.activeID = "v2"
	rotated.peppers["v2"] = bytes.Repeat([]byte{6}, 32)

	if _, err := security.NewTokenSigner(rotated).Verify(referenceJWT, issuedAt()); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

const familyID = "0192f0c4-7b1e-7c3a-9d2e-000000000001"

func TestNewRefreshToken_RSEC17_IsRandomNamesItsFamilyAndIsStoredOnlyAsSHA256(t *testing.T) {
	tokens := security.NewRefreshTokens(signingKeys())

	first, err := tokens.New(familyID)
	if err != nil {
		t.Fatalf("new token: %v", err)
	}
	second, err := tokens.New(familyID)
	if err != nil {
		t.Fatalf("new token: %v", err)
	}

	family, ok := security.FamilyOf(first.Plain)
	if first.Plain == second.Plain || !ok || family != familyID || len(first.Plain) < len(familyID)+44 {
		t.Fatalf("tokens = %q, %q", first.Plain, second.Plain)
	}
	if !bytes.Equal(first.Hash, security.HashRefreshToken(first.Plain)) || len(first.Hash) != 32 {
		t.Fatalf("hash = %x", first.Hash)
	}
}

func TestSuccessor_RSEC18_IsTheSameEveryTimeForOneToken(t *testing.T) {
	tokens := security.NewRefreshTokens(signingKeys())
	previous, err := tokens.New(familyID)
	if err != nil {
		t.Fatalf("new token: %v", err)
	}

	first := tokens.Successor(previous.Plain)
	again := tokens.Successor(previous.Plain)

	family, _ := security.FamilyOf(first.Plain)
	if first.Plain != again.Plain || first.Plain == previous.Plain || family != familyID {
		t.Fatalf("successors = %q, %q", first.Plain, again.Plain)
	}
}

func TestSuccessor_DependsOnTheServerKey(t *testing.T) {
	other := signingKeys()
	other.peppers["v1"] = bytes.Repeat([]byte{9}, 32)
	previous := familyID + ".c2VjcmV0"

	mine := security.NewRefreshTokens(signingKeys()).Successor(previous)
	theirs := security.NewRefreshTokens(other).Successor(previous)

	if mine.Plain == theirs.Plain {
		t.Fatal("successor must not be computable without the server key")
	}
}

func TestFamilyOf_RejectsMalformedTokens(t *testing.T) {
	for _, token := range []string{"", "no-separator", familyID + ".", "not-a-uuid.c2VjcmV0", ".c2VjcmV0"} {
		if _, ok := security.FamilyOf(token); ok {
			t.Errorf("%q was accepted", token)
		}
	}
}

func TestHashRefreshToken_IsPlainSHA256(t *testing.T) {
	want := "d9bb52af16886c9df8ec1ee28e0bdcf44f86cbb851ca979a89f3b132c0bd825f"

	if got := hex.EncodeToString(security.HashRefreshToken("refresh-token-sample")); got != want {
		t.Fatalf("hash = %s", got)
	}
}
