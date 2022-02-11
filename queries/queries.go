package queries

import (
	"fmt"
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
