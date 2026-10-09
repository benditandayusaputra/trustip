package handler

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

const (
	publicSealPath = "/public/seals/:code"
	sealKeysPath   = "/public/seal-keys"
	keyAlgorithm   = "Ed25519"
	statusValid    = "VALID"
	statusInvalid  = "INVALID"
)

type Verifier interface {
	Execute(ctx context.Context, rawCode string) (domain.PublicVerdict, error)
}

type KeyLister interface {
	Execute() []domain.PublicKey
}

type Deps struct {
	Verify      Verifier
	Keys        KeyLister
	Limiter     httpserver.RateLimiter
	SecurityLog *slog.Logger
}

func Mount(api fiber.Router, deps Deps) {
	verifyLimit := httpserver.RateLimit(deps.Limiter, httpserver.ByClientIP, deps.SecurityLog)
	api.Get(publicSealPath, verifyLimit, verify(deps.Verify, deps.SecurityLog))
	api.Get(sealKeysPath, listKeys(deps.Keys))
}

func PublicRoutes() []httpserver.PublicRoute {
	return []httpserver.PublicRoute{
		{Method: fiber.MethodGet, Path: httpserver.APIPrefix + publicSealPath},
		{Method: fiber.MethodGet, Path: httpserver.APIPrefix + sealKeysPath},
	}
}

type verificationResponse struct {
	Code           string `json:"code"`
	Status         string `json:"status"`
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
	EventType      string `json:"event_type"`
	OccurredAt     string `json:"occurred_at"`
	ConsignorName  string `json:"consignor_name"`
	StoreName      string `json:"store_name"`
}

func verify(verifier Verifier, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		verdict, err := verifier.Execute(c.Context(), c.Params("code"))
		if err != nil {
			return fmt.Errorf("verify seal: %w", err)
		}
		if !verdict.IsValid {
			securityLog.Warn("seal invalid", "event", "seal_invalid", "request_id", httpserver.RequestIDFrom(c),
				"ip", httpserver.ClientIP(c), "reason", verdict.Seal.Code)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: verificationOf(verdict)})
	}
}

func verificationOf(verdict domain.PublicVerdict) verificationResponse {
	status := statusInvalid
	if verdict.IsValid {
		status = statusValid
	}
	entry := verdict.Seal.Entry
	return verificationResponse{
		Code: verdict.Seal.Code, Status: status, DocumentType: entry.Document.Type,
		DocumentNumber: entry.Document.Number, EventType: string(entry.Event),
		OccurredAt:    entry.OccurredAt.UTC().Format(time.RFC3339),
		ConsignorName: verdict.Parties.ConsignorName, StoreName: verdict.Parties.StoreName,
	}
}

type publicKeyResponse struct {
	KeyID     string `json:"key_id"`
	Algorithm string `json:"algorithm"`
	PublicKey string `json:"public_key"`
}

func listKeys(lister KeyLister) fiber.Handler {
	return func(c fiber.Ctx) error {
		keys := lister.Execute()
		items := make([]publicKeyResponse, 0, len(keys))
		for _, key := range keys {
			items = append(items, publicKeyResponse{
				KeyID: key.KeyID, Algorithm: keyAlgorithm, PublicKey: base64.StdEncoding.EncodeToString(key.Bytes),
			})
		}
		return httpserver.RespondList(c, httpserver.Page[publicKeyResponse]{Items: items})
	}
}
