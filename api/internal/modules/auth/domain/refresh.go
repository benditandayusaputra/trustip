package domain

import (
	"errors"
	"time"
)

const RotationGrace = 10 * time.Second

var ErrRefreshReused = errors.New("a rotated refresh token was used again")

type RefreshDecision int

const (
	RefreshRejected RefreshDecision = iota
	RefreshRotate
	RefreshReplay
	RefreshReuseDetected
)

type StoredSession struct {
	SessionIdentity
	ExpiresAt time.Time
	RotatedAt time.Time
	RevokedAt time.Time
}

func DecideRefresh(session StoredSession, now time.Time) RefreshDecision {
	switch {
	case !session.RevokedAt.IsZero() || !now.Before(session.ExpiresAt):
		return RefreshRejected
	case session.RotatedAt.IsZero():
		return RefreshRotate
	case now.Sub(session.RotatedAt) <= RotationGrace:
		return RefreshReplay
	}
	return RefreshReuseDetected
}

func (s StoredSession) Rotate(successorID string, successorHash []byte, now time.Time) Session {
	identity := SessionIdentity{ID: successorID, UserID: s.UserID, FamilyID: s.FamilyID}
	return NewSession(identity, successorHash, now)
}
