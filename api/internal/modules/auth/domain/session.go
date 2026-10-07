package domain

import "time"

const (
	RefreshTokenLifetime = 30 * 24 * time.Hour
	AccessTokenLifetime  = 15 * time.Minute
)

type SessionIdentity struct {
	ID       string
	UserID   string
	FamilyID string
}

type Session struct {
	SessionIdentity
	TokenHash []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}

func NewSession(identity SessionIdentity, tokenHash []byte, now time.Time) Session {
	return Session{
		SessionIdentity: identity,
		TokenHash:       tokenHash,
		CreatedAt:       now,
		ExpiresAt:       now.Add(RefreshTokenLifetime),
	}
}

type RefreshToken struct {
	Plain string
	Hash  []byte
}

type AccessClaims struct {
	Subject       string
	SessionFamily string
	IssuedAt      time.Time
	ExpiresAt     time.Time
}

func NewAccessClaims(userID, familyID string, now time.Time) AccessClaims {
	return AccessClaims{
		Subject:       userID,
		SessionFamily: familyID,
		IssuedAt:      now,
		ExpiresAt:     now.Add(AccessTokenLifetime),
	}
}

type SessionTokens struct {
	Access         string
	AccessExpires  time.Time
	Refresh        string
	RefreshExpires time.Time
}
