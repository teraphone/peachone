package routes

import (
	"math/rand"
	"peachone/database"
	"peachone/models"
	"peachone/queries"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
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
	id, _ := getIDFromJWT(c)

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

	// get group_id from request
	group_id := c.Params("group_id")

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

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

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
type GroupUserInfo struct {
	UserID      uint      `json:"user_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	GroupRoleID uint      `json:"group_role_id"`
}

type CreateGroupUserRequest struct {
	UserID  uint `json:"user_id"`
	IsGuest bool `json:"is_guest"`
}

type CreateGroupUserResponse struct {
	Success   bool                  `json:"success"`
	GroupUser queries.GroupUserInfo `json:"group_user"`
}

func CreateGroupUser(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// get request body
	req := new(CreateGroupUserRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.UserID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user id.")
	}

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

	// verify group_user has access to group
	group_user := &models.GroupUser{}
	query := db.Where("user_id = ? AND group_id = ?", id, group_id).Find(group_user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have access to this group.")
	}

	// verify group_user is admin or owner
	if !(group_user.GroupRoleID == models.GroupRoleMap["admin"] || group_user.GroupRoleID == models.GroupRoleMap["owner"]) {
		return fiber.NewError(fiber.StatusUnauthorized, "You do not have permission to create users in this group.")
	}

	// verify new user is not already in group
	new_group_user := &models.GroupUser{
		GroupID: uint(group_id),
		UserID:  req.UserID,
	}
	query = db.Where("user_id = ? AND group_id = ?", req.UserID, group_id).Find(new_group_user)

	if query.RowsAffected != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "User is already in this group.")
	}

	// create new group user
	if req.IsGuest {
		new_group_user.GroupRoleID = models.GroupRoleMap["guest"]
	} else {
		new_group_user.GroupRoleID = models.GroupRoleMap["member"]
	}
	db.Create(new_group_user)

	// get group_user_info
	group_user_info, err := queries.GetGroupUserInfo(db, uint(group_id), new_group_user.UserID)
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
	Success    bool                    `json:"success"`
	GroupUsers []queries.GroupUserInfo `json:"group_users"`
}

func GetGroupUsers(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	Success   bool                  `json:"success"`
	GroupUser queries.GroupUserInfo `json:"group_user"`
}

func GetGroupUser(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

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

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	Success   bool                  `json:"success"`
	GroupUser queries.GroupUserInfo `json:"group_user"`
}

func UpdateGroupUser(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

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

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

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

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	db.Delete(target_group_user)

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
	id, _ := getIDFromJWT(c)

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

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

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

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

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	id, _ := getIDFromJWT(c)

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

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
	Success bool        `json:"success"`
	Room    models.Room `json:"room"`
}

func CreateRoom(c *fiber.Ctx) error {
	// extract user id from JWT claims
	id, _ := getIDFromJWT(c)

	// get group_id from request
	group_id_str := c.Params("group_id")
	group_id, err := strconv.ParseUint(group_id_str, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid group id.")
	}

	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

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
		if group_user.GroupRoleID != models.GroupRoleMap["banned"] {
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
	}
	if len(room_users) > 0 {
		query = db.Create(&room_users)
		if query.RowsAffected == 0 {
			return fiber.NewError(fiber.StatusInternalServerError, "Error creating room users.")
		}
	}

	// return response
	response := &CreateRoomResponse{
		Success: true,
		Room:    *room,
	}
	return c.JSON(response)
}

// TODO:
// - when a user is added to a group, add them to all public rooms
// - allow non-admin users to add themselves to a group using an invite code
