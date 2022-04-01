package routes

import (
	"math/rand"
	"peachone/database"
	"peachone/models"
	"peachone/queries"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	livekit "github.com/livekit/protocol/livekit"
)

// Private Welcome handler
func PrivateWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "private"})
}

// refresh expiration on JTW token
func RefreshToken(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

	// create refreshed JWT token
	user := new(models.User)
	user.ID = id
	fresh_token, expiration, err := createJWTToken(user)
	if err != nil {
		return err
	}

	// return response
	return c.JSON(fiber.Map{"token": fresh_token, "expiration": expiration})
}

// -----------------------------------------------------------------------------
// Create a new group
// -----------------------------------------------------------------------------
type CreateGroupRequest struct {
	Name string
}

type CreateGroupResponse struct {
	Success bool         `json:"success"`
	Group   models.Group `json:"group"`
}

func CreateGroup(c *fiber.Ctx) error {
	// get request body
	req := new(CreateGroupRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group name.")
	}

	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get database connection
	db := database.DB.DB

	// check if group name already exists
	query := db.Where("name = ?", req.Name).Find(&models.Group{})
	if query.RowsAffected != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Group name already exists.")
	}

	// create group
	group := &models.Group{Name: req.Name}
	tx := db.Create(group)
	if tx.Error != nil {
		return tx.Error
	}

	// create group_user
	group_user := &models.GroupUser{
		GroupID:     group.ID,
		UserID:      id,
		GroupRoleID: models.GroupRoleMap["owner"],
	}
	tx = db.Create(group_user)
	if tx.Error != nil {
		return tx.Error
	}

	// return response
	response := &CreateGroupResponse{
		Success: true,
		Group:   *group,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Get groups
// -----------------------------------------------------------------------------
type GetGroupsResponse struct {
	Success bool           `json:"success"`
	Groups  []models.Group `json:"groups"`
}

func GetGroups(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get database connection
	db := database.DB.DB

	// get group_users for user
	group_users := []models.GroupUser{}
	db.Where("user_id = ?", id).Find(&group_users)

	// get group for each group_id in group_users for user
	groups := []models.Group{}
	var ids []uint
	for _, group_user := range group_users {

		// only return groups that user is not banned from
		if group_user.GroupRoleID != models.GroupRoleMap["banned"] {
			ids = append(ids, group_user.GroupID)
		}
	}
	db.Where("id IN (?)", ids).Find(&groups)

	// return response
	response := &GetGroupsResponse{
		Success: true,
		Groups:  groups,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get group
// -----------------------------------------------------------------------------
type GetGroupResponse struct {
	Success bool         `json:"success"`
	Group   models.Group `json:"group"`
}

func GetGroup(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get group_id from request
	group_id := c.Params("group_id")

	// get database connection
	db := database.DB.DB

	// verify group_user has access to group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify group_user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this group.")
	}

	// get group
	group := &models.Group{}
	db.Where("id = ?", group_id).Find(group)

	// return response
	response := &GetGroupResponse{
		Success: true,
		Group:   *group,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Update group
// -----------------------------------------------------------------------------
type UpdateGroupRequest struct {
	Name string
}

type UpdateGroupResponse struct {
	Success bool         `json:"success"`
	Group   models.Group `json:"group"`
}

func UpdateGroup(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// get request body
	req := new(UpdateGroupRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group name.")
	}

	// get database connection
	db := database.DB.DB

	// verify group_user has access to group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify group_user is admin or owner
	if !(group_user.GroupRoleID == models.GroupRoleMap["admin"] || group_user.GroupRoleID == models.GroupRoleMap["owner"]) {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to update this group.")
	}

	// verify new group name does not already exist
	query = db.Where("name = ?", req.Name).Find(&models.Group{})
	if query.RowsAffected != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Group name already exists.")
	}

	// get group
	group := &models.Group{
		ID: uint(group_id),
	}
	db.Where("id = ?", group_id).Find(group)

	// update group
	group.Name = req.Name
	db.Model(group).Update("name", req.Name)

	// return response
	response := &UpdateGroupResponse{
		Success: true,
		Group:   *group,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Delete group
// -----------------------------------------------------------------------------
type DeleteGroupResponse struct {
	Success bool `json:"success"`
}

func DeleteGroup(c *fiber.Ctx) error {
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

	// get database connection
	db := database.DB.DB

	// verify group_user has access to group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify group_user is owner
	if !(group_user.GroupRoleID == models.GroupRoleMap["owner"]) {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to delete this group.")
	}

	// delete group
	db.Delete(&models.Group{}, group_id)

	// return response
	response := &DeleteGroupResponse{
		Success: true,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Create group user
// -----------------------------------------------------------------------------
type CreateGroupUserRequest struct {
	UserID uint `json:"user_id"`
}

type CreateGroupUserResponse struct {
	Success   bool                 `json:"success"`
	GroupUser models.GroupUserInfo `json:"group_user"`
}

func CreateGroupUser(c *fiber.Ctx) error {
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

	// get request body
	req := &CreateGroupUserRequest{}
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.UserID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user id.")
	}

	// get database connection
	db := database.DB.DB

	// verify requester is already in group
	requester := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(requester)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify requester is admin or owner
	if !(requester.GroupRoleID == models.GroupRoleMap["admin"] || requester.GroupRoleID == models.GroupRoleMap["owner"]) {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to create users in this group.")
	}

	// add user to group. also add user to rooms in group.
	err = queries.AddUserToGroupAndRooms(db, req.UserID, uint(group_id))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error adding user to group.")
	}

	// get group_user_info
	group_user_info, err := queries.GetGroupUserInfo(db, uint(group_id), req.UserID)
	if err != nil {
		return err
	}

	// return response
	response := &CreateGroupUserResponse{
		Success:   true,
		GroupUser: *group_user_info,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Get group users
// -----------------------------------------------------------------------------
type GetGroupUsersResponse struct {
	Success    bool                   `json:"success"`
	GroupUsers []models.GroupUserInfo `json:"group_users"`
}

func GetGroupUsers(c *fiber.Ctx) error {
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

	// get database connection
	db := database.DB.DB

	// verify group_user has access to group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify group_user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this group.")
	}

	// get group users
	group_users_info, err := queries.GetGroupUsersInfo(db, uint(group_id))
	if err != nil {
		return err
	}

	// return response
	response := &GetGroupUsersResponse{
		Success:    true,
		GroupUsers: group_users_info,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get group user
// -----------------------------------------------------------------------------
type GetGroupUserResponse struct {
	Success   bool                 `json:"success"`
	GroupUser models.GroupUserInfo `json:"group_user"`
}

func GetGroupUser(c *fiber.Ctx) error {
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

	// get user_id from request
	user_id_str := c.Params("user_id")
	user_id, err := strconv.ParseUint(user_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user id.")
	}

	// get database connection
	db := database.DB.DB

	// verify group_user is authorized to make request
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify group_user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this group.")
	}

	// get group user
	group_user_info, err := queries.GetGroupUserInfo(db, uint(group_id), uint(user_id))
	if err != nil {
		return err
	}

	// return response
	response := &GetGroupUserResponse{
		Success:   true,
		GroupUser: *group_user_info,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Update group user
// -----------------------------------------------------------------------------
type UpdateGroupUserRequest struct {
	GroupRoleID uint `json:"group_role_id"`
}

type UpdateGroupUserResponse struct {
	Success   bool                 `json:"success"`
	GroupUser models.GroupUserInfo `json:"group_user"`
}

func UpdateGroupUser(c *fiber.Ctx) error {
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

	// get user_id from request
	user_id_str := c.Params("user_id")
	user_id, err := strconv.ParseUint(user_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user id.")
	}

	// verify id is not requesting themselves
	if id == uint(user_id) {
		return fiber.NewError(fiber.StatusBadRequest, "You cannot change your own role.")
	}

	// get request body
	req := &UpdateGroupUserRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	// check group_role_id is in valid range
	if req.GroupRoleID < 1 || req.GroupRoleID > 6 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group_role_id.")
	}

	// get database connection
	db := database.DB.DB

	// verify group_user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// get target_group_user
	target_group_user := &models.GroupUser{}
	query = db.Where("user_id = ? AND group_id = ?", user_id, group_id).Find(target_group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "user_id not found in group.")
	}

	// logic to determine if requester_role_id can change current_role_id to target_role_id
	requester_role_id := group_user.GroupRoleID
	current_role_id := target_group_user.GroupRoleID
	target_role_id := req.GroupRoleID

	// can change role if...
	// you outrank or are an owner, AND
	// you are moderator or higher, AND
	// you outrank the new role or are an owner.
	condition1 := requester_role_id == models.GroupRoleMap["owner"] || (requester_role_id > current_role_id)
	condition2 := requester_role_id >= models.GroupRoleMap["moderator"]
	condition3 := requester_role_id == models.GroupRoleMap["owner"] || (requester_role_id > target_role_id)
	if condition1 && condition2 && condition3 {
		target_group_user.GroupRoleID = target_role_id
	} else {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to change this user's role.")
	}

	// update target_group_user
	db.Model(target_group_user).Update("group_role_id", target_group_user.GroupRoleID)

	// read back
	group_user_info, err := queries.GetGroupUserInfo(db, uint(group_id), uint(user_id))
	if err != nil {
		return err
	}

	// return response
	response := &UpdateGroupUserResponse{
		Success:   true,
		GroupUser: *group_user_info,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Delete group user
// -----------------------------------------------------------------------------
type DeleteGroupUserResponse struct {
	Success bool `json:"success"`
}

func DeleteGroupUser(c *fiber.Ctx) error {
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

	// get user_id from request
	user_id_str := c.Params("user_id")
	user_id, err := strconv.ParseUint(user_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user id.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify target user is in group
	target_group_user := &models.GroupUser{}
	query = db.Where("user_id = ? AND group_id = ?", user_id, group_id).Find(target_group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "user_id not found in group.")
	}

	// logic to check if group_user can delete target_group_user
	// rules...
	// to delete yourself:
	// - you must not be an owner OR
	// - there is another owner in the group
	// to delete another user:
	// - you must be admin or owner AND
	// - you must outrank the target_group_user OR be an owner
	can_delete := false
	if id == uint(user_id) {
		if group_user.GroupRoleID != models.GroupRoleMap["owner"] {
			can_delete = true
		} else {
			// check if there is another owner in the group
			role_count, err := queries.GetGroupUserRoleCount(db, uint(group_id), models.GroupRoleMap["owner"])
			if err != nil {
				return err
			} else if role_count.Count > 1 {
				can_delete = true
			} else {
				return fiber.NewError(fiber.StatusUnauthorized, "You must promote another owner in the group before you can remove yourself.")
			}
		}
	} else {
		requester_role_id := group_user.GroupRoleID
		target_role_id := target_group_user.GroupRoleID
		condition1 := requester_role_id >= models.GroupRoleMap["admin"]
		condition2 := requester_role_id == models.GroupRoleMap["owner"] || (requester_role_id > target_role_id)
		if condition1 && condition2 {
			can_delete = true
		}
	}

	if !can_delete {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to delete this user.")
	}

	// delete user
	tx := db.Delete(target_group_user)
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error deleting user from group.")
	}

	// also delete room_user entries for user_id in group_id
	rooms := []models.Room{}
	query = db.Where("group_id = ?", group_id).Find(&rooms)
	if query.RowsAffected != 0 {
		room_users := []models.RoomUser{}
		for _, room := range rooms {
			room_user := models.RoomUser{
				RoomID: room.ID,
				UserID: uint(user_id),
			}
			room_users = append(room_users, room_user)
		}
		tx = db.Delete(&room_users)
		if tx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error deleting room_user entries for user_id in group_id.")
		}
	}

	// return response
	response := &DeleteGroupUserResponse{
		Success: true,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Create group invite
// -----------------------------------------------------------------------------
type CreateGroupInviteResponse struct {
	Success     bool               `json:"success"`
	GroupInvite models.GroupInvite `json:"group_invite"`
}

func CreateGroupInvite(c *fiber.Ctx) error {
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

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify user is admin or owner
	if group_user.GroupRoleID < models.GroupRoleMap["admin"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to invite users to this group.")
	}

	// generate 16-digit random numeric code
	rand.Seed(time.Now().UnixNano())
	var code string
	for i := 0; i < 16; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}

	// create group invite
	group_invite := &models.GroupInvite{}
	group_invite.ExpiresAt = time.Now().Add(time.Hour * 730)
	group_invite.Code = code
	group_invite.GroupID = uint(group_id)
	group_invite.InviteStatusID = models.InviteStatusMap["pending"]
	group_invite.ReferrerID = id
	db.Create(group_invite)

	// return response
	response := &CreateGroupInviteResponse{
		Success:     true,
		GroupInvite: *group_invite,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get group invites
// -----------------------------------------------------------------------------
type GetGroupInvitesRequest struct {
	InviteStatusID uint `json:"invite_status_id,omitempty"`
}

type GetGroupInvitesResponse struct {
	Success bool                 `json:"success"`
	Invites []models.GroupInvite `json:"invites"`
}

func GetGroupInvites(c *fiber.Ctx) error {
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

	// get group_role_id from request
	req := &GetGroupInvitesRequest{}
	err = c.BodyParser(req)
	if err != nil {
		req.InviteStatusID = 0
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's invites.")
	}

	// verify user is admin or owner
	if group_user.GroupRoleID < models.GroupRoleMap["admin"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's invites.")
	}

	// get group invites
	group_invites := []models.GroupInvite{}
	if req.InviteStatusID == 0 {
		query = db.Where("group_id = ? and referrer_id = ?", group_id, id).Find(&group_invites)
	} else {
		query = db.Where("group_id = ? and referrer_id = ? and invite_status_id = ?", group_id, id, req.InviteStatusID).Find(&group_invites)
	}
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "No invites found.")
	}

	// return response
	response := &GetGroupInvitesResponse{
		Success: true,
		Invites: group_invites,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get group invite
// -----------------------------------------------------------------------------
type GetGroupInviteResponse struct {
	Success bool               `json:"success"`
	Invite  models.GroupInvite `json:"invite"`
}

func GetGroupInvite(c *fiber.Ctx) error {
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

	// get invite_id from request
	invite_id_str := c.Params("invite_id")
	invite_id, err := strconv.ParseUint(invite_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid invite id.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's invites.")
	}

	// verify user is admin or owner
	if group_user.GroupRoleID < models.GroupRoleMap["admin"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's invites.")
	}

	// get group invite
	group_invite := &models.GroupInvite{}
	query = db.Where("group_id = ? and id = ?", group_id, invite_id).Find(group_invite)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "No invite found.")
	}

	// return response
	response := &GetGroupInviteResponse{
		Success: true,
		Invite:  *group_invite,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Delete group invite
// -----------------------------------------------------------------------------
type DeleteGroupInviteResponse struct {
	Success bool `json:"success"`
}

func DeleteGroupInvite(c *fiber.Ctx) error {
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

	// get invite_id from request
	invite_id_str := c.Params("invite_id")
	invite_id, err := strconv.ParseUint(invite_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid invite id.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's invites.")
	}

	// verify user is admin or owner
	if group_user.GroupRoleID < models.GroupRoleMap["admin"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's invites.")
	}

	// get group invite
	group_invite := &models.GroupInvite{}
	query = db.Where("group_id = ? and id = ?", group_id, invite_id).Find(group_invite)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "No invite found.")
	}

	// delete group invite
	query = db.Delete(group_invite)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusInternalServerError, "Error deleting invite.")
	}

	// return response
	response := &DeleteGroupInviteResponse{
		Success: true,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Create room
// -----------------------------------------------------------------------------
type CreateRoomRequest struct {
	Name       string `json:"name"`
	Capacity   uint   `json:"capacity"`
	RoomTypeID uint   `json:"room_type_id"` // fk: RoomType.ID
}

type CreateRoomResponse struct {
	Success     bool         `json:"success"`
	Room        models.Room  `json:"room"`
	LiveKitRoom livekit.Room `json:"livekit_room"`
}

func CreateRoom(c *fiber.Ctx) error {

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

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify user is admin or owner
	if group_user.GroupRoleID < models.GroupRoleMap["admin"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// get request body
	req := &CreateRoomRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	// validate request body
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid room name.")
	}
	if !(req.Capacity == 8 || req.Capacity == 16) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid room capacity.")
	}
	if !(1 <= req.RoomTypeID && req.RoomTypeID <= 3) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid room type.")
	}

	// check if room name already exists in group
	query = db.Where("group_id = ? and name = ?", group_id, req.Name).Find(&models.Room{})
	if query.RowsAffected != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Room name already exists in group.")
	}

	// create room
	room := &models.Room{
		Name:              req.Name,
		GroupID:           uint(group_id),
		Capacity:          req.Capacity,
		RoomTypeID:        req.RoomTypeID,
		DeploymentZoneID:  models.DeploymentZoneMap["us-west1-b"],
		DeprecationCodeID: models.DeprecationCodeMap["active"],
	}
	query = db.Create(room)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating room.")
	}

	// Requester is room owner
	room_owner := &models.RoomUser{
		RoomID:     room.ID,
		UserID:     id,
		RoomRoleID: models.RoomRoleMap["owner"],
		CanJoin:    true,
		CanSee:     true,
	}

	// add the owner to the room.
	// add the other group members that aren't banned as members.
	// set can_join and can_see appropriately for secret and private rooms.
	group_users, err := queries.GetGroupUsersInfo(db, uint(group_id))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error getting group users.")
	}
	room_users := []models.RoomUser{}
	can_see := room.RoomTypeID != models.RoomTypeMap["secret"]
	can_join := room.RoomTypeID == models.RoomTypeMap["public"]
	for _, group_user := range group_users {
		if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
			continue
		}
		if group_user.UserID == id {
			room_users = append(room_users, *room_owner)
		} else {
			room_user := &models.RoomUser{
				RoomID:     room.ID,
				UserID:     group_user.UserID,
				RoomRoleID: group_user.GroupRoleID,
				CanJoin:    can_join,
				CanSee:     can_see,
			}
			room_users = append(room_users, *room_user)
		}

	}
	if len(room_users) > 0 {
		query = db.Create(&room_users)
		if query.RowsAffected == 0 {
			return fiber.NewError(fiber.StatusInternalServerError, "Error creating room users.")
		}
	}

	// create livekit room
	client := CreateRoomServiceClient()
	lkroom, err := client.CreateRoom(c.Context(), &livekit.CreateRoomRequest{
		Name:            EncodeRoomName(room.GroupID, room.ID),
		MaxParticipants: uint32(room.Capacity),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating livekit room.")
	}

	// return response
	response := &CreateRoomResponse{
		Success:     true,
		Room:        *room,
		LiveKitRoom: *lkroom,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get rooms
// -----------------------------------------------------------------------------
type GetRoomsResponse struct {
	Success bool          `json:"success"`
	Rooms   []models.Room `json:"rooms"`
}

func GetRooms(c *fiber.Ctx) error {
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

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify group_user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// get rooms
	rooms, err := queries.GetRoomsNotBanned(db, uint(group_id), id)
	if err != nil {
		return err
	}

	// return response
	response := &GetRoomsResponse{
		Success: true,
		Rooms:   rooms,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get room
// -----------------------------------------------------------------------------
type GetRoomResponse struct {
	Success bool        `json:"success"`
	Room    models.Room `json:"room"`
}

func GetRoom(c *fiber.Ctx) error {
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
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify group_user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify user is in room
	room_user := &models.RoomUser{}
	query = db.Where("user_id = ? AND room_id = ?", id, room_id).Find(room_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this room.")
	}

	// verify room_user is not banned
	if room_user.RoomRoleID == models.RoomRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this room.")
	}

	// get room
	room := &models.Room{}
	query = db.Where("id = ?", room_id).Find(room)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Room not found.")
	}

	// return response
	response := &GetRoomResponse{
		Success: true,
		Room:    *room,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Delete room
// -----------------------------------------------------------------------------
type DeleteRoomResponse struct {
	Success bool `json:"success"`
}

func DeleteRoom(c *fiber.Ctx) error {
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
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify group_user is admin or owner
	if group_user.GroupRoleID != models.GroupRoleMap["admin"] && group_user.GroupRoleID != models.GroupRoleMap["owner"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to delete this room.")
	}

	// verify user is in room
	room_user := &models.RoomUser{}
	query = db.Where("user_id = ? AND room_id = ?", id, room_id).Find(room_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to delete this room.")
	}

	// delete room
	tx := db.Delete(&models.Room{}, room_id)
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error deleting room.")
	}

	// get roomservice client
	client := CreateRoomServiceClient()

	// delete livekit room
	client.DeleteRoom(c.Context(), &livekit.DeleteRoomRequest{
		Room: EncodeRoomName(uint(group_id), uint(room_id)),
	})

	// return response
	response := &DeleteRoomResponse{
		Success: true,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Update room
// -----------------------------------------------------------------------------
type UpdateRoomRequest struct {
	Name string `json:"name"`
}

type UpdateRoomResponse struct {
	Success bool        `json:"success"`
	Room    models.Room `json:"room"`
}

func UpdateRoom(c *fiber.Ctx) error {
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

	// get request body
	req := new(UpdateRoomRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid room name.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify group_user is admin or owner
	if group_user.GroupRoleID != models.GroupRoleMap["admin"] && group_user.GroupRoleID != models.GroupRoleMap["owner"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to update this room.")
	}

	// get room
	room := &models.Room{}
	query = db.Where("id = ?", room_id).Find(room)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Room not found.")
	}

	// verify room is in group
	if room.GroupID != uint(group_id) {
		return fiber.NewError(fiber.StatusUnauthorized, "Room not found.")
	}

	// update room
	tx := db.Model(room).Update("name", req.Name)
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error updating room.")
	}

	// return response
	response := &UpdateRoomResponse{
		Success: true,
		Room:    *room,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get room users
// -----------------------------------------------------------------------------
type GetRoomUsersResponse struct {
	Success   bool                  `json:"success"`
	RoomUsers []models.RoomUserInfo `json:"room_users"`
}

func GetRoomUsers(c *fiber.Ctx) error {
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
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify group_user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this group.")
	}

	// get room users' info
	room_users_info, err := queries.GetRoomUsersInfo(db, uint(room_id))
	if err != nil {
		return err
	}

	// return response
	response := &GetRoomUsersResponse{
		Success:   true,
		RoomUsers: room_users_info,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Get room user
// -----------------------------------------------------------------------------
type GetRoomUserResponse struct {
	Success  bool                `json:"success"`
	RoomUser models.RoomUserInfo `json:"room_user"`
}

func GetRoomUser(c *fiber.Ctx) error {
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

	// get user_id from request
	user_id_str := c.Params("user_id")
	user_id, err := strconv.ParseUint(user_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user id.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify group_user is not banned
	if group_user.GroupRoleID == models.GroupRoleMap["banned"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You are banned from this group.")
	}

	// get room user's info
	room_user_info, err := queries.GetRoomUserInfo(db, uint(room_id), uint(user_id))
	if err != nil {
		return err
	}

	// return response
	response := &GetRoomUserResponse{
		Success:  true,
		RoomUser: *room_user_info,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Update room user
// -----------------------------------------------------------------------------
type UpdateRoomUserRequest struct {
	RoomRoleID uint `json:"room_role_id"`
	CanSee     bool `json:"can_see"`
	CanJoin    bool `json:"can_join"`
}

type UpdateRoomUserResponse struct {
	Success  bool                `json:"success"`
	RoomUser models.RoomUserInfo `json:"room_user"`
}

func UpdateRoomUser(c *fiber.Ctx) error {
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

	// get user_id from request
	user_id_str := c.Params("user_id")
	user_id, err := strconv.ParseUint(user_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user id.")
	}

	// get request body
	req := &UpdateRoomUserRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	// validate request body
	if req.RoomRoleID < 1 || req.RoomRoleID > 6 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid room_role_id.")
	}

	// get database connection
	db := database.DB.DB

	// verify user is in group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group's rooms.")
	}

	// verify user has moderator, admin, or owner role in room
	room_user := &models.RoomUser{}
	query = db.Where("user_id = ? AND room_id = ?", id, room_id).Find(room_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this room.")
	}
	if room_user.RoomRoleID < 4 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to update this room user.")
	}

	// get target room user
	target_room_user := &models.RoomUser{}
	query = db.Where("user_id = ? AND room_id = ?", user_id, room_id).Find(target_room_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Room user not found.")
	}

	// can only update target room user if you outrank them or are room owner
	if target_room_user.RoomRoleID >= room_user.RoomRoleID && room_user.RoomRoleID != models.RoomRoleMap["owner"] {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to update this room user.")
	}

	// cannot update target room user role beyond your own role
	if req.RoomRoleID > room_user.RoomRoleID {
		return fiber.NewError(fiber.StatusUnauthorized, "You cannot promote user_id beyond your own role.")
	}

	// cannot demote yourself
	if req.RoomRoleID < room_user.RoomRoleID && id == uint(user_id) {
		return fiber.NewError(fiber.StatusUnauthorized, "You cannot demote yourself.")
	}

	// update room user
	tx := db.Model(target_room_user).Updates(models.RoomUser{
		RoomRoleID: req.RoomRoleID,
		CanSee:     req.CanSee,
		CanJoin:    req.CanJoin,
	})
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error updating room user.")
	}

	// get room user info
	room_user_info, err := queries.GetRoomUserInfo(db, uint(room_id), uint(user_id))
	if err != nil {
		return err
	}

	// return response
	response := &UpdateRoomUserResponse{
		Success:  true,
		RoomUser: *room_user_info,
	}
	return c.JSON(response)
}

// -----------------------------------------------------------------------------
// Accept group invite
// -----------------------------------------------------------------------------
type AcceptGroupInviteRequest struct {
	InviteCode string `json:"invite_code"`
}

type AcceptGroupInviteResponse struct {
	Success       bool                 `json:"success"`
	GroupUserInfo models.GroupUserInfo `json:"group_user_info"`
	Group         models.Group         `json:"group"`
}

func AcceptGroupInvite(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// validate request body
	req := &AcceptGroupInviteRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body.")
	}

	// get database connection
	db := database.DB.DB

	// validate invite code
	group_invite, err := queries.GetGroupInviteCode(db, req.InviteCode)
	if err != nil {
		return err
	}

	// add user to group and rooms
	err = queries.AddUserToGroupAndRooms(db, id, group_invite.GroupID)
	if err != nil {
		return err
	}

	// accept invite, create referral
	err = queries.AcceptInviteAndCreateReferral(db, group_invite, id)
	if err != nil {
		return err
	}

	// get group
	group := &models.Group{}
	query := db.Where("id = ?", group_invite.GroupID).Find(group)
	if query.Error != nil {
		return query.Error
	}

	// get group_user_info
	group_user_info, err := queries.GetGroupUserInfo(db, group_invite.GroupID, id)
	if err != nil {
		return err
	}

	// return response
	response := &AcceptGroupInviteResponse{
		Success:       true,
		GroupUserInfo: *group_user_info,
		Group:         *group,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Get World
// -----------------------------------------------------------------------------

type GetWorldResponse struct {
	Success    bool               `json:"success"`
	GroupsInfo []models.GroupInfo `json:"groups_info"`
}

func GetWorld(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, err := getIDFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT token.")
	}

	// get database connection
	db := database.DB.DB

	// get group_users for user
	group_users := []models.GroupUser{}
	db.Where("user_id = ?", id).Find(&group_users)

	// get group for each group_id in group_users for user
	groups := []models.Group{}
	var ids []uint
	for _, group_user := range group_users {

		// only return groups that user is not banned from
		if group_user.GroupRoleID != models.GroupRoleMap["banned"] {
			ids = append(ids, group_user.GroupID)
		}
	}
	db.Where("id IN (?)", ids).Find(&groups)

	// for each group...
	// - get rooms. for each room...
	// -- get room_user
	// -- get room_users_info
	// -- get token
	// -- assemble room_info
	// - get group_users_info
	// - assemble group_info

	groups_info := []models.GroupInfo{}
	for _, group := range groups {
		// get rooms
		rooms, err := queries.GetRoomsNotBanned(db, group.ID, id)
		if err != nil {
			return err
		}

		rooms_info := []models.RoomInfo{}
		for _, room := range rooms {
			room_info := &models.RoomInfo{}

			// get room_user
			room_user, err := queries.GetRoomUser(db, room.ID, id)
			if err != nil {
				return err
			}

			// get room_users_info
			room_users_info, err := queries.GetRoomUsersInfo(db, room.ID)
			if err != nil {
				return err
			}

			// get token
			token, err := createLiveKitJoinToken(room_user, group.ID, room.ID, id)
			if err != nil {
				return err
			}

			// assemble room_info
			room_info.Room = room
			room_info.Users = room_users_info
			room_info.Token = token

			// add room_info to rooms_info
			rooms_info = append(rooms_info, *room_info)

		}

		// get group_users_info
		group_users_info, err := queries.GetGroupUsersInfo(db, group.ID)
		if err != nil {
			return err
		}

		// assemble group_info
		group_info := &models.GroupInfo{
			Group: group,
			Users: group_users_info,
			Rooms: rooms_info,
		}

		// add group_info to groups_info
		groups_info = append(groups_info, *group_info)

	}

	// return response
	response := &GetWorldResponse{
		Success:    true,
		GroupsInfo: groups_info,
	}
	return c.JSON(response)
}

// TODO:
// - when a group user is banned,
// -- update their room_user roles and can_join/can_see,
// -- and drop them from any livekit rooms
// - when a group is deleted, delete the rooms and livekit rooms
