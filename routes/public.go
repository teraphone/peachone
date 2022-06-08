package routes

import (
	"fmt"
	"peachone/database"
	"peachone/models"
	"peachone/queries"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
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
	Success           bool        `json:"success"`
	Token             string      `json:"token"`
	Expiration        int64       `json:"expiration"`
	FirebaseAuthToken string      `json:"firebase_auth_token"`
	User              models.User `json:"user"`
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
	query := db.Where("email = ?", strings.ToLower(req.Email)).Find(user)
	if query.RowsAffected > 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid signup credentials.")
	}

	// hash password, populate user model, save to db
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Name = req.Name
	user.Email = strings.ToLower(req.Email)
	user.Password = string(hash)

	db.Create(user)

	// create JWT token
	token, expiration, err := createJWTToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating JWT token.")
	}

	// create firebase auth token
	firebase_auth_token, err := createFirebaseAuthToken(c.Context(), user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating Firebase auth token.")
	}

	// return response
	response := &SignupResponse{
		Success:           true,
		Token:             token,
		Expiration:        expiration,
		FirebaseAuthToken: firebase_auth_token,
		User:              *user,
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
	Success           bool        `json:"success"`
	Token             string      `json:"token"`
	Expiration        int64       `json:"expiration"`
	FirebaseAuthToken string      `json:"firebase_auth_token"`
	User              models.User `json:"user"`
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
	query := db.Where("email = ?", strings.ToLower(req.Email)).Find(user)
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
	user.Email = strings.ToLower(req.Email)
	user.Password = string(hash)

	db.Create(user)

	// create JWT token
	token, expiration, err := createJWTToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating JWT token.")
	}

	// create firebase auth token
	firebase_auth_token, err := createFirebaseAuthToken(c.Context(), user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating Firebase auth token.")
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
		Success:           true,
		Token:             token,
		Expiration:        expiration,
		FirebaseAuthToken: firebase_auth_token,
		User:              *user,
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
	Success           bool        `json:"success"`
	Token             string      `json:"token"`
	Expiration        int64       `json:"expiration"`
	FirebaseAuthToken string      `json:"firebase_auth_token"`
	User              models.User `json:"user"`
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
	query := db.Where("email = ?", strings.ToLower(req.Email)).Find(user)
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
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating JWT token.")
	}

	// create firebase auth token
	firebase_auth_token, err := createFirebaseAuthToken(c.Context(), user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating Firebase auth token.")
	}

	// return response
	response := &LoginResponse{
		Success:           true,
		Token:             token,
		Expiration:        expiration,
		FirebaseAuthToken: firebase_auth_token,
		User:              *user,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// Email Verification request handler
// --------------------------------------------------------------------------------
type EmailVerificationRequest struct {
	Code string `json:"code"`
}

type EmailVerificationResponse struct {
	Success bool `json:"success"`
}

func EmailVerification(c *fiber.Ctx) error {
	// get request body
	req := new(EmailVerificationRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid verification code.")
	}

	// get database connection
	db := database.DB.DB

	// check if code exists in db
	evcode := new(models.EmailVerificationCode)
	query := db.Where("code = ?", req.Code).Find(evcode)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid verification code.")
	}

	// check if code has expired
	if time.Now().Unix() > evcode.ExpiresAt.Unix() {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid verification code.")
	}

	// get user
	user := new(models.User)
	db.Where("id = ?", evcode.UserID).Find(user)

	// update user
	user.IsVerified = true
	db.Save(user)

	// return response
	response := &EmailVerificationResponse{
		Success: true,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// Password Reset request handler
// --------------------------------------------------------------------------------
type PasswordResetRequest struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type PasswordResetResponse struct {
	Success bool `json:"success"`
}

func PasswordReset(c *fiber.Ctx) error {
	// get request body
	req := new(PasswordResetRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Code == "" || req.NewPassword == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid password reset request.")
	}

	// get database connection
	db := database.DB.DB

	// check if code exists in db
	prcode := new(models.PasswordResetCode)
	query := db.Where("code = ?", req.Code).Find(prcode)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid password reset request.")
	}

	// check if code has expired
	if time.Now().Unix() > prcode.ExpiresAt.Unix() {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid password reset request.")
	}

	// get user
	user := new(models.User)
	db.Where("id = ?", prcode.UserID).Find(user)

	// hash password, populate user model, save to db
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	db.Save(user)

	// return response
	response := &PasswordResetResponse{
		Success: true,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// Forgot Password request handler
// --------------------------------------------------------------------------------
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ForgotPasswordResponse struct {
	Success bool `json:"success"`
}

func ForgotPassword(c *fiber.Ctx) error {
	// get request body
	req := new(ForgotPasswordRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid forgot password request.")
	}

	// get database connection
	db := database.DB.DB

	// check if email exists in db
	user := new(models.User)
	query := db.Where("email = ?", strings.ToLower(req.Email)).Find(user)
	if query.RowsAffected == 0 {
		response := &ForgotPasswordResponse{
			Success: true,
		}
		return c.JSON(response)
	}

	// create password reset code
	expiresInHours := uint(2)
	prcode := new(models.PasswordResetCode)
	prcode.UserID = user.ID
	code, err := uuid.NewV4()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	prcode.Code = code.String()
	prcode.ExpiresAt = time.Now().Add(time.Hour * time.Duration(expiresInHours))
	db.Create(prcode)

	// send password reset email
	passwordRestVars := &PasswordResetVars{
		SenderEmail: "david@teraphone.app",
		Subject:     "[Teraphone]: Instructions for changing your Teraphone password",
		TemplateVars: &PasswordResetTemplateVars{
			Name:           user.Name,
			Email:          user.Email,
			Code:           prcode.Code,
			ExpiresInHours: expiresInHours,
			SenderName:     "David Wurtz",
		},
	}
	_, _, err = SendPasswordResetEmail(c.Context(), passwordRestVars)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// return response
	response := &ForgotPasswordResponse{
		Success: true,
	}
	return c.JSON(response)

}
