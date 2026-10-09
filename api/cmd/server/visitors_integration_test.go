package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestVisitor_FRPUB07_BrowsesTheBoardAndTheStoreDirectoryWithoutSigningIn(t *testing.T) {
	scene := newBoardScene(t)
	visitor := scene.consignor.visitorOnTheSameServer()
	contract := loadContract(t)

	board := visitor.get("/api/v1/directory/needs?origin=CITY&city_code=33.74&radius_km=all")
	stores := visitor.get(directoryStoresPath + "?origin=CITY&city_code=33.74&radius_km=all")

	requireCode(t, board, "OK")
	requireCode(t, stores, "OK")
	contract.requireConforms(t, board)
	contract.requireConforms(t, stores)
	if !strings.Contains(string(board.body), scene.needID) || !strings.Contains(string(stores.body), scene.storeID) ||
		strings.Contains(string(stores.body), `"score":`) {
		t.Fatalf("board = %s\nstores = %s", board.body, stores.body)
	}
}

func TestVisitor_FRNED08_FRPUB06_OpensASharedNeed(t *testing.T) {
	scene := newBoardScene(t)

	got := scene.consignor.visitorOnTheSameServer().get("/api/v1/public/needs/" + scene.needID)

	requireCode(t, got, "OK")
	loadContract(t).requireConforms(t, got)
	if !strings.Contains(string(got.body), `"title":"Butuh roti untuk sarapan"`) ||
		!strings.Contains(string(got.body), `"status":"OPEN"`) {
		t.Fatalf("got = %s", got.body)
	}
	requireNoneOf(t, got, "whatsapp", "6281234567890", "Jl. Sirojudin")
}

func TestVisitor_UnknownNeedIsNotFound(t *testing.T) {
	scene := newBoardScene(t)

	got := scene.consignor.visitorOnTheSameServer().get("/api/v1/public/needs/0192f0c4-0000-7000-8000-00000000ffff")

	if got.status != http.StatusNotFound {
		t.Fatalf("got = %d %s", got.status, got.body)
	}
}

func TestVisitor_BRAUTHZ01_SignedInUserWithoutOrganizationStillNeedsOne(t *testing.T) {
	newcomer := databaseBrowser(t)

	got := newcomer.get(directoryStoresPath)

	requireCode(t, got, "FORBIDDEN")
}
