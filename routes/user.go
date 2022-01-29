package routes

import (
	"os"
	"peachone/database"
	"peachone/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
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

// Signup request handler
func Signup(c *fiber.Ctx) error {
	// get request body
	req := new(SignupRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signup credentials.")
	}

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

	// check if email already exists in db
	user := new(models.User)
	query := db.Where("email = ?", req.Email).Find(user)
	if query.RowsAffected > 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signup credentials.")
	}

	// hash password, populate user model, save to db
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Password = string(hash)

	db.Create(user)

	// create JWT token
	token, expiration, err := createJWTToken(user)
	if err != nil {
		return err
	}

	// return response
	return c.JSON(fiber.Map{"token": token, "expiration": expiration, "user": user})

}

// Login request handler
func Login(c *fiber.Ctx) error {
	// get request body
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid login credentials.")
	}

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

	// check if email exists in db
	user := new(models.User)
	query := db.Where("email = ?", req.Email).Find(user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid login credentials.")
	}

	// check if password matches
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid login credentials.")
	}

	// create JWT token
	token, expiration, err := createJWTToken(user)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"token": token, "expiration": expiration, "user": user})

}

// Public Welcome handler
func PublicWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "public"})
}

// Private Welcome handler
func PrivateWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "private"})
}

func createJWTToken(user *models.User) (string, int64, error) {
	expiration := time.Now().Add(time.Hour * 24).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userid"] = user.ID
	claims["expiration"] = expiration
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	tokenString, err := token.SignedString([]byte(SIGNING_KEY))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiration, nil
}
