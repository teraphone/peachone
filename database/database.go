package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"peachone/models"
)

func InitDBTables(db *gorm.DB) {

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
		db.Create(&models.RoomRole{Role: "base"})
		db.Create(&models.RoomRole{Role: "moderator"})
		db.Create(&models.RoomRole{Role: "admin"})
		db.Create(&models.RoomRole{Role: "owner"})
		db.Create(&models.RoomRole{Role: "guest"})
		db.Create(&models.RoomRole{Role: "banned"})
	}

	// check if db has group_role table, create and populate if not
	if !db.Migrator().HasTable(&models.GroupRole{}) {
		// create group_role table
		db.Migrator().CreateTable(&models.GroupRole{})

		// add group roles
		db.Create(&models.GroupRole{Role: "base"})
		db.Create(&models.GroupRole{Role: "moderator"})
		db.Create(&models.GroupRole{Role: "admin"})
		db.Create(&models.GroupRole{Role: "owner"})
		db.Create(&models.GroupRole{Role: "guest"})
		db.Create(&models.GroupRole{Role: "banned"})
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

	db.SetupJoinTable(&models.Group{}, "Users", &models.GroupUser{})
	db.SetupJoinTable(&models.Room{}, "Users", &models.RoomUser{})

	db.AutoMigrate(&models.Group{})
	db.AutoMigrate(&models.Room{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.GroupUser{})
	db.AutoMigrate(&models.RoomUser{})
	db.AutoMigrate(&models.GroupInvite{})

}

func CreateDBConnection() (*gorm.DB, error) {
	// get environment variables for db connection
	DB_HOST := os.Getenv("DB_HOST")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_NAME := os.Getenv("DB_NAME")
	DB_PORT := os.Getenv("DB_PORT")

	// set up db connection string
	connectionInfoFmt := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=US/Pacific"
	connectionInfo := fmt.Sprintf(connectionInfoFmt, DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT)
	fmt.Println("connectionInfo: ", connectionInfo)

	// open db connection
	db, err := gorm.Open(postgres.Open(connectionInfo), &gorm.Config{})
	if err != nil {
		return nil, err
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
		// is this necessary?
		db.SetupJoinTable(&models.Group{}, "Users", &models.GroupUser{})
		db.SetupJoinTable(&models.Room{}, "Users", &models.RoomUser{})
	}

	return db, nil
}
