package httpserver

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
)

const codeOK = "OK"

type Reply struct {
	Status  int
	Message string
	Data    any
}

type Page[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
	Search     any
}

type meta struct {
	RequestID string `json:"request_id"`
}

type listMeta struct {
	RequestID  string  `json:"request_id"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
	Search     any     `json:"search,omitempty"`
}

type successEnvelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    any    `json:"meta"`
}

func Respond(c fiber.Ctx, reply Reply) error {
	status := reply.Status
	if status == 0 {
		status = fiber.StatusOK
	}
	return writeJSON(c, status, successEnvelope{
		Code:    codeOK,
		Message: reply.Message,
		Data:    reply.Data,
		Meta:    meta{RequestID: RequestIDFrom(c)},
	})
}

func RespondList[T any](c fiber.Ctx, page Page[T]) error {
	items := page.Items
	if items == nil {
		items = []T{}
	}
	return writeJSON(c, fiber.StatusOK, successEnvelope{
		Code: codeOK,
		Data: items,
		Meta: listMeta{
			RequestID: RequestIDFrom(c), NextCursor: cursorOrNull(page.NextCursor), HasMore: page.HasMore,
			Search: page.Search,
		},
	})
}

func cursorOrNull(cursor string) *string {
	if cursor == "" {
		return nil
	}
	return &cursor
}

func writeJSON(c fiber.Ctx, status int, body any) error {
	c.Set(fiber.HeaderCacheControl, "no-store")
	if err := c.Status(status).JSON(body); err != nil {
		return fmt.Errorf("write json response: %w", err)
	}
	return nil
}

type FileReply struct {
	ContentType string
	Data        []byte
	IsPublic    bool
}

func RespondFile(c fiber.Ctx, reply FileReply) error {
	c.Set(fiber.HeaderContentType, reply.ContentType)
	c.Set(fiber.HeaderContentDisposition, "inline")
	c.Set(fiber.HeaderCacheControl, "no-store")
	if reply.IsPublic {
		c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")
		c.Set(fiber.HeaderCrossOriginResourcePolicy, "cross-origin")
	}
	if err := c.Status(fiber.StatusOK).Send(reply.Data); err != nil {
		return fmt.Errorf("write file response: %w", err)
	}
	return nil
}
