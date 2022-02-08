package models

import (
	"gorm.io/gorm"
)

var RoomTypeMap = map[string]uint{
	"public":  1,
	"private": 2,
	"secret":  3,
}

type RoomType struct {
	gorm.Model
	Type string // must initialize: "public", "private", "secret"
}

var DeploymentZoneMap = map[string]uint{
	"us-west1-b": 1,
}

type DeploymentZone struct {
	gorm.Model
	Zone string // must initialize: "us-west1-b"
}

var DeprecationCodeMap = map[string]uint{
	"active":   1,
	"inactive": 2,
}

type DeprecationCode struct {
	gorm.Model
	Code string // must initialize: "active", "inactive"
}

var RoomRoleMap = map[string]uint{
	"base":      1,
	"moderator": 2,
	"admin":     3,
	"owner":     4,
	"guest":     5,
	"banned":    6,
}

type RoomRole struct {
	gorm.Model
	Role string // must initialize: "base", "moderator", "admin", "owner", "guest", "banned"
}

var GroupRoleMap = map[string]uint{
	"base":      1,
	"moderator": 2,
	"admin":     3,
	"owner":     4,
	"guest":     5,
	"banned":    6,
}

type GroupRole struct {
	gorm.Model
	Role string // must initialize: "base", "moderator", "admin", "owner", "guest", "banned"
}

var InviteStatusMap = map[string]uint{
	"pending":  1,
	"accepted": 2,
	"expired":  3,
}

type InviteStatus struct {
	gorm.Model
	Status string // must initialize: "pending", "accepted", "expired"
}
