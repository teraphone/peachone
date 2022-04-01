package routes

import (
	"peachone/database"
	"peachone/models"
	"peachone/queries"
	"strconv"

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

func GetLiveKitRooms(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get database connection
	db := database.DB.DB

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
	rooms, err := client.ListRooms(c.Context(), &livekit.ListRoomsRequest{
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

// -----------------------------------------------------------------------------
// Join livekit room
// -----------------------------------------------------------------------------
type JoinLiveKitRoomResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
}

func JoinLiveKitRoom(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// get room_id from request
	room_id_str := c.Params("room_id")
	room_id, err := strconv.ParseUint(room_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid room id.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("group_id = ? AND user_id = ?", group_id, id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this group.")
	}

	// verify user is in room
	room_user := &models.RoomUser{}
	query = db.Where("room_id = ? AND user_id = ?", room_id, id).Find(room_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this room.")
	}

	// verify user is not banned
	if room_user.RoomRoleID == models.RoomRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this room.")
	}

	// construct access token
	token, err := createLiveKitJoinToken(room_user, uint(group_id), uint(room_id), id)
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
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// get room_id from request
	room_id_str := c.Params("room_id")
	room_id, err := strconv.ParseUint(room_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid room id.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("group_id = ? AND user_id = ?", group_id, id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this group.")
	}

	// verify user is in room
	room_user := &models.RoomUser{}
	query = db.Where("room_id = ? AND user_id = ?", room_id, id).Find(room_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this room.")
	}

	// verify user is not banned
	if room_user.RoomRoleID == models.RoomRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this room.")
	}

	// get roomservice client
	client := CreateRoomServiceClient()

	// get room participants
	participants, err := client.ListParticipants(c.Context(), &livekit.ListParticipantsRequest{
		Room: EncodeRoomName(uint(group_id), uint(room_id)),
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
