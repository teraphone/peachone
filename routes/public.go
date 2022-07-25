package routes

import (
	"encoding/json"
	"fmt"
	"peachone/auth"
	"peachone/database"
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
	cred, client, err := auth.NewMSGraphClient(req.MSAccessToken)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not authenticate.")
	}

	// get joined teams
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

	// get user from IDToken
	user := &models.TenantUser{
		Oid:   cred.UserAuth.IDToken.Oid,
		Name:  cred.UserAuth.IDToken.Name,
		Email: cred.UserAuth.IDToken.Email,
		Tid:   cred.UserAuth.IDToken.TenantID,
	}
	fmt.Println("user from cred.UserAuth:", user)

	// get database connection
	db := database.DB.DB

	// check if user exists
	query := db.Where("oid = ?", user.Oid).Find(user)
	if query.RowsAffected == 0 {
		// SetUpNewUserAndLicense(db, user)
		fmt.Println("create user:", user)
		fmt.Println("create user license")
	}

	// for each team (todo: finish this)
	for _, team := range teams {
		// check if team exists
		query := db.Where("id = ?", team.Id).Find(team)
		if query.RowsAffected == 0 {
			// SetUpNewTeamAndRooms(db, team)
			fmt.Println("create team:", team)
			fmt.Println("create team rooms")
		}

		// check if user exists in team
		teamUser := &models.TeamUser{
			Id:  team.Id,
			Oid: user.Oid,
		}
		query = db.Where("id = ? AND oid = ?", team.Id, user.Oid).Find(teamUser)
		if query.RowsAffected == 0 {
			// db.Create(teamUser)
			fmt.Println("create team user:", teamUser)
		}
	}

	// return response
	response := &LoginResponse{
		Success: true,
	}
	return c.JSON(response)

}
