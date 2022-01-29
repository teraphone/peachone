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

type DbInstance struct {
	Db *gorm.DB
}

var Database DbInstance

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
		db.AutoMigrate(new(models.User))
	} else {
		log.Println("Skipping Migrations")
	}

	return db, nil
}
