package response

import "github.com/gofiber/fiber/v2"

type Response struct {
	Status  string      `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func Success(c *fiber.Ctx, code string, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Response{
		Status:  "success",
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Created(c *fiber.Ctx, code string, message string, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Status:  "success",
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *fiber.Ctx, code string, message string, details interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Status:  "error",
		Code:    code,
		Message: message,
		Details: details,
	})
}

func Unauthorized(c *fiber.Ctx, code string, message string, details interface{}) error {
	return c.Status(fiber.StatusUnauthorized).JSON(Response{
		Status:  "error",
		Code:    code,
		Message: message,
		Details: details,
	})
}

func InternalError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(Response{
		Status:  "error",
		Code:    "INTERNAL_ERROR",
		Message: "An unexpected error occurred",
	})
}

func ValidationError(c *fiber.Ctx, details interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Status:  "error",
		Code:    "VALIDATION_ERROR",
		Message: "Invalid request data",
		Details: details,
	})
}

func InvalidFormat(c *fiber.Ctx, err error, format interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(Response{
		Status:  "error",
		Code:    "BAD_REQUEST",
		Message: "Invalid request format",
		Details: fiber.Map{
			"error":           err.Error(),
			"required_format": format,
		},
	})
}
