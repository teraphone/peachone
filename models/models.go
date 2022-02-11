package models

import (
	"time"
)

type Group struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
}

type Room struct {
	ID                uint      `gorm:"primary_key" json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Name              string    `json:"name"`
	GroupID           uint      `json:"group_id"` // fk: Group.ID
	Capacity          uint      `json:"capacity"`
	RoomTypeID        uint      `json:"room_type_id"`        // fk: RoomType.ID
	DeploymentZoneID  uint      `json:"deployment_zone_id"`  // fk: DeploymentZone.ID
	DeprecationCodeID uint      `json:"deprecation_code_id"` // fk: DeprecationCode.ID
}

type User struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
}

type GroupUser struct {
	ID          uint      `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	GroupID     uint      `json:"group_id"`      // fk: Group.ID
	UserID      uint      `json:"user_id"`       // fk: User.ID
	GroupRoleID uint      `json:"group_role_id"` // fk: GroupRole.ID
}

type RoomUser struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	RoomID     uint      `json:"room_id"`      // fk: Room.ID
	UserID     uint      `json:"user_id"`      // fk: User.ID
	RoomRoleID uint      `json:"room_role_id"` // fk: RoomRole.ID
	CanJoin    bool      `json:"can_join"`
	CanSee     bool      `json:"can_see"`
}

type GroupInvite struct {
	ID             uint      `gorm:"primary_key" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Code           string    `json:"code"`
	GroupID        uint      `json:"group_id"`         // fk: Group.ID
	InviteStatusID uint      `json:"invite_status_id"` // fk: InviteStatus.ID
	ReferrerID     uint      `json:"referrer_id"`      // fk: User.ID
}

type Referral struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UserID     uint      `json:"user_id"`     // fk: User.ID
	ReferrerID uint      `json:"referrer_id"` // fk: User.ID
}
