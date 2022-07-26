package routes

import (
	"fmt"
	"peachone/database"
	"peachone/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Private Welcome handler
func PrivateWelcome(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"success": true, "path": "private"})
}

// --------------------------------------------------------------------------------
// Update License request handler
// --------------------------------------------------------------------------------
type UpdateLicenseResponse struct {
	Success bool               `json:"success"`
	License models.UserLicense `json:"license"`
}

func UpdateLicense(c *fiber.Ctx) error {
	// extract claims from JWT
	claims, err := getClaimsFromJWT(c)
	if err != nil {
		fmt.Println("error extracting claims from JWT:", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Expired JWT.")
	}

	// get database connection
	db := database.DB.DB

	// get license
	license := &models.UserLicense{
		Oid: claims.Oid,
	}
	query := db.Where("oid = ?", license.Oid).Find(license)
	if query.RowsAffected == 0 {
		fmt.Println("license not found for user:", license.Oid)
		return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
	}

	// update license
	if !license.TrialActivated {
		tx := db.Model(license).Updates(models.UserLicense{
			TrialActivated: true,
			TrialExpiresAt: time.Now().Add(time.Hour * 24 & 30),
		})
		if tx.Error != nil {
			fmt.Println("error updating license:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "Error processing request.")
		}
	}

	// return response
	response := &UpdateLicenseResponse{
		Success: true,
		License: *license,
	}
	return c.JSON(response)

}
