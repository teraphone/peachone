package routes

import (
	"peachone/database"
	"peachone/models"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

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
