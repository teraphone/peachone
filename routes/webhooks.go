package routes

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/livekit/protocol/auth"
	livekit "github.com/livekit/protocol/livekit"
	"google.golang.org/protobuf/encoding/protojson"
)

// --------------------------------------------------------------------------------
// livekit webhook handler
// --------------------------------------------------------------------------------
type LivekitHandlerResponse struct {
	Success bool `json:"success"`
}

const (
	RoomStarted       string = "room_started"
	RoomFinished      string = "room_finished"
	ParticipantJoined string = "participant_joined"
	ParticipantLeft   string = "participant_left"
)

func LivekitHandler(c *fiber.Ctx) error {
	keys := map[string]string{os.Getenv("LIVEKIT_KEY"): os.Getenv("LIVEKIT_SECRET")}
	provider := auth.NewFileBasedKeyProviderFromMap(keys)

	// get raw body
	ctx := c.Context()
	data := ctx.PostBody()

	// get request header
	authToken := c.Get("Authorization")
	if authToken == "" {
		log.Println("No authorization token found")
		return fiber.NewError(fiber.StatusUnauthorized, "No authorization token found")
	}

	// parse auth token
	v, err := auth.ParseAPIToken(authToken)
	if err != nil {
		log.Println("Error parsing authorization token:", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Error parsing authorization token")
	}

	secret := provider.GetSecret(v.APIKey())
	if secret == "" {
		log.Println("API secret not found")
		return fiber.NewError(fiber.StatusUnauthorized, "API secret not found")
	}

	claims, err := v.Verify(secret)
	if err != nil {
		log.Println("Error verifying authorization token:", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Error verifying authorization token")
	}

	// verify checksum
	sha := sha256.Sum256(data)
	hash := base64.StdEncoding.EncodeToString(sha[:])
	if claims.Sha256 != hash {
		log.Println("Invalid checksum")
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid checksum")
	}

	unmarshalOpts := protojson.UnmarshalOptions{
		DiscardUnknown: true,
		AllowPartial:   true,
	}
	event := livekit.WebhookEvent{}
	if err = unmarshalOpts.Unmarshal(data, &event); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error unmarshaling webhook event")
	}

	// dispatch event to appropriate handler
	switch event.Event {
	case RoomStarted:
		handleRoomStarted(ctx, &event)
	case RoomFinished:
		handleRoomFinished(ctx, &event)
	case ParticipantJoined:
		handleParticipantJoined(ctx, &event)
	case ParticipantLeft:
		handleParticipantLeft(ctx, &event)
	default:
		log.Println("Ignoring event:", event.Event)
	}

	// return response
	response := &LivekitHandlerResponse{
		Success: true,
	}
	return c.JSON(response)
}

func handleRoomStarted(ctx context.Context, event *livekit.WebhookEvent) {
	log.Println("Handling event:", event.Event)
	log.Println("Room name:", event.Room.Name)
}

func handleRoomFinished(ctx context.Context, event *livekit.WebhookEvent) {
	log.Println("Handling event:", event.Event)
	log.Println("Room name:", event.Room.Name)
}

func handleParticipantJoined(ctx context.Context, event *livekit.WebhookEvent) {
	log.Println("Handling event:", event.Event)
	log.Println("Room name:", event.Room.Name)
	log.Println("Participant identity:", event.Participant.Identity)
}

func handleParticipantLeft(ctx context.Context, event *livekit.WebhookEvent) {
	log.Println("Handling event:", event.Event)
	log.Println("Room name:", event.Room.Name)
	log.Println("Participant identity:", event.Participant.Identity)
}
