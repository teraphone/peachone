package routes

import (
	"context"
	"peachone/database"
	"peachone/models"

	"github.com/gofiber/fiber/v2"

	livekit "github.com/livekit/protocol/livekit"
)

// -----------------------------------------------------------------------------
// Get livekit rooms
// -----------------------------------------------------------------------------
type GetLivekitRoomsResponse struct {
	Success bool `json:"success"`
	livekit.ListRoomsResponse
}

func GetLivekitRooms(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

	// get room_users for user id
	room_users := []models.RoomUser{}
	query := db.Where("user_id = ?", id).Find(&room_users)
	if query.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error getting room_user records.")
	}

	// get roomservice client
	client := CreateRoomServiceClient()

	// list rooms (only returns "active" rooms)
	// TODO: only request rooms that the user is in
	rooms, err := client.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	if err != nil {
		return err
	}

	// return response
	response := &GetLivekitRoomsResponse{
		Success:           true,
		ListRoomsResponse: *rooms,
	}
	return c.JSON(response)

}
