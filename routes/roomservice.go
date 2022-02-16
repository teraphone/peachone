package routes

import (
	"context"
	"os"
	"peachone/database"
	"peachone/models"

	"github.com/gofiber/fiber/v2"

	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
)

// -----------------------------------------------------------------------------
// Get livekit rooms
// -----------------------------------------------------------------------------
type GetLivekitRoomsResponse struct {
	Success bool                      `json:"success"`
	Rooms   livekit.ListRoomsResponse `json:"rooms"`
}

func GetLivekitRooms(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

	// get user's rooms
	user_rooms := []models.RoomUser{}
	query := db.Where("user_id = ?", id).Find(&user_rooms)
	if query.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error getting user's rooms.")
	}

	LIVEKIT_KEY := os.Getenv("LIVEKIT_KEY")
	LIVEKIT_SECRET := os.Getenv("LIVEKIT_SECRET")
	LIVEKIT_HOST := os.Getenv("LIVEKIT_HOST")

	roomClient := lksdk.NewRoomServiceClient(LIVEKIT_HOST, LIVEKIT_KEY, LIVEKIT_SECRET)

	// list rooms
	rooms, err := roomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{})
	if err != nil {
		return err
	}

	// return response
	response := &GetLivekitRoomsResponse{
		Success: true,
		Rooms:   *rooms,
	}
	return c.JSON(response)

}
