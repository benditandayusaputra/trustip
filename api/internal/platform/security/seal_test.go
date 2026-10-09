package security_test

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const (
	rfc8032Seed      = "9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60"
	rfc8032Public    = "d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a"
	rfc8032Signature = "e5564300c360ac729086e2cc806e828a84877f1eb8e5d974d873e06522490155" +
		"5fb8821590a33bacc61e39701cf9b46bd25bf5f0595bbe24655141438e7a100b"
)

type fixedKeys struct {
	keys map[string][]byte
}

func (f fixedKeys) Active() (string, []byte) { return "v1", f.keys["v1"] }

func (f fixedKeys) IDs() []string { return []string{"v1"} }

func (f fixedKeys) Lookup(id string) ([]byte, bool) {
	key, ok := f.keys[id]
	return key, ok
}

func rfcSealer(t *testing.T) *security.Sealer {
	t.Helper()
	seed, err := hex.DecodeString(rfc8032Seed)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return security.NewSealer(fixedKeys{keys: map[string][]byte{"v1": seed}})
}

func TestSealer_RCC51_MatchesTheFirstRFC8032Vector(t *testing.T) {
	sealer := rfcSealer(t)

	signature := sealer.Sign([]byte{})
	public, ok := sealer.PublicKey("v1")

	if hex.EncodeToString(signature.Bytes) != rfc8032Signature || signature.KeyID != "v1" || !ok ||
		hex.EncodeToString(public) != rfc8032Public {
		t.Fatalf("signature = %x, public = %x", signature.Bytes, public)
	}
}

func TestSealer_RSEC29_ChangedDigestFailsVerification(t *testing.T) {
	sealer := rfcSealer(t)
	signature := sealer.Sign([]byte("entry hash"))

	same, sameErr := sealer.Verify([]byte("entry hash"), signature)
	changed, changedErr := sealer.Verify([]byte("entry hasH"), signature)

	if !same || changed || errors.Join(sameErr, changedErr) != nil {
		t.Fatalf("same = %v, changed = %v", same, changed)
	}
}

func TestSealer_UnknownKeyIsAnError(t *testing.T) {
	_, err := rfcSealer(t).Verify([]byte("x"), security.Signature{KeyID: "v9"})

	if !errors.Is(err, security.ErrSealUnknownKey) {
		t.Fatalf("err = %v", err)
	}
}

type sealVector struct {
	SeedHex   string `json:"seed_hex"`
	EntryHash string `json:"entry_hash"`
	Signature string `json:"signature"`
}

func TestSealer_RSEC29_SharedSealVectorSignsTheSame(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "vectors", "seal.json"))
	if err != nil {
		t.Fatalf("read vector: %v", err)
	}
	var vector sealVector
	if err := json.Unmarshal(raw, &vector); err != nil {
		t.Fatalf("decode vector: %v", err)
	}
	seed, seedErr := hex.DecodeString(vector.SeedHex)
	digest, digestErr := hex.DecodeString(vector.EntryHash)
	if errors.Join(seedErr, digestErr) != nil {
		t.Fatalf("decode hex: %v %v", seedErr, digestErr)
	}

	signature := security.NewSealer(fixedKeys{keys: map[string][]byte{"v1": seed}}).Sign(digest)

	if got := hex.EncodeToString(signature.Bytes); got != vector.Signature {
		t.Fatalf("signature = %s", got)
	}
}

func TestPublicKeys_RSEC32_PublishesEveryKeyOfTheRing(t *testing.T) {
	keys := rfcSealer(t).PublicKeys()

	if len(keys) != 1 || keys[0].KeyID != "v1" || hex.EncodeToString(keys[0].Bytes) != rfc8032Public {
		t.Fatalf("keys = %+v", keys)
	}
}
