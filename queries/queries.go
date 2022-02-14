package queries

import (
	"fmt"
	"peachone/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type GroupUserInfo struct {
	UserID      uint      `json:"user_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	GroupRoleID uint      `json:"group_role_id"`
}

type GroupUsersInfo struct {
	GroupUsers []GroupUserInfo `json:"group_users"`
}

func GetGroupUserInfo(db *gorm.DB, group_id uint, user_id uint) (*GroupUserInfo, error) {
	sql_fmt := "SELECT users.id as user_id, users.name, group_users.created_at, group_users.updated_at, group_users.group_role_id " +
		"FROM group_users " +
		"JOIN users " +
		"ON group_users.user_id = users.id " +
		"WHERE group_users.group_id = %d AND group_users.user_id = %d;"
	sql := fmt.Sprintf(sql_fmt, group_id, user_id)
	group_user_info := &GroupUserInfo{}
	tx := db.Raw(sql).Scan(group_user_info)
	if tx.RowsAffected == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "Group user not found.")
	}

	return group_user_info, nil
}

func GetGroupUsersInfo(db *gorm.DB, group_id uint) ([]GroupUserInfo, error) {
	sql_fmt := "SELECT users.id as user_id, users.name, group_users.created_at, group_users.updated_at, group_users.group_role_id " +
		"FROM group_users " +
		"JOIN users " +
		"ON group_users.user_id = users.id " +
		"WHERE group_users.group_id = %d;"
	sql := fmt.Sprintf(sql_fmt, group_id)
	group_users_info := []GroupUserInfo{}
	tx := db.Raw(sql).Scan(&group_users_info)
	if tx.RowsAffected == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "Group users not found.")
	}

	return group_users_info, nil
}

type GroupUserRoleCount struct {
	Count uint `json:"count"`
}

func GetGroupUserRoleCount(db *gorm.DB, group_id uint, group_role_id uint) (*GroupUserRoleCount, error) {
	sql_fmt := "SELECT COUNT(*) as count " +
		"FROM group_users " +
		"WHERE group_id = %d AND group_role_id = %d;"
	sql := fmt.Sprintf(sql_fmt, group_id, group_role_id)
	group_user_role_count := GroupUserRoleCount{}
	tx := db.Raw(sql).Scan(&group_user_role_count)
	if tx.RowsAffected == 0 {
		return nil, fiber.NewError(fiber.StatusNotFound, "Ruh-roh! Can't count group users with that role.")
	}

	return &group_user_role_count, nil
}

func AddUserToGroupAndRooms(db *gorm.DB, user_id uint, group_id uint) error {
	// verify user exists
	user := &models.User{}
	query := db.Where("id = ?", user_id).Find(user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user_id.")
	}

	// verify new user is not already in group
	new_group_user := &models.GroupUser{
		GroupID:     group_id,
		UserID:      user_id,
		GroupRoleID: models.GroupRoleMap["member"],
	}
	query = db.Where("user_id = ? AND group_id = ?", user_id, group_id).Find(new_group_user)
	if query.RowsAffected != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is already in this group.")
	}

	// create new group user
	tx := db.Create(new_group_user)
	if tx.Error != nil {
		return tx.Error
	}

	// add user to group's rooms
	rooms := []models.Room{}
	tx = db.Where("group_id = ?", group_id).Find(&rooms)
	if tx.Error != nil {
		return tx.Error
	}
	room_users := []models.RoomUser{}
	for _, room := range rooms {
		room_user := &models.RoomUser{
			RoomID:     room.ID,
			UserID:     new_group_user.UserID,
			RoomRoleID: new_group_user.GroupRoleID,
			CanJoin:    room.RoomTypeID == models.RoomTypeMap["public"],
			CanSee:     room.RoomTypeID != models.RoomTypeMap["secret"],
		}
		room_users = append(room_users, *room_user)
	}
	tx = db.Create(&room_users)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func ValidateGroupInviteCode(db *gorm.DB, group_id uint, invite_code string) (bool, *models.GroupInvite, error) {
	is_valid := false
	group_invite := &models.GroupInvite{}
	if invite_code != "" {
		query := db.Where("code = ? AND group_id = ? AND invite_status_id = ?",
			invite_code, group_id, models.InviteStatusMap["pending"]).Find(group_invite)
		if query.RowsAffected == 0 {
			return is_valid, group_invite, fiber.NewError(fiber.StatusBadRequest, "Invalid invite_code.")
		}
		is_valid = true
	}

	return is_valid, group_invite, nil
}

func AcceptInviteAndCreateReferral(db *gorm.DB, group_invite *models.GroupInvite, user_id uint) error {
	// set invite status to accepted
	group_invite.InviteStatusID = models.InviteStatusMap["accepted"]
	tx := db.Model(group_invite).Update("invite_status_id", models.InviteStatusMap["accepted"])
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error updating invite status.")
	}

	// create referral
	referral := &models.Referral{
		UserID:     user_id,
		ReferrerID: group_invite.ReferrerID,
	}
	tx = db.Create(referral)
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error creating referral.")
	}

	return nil
}
