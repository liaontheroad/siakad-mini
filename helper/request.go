package helper

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func ParamID(c *fiber.Ctx) (int, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, BadRequest("ID harus berupa angka positif")
	}
	return id, nil
}

func BindAndValidate(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return BadRequest("Body harus berupa JSON yang valid")
	}
	if errs := Validate(dst); errs != nil {
		return Validation(errs)
	}
	return nil
}
