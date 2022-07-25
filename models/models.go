package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type TenantUser struct {
	Oid       string    `gorm:"primary_key" json:"oid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Tid       string    `json:"tid"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserLicense struct {
	Oid                string        `gorm:"primary_key" json:"oid"` // fk: TenantUser.Oid
	Tid                string        `json:"tid"`
	LicenseExpiresAt   time.Time     `json:"licenseExpiresAt"`
	LicenseStatus      LicenseStatus `json:"licenseStatus"`
	LicensePlan        LicensePlan   `json:"licensePlan"`
	LicenseAutoRenew   bool          `json:"licenseAutoRenew"`
	LicenseRequested   bool          `json:"licenseRequested"`
	LicenseRequestedAt time.Time     `json:"licenseRequestedAt"`
	TrialActivated     bool          `json:"trialActivated"`
	TrialExpiresAt     time.Time     `json:"trialExpiresAt"`
	CreatedAt          time.Time     `json:"createdAt"`
	UpdatedAt          time.Time     `json:"updatedAt"`
}

type TenantTeam struct {
	Id          string    `gorm:"primary_key" json:"id"`
	Tid         string    `json:"tid"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TeamUser struct {
	Id  string `gorm:"primary_key" json:"id"`  // fk: TenantTeam.Id
	Oid string `gorm:"primary_key" json:"oid"` // fk: TenantUser.Oid
}

type TeamRoom struct {
	Id             uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	TeamId         string         `json:"teamId"` // fk: TenantTeam.Id
	DisplayName    string         `json:"displayName"`
	Description    string         `json:"description"`
	Capacity       int            `json:"capacity"`
	DeploymentZone DeploymentZone `json:"deploymentZone"`
	RoomType       RoomType       `json:"roomType"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

type RoomInfo struct {
	Room      TeamRoom `json:"room"`
	RoomToken string   `json:"roomToken"`
}

type TeamInfo struct {
	Team  TenantTeam   `json:"team"`
	Rooms []RoomInfo   `json:"rooms"`
	Users []TenantUser `json:"users"`
}
