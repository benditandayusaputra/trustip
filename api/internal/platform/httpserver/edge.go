package httpserver

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/platform/security"
)

const (
	HeaderEdgeKeyID     = "X-Edge-Key-Id"
	HeaderEdgeTimestamp = "X-Edge-Timestamp"
	HeaderEdgeNonce     = "X-Edge-Nonce"
	HeaderEdgeClientIP  = "X-Edge-Client-Ip"
	HeaderEdgeSignature = "X-Edge-Signature"
)

type EdgeVerifier interface {
	Verify(request security.SignedRequest) error
}

type clientIPKey struct{}

func EdgeSignature(verifier EdgeVerifier, securityLog *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		if isUnsigned(c) {
			return c.Next()
		}
		request := signedRequestFrom(c)
		if err := verifier.Verify(request); err != nil {
			securityLog.Warn("edge signature invalid",
				"event", "edge_signature_invalid", "request_id", RequestIDFrom(c), "ip", c.IP(), "reason", err.Error())
			return apperror.New(apperror.CodeEdgeSignatureInvalid)
		}
		c.Locals(clientIPKey{}, request.ClientIP)
		return c.Next()
	}
}

const PublicFilesPrefix = "/files/"

func isUnsigned(c fiber.Ctx) bool {
	isPublicFile := c.Method() == fiber.MethodGet && strings.HasPrefix(c.Path(), PublicFilesPrefix)
	return c.Path() == healthPath || isPublicFile
}

func signedRequestFrom(c fiber.Ctx) security.SignedRequest {
	return security.SignedRequest{
		KeyID:     c.Get(HeaderEdgeKeyID),
		Method:    c.Method(),
		URI:       string(c.Request().RequestURI()),
		Timestamp: c.Get(HeaderEdgeTimestamp),
		Nonce:     c.Get(HeaderEdgeNonce),
		ClientIP:  c.Get(HeaderEdgeClientIP),
		Body:      c.Body(),
		Signature: c.Get(HeaderEdgeSignature),
	}
}

func ClientIP(c fiber.Ctx) string {
	if verified, ok := c.Locals(clientIPKey{}).(string); ok {
		return verified
	}
	return c.IP()
}
