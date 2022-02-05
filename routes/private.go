package routes

import (
	"peachone/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// Private Welcome handler
func PrivateWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "private"})
}

// refresh expiration on JTW token
func RefreshToken(c *fiber.Ctx) error {
	// get claims from JWT token
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)

	// extract userid from claims
	user := new(models.User)
	user.ID = uint(claims["id"].(float64)) // why float64?

	// create refreshed JWT token
	fresh_token, expiration, err := createJWTToken(user)
	if err != nil {
		return err
	}

	// return response
	return c.JSON(fiber.Map{"token": fresh_token, "expiration": expiration})
}
