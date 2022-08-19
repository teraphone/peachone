package routes

import (
	"peachone/saasapi"

	"github.com/gofiber/fiber/v2"
)

// --------------------------------------------------------------------------------
// Resolve handler
// --------------------------------------------------------------------------------
type ResolveRequest struct {
	Token string `json:"token"`
}

type ResolveResponse struct {
	Success              bool                         `json:"success"`
	ResolvedSubscription saasapi.ResolvedSubscription `json:"resolvedSubscription"`
}

func Resolve(c *fiber.Ctx) error {
	// get request body
	req := new(ResolveRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid token.")
	}

	// create fulfillment api client
	client, err := saasapi.NewDefaultFulfillmentOperationsClient()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create fulfillment api client")
	}

	resp, err := client.Resolve(c.Context(), req.Token, nil)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not resolve token")
	}

	// return response
	response := &ResolveResponse{
		Success:              true,
		ResolvedSubscription: resp.ResolvedSubscription,
	}
	return c.JSON(response)
}
