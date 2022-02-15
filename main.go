package main

import (
	"fmt"
	"log"
	"os"

	"peachone/database"
	"peachone/routes"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

func setupRoutes(app *fiber.App) {
	setupPublic(app)
	setupPrivate(app)

}

func setupPublic(app *fiber.App) {
	public := app.Group("/v1/public")

	// Welcome endpoint
	public.Get("/", routes.PublicWelcome)

	// User endpoints
	public.Post("/signup", routes.Signup)
	public.Post("/login", routes.Login)
}

func setupPrivate(app *fiber.App) {
	private := app.Group("/v1/private")
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	private.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(SIGNING_KEY),
	}))

	// Welcome endpoint
	private.Get("/", routes.PrivateWelcome)

	// Auth endpoint
	private.Get("/auth", routes.RefreshToken)

	// Groups endpoints
	private.Post("/groups", routes.CreateGroup)
	private.Get("/groups", routes.GetGroups)

	private.Get("/groups/:group_id", routes.GetGroup)
	private.Patch("/groups/:group_id", routes.UpdateGroup)
	private.Delete("/groups/:group_id", routes.DeleteGroup)

	private.Post("/groups/:group_id/users", routes.CreateGroupUser)
	private.Get("/groups/:group_id/users", routes.GetGroupUsers)

	private.Get("/groups/:group_id/users/:user_id", routes.GetGroupUser)
	private.Patch("/groups/:group_id/users/:user_id", routes.UpdateGroupUser)
	private.Delete("/groups/:group_id/users/:user_id", routes.DeleteGroupUser)

	private.Post("/groups/:group_id/invites", routes.CreateGroupInvite)
	private.Get("/groups/:group_id/invites", routes.GetGroupInvites)

	private.Get("/groups/:group_id/invites/:invite_id", routes.GetGroupInvite)
	private.Delete("/groups/:group_id/invites/:invite_id", routes.DeleteGroupInvite)

	private.Post("/groups/:group_id/rooms", routes.CreateRoom)
	private.Get("/groups/:group_id/rooms", routes.GetRooms)

	private.Get("/groups/:group_id/rooms/:room_id", routes.GetRoom)
	private.Delete("/groups/:group_id/rooms/:room_id", routes.DeleteRoom)
	private.Patch("/groups/:group_id/rooms/:room_id", routes.UpdateRoom)

	private.Get("/groups/:group_id/rooms/:room_id/users", routes.GetRoomUsers)

	private.Get("/groups/:group_id/rooms/:room_id/users/:user_id", routes.GetRoomUser)
	private.Patch("/groups/:group_id/rooms/:room_id/users/:user_id", routes.UpdateRoomUser)

	// Invites endpoints
	private.Post("/invites", routes.AcceptGroupInvite)
}

func main() {
	// Optionally automigrate at startup
	DB_AUTOMIGRATE := os.Getenv("DB_AUTOMIGRATE")
	if DB_AUTOMIGRATE == "true" {
		_, err := database.CreateDBConnection()
		if err != nil {
			panic(err)
		}
	}

	// Create app
	app := fiber.New()
	setupRoutes(app)

	// Determine port for HTTP service.
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "3000"
		log.Printf("defaulting to port %s", PORT)
	}

	// Start server
	log.Fatal(app.Listen(fmt.Sprintf(":%s", PORT)))

}
