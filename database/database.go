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

func ConnectDb() {
	// get environment variables for db connection
	DB_HOST := os.Getenv("DB_HOST")
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_NAME := os.Getenv("DB_NAME")
	DB_PORT := os.Getenv("DB_PORT")

	// set up db connection string
	DNS_fmt := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=US/Pacific"
	DNS := fmt.Sprintf(DNS_fmt, DB_HOST, DB_USER, DB_PASSWORD, DB_NAME, DB_PORT)
	fmt.Println("DNS: ", DNS)

	// open db connection
	db, err := gorm.Open(postgres.Open(DNS), &gorm.Config{})
	if err != nil {
		fmt.Println(err.Error())
		panic("Cannot connect to DB")
	}

	if err != nil {
		log.Fatal("Failed to connect to the database! \n", err)
		os.Exit(2)
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

	Database.Db = db
}
