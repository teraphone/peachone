package routes

import (
	"encoding/json"
	"fmt"
	"peachone/auth"
	"peachone/models"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
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
	if req.MSAccessToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid login credentials.")
	}

	// authenticate with on-behalf-of flow
	cred, err := confidential.NewCredFromSecret(auth.Config.ClientSecret)
	if err != nil {
		fmt.Println("Error creating credential:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Authentication failed")
	}

	app, err := confidential.New(
		auth.Config.ClientID, cred,
		confidential.WithAuthority(auth.Config.Authority),
	)
	if err != nil {
		fmt.Println("Error creating auth client:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Authentication failed")
	}

	authResult, err := app.AcquireTokenOnBehalfOf(c.Context(), req.MSAccessToken, auth.Config.Scopes)
	if err != nil {
		fmt.Println("Error acquiring token on-behalf-of user:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Authentication failed")
	}

	authResultJSON, err := json.MarshalIndent(authResult, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling auth result:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Authentication failed")
	}

	// log authResultJSON
	fmt.Println("Auth result:", string(authResultJSON))

	// return response
	response := &LoginResponse{
		Success: true,
	}
	return c.JSON(response)

}
