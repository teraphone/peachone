package routes

import (
	"peachone/database"
	"peachone/models"
	"peachone/queries"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Public Welcome handler
func PublicWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "public"})
}

// --------------------------------------------------------------------------------
// Signup request handler
// --------------------------------------------------------------------------------
type SignupRequest struct {
	Name     string
	Email    string
	Password string
}

type SignupResponse struct {
	Success    bool        `json:"success"`
	Token      string      `json:"token"`
	Expiration int64       `json:"expiration"`
	User       models.User `json:"user"`
}

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

	// get database connection
	db := database.DB.DB

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
	response := &SignupResponse{
		Success:    true,
		Token:      token,
		Expiration: expiration,
		User:       *user,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// SignupWithInvite request handler
// --------------------------------------------------------------------------------
type SignupWithInviteRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"`
}

type SignupWithInviteResponse struct {
	Success    bool        `json:"success"`
	Token      string      `json:"token"`
	Expiration int64       `json:"expiration"`
	User       models.User `json:"user"`
}

func SignupWithInvite(c *fiber.Ctx) error {
	// get request body
	req := new(SignupWithInviteRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Name == "" || req.Email == "" || req.Password == "" || req.InviteCode == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signup credentials.")
	}

	// get database connection
	db := database.DB.DB

	// check if email already exists in db
	user := new(models.User)
	query := db.Where("email = ?", req.Email).Find(user)
	if query.RowsAffected > 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signup credentials.")
	}

	// validate invite code
	group_invite, err := queries.GetGroupInviteCode(db, req.InviteCode)
	if err != nil {
		return err
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

	// add user to group and rooms
	err = queries.AddUserToGroupAndRooms(db, user.ID, group_invite.GroupID)
	if err != nil {
		return err
	}

	// accept invite, create referral
	err = queries.AcceptInviteAndCreateReferral(db, group_invite, user.ID)
	if err != nil {
		return err
	}

	// return response
	response := &SignupWithInviteResponse{
		Success:    true,
		Token:      token,
		Expiration: expiration,
		User:       *user,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// Login request handler
// --------------------------------------------------------------------------------
type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	Success    bool        `json:"success"`
	Token      string      `json:"token"`
	Expiration int64       `json:"expiration"`
	User       models.User `json:"user"`
}

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

	// get database connection
	db := database.DB.DB

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

	// return response
	response := &LoginResponse{
		Success:    true,
		Token:      token,
		Expiration: expiration,
		User:       *user,
	}
	return c.JSON(response)

}
