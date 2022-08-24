package routes

import (
	"fmt"
	"peachone/database"
	"peachone/models"
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
	// check JWT
	_, err := getClaimsFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "expired jwt")
	}

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

// --------------------------------------------------------------------------------
// Activate handler
// --------------------------------------------------------------------------------
type ActivateRequest struct {
	SubscriptionId string `json:"subscriptionId"`
}

type ActivateResponse struct {
	Success      bool                `json:"success"`
	Subscription models.Subscription `json:"subscription"`
}

func Activate(c *fiber.Ctx) error {
	// check JWT
	_, err := getClaimsFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "expired jwt")
	}

	// get request body
	req := new(ActivateRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.SubscriptionId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid subscriptionId")
	}

	// create fulfillment api client
	client, err := saasapi.NewDefaultFulfillmentOperationsClient()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create fulfillment api client")
	}

	// get subscription
	subscriptionResponse, err := client.GetSubscription(c.Context(), req.SubscriptionId, nil)
	if err != nil {
		fmt.Println("error getting subscription:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve subscription")
	}

	// activate subscription
	quantity := int64(*subscriptionResponse.Subscription.Quantity)
	_, err = client.ActivateSubscription(c.Context(), req.SubscriptionId, saasapi.SubscriberPlan{
		PlanID:   subscriptionResponse.Subscription.PlanID,
		Quantity: &quantity,
	}, nil)
	if err != nil {
		fmt.Println("error activating subscription:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not activate subscription")
	}

	// get activated subscription
	activatedSubscriptionResponse, err := client.GetSubscription(c.Context(), req.SubscriptionId, nil)
	if err != nil {
		fmt.Println("error getting activated subscription:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve activated subscription")
	}

	// get database connection
	db := database.DB.DB

	// populate new subscription
	newSubscription := &models.Subscription{
		AutoRenew:                 *activatedSubscriptionResponse.Subscription.AutoRenew,
		BeneficiaryEmail:          *activatedSubscriptionResponse.Subscription.Beneficiary.EmailID,
		BeneficiaryOid:            *activatedSubscriptionResponse.Subscription.Beneficiary.ObjectID,
		BeneficiaryTid:            *activatedSubscriptionResponse.Subscription.Beneficiary.TenantID,
		BeneficiaryPuid:           *activatedSubscriptionResponse.Subscription.Beneficiary.Puid,
		Created:                   *activatedSubscriptionResponse.Subscription.Created,
		Id:                        *activatedSubscriptionResponse.Subscription.ID,
		IsTest:                    *activatedSubscriptionResponse.Subscription.IsTest,
		Name:                      *activatedSubscriptionResponse.Subscription.Name,
		OfferId:                   *activatedSubscriptionResponse.Subscription.OfferID,
		PlanId:                    *activatedSubscriptionResponse.Subscription.PlanID,
		PublisherId:               *activatedSubscriptionResponse.Subscription.PublisherID,
		PurchaserEmail:            *activatedSubscriptionResponse.Subscription.Purchaser.EmailID,
		PurchaserOid:              *activatedSubscriptionResponse.Subscription.Purchaser.ObjectID,
		PurchaserTid:              *activatedSubscriptionResponse.Subscription.Purchaser.TenantID,
		PurchaserPuid:             *activatedSubscriptionResponse.Subscription.Purchaser.Puid,
		Quantity:                  int(*activatedSubscriptionResponse.Quantity),
		SaaSSubscriptionStatus:    models.SubscriptionStatusEnum(*activatedSubscriptionResponse.SaasSubscriptionStatus),
		SandboxType:               models.SandboxTypeEnum(*activatedSubscriptionResponse.Subscription.SandboxType),
		SessionId:                 *activatedSubscriptionResponse.Subscription.SessionID,
		SessionMode:               models.SessionModeEnum(*activatedSubscriptionResponse.Subscription.SessionMode),
		StoreFront:                *activatedSubscriptionResponse.Subscription.StoreFront,
		SubscriptionTermStartDate: *activatedSubscriptionResponse.Subscription.Term.StartDate,
		SubscriptionTermEndDate:   *activatedSubscriptionResponse.Subscription.Term.EndDate,
	}

	// check if subscription in db... if no, create it; if so, update it
	query := db.Where("id = ?", req.SubscriptionId).Find(&models.Subscription{})
	if query.RowsAffected == 0 {
		tx := db.Create(newSubscription)
		if tx.Error != nil {
			fmt.Println("db error creating subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "db could not create subscription")
		}
	} else {
		tx := db.Model(&models.Subscription{}).Updates(*newSubscription)
		if tx.Error != nil {
			fmt.Println("db error updating subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "db could not update subscription")
		}
	}

	// return response
	response := &ActivateResponse{
		Success:      true,
		Subscription: *newSubscription,
	}
	return c.JSON(response)
}
