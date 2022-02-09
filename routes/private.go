package routes

import (
	"peachone/database"
	"peachone/models"
	"strconv"

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
	Success   bool             `json:"success"`
	Group     models.Group     `json:"group"`
	GroupUser models.GroupUser `json:"group_user"`
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
		Success:   true,
		Group:     *group,
		GroupUser: *group_user,
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
type CreateGroupUserRequest struct {
	UserID uint `json:"user_id"`
}

type CreateGroupUserResponse struct {
	Success   bool             `json:"success"`
	GroupUser models.GroupUser `json:"group_user"`
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
		GroupID:     uint(group_id),
		UserID:      req.UserID,
		GroupRoleID: models.GroupRoleMap["base"],
	}
	query = db.Where("user_id = ? AND group_id = ?", req.UserID, group_id).Find(new_group_user)

	if query.RowsAffected != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "User is already in this group.")
	}

	// create new group user
	db.Create(new_group_user)

	// return response
	response := &CreateGroupUserResponse{
		Success:   true,
		GroupUser: *new_group_user,
	}
	return c.JSON(response)

}

// -----------------------------------------------------------------------------
// Get group users
// -----------------------------------------------------------------------------
type GroupUserInfo struct {
	models.GroupUser
	Name string `json:"name"`
}

type GetGroupUsersResponse struct {
	Success    bool            `json:"success"`
	GroupUsers []GroupUserInfo `json:"group_users"`
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

	// get group users
	group_users := []models.GroupUser{}
	db.Where("group_id = ?", group_id).Find(&group_users)

	// get user_ids from group_users
	user_ids := []uint{}
	for _, group_user := range group_users {
		user_ids = append(user_ids, group_user.UserID)
	}

	// get users
	users := []models.User{}
	db.Where("id IN ?", user_ids).Find(&users)

	// create user_id:user_name map
	user_id_name_map := map[uint]string{}
	for _, user := range users {
		user_id_name_map[user.ID] = user.Name
	}

	// build GroupUserInfo objects
	group_users_info := []GroupUserInfo{}
	for _, group_user := range group_users {
		group_users_info = append(group_users_info, GroupUserInfo{
			GroupUser: group_user,
			Name:      user_id_name_map[group_user.UserID],
		})
	}

	// return response
	response := &GetGroupUsersResponse{
		Success:    true,
		GroupUsers: group_users_info,
	}
	return c.JSON(response)
}
