package main

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/clock"
)

const (
	signedInUserID   = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	signedInFamilyID = "0192f0c4-8a11-7d55-b0aa-3c1f2e4d5a6b"
)

func TestAccessVerifier_BRAUTHZ05_ActorMembershipComesFromTheSignedToken(t *testing.T) {
	now := clock.NewManual(testNow())
	components := testAuthComponents(now)
	membership := actor.Membership{
		OrganizationID:   "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c",
		OrganizationType: actor.Consignor,
		Role:             actor.Staff,
	}
	claims := domain.NewAccessClaims(signedInUserID, signedInFamilyID, testNow())
	claims.Membership = membership
	token, err := components.tokens.SignAccessToken(claims)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	signedIn, err := newAccessVerifier(components).VerifyAccess(token)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}

	if signedIn.Membership != membership || signedIn.UserID != claims.Subject {
		t.Fatalf("actor = %+v", signedIn)
	}
}

func TestAccessVerifier_FRAUTH05_TokenWithoutOrganizationGivesAnActorWithoutMembership(t *testing.T) {
	now := clock.NewManual(testNow())
	components := testAuthComponents(now)
	claims := domain.NewAccessClaims(signedInUserID, signedInFamilyID, testNow())
	token, err := components.tokens.SignAccessToken(claims)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	signedIn, err := newAccessVerifier(components).VerifyAccess(token)

	if err != nil || signedIn.IsMember() {
		t.Fatalf("actor = %+v, err = %v", signedIn, err)
	}
}
