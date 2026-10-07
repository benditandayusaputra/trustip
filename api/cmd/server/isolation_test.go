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
