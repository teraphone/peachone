package routes

import (
	"encoding/json"
	"fmt"
	"peachone/auth"
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
	client, err := auth.NewMSGraphClient(req.MSAccessToken)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not authenticate.")
	}

	// GET https://graph.microsoft.com/v1.0/me
	me := client.Me()
	joinedTeamsReq := me.JoinedTeams()
	result, errObj := joinedTeamsReq.Get()
	if errObj != nil {
		errJSON, _ := json.MarshalIndent(errObj, "", "  ")
		fmt.Println("Error making request:", errJSON)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}
	// process result
	teamables := result.GetValue()
	teams := make([]models.TenantTeam, len(teamables))
	for i, teamable := range teamables {
		teams[i] = models.TenantTeam{
			Id:          ReadString(teamable.GetId()),
			Tid:         ReadString(teamable.GetTenantId()),
			DisplayName: ReadString(teamable.GetDisplayName()),
			Description: ReadString(teamable.GetDescription()),
		}
	}

	fmt.Println("teams:", teams)

	// return response
	response := &LoginResponse{
		Success: true,
	}
	return c.JSON(response)

}
