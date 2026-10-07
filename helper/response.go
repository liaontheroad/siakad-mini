package helper

import (
	"github.com/gofiber/fiber/v2"

	"siakad-mini/app/model"
)

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func SuccessList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true, Message: message, Data: data, Meta: meta,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	c.Set("Location", location)
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true, Message: message, Data: data,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func WriteError(c *fiber.Ctx, status int, message string, errs map[string][]string) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false, Message: message, Errors: errs,
	})
}

func NewMeta(page, perPage, total int) *model.Meta {
	lastPage := 0
	if perPage > 0 {
		lastPage = (total + perPage - 1) / perPage
	}
	return &model.Meta{
		CurrentPage: page, PerPage: perPage, Total: total, LastPage: lastPage,
	}
}
