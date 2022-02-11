package routes

import (
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

	// create group
	group := &models.Group{Name: req.Name}
	db.Create(group)
	group_user := &models.GroupUser{
		GroupID:     group.ID,
		UserID:      id,
		GroupRoleID: models.GroupRoleMap["owner"],
	}
	db.Create(group_user)

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
