package security

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	phcAlgorithm  = "argon2id"
	phcPartsCount = 6
)

type phcHash struct {
	params Argon2Params
	salt   []byte
	key    []byte
}

func (p phcHash) String() string {
	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s", phcAlgorithm, argon2.Version,
		p.params.MemoryKiB, p.params.Iterations, p.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(p.salt), base64.RawStdEncoding.EncodeToString(p.key))
}

func parsePHC(encoded string) (phcHash, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != phcPartsCount || parts[0] != "" || parts[1] != phcAlgorithm {
		return phcHash{}, ErrMalformedPasswordHash
	}
	params, err := parsePHCParams(parts[2], parts[3])
	if err != nil {
		return phcHash{}, err
	}
	salt, err := decodePHCBytes(parts[4], argon2SaltBytes)
	if err != nil {
		return phcHash{}, err
	}
	key, err := decodePHCBytes(parts[5], argon2KeyBytes)
	if err != nil {
		return phcHash{}, err
	}
	return phcHash{params: params, salt: salt, key: key}, nil
}

func parsePHCParams(versionPart, paramsPart string) (Argon2Params, error) {
	var version int
	if _, err := fmt.Sscanf(versionPart, "v=%d", &version); err != nil || version != argon2.Version {
		return Argon2Params{}, ErrMalformedPasswordHash
	}
	var params Argon2Params
	_, err := fmt.Sscanf(paramsPart, "m=%d,t=%d,p=%d", &params.MemoryKiB, &params.Iterations, &params.Parallelism)
	if err != nil {
		return Argon2Params{}, ErrMalformedPasswordHash
	}
	return params, nil
}

func decodePHCBytes(encoded string, length int) ([]byte, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) != length {
		return nil, ErrMalformedPasswordHash
	}
	return decoded, nil
}
