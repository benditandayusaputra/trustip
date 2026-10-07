package idgen

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

type UUIDv7 struct{}

func (UUIDv7) NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate uuid v7: %w", err)
	}
	return id.String(), nil
}

type Sequence struct {
	last atomic.Uint64
}

func NewSequence() *Sequence {
	return &Sequence{}
}

func (s *Sequence) NewID() (string, error) {
	return fmt.Sprintf("00000000-0000-7000-8000-%012x", s.last.Add(1)), nil
}
