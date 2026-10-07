package policy_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
)

func member(orgType actor.OrganizationType, role actor.Role) actor.Actor {
	return actor.Actor{
		UserID:        "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e",
		SessionFamily: "0192f0c4-8a11-7d55-b0aa-3c1f2e4d5a6b",
		Membership: actor.Membership{
			OrganizationID:   "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c",
			OrganizationType: orgType,
			Role:             role,
		},
	}
}

func TestRequire_BRAUTHZ01_AllowedRolePasses(t *testing.T) {
	if err := policy.Require(member(actor.Store, actor.Staff), policy.RecordCounts); err != nil {
		t.Fatalf("store staff counting = %v", err)
	}
}

func TestRequire_BRAUTHZ01_RoleWithoutGrantIsForbidden(t *testing.T) {
	err := policy.Require(member(actor.Consignor, actor.Staff), policy.EditProfile)

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestRequire_BRAUTHZ01_WrongOrganizationTypeIsForbidden(t *testing.T) {
	err := policy.Require(member(actor.Consignor, actor.Owner), policy.RecordDrops)

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestRequire_BRAUTHZ01_UserWithoutOrganizationIsForbidden(t *testing.T) {
	visitor := actor.Actor{UserID: "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"}

	err := policy.Require(visitor, policy.ReportListings)

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestRequire_DelegatedGrantStaysClosedUntilOwnersCanDelegate(t *testing.T) {
	err := policy.Require(member(actor.Store, actor.Staff), policy.RecordPayments)

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestRequire_UnknownActionIsForbidden(t *testing.T) {
	err := policy.Require(member(actor.Store, actor.Owner), policy.Action("DROP_DATABASE"))

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestRequire_UnknownRoleIsForbidden(t *testing.T) {
	err := policy.Require(member(actor.Store, actor.Role("ADMIN")), policy.RecordCounts)

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestRequireNewcomer_BRAUTH02_UserWithoutOrganizationMayFoundOne(t *testing.T) {
	visitor := actor.Actor{UserID: "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"}

	if err := policy.RequireNewcomer(visitor); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestRequireNewcomer_MemberCannotFoundASecondOrganization(t *testing.T) {
	err := policy.RequireNewcomer(member(actor.Store, actor.Owner))

	if apperror.From(err).Code() != apperror.CodeOrganizationExists {
		t.Fatalf("err = %v", err)
	}
}

func TestRequireMember_BRAUTHZ01_AnyRoleOfAnOrganizationPasses(t *testing.T) {
	if err := policy.RequireMember(member(actor.Consignor, actor.Staff)); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := policy.RequireMember(actor.Actor{UserID: "u"}); apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("newcomer err = %v", err)
	}
}
