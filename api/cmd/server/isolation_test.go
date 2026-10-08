package main

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type isolationWorld struct {
	store             organizationFixture
	consignor         organizationFixture
	storeStranger     *browser
	consignorStranger *browser
}

type organizationFixture struct {
	owner          *browser
	id             string
	fileID         string
	notificationID string
	productID      string
	inviteCode     string
	proposalID     string
	partnershipID  string
	dropID         string
	countID        string
	settlementID   string
	paymentID      string
}

type isolationCase func(world isolationWorld) exchange

func isolationCases() map[string]isolationCase {
	return map[string]isolationCase{
		"GET /api/v1/organizations/:id": func(w isolationWorld) exchange {
			return w.storeStranger.get(organizationsPath + "/" + w.store.id)
		},
		"PATCH /api/v1/organizations/:id": func(w isolationWorld) exchange {
			return w.storeStranger.patch(organizationsPath+"/"+w.store.id, profileUpdate(""))
		},
		"PATCH /api/v1/organizations/:id/store-profile": func(w isolationWorld) exchange {
			return w.storeStranger.patch(organizationsPath+"/"+w.store.id+"/store-profile", storeProfileBody(""))
		},
		"PATCH /api/v1/organizations/:id/consignor-profile": func(w isolationWorld) exchange {
			return w.consignorStranger.patch(organizationsPath+"/"+w.consignor.id+"/consignor-profile",
				consignorProfileBody)
		},
		"PATCH /api/v1/organizations/:id/payment-info": func(w isolationWorld) exchange {
			return w.consignorStranger.patch(organizationsPath+"/"+w.consignor.id+"/payment-info",
				`{"payment_info":"","payment_qr_file_id":null}`)
		},
		"PATCH /api/v1/organizations/:id/listing": func(w isolationWorld) exchange {
			return w.storeStranger.patch(organizationsPath+"/"+w.store.id+"/listing", `{"is_listed":false}`)
		},
		"GET /api/v1/files/:id": func(w isolationWorld) exchange {
			return w.consignorStranger.get("/api/v1/files/" + w.consignor.fileID)
		},
		"POST /api/v1/notifications/:id/read": func(w isolationWorld) exchange {
			return w.consignorStranger.post("/api/v1/notifications/" + w.consignor.notificationID + "/read")
		},
		"GET /api/v1/products/:id": func(w isolationWorld) exchange {
			return w.consignorStranger.get(productsPath + "/" + w.consignor.productID)
		},
		"PATCH /api/v1/products/:id": func(w isolationWorld) exchange {
			return w.consignorStranger.patch(productsPath+"/"+w.consignor.productID, productBody(""))
		},
		"POST /api/v1/products/:id/archive": func(w isolationWorld) exchange {
			return w.consignorStranger.post(productsPath + "/" + w.consignor.productID + "/archive")
		},
		"POST /api/v1/products/:id/reactivate": func(w isolationWorld) exchange {
			return w.consignorStranger.post(productsPath + "/" + w.consignor.productID + "/reactivate")
		},
		"PATCH /api/v1/products/:id/seeking": func(w isolationWorld) exchange {
			return w.consignorStranger.patch(productsPath+"/"+w.consignor.productID+"/seeking",
				`{"is_seeking_stores":true}`)
		},
		"GET /api/v1/invitations/:code": func(w isolationWorld) exchange {
			return w.storeStranger.get(invitationsPath + "/" + w.consignor.inviteCode)
		},
		"POST /api/v1/invitations/:code/claim": func(w isolationWorld) exchange {
			return w.storeStranger.post(invitationsPath + "/" + w.consignor.inviteCode + "/claim")
		},
		"GET /api/v1/directory/organizations/:id": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/directory/organizations/" + w.consignor.id)
		},
		"GET /api/v1/proposals/:id": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/proposals/" + w.consignor.proposalID)
		},
		"POST /api/v1/proposals/:id/accept": func(w isolationWorld) exchange {
			return w.storeStranger.post("/api/v1/proposals/" + w.consignor.proposalID + "/accept")
		},
		"POST /api/v1/proposals/:id/decline": func(w isolationWorld) exchange {
			return w.storeStranger.postJSON("/api/v1/proposals/"+w.consignor.proposalID+"/decline", `{"reason":"x"}`)
		},
		"POST /api/v1/proposals/:id/withdraw": func(w isolationWorld) exchange {
			return w.consignorStranger.post("/api/v1/proposals/" + w.consignor.proposalID + "/withdraw")
		},
		"GET /api/v1/partnerships/:id/agreement": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/agreement")
		},
		"GET /api/v1/partnerships/:id": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID)
		},
		"GET /api/v1/partnerships/:id/catalog": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/catalog")
		},
		"GET /api/v1/partnerships/:id/agreements": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/agreements")
		},
		"GET /api/v1/partnerships/:id/agreements/:version": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/agreements/1")
		},
		"GET /api/v1/drops/:id": func(w isolationWorld) exchange {
			return w.storeStranger.get(dropsPath + "/" + w.consignor.dropID)
		},
		"GET /api/v1/partnerships/:id/stock": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/stock")
		},
		"GET /api/v1/partnerships/:id/stock/:product_id/card": func(w isolationWorld) exchange {
			return w.consignorStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/stock/" +
				w.consignor.productID + "/card")
		},
		"GET /api/v1/partnerships/:id/count-sheet": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/count-sheet")
		},
		"GET /api/v1/counts/:id": func(w isolationWorld) exchange {
			return w.consignorStranger.get("/api/v1/counts/" + w.consignor.countID)
		},
		"POST /api/v1/counts/:id/confirm": func(w isolationWorld) exchange {
			return w.consignorStranger.post("/api/v1/counts/" + w.consignor.countID + "/confirm")
		},
		"GET /api/v1/settlements/:id": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/settlements/" + w.consignor.settlementID)
		},
		"POST /api/v1/settlements/:id/payments": func(w isolationWorld) exchange {
			return w.storeStranger.recordPayment(w.consignor.settlementID, paymentBody(1750, "CASH", ""), paymentKey)
		},
		"POST /api/v1/settlements/:id/mark-paid": func(w isolationWorld) exchange {
			return w.consignorStranger.post("/api/v1/settlements/" + w.consignor.settlementID + "/mark-paid")
		},
		"POST /api/v1/payments/:id/confirm": func(w isolationWorld) exchange {
			return w.consignorStranger.post("/api/v1/payments/" + w.consignor.paymentID + "/confirm")
		},
		"POST /api/v1/payments/:id/reject": func(w isolationWorld) exchange {
			return w.consignorStranger.postJSON("/api/v1/payments/"+w.consignor.paymentID+"/reject", `{"reason":"x"}`)
		},
		"GET /api/v1/partnerships/:id/blockers": func(w isolationWorld) exchange {
			return w.storeStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/blockers")
		},
		"POST /api/v1/partnerships/:id/end": func(w isolationWorld) exchange {
			return w.consignorStranger.postJSON("/api/v1/partnerships/"+w.consignor.partnershipID+"/end",
				`{"reason":""}`)
		},
		"PATCH /api/v1/partnerships/:id/public-listing": func(w isolationWorld) exchange {
			return w.storeStranger.patch("/api/v1/partnerships/"+w.consignor.partnershipID+"/public-listing",
				`{"show_on_public":true}`)
		},
		"GET /api/v1/partnerships/:id/running-sales": func(w isolationWorld) exchange {
			return w.consignorStranger.get("/api/v1/partnerships/" + w.consignor.partnershipID + "/running-sales")
		},
		"POST /api/v1/drops/:id/confirm": func(w isolationWorld) exchange {
			return w.consignorStranger.post(dropsPath + "/" + w.consignor.dropID + "/confirm")
		},
		"POST /api/v1/drops/:id/cancel": func(w isolationWorld) exchange {
			return w.storeStranger.post(dropsPath + "/" + w.consignor.dropID + "/cancel")
		},
		"POST /api/v1/partnerships/:id/amendments": func(w isolationWorld) exchange {
			return w.consignorStranger.postJSON("/api/v1/partnerships/"+w.consignor.partnershipID+"/amendments",
				invitationBody(""))
		},
	}
}

func routesWithIDs(t *testing.T) []string {
	t.Helper()
	public := publicRoutes()
	var routes []string
	for _, route := range newTestServer(alwaysReady).App().GetRoutes(true) {
		isPublic := slices.ContainsFunc(public, func(p httpserver.PublicRoute) bool {
			return p.Method == route.Method && p.Path == route.Path
		})
		if strings.Contains(route.Path, "/:") && route.Method != http.MethodHead && !isPublic {
			routes = append(routes, route.Method+" "+route.Path)
		}
	}
	return routes
}

func TestIsolation_RSEC41_EveryRouteWithAnIDHasAnIsolationCase(t *testing.T) {
	routes := routesWithIDs(t)
	cases := isolationCases()

	for _, route := range routes {
		if _, ok := cases[route]; !ok {
			t.Errorf("route %s has no isolation case in isolationCases", route)
		}
	}
	for route := range cases {
		if !slices.Contains(routes, route) {
			t.Errorf("isolation case %s matches no route", route)
		}
	}
}
