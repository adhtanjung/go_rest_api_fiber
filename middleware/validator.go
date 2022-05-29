package middleware

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var Validator = validator.New()

type IError struct {
	Field string
	Tag   string
	Value string
}
type LoginInput struct {
	Identity string `json:"identity" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func ValidateUserLogin(c *fiber.Ctx) error {
	var errors []*IError
	user := new(LoginInput)
	c.BodyParser(&user)

	err := Validator.Struct(user)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var el IError
			el.Field = err.Field()
			el.Tag = err.Tag()
			el.Value = fmt.Sprint(err.Value())
			errors = append(errors, &el)
		}
		return c.Status(fiber.StatusBadRequest).JSON(errors)
	}

	return c.Next()
}
