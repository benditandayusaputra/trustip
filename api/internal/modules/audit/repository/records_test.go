package repository_test

import (
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/audit/repository"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database/dbtest"
)

const (
	userID     = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"
	storedHash = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA"
)

func TestRecord_NFR14_KeepsTheEntryEvenWithoutAnOrganization(t *testing.T) {
	pool := dbtest.MigratedPool(t)
	queries := sqlc.New(pool.Querier(t.Context()))
	recordedAt := time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
	if err := queries.CreateUser(t.Context(), sqlc.CreateUserParams{
		ID: userID, Email: "sari@contoh.id", PasswordHash: storedHash, PepperVersion: "v1",
		FullName: "Sari Wulandari", Whatsapp: "6281234567890", PrivacyAcceptedAt: recordedAt,
	}); err != nil {
		t.Fatalf("user: %v", err)
	}

	err := repository.NewRecords(pool).Record(t.Context(), audit.Entry{
		ID: "0192f0c4-bb00-7e66-a1bb-4d2f3e5d6a01", ActorID: userID, Action: audit.SignedIn,
		EntityType: audit.UserEntity, EntityID: userID, IP: "203.0.113.7", At: recordedAt,
	})
	if err != nil {
		t.Fatalf("record: %v", err)
	}

	rows, err := queries.ListAuditActions(t.Context(), userID)
	if err != nil || len(rows) != 1 || rows[0].Action != "SIGNED_IN" || rows[0].OrganizationID != "" ||
		rows[0].Ip != "203.0.113.7" {
		t.Fatalf("rows = %+v, err = %v", rows, err)
	}
}
