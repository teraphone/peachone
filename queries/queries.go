package queries

import (
	"peachone/models"

	"github.com/gofiber/fiber/v2"
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
