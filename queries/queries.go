package queries

import (
	"errors"
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
