package routes

import (
	"context"
	"peachone/database"
	"peachone/queries"

	"github.com/gofiber/fiber/v2"

	livekit "github.com/livekit/protocol/livekit"
)

// -----------------------------------------------------------------------------
// Get livekit rooms
// -----------------------------------------------------------------------------
type GetLivekitRoomsResponse struct {
	livekit.ListRoomsResponse      // will be absent if empty
	Success                   bool `json:"success"`
}

func GetLivekitRooms(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

	// get user rooms
	user_rooms, err := queries.GetUserRooms(db, uint(id))
	if err != nil {
		return err
	}

	// get list of livekit room names that the user is a member of
	var livekit_room_names []string
	for _, user_room := range user_rooms {
		name := EncodeRoomName(user_room.GroupID, user_room.ID)
		livekit_room_names = append(livekit_room_names, name)
	}

	// get roomservice client
	client := CreateRoomServiceClient()

	// list rooms (only returns "active" rooms)
	rooms, err := client.ListRooms(context.Background(), &livekit.ListRoomsRequest{
		Names: livekit_room_names,
	})
	if err != nil {
		return err
	}

	response := &GetLivekitRoomsResponse{
		ListRoomsResponse: *rooms,
		Success:           true,
	}
	return c.JSON(response)
}
