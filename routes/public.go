package routes

import (
	"peachone/models"

	"github.com/gofiber/fiber/v2"
)

// Public Welcome handler
func PublicWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "public"})
}

// --------------------------------------------------------------------------------
// Login request handler
// --------------------------------------------------------------------------------
type LoginRequest struct {
	MSAccessToken string `json:"msAccessToken"`
	IdToken       string `json:"idToken"`
}

type LoginResponse struct {
	Success           bool              `json:"success"`
	AccessToken       string            `json:"accessToken"`
	Expiration        int64             `json:"expiration"`
	RefreshToken      string            `json:"refreshToken"`
	FirebaseAuthToken string            `json:"firebaseAuthToken"`
	User              models.TenantUser `json:"user"`
}

func Login(c *fiber.Ctx) error {
	// get request body
	req := new(LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.MSAccessToken == "" || req.IdToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid login credentials.")
	}

	// return response
	response := &LoginResponse{
		Success: true,
	}
	return c.JSON(response)

}
