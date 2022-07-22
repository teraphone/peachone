package routes

import (
	"peachone/database"
	"peachone/models"

	"github.com/gofiber/fiber/v2"

	livekit "github.com/livekit/protocol/livekit"
)

// -----------------------------------------------------------------------------
// Join livekit room
// -----------------------------------------------------------------------------
type JoinLiveKitRoomResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

func JoinLiveKitRoom(c *fiber.Ctx) error {
	// extract userId from JWT claims
	tokenClaims, err := getClaimsFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT.")
	}
	userId := tokenClaims.Oid

	// get teamId, roomId from request
	teamId := c.Params("teamId")
	roomId := c.Params("room_id")

	// get database connection
	db := database.DB.DB

	// verify user is in team
	teamUser := &models.TeamUser{}
	query := db.Where("id = ? AND oid = ?", teamId, userId).Find(teamUser)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this team.")
	}

	// construct access token
	token, err := createLiveKitJoinToken(teamId, roomId, userId)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error generating access token.")
	}

	// return response
	response := &JoinLiveKitRoomResponse{
		Success: true,
		Token:   token,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Get livekit room participants
// -----------------------------------------------------------------------------
type GetLiveKitRoomParticipantsResponse struct {
	livekit.ListParticipantsResponse
	Success bool `json:"success"`
}

func GetLiveKitRoomParticipants(c *fiber.Ctx) error {
	// extract userId from JWT claims
	tokenClaims, err := getClaimsFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT.")
	}
	userId := tokenClaims.Oid

	// get teamId, roomId from request
	teamId := c.Params("teamId")
	roomId := c.Params("room_id")

	// get database connection
	db := database.DB.DB

	// verify user is in team
	teamUser := &models.TeamUser{}
	query := db.Where("id = ? AND oid = ?", teamId, userId).Find(teamUser)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this team.")
	}

	// get roomservice client
	client := CreateRoomServiceClient()

	// get room participants
	participants, err := client.ListParticipants(c.Context(), &livekit.ListParticipantsRequest{
		Room: EncodeRoomName(teamId, roomId),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error getting room participants.")
	}

	// return response
	response := &GetLiveKitRoomParticipantsResponse{
		ListParticipantsResponse: *participants,
		Success:                  true,
	}
	return c.JSON(response)
}
