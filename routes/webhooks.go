package routes

import (
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
)

// --------------------------------------------------------------------------------
// livekit webhook handler
// --------------------------------------------------------------------------------
type LivekitHandlerResponse struct {
	Success bool `json:"success"`
}

func LivekitHandler(c *fiber.Ctx) error {
	log.Println("Received livekit webhook", c)
	keys := map[string]string{os.Getenv("LIVEKIT_KEY"): os.Getenv("LIVEKIT_SECRET")}
	provider := auth.NewFileBasedKeyProviderFromMap(keys)

	// get request body
	event := &livekit.WebhookEvent{}
	if err := c.BodyParser(event); err != nil {
		log.Println("Error parsing body:", err)
		return err
	}

	// get raw body
	data, err := ioutil.ReadAll(c.Context().RequestBodyStream())
	if err != nil {
		log.Println("Error reading body:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Error reading body.")
	}
	log.Println("Raw body:", string(data))

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

	log.Println("Received valid webhook", event)
	log.Println("can handle as desired")

	// return response
	response := &LivekitHandlerResponse{
		Success: true,
	}
	return c.JSON(response)
}
