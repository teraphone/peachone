package models

import (
	"time"
)

type Group struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
}

type Room struct {
	ID                uint `gorm:"primary_key"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Name              string
	GroupID           uint // fk: Group.ID
	Capacity          uint
	RoomTypeID        uint // fk: RoomType.ID
	DeploymentZoneID  uint // fk: DeploymentZone.ID
	DeprecationCodeID uint // fk: DeprecationCode.ID
}

type User struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Email     string
	Password  string `json:"-"`
}

type GroupUser struct {
	ID          uint `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	GroupID     uint // fk: Group.ID
	UserID      uint // fk: User.ID
	GroupRoleID uint // fk: GroupRole.ID
}

type RoomUser struct {
	ID         uint `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	RoomID     uint // fk: Room.ID
	UserID     uint // fk: User.ID
	RoomRoleID uint // fk: RoomRole.ID
	CanJoin    bool
	CanSee     bool
}

type GroupInvite struct {
	ID             uint `gorm:"primary_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      time.Time
	Code           string
	GroupID        uint // fk: Group.ID
	InviteStatusID uint // fk: InviteStatus.ID
	ReferrerID     uint // fk: User.ID
	RoomID         uint // fk: Room.ID (optional)
}

type Referral struct {
	ID         uint `gorm:"primary_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UserID     uint // fk: User.ID
	ReferrerID uint // fk: User.ID
}
