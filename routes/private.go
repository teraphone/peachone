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

// Create a new group
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

	// create group and group user db entries
	group := new(models.Group)
	group.Name = req.Name
	db.Create(group)
	group_user := new(models.GroupUser)
	group_user.GroupID = group.ID
	group_user.UserID = id
	group_user.RoleID = models.GroupRoleMap["owner"]
	db.Create(group_user)

	// return response
	return c.JSON(fiber.Map{"success": true, "group": group, "group_user": group_user})

	// TODO: this response is pretty ugly. probably need to simplify models.
}
