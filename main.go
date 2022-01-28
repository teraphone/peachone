package main

import (
	"fmt"
	"log"
	"os"

	"peachone/database"
	"peachone/routes"

	"github.com/gofiber/fiber/v2"
)

func welcome(c *fiber.Ctx) error {
	return c.SendString("Welcome to an Awesome API")
}

func setupRoutes(app *fiber.App) {
	// Welcome endpoint
	app.Get("/v1", welcome)
	// User endpoints
	app.Post("/v1/users", routes.CreateUser)
	app.Get("/v1/users", routes.GetUsers)
	app.Get("/v1/users/:id", routes.GetUser)
	app.Delete("/v1/users/:id", routes.DeleteUser)

}

func main() {
	database.ConnectDb()

	app := fiber.New()
	setupRoutes(app)

	// Determine port for HTTP service.
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "3000"
		log.Printf("defaulting to port %s", PORT)
	}

	log.Fatal(app.Listen(fmt.Sprintf(":%s", PORT)))

}
