package httpserver

import (
	"fmt"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
)

type panicError struct {
	value any
	stack []byte
}

func (p *panicError) Error() string {
	return fmt.Sprintf("panic: %v", p.value)
}

func Recover() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = apperror.Internal(&panicError{value: recovered, stack: debug.Stack()})
			}
		}()
		return c.Next()
	}
}
