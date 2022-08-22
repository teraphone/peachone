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
		"ALTER TABLE team_users DROP CONSTRAINT fk_team_users_id;",
		"ALTER TABLE team_users DROP CONSTRAINT fk_team_users_oid;",
		"ALTER TABLE team_rooms DROP CONSTRAINT fk_team_rooms_team_id;",
		"ALTER TABLE tenant_users DROP CONSTRAINT fk_tenant_users_subscription_id;",
	}
	// run sql statements
	for _, sql := range sql_drop_constraints {
		err := db.Exec(sql).Error
		if err != nil {
			log.Println("error:", err)
		}
	}

	db.AutoMigrate(&models.TenantUser{})
	db.AutoMigrate(&models.TenantTeam{})
	db.AutoMigrate(&models.TeamUser{})
	db.AutoMigrate(&models.TeamRoom{})
	db.AutoMigrate(&models.Subscription{})

	// define foreign key relationships
	sql_add_constraints := []string{
		"ALTER TABLE team_users ADD CONSTRAINT fk_team_users_id FOREIGN KEY (id) REFERENCES tenant_teams(id) ON DELETE CASCADE;",
		"ALTER TABLE team_users ADD CONSTRAINT fk_team_users_oid FOREIGN KEY (oid) REFERENCES tenant_users(oid) ON DELETE CASCADE;",
		"ALTER TABLE team_rooms ADD CONSTRAINT fk_team_rooms_team_id FOREIGN KEY (team_id) REFERENCES tenant_teams(id) ON DELETE CASCADE;",
		"ALTER TABLE tenant_users ADD CONSTRAINT fk_tenant_users_subscription_id FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL;",
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
