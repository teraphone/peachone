package models

var RoomTypeMap = map[string]uint{
	"public":  1,
	"private": 2,
	"secret":  3,
}

type RoomType struct {
	ID   uint
	Type string // must initialize: "public", "private", "secret"
}

var DeploymentZoneMap = map[string]uint{
	"us-west1-b": 1,
}

type DeploymentZone struct {
	ID   uint
	Zone string // must initialize: "us-west1-b"
}

var DeprecationCodeMap = map[string]uint{
	"active":   1,
	"inactive": 2,
}

type DeprecationCode struct {
	ID   uint
	Code string // must initialize: "active", "inactive"
}

var RoomRoleMap = map[string]uint{
	"banned":    1,
	"guest":     2,
	"member":    3,
	"moderator": 4,
	"admin":     5,
	"owner":     6,
}

type RoomRole struct {
	ID   uint
	Role string // must initialize: "banned", "guest", "member", "moderator", "admin", "owner"
}

var GroupRoleMap = map[string]uint{
	"banned":    1,
	"guest":     2,
	"member":    3,
	"moderator": 4,
	"admin":     5,
	"owner":     6,
}

type GroupRole struct {
	ID   uint
	Role string // must initialize: "banned", "guest", "member", "moderator", "admin", "owner"
}

var InviteStatusMap = map[string]uint{
	"pending":  1,
	"accepted": 2,
	"expired":  3,
}

type InviteStatus struct {
	ID     uint
	Status string // must initialize: "pending", "accepted", "expired"
}
