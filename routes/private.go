package routes

import (
	"github.com/gofiber/fiber/v2"
)

// Private Welcome handler
func PrivateWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "private"})
}
