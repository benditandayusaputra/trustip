package main

import (
	"net/http"
	"strings"
	"testing"
)

type boardScene struct {
	consignorDirectoryScene
	consignor *browser
	storeID   string
	needID    string
}

func newBoardScene(t *testing.T) boardScene {
	t.Helper()
	scene := listedConsignorScene(t)
	account := scene.store.getAccount()
	storeID := quotedAfter(string(account.body), `"organization_id":"`)
	listed := scene.store.patch(organizationsPath+"/"+storeID+"/listing", `{"is_listed":true}`)
	if listed.status != http.StatusOK || storeID == "" {
		t.Fatalf("listing = %d %s", listed.status, listed.body)
	}
	posted := scene.store.postJSON(needsPath, needBody("Butuh roti untuk sarapan"))
	return boardScene{
		consignorDirectoryScene: scene, consignor: scene.consignor, storeID: storeID,
		needID: quotedAfter(string(posted.body), `"id":"`),
	}
}

func TestNeedBoard_FRNED03_FRNED07_ConsignorNearbyIsToldAndSeesTheNeed(t *testing.T) {
	scene := newBoardScene(t)
	contract := loadContract(t)

	board := scene.consignor.get("/api/v1/directory/needs?origin=CITY&city_code=33.74&radius_km=all")
	inbox := scene.consignor.get("/api/v1/notifications")

	requireCode(t, board, "OK")
	contract.requireConforms(t, board)
	if !strings.Contains(string(board.body), scene.needID) ||
		!strings.Contains(string(inbox.body), `"type":"NEED_POSTED"`) {
		t.Fatalf("board = %s\ninbox = %s", board.body, inbox.body)
	}
}

func TestNeedBoard_FRNED04_StoreCardShowsTheOpenNeedAndFilters(t *testing.T) {
	scene := newBoardScene(t)

	withNeed := scene.consignor.get(directoryStoresPath + "?radius_km=all&category_id=any&has_open_need=true")

	requireCode(t, withNeed, "OK")
	loadContract(t).requireConforms(t, withNeed)
	if !strings.Contains(string(withNeed.body), `"open_need_category_ids":["`+breadCategoryID+`"]`) ||
		!strings.Contains(string(withNeed.body), `"code":"NEEDS_CATEGORY"`) {
		t.Fatalf("got = %s", withNeed.body)
	}
}

func TestNeedResponse_FRNED05_ProposalLinkedToTheNeedShowsUpAsAResponse(t *testing.T) {
	scene := newBoardScene(t)
	body := strings.Replace(directBody(scene.storeID, scene.productID), `{"target_organization_id"`,
		`{"need_id":"`+scene.needID+`","target_organization_id"`, 1)

	sent := scene.consignor.postJSON("/api/v1/proposals", body)
	detail := scene.store.get(needsPath + "/" + scene.needID)

	if sent.status != http.StatusCreated {
		t.Fatalf("sent = %d %s", sent.status, sent.body)
	}
	loadContract(t).requireConforms(t, sent)
	if !strings.Contains(string(detail.body), `"response_count":1`) ||
		!strings.Contains(string(detail.body), quotedAfter(string(sent.body), `"id":"`)) {
		t.Fatalf("detail = %s", detail.body)
	}
}

func TestNeedResponse_FRNED05_ClosedNeedRefusesResponses(t *testing.T) {
	scene := newBoardScene(t)
	scene.store.post(needsPath + "/" + scene.needID + "/close")
	body := strings.Replace(directBody(scene.storeID, ""), `{"target_organization_id"`,
		`{"need_id":"`+scene.needID+`","target_organization_id"`, 1)

	got := scene.consignor.postJSON("/api/v1/proposals", body)

	requireCode(t, got, "NEED_NOT_OPEN")
	if !strings.Contains(string(got.body), "Kebutuhan ini sudah ditutup.") {
		t.Fatalf("got = %s", got.body)
	}
}
