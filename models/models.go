package models

import (
	"time"
)

type Group struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `gorm:"unique" json:"name"`
}

type Room struct {
	ID                uint      `gorm:"unique;autoIncrement" json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Name              string    `gorm:"primary_key" json:"name"`
	GroupID           uint      `gorm:"primary_key" json:"group_id"` // fk: Group.ID
	Capacity          uint      `json:"capacity"`
	RoomTypeID        uint      `json:"room_type_id"`        // fk: RoomType.ID
	DeploymentZoneID  uint      `json:"deployment_zone_id"`  // fk: DeploymentZone.ID
	DeprecationCodeID uint      `json:"deprecation_code_id"` // fk: DeprecationCode.ID
}

type User struct {
	ID         uint      `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Name       string    `json:"name"`
	Email      string    `gorm:"unique" json:"email"`
	Password   string    `json:"-"`
	IsVerified bool      `json:"is_verified"`
}

type EmailVerificationCode struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Code      string    `json:"code"`
	UserID    uint      `json:"user_id"` // fk: User.ID
	Uses      uint      `json:"uses"`
}

type PasswordResetCode struct {
	ID        uint      `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Code      string    `json:"code"`
	UserID    uint      `json:"user_id"` // fk: User.ID
	Uses      uint      `json:"uses"`
}

type GroupUser struct {
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	GroupID     uint      `gorm:"primary_key" json:"group_id"` // fk: Group.ID
	UserID      uint      `gorm:"primary_key" json:"user_id"`  // fk: User.ID
	GroupRoleID uint      `json:"group_role_id"`               // fk: GroupRole.ID
}

type RoomUser struct {
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	RoomID     uint      `gorm:"primary_key" json:"room_id"` // fk: Room.ID
	UserID     uint      `gorm:"primary_key" json:"user_id"` // fk: User.ID
	RoomRoleID uint      `json:"room_role_id"`               // fk: RoomRole.ID
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
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	UserID     uint      `gorm:"primary_key" json:"user_id"`
	ReferrerID uint      `gorm:"primary_key" json:"referrer_id"`
}

type GroupUserInfo struct {
	UserID      uint      `json:"user_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	GroupRoleID uint      `json:"group_role_id"`
}

type RoomUserInfo struct {
	UserID     uint `json:"user_id"`
	RoomRoleID uint `json:"room_role_id"`
	CanJoin    bool `json:"can_join"`
	CanSee     bool `json:"can_see"`
}

type RoomInfo struct {
	Room  Room           `json:"room"`
	Users []RoomUserInfo `json:"users"`
	Token string         `json:"token"`
}

type GroupInfo struct {
	Group Group           `json:"group"`
	Users []GroupUserInfo `json:"users"`
	Rooms []RoomInfo      `json:"rooms"`
}
