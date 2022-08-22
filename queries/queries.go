package queries

import (
	"errors"
	"fmt"
	"peachone/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

func AddUserToTeam(db *gorm.DB, userId string, teamId string) error {
	// verify user exists
	user := &models.TenantUser{}
	query := db.Where("oid = ?", userId).Find(user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "userId not found.")
	}

	// verify new user is not already in team
	newTeamUser := &models.TeamUser{
		Id:  teamId,
		Oid: userId,
	}
	query = db.Where("id = ? AND oid = ?", teamId, userId).Find(newTeamUser)
	if query.RowsAffected != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "userId is already in this team.")
	}

	// create new team user
	tx := db.Create(newTeamUser)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func SetUpNewUser(db *gorm.DB, user *models.TenantUser) error {
	// make sure user isn't empty
	if user.Oid == "" || user.Tid == "" {
		return errors.New("missing fields in user")
	}

	// create user
	tx := db.Create(user)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

type DefaultRoomConfig struct {
	DisplayName    string                `json:"name"`
	Description    string                `json:"description"`
	Capacity       int                   `json:"capacity"`
	DeploymentZone models.DeploymentZone `json:"deploymentZone"`
	RoomType       models.RoomType       `json:"roomType"`
}

const DefaultRoomCapacity = 16

var DefaultRoomConfigs = []DefaultRoomConfig{
	{
		DisplayName:    "Hangout",
		Description:    "Just chatting",
		Capacity:       DefaultRoomCapacity,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Co-Work",
		Description:    "Working together",
		Capacity:       DefaultRoomCapacity,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Meeting Room Apple",
		Description:    "Inventing the future",
		Capacity:       DefaultRoomCapacity,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Meeting Room Banana",
		Description:    "Solving hard problems",
		Capacity:       DefaultRoomCapacity,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Office Hours",
		Description:    "Helping each other",
		Capacity:       DefaultRoomCapacity,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
}

func SetUpNewTeamAndRooms(db *gorm.DB, team *models.TenantTeam) error {
	// make sure team isn't empty
	if team.Id == "" || team.Tid == "" {
		return errors.New("missing fields in team")
	}

	// create team
	tx := db.Create(team)
	if tx.Error != nil {
		return tx.Error
	}

	// create rooms
	for _, roomConfig := range DefaultRoomConfigs {
		room := &models.TeamRoom{
			Id:             uuid.Must(uuid.NewV4()),
			TeamId:         team.Id,
			DisplayName:    roomConfig.DisplayName,
			Description:    roomConfig.Description,
			Capacity:       roomConfig.Capacity,
			DeploymentZone: roomConfig.DeploymentZone,
			RoomType:       roomConfig.RoomType,
		}
		tx = db.Create(room)
		if tx.Error != nil {
			return tx.Error
		}
	}

	return nil
}

func GetUsersForTeam(db *gorm.DB, teamId string) ([]models.TenantUser, error) {
	sql_fmt := "SELECT tenant_users.* " +
		"FROM tenant_users " +
		"JOIN team_users ON tenant_users.oid = team_users.oid " +
		"WHERE team_users.id = '%s' "
	sql := fmt.Sprintf(sql_fmt, teamId)
	users := []models.TenantUser{}
	query := db.Raw(sql).Scan(&users)
	if query.RowsAffected == 0 {
		return nil, errors.New("could not find users for team")
	}

	return users, nil
}
