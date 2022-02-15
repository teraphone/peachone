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

func ValidateGroupInviteCode(db *gorm.DB, group_id uint, invite_code string) (*models.GroupInvite, error) {
	group_invite := &models.GroupInvite{}
	if invite_code != "" {

		if group_id != 0 {
			query := db.Where("code = ? AND group_id = ? AND invite_status_id = ?",
				invite_code, group_id, models.InviteStatusMap["pending"]).Find(group_invite)
			if query.RowsAffected == 0 {
				return group_invite, fiber.NewError(fiber.StatusBadRequest, "Invalid invite_code.")
			}
		} else {
			query := db.Where("code = ? AND invite_status_id = ?",
				invite_code, models.InviteStatusMap["pending"]).Find(group_invite)
			if query.RowsAffected == 0 {
				return group_invite, fiber.NewError(fiber.StatusBadRequest, "Invalid invite_code.")
			}
		}

	} else {
		return group_invite, fiber.NewError(fiber.StatusBadRequest, "Invalid invite_code.")
	}

	return group_invite, nil
}

func AcceptInviteAndCreateReferral(db *gorm.DB, group_invite *models.GroupInvite, user_id uint) error {
	// set invite status to accepted
	group_invite.InviteStatusID = models.InviteStatusMap["accepted"]
	tx := db.Model(group_invite).Update("invite_status_id", models.InviteStatusMap["accepted"])
	if tx.Error != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error updating invite status.")
	}

	// check if referral exists. if not, create one
	referral := &models.Referral{
		UserID:     user_id,
		ReferrerID: group_invite.ReferrerID,
	}
	query := db.Where("user_id = ? AND referrer_id = ?", user_id, group_invite.ReferrerID).Find(referral)
	if query.RowsAffected == 0 {
		tx = db.Create(referral)
		if tx.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Error creating referral.")
		}
	}
	// if referrer_id previously invited user_id to any group, then we already have the referral.

	return nil
}

func GetRoomsNotBanned(db *gorm.DB, group_id uint, user_id uint) ([]models.Room, error) {
	sql_fmt := "SELECT rooms.* " +
		"FROM rooms " +
		"JOIN room_users " +
		"ON room_users.room_id = rooms.id " +
		"WHERE room_users.user_id = %d AND rooms.group_id = %d AND room_users.room_role_id > %d;"
	sql := fmt.Sprintf(sql_fmt, user_id, group_id, models.RoomRoleMap["banned"])
	rooms := []models.Room{}
	tx := db.Raw(sql).Scan(&rooms)
	if tx.Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Error finding rooms.")
	}

	return rooms, nil
}

type RoomUserInfo struct {
	Name string `json:"name"`
	models.RoomUser
}

func GetRoomUserInfo(db *gorm.DB, room_id uint, user_id uint) (*RoomUserInfo, error) {
	sql_fmt := "SELECT users.name, room_users.* " +
		"FROM room_users " +
		"JOIN users " +
		"ON users.id = room_users.user_id " +
		"WHERE room_users.room_id = %d AND room_users.user_id = %d;"
	sql := fmt.Sprintf(sql_fmt, room_id, user_id)
	room_user_info := &RoomUserInfo{}
	tx := db.Raw(sql).Scan(room_user_info)
	if tx.Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Error finding room user info.")
	}

	return room_user_info, nil
}

func GetRoomUsersInfo(db *gorm.DB, room_id uint) ([]RoomUserInfo, error) {
	sql_fmt := "SELECT users.name, room_users.* " +
		"FROM room_users " +
		"JOIN users " +
		"ON users.id = room_users.user_id " +
		"WHERE room_users.room_id = %d;"
	sql := fmt.Sprintf(sql_fmt, room_id)
	room_users_info := []RoomUserInfo{}
	tx := db.Raw(sql).Scan(&room_users_info)
	if tx.Error != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "Error finding room users' info.")
	}

	return room_users_info, nil
}
