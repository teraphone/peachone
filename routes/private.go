package routes

import (
	"peachone/database"
	"peachone/models"

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

func _CreateGroup(userid uint, group_name string) (*CreateGroupResponse, error) {
	// create database connection
	db, err := database.CreateDBConnection()
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Error connecting to database.")
	}

	group := &models.Group{Name: group_name}
	db.Create(group)
	group_user := &models.GroupUser{
		GroupID:     group.ID,
		UserID:      userid,
		GroupRoleID: models.GroupRoleMap["owner"],
	}
	db.Create(group_user)

	return &CreateGroupResponse{
		Success:   true,
		Group:     *group,
		GroupUser: *group_user,
	}, nil
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

	// create group and group user db entries
	response, err := _CreateGroup(id, req.Name)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating group.")
	}

	// return response
	return c.JSON(response)

}
