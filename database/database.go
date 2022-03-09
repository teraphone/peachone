package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"peachone/models"
)

type DBInstance struct {
	DB *gorm.DB
}

var DB DBInstance

func InitDBTables(db *gorm.DB) {
	// drop constraints
	sql_drop_constraints := []string{
		"ALTER TABLE rooms DROP CONSTRAINT fk_rooms_group_id;",
		"ALTER TABLE rooms DROP CONSTRAINT fk_rooms_room_type_id;",
		"ALTER TABLE rooms DROP CONSTRAINT fk_rooms_deployment_zone_id;",
		"ALTER TABLE rooms DROP CONSTRAINT fk_rooms_deprecation_code_id;",
		"ALTER TABLE group_users DROP CONSTRAINT fk_group_users_group_id;",
		"ALTER TABLE group_users DROP CONSTRAINT fk_group_users_user_id;",
		"ALTER TABLE group_users DROP CONSTRAINT fk_group_users_group_role_id;",
		"ALTER TABLE room_users DROP CONSTRAINT fk_room_users_room_id;",
		"ALTER TABLE room_users DROP CONSTRAINT fk_room_users_user_id;",
		"ALTER TABLE room_users DROP CONSTRAINT fk_room_users_room_role_id;",
		"ALTER TABLE group_invites DROP CONSTRAINT fk_group_invites_group_id;",
		"ALTER TABLE group_invites DROP CONSTRAINT fk_group_invites_invite_status_id;",
		"ALTER TABLE group_invites DROP CONSTRAINT fk_group_invites_referrer_id;",
	}
	// run sql statements
	for _, sql := range sql_drop_constraints {
		err := db.Exec(sql).Error
		if err != nil {
			log.Println("error:", err)
		}
	}

	// check if db has room_type table, create and populate if not
	if !db.Migrator().HasTable(&models.RoomType{}) {
		// create room_type table
		db.Migrator().CreateTable(&models.RoomType{})

		// add room types
		db.Create(&models.RoomType{Type: "public"})
		db.Create(&models.RoomType{Type: "private"})
		db.Create(&models.RoomType{Type: "secret"})
	}

	// check if db has deployment_zone table, create and populate if not
	if !db.Migrator().HasTable(&models.DeploymentZone{}) {
		// create deployment_zone table
		db.Migrator().CreateTable(&models.DeploymentZone{})

		// add deployment zones
		db.Create(&models.DeploymentZone{Zone: "us-west1-b"})
	}

	// check if db has deprecation_code table, create and populate if not
	if !db.Migrator().HasTable(&models.DeprecationCode{}) {
		// create deprecation_code table
		db.Migrator().CreateTable(&models.DeprecationCode{})

		// add deprecation codes
		db.Create(&models.DeprecationCode{Code: "active"})
		db.Create(&models.DeprecationCode{Code: "inactive"})
	}

	// check if db has room_role table, create and populate if not
	if !db.Migrator().HasTable(&models.RoomRole{}) {
		// create room_role table
		db.Migrator().CreateTable(&models.RoomRole{})

		// add room roles
		db.Create(&models.RoomRole{Role: "banned"})
		db.Create(&models.RoomRole{Role: "guest"})
		db.Create(&models.RoomRole{Role: "member"})
		db.Create(&models.RoomRole{Role: "moderator"})
		db.Create(&models.RoomRole{Role: "admin"})
		db.Create(&models.RoomRole{Role: "owner"})
	}

	// check if db has group_role table, create and populate if not
	if !db.Migrator().HasTable(&models.GroupRole{}) {
		// create group_role table
		db.Migrator().CreateTable(&models.GroupRole{})

		// add group roles
		db.Create(&models.GroupRole{Role: "banned"})
		db.Create(&models.GroupRole{Role: "guest"})
		db.Create(&models.GroupRole{Role: "member"})
		db.Create(&models.GroupRole{Role: "moderator"})
		db.Create(&models.GroupRole{Role: "admin"})
		db.Create(&models.GroupRole{Role: "owner"})
	}

	// check if db has invite_status table, create and populate if not
	if !db.Migrator().HasTable(&models.InviteStatus{}) {
		// create invite_status table
		db.Migrator().CreateTable(&models.InviteStatus{})

		// add invite statuses
		db.Create(&models.InviteStatus{Status: "pending"})
		db.Create(&models.InviteStatus{Status: "accepted"})
		db.Create(&models.InviteStatus{Status: "expired"})
	}

	db.AutoMigrate(&models.Group{})
	db.AutoMigrate(&models.Room{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.GroupUser{})
	db.AutoMigrate(&models.RoomUser{})
	db.AutoMigrate(&models.GroupInvite{})
	db.AutoMigrate(&models.Referral{})

	// define foreign key relationships
	sql_add_constraints := []string{
		"ALTER TABLE rooms ADD CONSTRAINT fk_rooms_group_id FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;",
		"ALTER TABLE rooms ADD CONSTRAINT fk_rooms_room_type_id FOREIGN KEY (room_type_id) REFERENCES room_types(id);",
		"ALTER TABLE rooms ADD CONSTRAINT fk_rooms_deployment_zone_id FOREIGN KEY (deployment_zone_id) REFERENCES deployment_zones(id);",
		"ALTER TABLE rooms ADD CONSTRAINT fk_rooms_deprecation_code_id FOREIGN KEY (deprecation_code_id) REFERENCES deprecation_codes(id);",
		"ALTER TABLE group_users ADD CONSTRAINT fk_group_users_group_id FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;",
		"ALTER TABLE group_users ADD CONSTRAINT fk_group_users_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;",
		"ALTER TABLE group_users ADD CONSTRAINT fk_group_users_group_role_id FOREIGN KEY (group_role_id) REFERENCES group_roles(id);",
		"ALTER TABLE room_users ADD CONSTRAINT fk_room_users_room_id FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE;",
		"ALTER TABLE room_users ADD CONSTRAINT fk_room_users_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;",
		"ALTER TABLE room_users ADD CONSTRAINT fk_room_users_room_role_id FOREIGN KEY (room_role_id) REFERENCES room_roles(id);",
		"ALTER TABLE group_invites ADD CONSTRAINT fk_group_invites_group_id FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE;",
		"ALTER TABLE group_invites ADD CONSTRAINT fk_group_invites_invite_status_id FOREIGN KEY (invite_status_id) REFERENCES invite_statuses(id);",
		"ALTER TABLE group_invites ADD CONSTRAINT fk_group_invites_referrer_id FOREIGN KEY (referrer_id) REFERENCES users(id) ON DELETE CASCADE;",
	}
	// run sql statements
	for _, sql := range sql_add_constraints {
		err := db.Exec(sql).Error
		if err != nil {
			log.Println("error:", err)
		}
	}

}

func CreateDBConnection(ctx context.Context) {
	// get environment variables for db connection
	DB_HOST := os.Getenv("DB_HOST")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_NAME := os.Getenv("DB_NAME")
	DB_PORT := os.Getenv("DB_PORT")

	// set up db connection string
	connectionInfoFmt := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC"
	connectionInfo := fmt.Sprintf(connectionInfoFmt, DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT)
	fmt.Println("connectionInfo: ", connectionInfo)

	// open db connection
	db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
		os.Exit(2)
	}

	log.Println("Connected Successfully to Database")
	db.Logger = logger.Default.LogMode(logger.Info)

	// auto migrate tables
	DB_AUTOMIGRATE := os.Getenv("DB_AUTOMIGRATE")
	if DB_AUTOMIGRATE == "true" {
		log.Println("Running Migrations")
		InitDBTables(db)
	} else {
		log.Println("Skipping Migrations")

	}

	DB = DBInstance{
		DB: db.WithContext(ctx),
	}
}
