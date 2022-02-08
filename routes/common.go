package routes

import (
	"os"
	"peachone/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func getIDFromJWT(c *fiber.Ctx) (uint, error) {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	id := uint(claims["id"].(float64))
	return id, nil
}

func createJWTToken(user *models.User) (string, int64, error) {
	expiration := time.Now().Add(time.Hour * 24).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["expiration"] = expiration
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	tokenString, err := token.SignedString([]byte(SIGNING_KEY))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiration, nil
}
