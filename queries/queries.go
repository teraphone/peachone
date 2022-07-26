package queries

import (
	"errors"
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

func SetUpNewUserAndLicense(db *gorm.DB, user *models.TenantUser, license *models.UserLicense) error {
	// make sure user isn't empty
	if user.Oid == "" || user.Tid == "" {
		return errors.New("missing fields in user")
	}

	// create user
	tx := db.Create(user)
	if tx.Error != nil {
		return tx.Error
	}

	// create license
	license.Oid = user.Oid
	license.Tid = user.Tid
	license.LicenseStatus = models.Inactive
	license.LicensePlan = models.None
	license.LicenseAutoRenew = false
	license.LicenseRequested = false
	license.TrialActivated = false
	tx = db.Create(license)
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

var DefaultRoomConfigs = []DefaultRoomConfig{
	{
		DisplayName:    "Hangout",
		Description:    "Just chatting",
		Capacity:       16,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Co-Work",
		Description:    "Working together",
		Capacity:       16,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Meeting Room Apple",
		Description:    "Inventing the future",
		Capacity:       16,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Meeting Room Banana",
		Description:    "Solving hard problems",
		Capacity:       16,
		DeploymentZone: models.USWest1B,
		RoomType:       models.Public,
	},
	{
		DisplayName:    "Office Hours",
		Description:    "Helping each other",
		Capacity:       16,
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
			TeamId:         team.Tid,
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
