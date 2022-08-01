package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"peachone/database"
	"peachone/fbadmin"
	"peachone/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
)

func setupRoutes(app *fiber.App) {
	app.Use(cors.New())
	setupPublic(app)
	setupPrivate(app)
	setupRoomService(app)
	setupWebhooks(app)

}

func setupPublic(app *fiber.App) {
	public := app.Group("/v1/public")

	// Welcome endpoint
	public.Get("/", routes.PublicWelcome)

	// Public endpoints
	public.Post("/login", routes.Login)
}

func setupPrivate(app *fiber.App) {
	private := app.Group("/v1/private")
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	private.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(SIGNING_KEY),
	}))

	// Private endpoints
	private.Get("/", routes.PrivateWelcome)
	private.Patch("/license", routes.UpdateLicense)
	private.Get("/world", routes.GetWorld)
	private.Post("/auth", routes.GetRefreshedAccessToken)

}

func setupRoomService(app *fiber.App) {
	roomservice := app.Group("/v1/roomservice")
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	roomservice.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(SIGNING_KEY),
	}))

	// Rooms endpoints
	roomservice.Get("/rooms/:teamId/:roomId/join", routes.JoinLiveKitRoom)
	roomservice.Get("/rooms/:teamId/:roomId", routes.GetLiveKitRoomParticipants)

}

func setupWebhooks(app *fiber.App) {
	webhooks := app.Group("/v1/webhooks")

	// Livekit webhook handler
	webhooks.Post("/livekit", routes.LivekitHandler)
}

func main() {
	// Init Firebase Admin SDK
	ctx := context.Background()
	fbadmin.InitFirebaseApp(ctx)
	fbadmin.InitFirebaseAuthClient(ctx)

	// Connect to DB
	database.CreateDBConnection(ctx)

	// Create app
	app := fiber.New()
	setupRoutes(app)

	// Determine port for HTTP service.
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
		log.Printf("defaulting to port %s", PORT)
	}

	// Listen from a different goroutine
	go func() {
		if err := app.Listen(fmt.Sprintf(":%s", PORT)); err != nil {
			log.Panic(err)
		}
	}()

	// Create channel to signify a signal being sent
	c := make(chan os.Signal, 1)

	// When an interrupt or termination signal is sent, notify the channel
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until an interrupt is received
	<-c
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	// Cleanup
	fmt.Println("Running cleanup tasks...")
	db := database.DB.DB
	conn, err := db.DB()
	if err != nil {
		log.Panic(err)
	}
	err = conn.Close()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Shutdown successful.")

}
