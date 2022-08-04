package routes

import (
	"encoding/json"
	"fmt"
	"peachone/auth"
	"peachone/database"
	"peachone/models"
	"peachone/queries"

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
	Success                bool               `json:"success"`
	AccessToken            string             `json:"accessToken"`
	AccessTokenExpiration  int64              `json:"accessTokenExpiration"`
	RefreshToken           string             `json:"refreshToken"`
	RefreshTokenExpiration int64              `json:"refreshTokenExpiration"`
	FirebaseAuthToken      string             `json:"firebaseAuthToken"`
	User                   models.TenantUser  `json:"user"`
	License                models.UserLicense `json:"license"`
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
			Tid:         ReadString(teamable.GetTenantId()), // <-- why is this empty?
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

	// user license
	license := &models.UserLicense{}

	// get database connection
	db := database.DB.DB

	// check if user exists
	query := db.Where("oid = ?", user.Oid).Find(user)
	if query.RowsAffected == 0 {
		err := queries.SetUpNewUserAndLicense(db, user, license)
		if err != nil {
			fmt.Println("error setting up new user and license:", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}

	} else {
		// get user license
		query = db.Where("oid = ?", user.Oid).Find(license)
		if query.RowsAffected == 0 {
			fmt.Println("license not found for user:", user)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}
	}

	// for each team
	for _, team := range teams {
		// fix empty team.Tid
		team.Tid = user.Tid

		// check if team exists
		query := db.Where("id = ?", team.Id).Find(&team)
		if query.RowsAffected == 0 {
			err := queries.SetUpNewTeamAndRooms(db, &team)
			if err != nil {
				fmt.Println("error setting up new team and rooms:", err, team)
				return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
			}
		}

		// check if user exists in team
		teamUser := &models.TeamUser{
			Id:  team.Id,
			Oid: user.Oid,
		}
		query = db.Where("id = ? AND oid = ?", team.Id, user.Oid).Find(teamUser)
		if query.RowsAffected == 0 {
			db.Create(teamUser)
			fmt.Println("create team user:", teamUser)
		}
	}

	// create access token
	accessToken, accessTokenExp, err := createAccessToken(user)
	if err != nil {
		fmt.Println("error creating access token:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	// create refresh token
	refreshToken, refreshTokenExpiration, err := createRefreshToken(user)
	if err != nil {
		fmt.Println("error creating refresh token:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	// create firebase auth token
	firebaseAuthToken, err := createFirebaseAuthToken(c.Context(), user.Oid)
	if err != nil {
		fmt.Println("error creating firebase auth token:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	// return response
	response := &LoginResponse{
		Success:                true,
		AccessToken:            accessToken,
		AccessTokenExpiration:  accessTokenExp,
		RefreshToken:           refreshToken,
		RefreshTokenExpiration: refreshTokenExpiration,
		FirebaseAuthToken:      firebaseAuthToken,
		User:                   *user,
		License:                *license,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// EmailSignup
// --------------------------------------------------------------------------------
type EmailSignupRequest struct {
	Email string `json:"email"`
}

type EmailSignupResponse struct {
	Success bool `json:"success"`
}

func EmailSignup(c *fiber.Ctx) error {
	// get request body
	req := &EmailSignupRequest{}
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid email.")
	}

	// send alert email
	alertVars := &EmailSignupAlertVars{
		SenderEmail:     "alerts@teraphone.app",
		Subject:         "[email-signup] New Email Signup",
		RecipientEmails: []string{"david@teraphone.app", "nathan@teraphone.app"},
		TemplateVars: &EmailSignupAlertTemplateVars{
			Email: req.Email,
		},
	}
	message, id, err := SendEmailSignupAlert(c.Context(), alertVars)
	if err != nil {
		fmt.Println("error sending email signup alert:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}
	fmt.Println("send email signup alert for:", req.Email)
	fmt.Println("message:", message)
	fmt.Println("id:", id)

	// return response
	response := &EmailSignupResponse{
		Success: true,
	}
	return c.JSON(response)

}
