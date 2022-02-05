package routes

import (
	"os"
	"peachone/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type SignupRequest struct {
	Name     string
	Email    string
	Password string
}

type LoginRequest struct {
	Email    string
	Password string
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
