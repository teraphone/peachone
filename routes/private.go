package routes

import (
	"fmt"
	"peachone/database"
	"peachone/models"
	"peachone/queries"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Private Welcome handler
func PrivateWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "private"})
}

// --------------------------------------------------------------------------------
// Update License request handler
// --------------------------------------------------------------------------------
type UpdateLicenseResponse struct {
	Success bool               `json:"success"`
	License models.UserLicense `json:"license"`
}

func UpdateLicense(c *fiber.Ctx) error {
	// extract claims from JWT
	claims, err := getClaimsFromJWT(c)
	if err != nil {
		fmt.Println("error extracting claims from JWT:", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT.")
	}

	// get database connection
	db := database.DB.DB

	// get license
	license := &models.UserLicense{
		Oid: claims.Oid,
	}
	query := db.Where("oid = ?", license.Oid).Find(license)
	if query.RowsAffected == 0 {
		fmt.Println("license not found for user:", license.Oid)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	// update license
	if !license.TrialActivated {
		tx := db.Model(license).Updates(models.UserLicense{
			TrialActivated: true,
			TrialExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		})
		if tx.Error != nil {
			fmt.Println("error updating license:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}
	}

	// return response
	response := &UpdateLicenseResponse{
		Success: true,
		License: *license,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// Get World request handler
// --------------------------------------------------------------------------------
type GetWorldResponse struct {
	Teams []models.TeamInfo `json:"teams"`
}

func GetWorld(c *fiber.Ctx) error {
	// extract claims from JWT
	claims, err := getClaimsFromJWT(c)
	if err != nil {
		fmt.Println("error extracting claims from JWT:", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT.")
	}

	// get database connection
	db := database.DB.DB

	// get user
	user := &models.TenantUser{}
	query := db.Where("oid = ?", claims.Oid).Find(user)
	if query.RowsAffected == 0 {
		fmt.Println("user not found for user:", claims.Oid)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}
	fmt.Println("found user:", user)

	// get subscription
	subscription := &models.Subscription{}
	if user.SubscriptionId != "" {
		query = db.Where("id = ?", user.SubscriptionId).Find(subscription)
		if query.RowsAffected == 0 {
			fmt.Println("subscription not found:", user.SubscriptionId)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}
	}

	teamInfos := []models.TeamInfo{}

	// get the user's teams
	usersTeams := []models.TeamUser{}
	query = db.Where("oid = ?", user.Oid).Find(&usersTeams)
	if query.RowsAffected == 0 {
		fmt.Println("no teams found for user:", user.Oid)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	// get TeamInfo for each team
	for _, userTeam := range usersTeams {
		roomInfos := []models.RoomInfo{}

		// get TenantTeam
		team := &models.TenantTeam{}
		query := db.Where("id = ?", userTeam.Id).Find(team)
		if query.RowsAffected == 0 {
			fmt.Println("team not found:", userTeam.Id)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}

		// get TeamRooms for team
		rooms := []models.TeamRoom{}
		query = db.Where("team_id = ?", team.Id).Find(&rooms)
		if query.RowsAffected == 0 {
			fmt.Println("rooms not found for team:", team.Id)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}

		// check if subscription is active or trial is active
		subscriptionActive := subscription.SaaSSubscriptionStatus == models.SubscriptionStatusEnumSubscribed
		trialActive := user.TrialActivated && (time.Now().Unix() < user.TrialExpiresAt.Unix())
		canJoin := subscriptionActive || trialActive

		// for each room, get LivekitJoinToken
		for _, room := range rooms {
			token, err := createLiveKitJoinToken(room.TeamId, room.Id.String(), userTeam.Oid)
			if err != nil {
				fmt.Println("error creating LiveKitJoinToken:", err)
				return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
			}
			roomInfo := models.RoomInfo{
				Room: room,
			}
			if canJoin {
				roomInfo.RoomToken = token
			}
			roomInfos = append(roomInfos, roomInfo)
		}

		// get users for team
		users, err := queries.GetUsersForTeam(db, team.Id)
		if err != nil {
			fmt.Println("error getting users for team:", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}

		// create TeamInfo
		teamInfo := &models.TeamInfo{
			Team:  *team,
			Rooms: roomInfos,
			Users: users,
		}

		teamInfos = append(teamInfos, *teamInfo)
	}

	// return response
	response := &GetWorldResponse{
		Teams: teamInfos,
	}
	return c.JSON(response)
}

// --------------------------------------------------------------------------------
// Get Refreshed Access Token request handler
// --------------------------------------------------------------------------------
type GetRefreshedAccessTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type GetRefreshedAccessTokenResponse struct {
	Success                bool   `json:"success"`
	AccessToken            string `json:"accessToken"`
	AccessTokenExpiration  int64  `json:"accessTokenExpiration"`
	RefreshToken           string `json:"refreshToken"`
	RefreshTokenExpiration int64  `json:"refreshTokenExpiration"`
}

func GetRefreshedAccessToken(c *fiber.Ctx) error {
	// get request body
	req := &GetRefreshedAccessTokenRequest{}
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate refresh token
	claims, err := validateToken(req.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid refresh token.")
	}

	// create new access and refresh tokens
	user := &models.TenantUser{
		Oid: claims.Oid,
		Tid: claims.Tid,
	}
	accessToken, accessTokenExpiration, err := createAccessToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	refreshToken, refreshTokenExpiration, err := createRefreshToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	// return response
	response := &GetRefreshedAccessTokenResponse{
		Success:                true,
		AccessToken:            accessToken,
		AccessTokenExpiration:  accessTokenExpiration,
		RefreshToken:           refreshToken,
		RefreshTokenExpiration: refreshTokenExpiration,
	}
	return c.JSON(response)
}
