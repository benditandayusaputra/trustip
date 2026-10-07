package main

import (
	"slices"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver/edgetest"
)

func TestAudit_NFR14_SignInAndProfileChangesAreRecordedWithTheClientIP(t *testing.T) {
	visitor := databaseBrowser(t)
	founded := visitor.postOrganization(onboarding(semarangCityCode))
	organizationID := quotedAfter(string(founded.body), `"id":"`)
	visitor.post("/api/v1/sessions/refresh")
	visitor.patch(organizationsPath+"/"+organizationID, profileUpdate(""))
	visitor.patch(organizationsPath+"/"+organizationID+"/listing", `{"is_listed":false}`)
	account := visitor.getAccount()
	userID := quotedAfter(string(account.body), `"id":"`)

	rows, err := sqlc.New(visitor.pool.Querier(t.Context())).ListAuditActions(t.Context(), userID)
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}

	actions := make([]string, 0, len(rows))
	for _, row := range rows {
		actions = append(actions, row.Action)
	}
	want := []string{"SIGNED_IN", "ORGANIZATION_CREATED", "PROFILE_UPDATED", "LISTING_CHANGED"}
	if !slices.Equal(actions, want) || rows[0].Ip != edgetest.ClientIP || rows[1].OrganizationID != organizationID {
		t.Fatalf("rows = %+v", rows)
	}
}
