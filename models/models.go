package models

import (
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	Name  string
	Rooms []Room
	Users []User `gorm:"many2many:group_users;"`
}

type Room struct {
	gorm.Model
	Name              string
	GroupID           uint
	Group             Group
	Capacity          uint
	RoomTypeID        uint
	RoomType          RoomType
	DeploymentZoneID  uint
	DeploymentZone    DeploymentZone
	DeprecationCodeID uint
	DeprecationCode   DeprecationCode
	Users             []User `gorm:"many2many:room_users;"`
}

type User struct {
	gorm.Model
	Name       string
	Email      string
	Password   string `json:"-"`
	ReferrerID *uint
	Referrer   *User
}

type GroupUser struct {
	GroupID uint `gorm:"primary_key"`
	UserID  uint `gorm:"primary_key"`
	RoleID  uint
	Role    GroupRole
}

type RoomUser struct {
	RoomID  uint `gorm:"primary_key"`
	UserID  uint `gorm:"primary_key"`
	RoleID  uint
	Role    RoomRole
	CanJoin bool
	CanSee  bool
}

type GroupInvite struct {
	ID         uint `gorm:"primary_key"`
	Code       string
	GroupID    uint
	Group      Group
	Expiration int64
	StatusID   uint
	Status     InviteStatus
	ReferrerID uint
	Referrer   User
	RoomID     uint
	Room       Room // optional
}
