package routes

import (
	"fmt"
	"peachone/database"
	"peachone/models"
	"peachone/saasapi"

	"github.com/gofiber/fiber/v2"
)

func makeSubscription(resp saasapi.FulfillmentOperationsClientGetSubscriptionResponse) *models.Subscription {
	return &models.Subscription{
		AutoRenew:                 *resp.Subscription.AutoRenew,
		BeneficiaryEmail:          *resp.Subscription.Beneficiary.EmailID,
		BeneficiaryOid:            *resp.Subscription.Beneficiary.ObjectID,
		BeneficiaryTid:            *resp.Subscription.Beneficiary.TenantID,
		BeneficiaryPuid:           *resp.Subscription.Beneficiary.Puid,
		Created:                   *resp.Subscription.Created,
		Id:                        *resp.Subscription.ID,
		IsTest:                    *resp.Subscription.IsTest,
		Name:                      *resp.Subscription.Name,
		OfferId:                   *resp.Subscription.OfferID,
		PlanId:                    *resp.Subscription.PlanID,
		PublisherId:               *resp.Subscription.PublisherID,
		PurchaserEmail:            *resp.Subscription.Purchaser.EmailID,
		PurchaserOid:              *resp.Subscription.Purchaser.ObjectID,
		PurchaserTid:              *resp.Subscription.Purchaser.TenantID,
		PurchaserPuid:             *resp.Subscription.Purchaser.Puid,
		Quantity:                  int(*resp.Quantity),
		SaaSSubscriptionStatus:    models.SubscriptionStatusEnum(*resp.SaasSubscriptionStatus),
		SandboxType:               models.SandboxTypeEnum(*resp.Subscription.SandboxType),
		SessionId:                 ReadString(resp.Subscription.SessionID),
		SessionMode:               models.SessionModeEnum(*resp.Subscription.SessionMode),
		StoreFront:                ReadString(resp.Subscription.StoreFront),
		SubscriptionTermStartDate: *resp.Subscription.Term.StartDate,
		SubscriptionTermEndDate:   *resp.Subscription.Term.EndDate,
	}
}

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
	subscriptionResponse, err := client.GetSubscription(
		c.Context(),
		req.SubscriptionId,
		&saasapi.FulfillmentOperationsClientGetSubscriptionOptions{},
	)
	if err != nil {
		fmt.Println("error getting subscription:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve subscription")
	}

	// activate subscription
	quantity := int64(*subscriptionResponse.Subscription.Quantity)
	_, err = client.ActivateSubscription(c.Context(), req.SubscriptionId, saasapi.SubscriberPlan{
		PlanID:   subscriptionResponse.Subscription.PlanID,
		Quantity: &quantity,
	}, &saasapi.FulfillmentOperationsClientActivateSubscriptionOptions{})
	if err != nil {
		fmt.Println("error activating subscription:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not activate subscription")
	}

	// get activated subscription
	activatedSubscriptionResponse, err := client.GetSubscription(
		c.Context(),
		req.SubscriptionId,
		&saasapi.FulfillmentOperationsClientGetSubscriptionOptions{},
	)
	if err != nil {
		fmt.Println("error getting activated subscription:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve activated subscription")
	}

	// get database connection
	db := database.DB.DB

	// populate new subscription
	newSubscription := makeSubscription(activatedSubscriptionResponse)

	// check if subscription in db... if no, create it; if so, update it
	query := db.Where("id = ?", req.SubscriptionId).Find(&models.Subscription{})
	if query.RowsAffected == 0 {
		tx := db.Create(newSubscription)
		if tx.Error != nil {
			fmt.Println("db error creating subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "db could not create subscription")
		}
	} else {
		tx := db.Model(&models.Subscription{}).Where("id = ?", req.SubscriptionId).Updates(newSubscription)
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

// --------------------------------------------------------------------------------
// Subscriptions Webhook Dispatch
// --------------------------------------------------------------------------------
type SubscriptionsWebhookRequest struct {
	Action saasapi.OperationActionEnum `json:"action"`
}

func SubscriptionsWebhook(c *fiber.Ctx) error {
	// get request body
	req := new(SubscriptionsWebhookRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.Action == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	return GenericAction(c)

}

// --------------------------------------------------------------------------------
// Webhook Generic Action handler
// --------------------------------------------------------------------------------
type GenericActionRequest struct {
	Id             string                      `json:"id"`
	SubscriptionId string                      `json:"subscriptionId"`
	Action         saasapi.OperationActionEnum `json:"action"`
}

type GenericActionResponse struct {
	Success bool `json:"success"`
}

func GenericAction(c *fiber.Ctx) error {
	// get request body
	req := new(GenericActionRequest)
	if err := c.BodyParser(req); err != nil {
		return err
	}

	// validate request body
	if req.SubscriptionId == "" || req.Id == "" || req.Action == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// create subscription operations client
	operationsClient, err := saasapi.NewDefaultSubscriptionOperationsClient()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create subscription operations client")
	}

	// get operation
	operationStatusResponse, err := operationsClient.GetOperationStatus(
		c.Context(),
		req.SubscriptionId,
		req.Id,
		&saasapi.SubscriptionOperationsClientGetOperationStatusOptions{},
	)
	if err != nil {
		fmt.Println("error getting operation:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve operation")
	}
	operationJSON, err := operationStatusResponse.MarshalJSON()
	if err != nil {
		fmt.Println("error marshalling operation:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not marshal operation")
	}
	fmt.Println("operation:", string(operationJSON))

	// update operation status if action is "Reinstate", "ChangePlan", or "ChangeQuantity"
	action := *operationStatusResponse.Action
	if action == saasapi.OperationActionEnumReinstate ||
		action == saasapi.OperationActionEnumChangePlan ||
		action == saasapi.OperationActionEnumChangeQuantity {
		quantity := int64(*operationStatusResponse.Quantity)
		status := saasapi.UpdateOperationStatusEnumSuccess
		updateOperation := &saasapi.UpdateOperation{
			PlanID:   operationStatusResponse.PlanID,
			Quantity: &quantity,
			Status:   &status,
		}
		_, err = operationsClient.UpdateOperationStatus(
			c.Context(),
			*operationStatusResponse.SubscriptionID,
			*operationStatusResponse.ID,
			*updateOperation,
			&saasapi.SubscriptionOperationsClientUpdateOperationStatusOptions{},
		)
		if err != nil {
			fmt.Println("error updating operation:", err)
			return fiber.NewError(fiber.StatusInternalServerError, "could not update operation")
		}
	}

	// create fulfillment api client
	fulfillmentClient, err := saasapi.NewDefaultFulfillmentOperationsClient()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create fulfillment api client")
	}

	// get subscription
	subscriptionResponse, err := fulfillmentClient.GetSubscription(
		c.Context(),
		*operationStatusResponse.SubscriptionID,
		&saasapi.FulfillmentOperationsClientGetSubscriptionOptions{},
	)
	if err != nil {
		fmt.Println("error getting subscription:", err)
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve subscription")
	}

	// get database connection
	db := database.DB.DB

	// populate new subscription
	newSubscription := makeSubscription(subscriptionResponse)

	// check if subscription in db... if no, create it; if so, update it
	currentSubscription := &models.Subscription{}
	query := db.Where("id = ?", newSubscription.Id).Find(currentSubscription)
	if query.RowsAffected == 0 {
		tx := db.Create(newSubscription)
		if tx.Error != nil {
			fmt.Println("db error creating subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "db could not create subscription")
		}
	} else {
		// check if new sub has lower quantity. if so, send email.
		if newSubscription.Quantity < currentSubscription.Quantity {
			// send email
			_, _, err := SendSubscriptionDowngradeAlert(c.Context(), newSubscription, currentSubscription)
			if err != nil {
				fmt.Println("error sending subscription downgrade alert for subscriptionId:", newSubscription.Id, err)
			}
		}
		tx := db.Model(&models.Subscription{}).Where("id = ?", newSubscription.Id).Updates(newSubscription)
		if tx.Error != nil {
			fmt.Println("db error updating subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "db could not update subscription")
		}
	}

	// return response
	response := &GenericActionResponse{
		Success: true,
	}
	return c.JSON(response)
}

// --------------------------------------------------------------------------------
// Get Subscriptions Request
// --------------------------------------------------------------------------------
type TenantSubscriptions map[string][]models.Subscription

type GetSubscriptionsResponse struct {
	Success       bool                `json:"success"`
	Subscriptions TenantSubscriptions `json:"subscriptions"`
}

func GetSubscriptions(c *fiber.Ctx) error {
	// check JWT
	claims, err := getClaimsFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "expired jwt")
	}

	// get database connection
	db := database.DB.DB

	// get user, user subscription (if exists)
	user := &models.TenantUser{}
	hasUserSubscription := false
	userSubscription := &models.Subscription{}
	query := db.Where("oid = ?", claims.Oid).Find(user)
	if query.RowsAffected != 0 {
		query = db.Where("id = ?", user.SubscriptionId).Find(userSubscription)
		if query.RowsAffected != 0 {
			hasUserSubscription = true
		}
	}

	// get admin (purchaser/beneficiary) subscriptions?
	hasAdminSubscriptions := false
	adminSubscriptions := []models.Subscription{}
	query = db.Where("beneficiary_oid = ? OR purchaser_oid = ?", claims.Oid, claims.Oid).Find(&adminSubscriptions)
	if query.RowsAffected != 0 {
		hasAdminSubscriptions = true
	}

	// populate TenantSubscriptions
	tenantSubscriptions := make(TenantSubscriptions)
	if hasUserSubscription {
		tenantSubscriptions[user.Tid] = []models.Subscription{}
		tenantSubscriptions[user.Tid] = append(tenantSubscriptions[user.Tid], *userSubscription)
	}
	if hasAdminSubscriptions {
		for _, adminSubscription := range adminSubscriptions {
			var tid string
			if adminSubscription.PurchaserOid == claims.Oid {
				tid = adminSubscription.PurchaserTid
			} else {
				tid = adminSubscription.BeneficiaryTid
			}
			if _, ok := tenantSubscriptions[tid]; !ok {
				tenantSubscriptions[tid] = []models.Subscription{}
			}
			tenantSubscriptions[tid] = append(tenantSubscriptions[tid], adminSubscription)
		}
	}

	// todo: finish this
	// - should TenantSubscriptions be map[string]map[string]models.Subscription instead?

	// create response
	response := &GetSubscriptionsResponse{
		Success:       true,
		Subscriptions: make(TenantSubscriptions),
	}

	return c.JSON(response)
}
