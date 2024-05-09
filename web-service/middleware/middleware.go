package middleware

import "github.com/gofiber/fiber/v2"

func CheckClient(c *fiber.Ctx) error {
	client := string(c.Request().Header.Peek("Client"))
	if client == "Mobile" {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
		"message": "User Unauthorized",
	})
}

func CheckRole(c *fiber.Ctx) error {
	client := string(c.Request().Header.Peek("Role"))
	if client == "Admin" {
		return c.Next()
	}
	return c.Status(fiber.StatusUnauthorized).JSON(map[string]any{
		"message": "User Unauthorized",
	})
}
