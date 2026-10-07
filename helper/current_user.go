package helper

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
)

const localsAuthUserKey = "auth_user"

func SetCurrentUser(c *fiber.Ctx, u model.AuthUser) {
	c.Locals(localsAuthUserKey, u)
}

func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	u, ok := c.Locals(localsAuthUserKey).(model.AuthUser)
	return u, ok
}
