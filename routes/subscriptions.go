package routes

import (
	"fmt"
	"peachone/database"
	"peachone/models"
	"peachone/saasapi"
	"time"

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
		SubscriptionTermStartDate: ReadDate(resp.Subscription.Term.StartDate),
		SubscriptionTermEndDate:   ReadDate(resp.Subscription.Term.EndDate),
	}
}

// --------------------------------------------------------------------------------
// Resolve handler
// --------------------------------------------------------------------------------
type ResolveRequest struct {
	Token string `json:"token"`
}

type ResolveResponse struct {
	Success        bool   `json:"success"`
	SubscriptionId string `json:"subscriptionId"`
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
		Success:        true,
		SubscriptionId: *resp.ResolvedSubscription.ID,
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

	// populate new subscription
	newSubscription := makeSubscription(activatedSubscriptionResponse)

	// make sure subscription has valid start dates
	maxTries := 24
	for i := 0; i < maxTries; i++ {
		if newSubscription.SubscriptionTermEndDate.IsZero() ||
			newSubscription.SubscriptionTermStartDate.IsZero() {
			time.Sleep(5 * time.Second)
			activatedSubscriptionResponse, err = client.GetSubscription(
				c.Context(),
				req.SubscriptionId,
				&saasapi.FulfillmentOperationsClientGetSubscriptionOptions{},
			)
			if err != nil {
				fmt.Println("error getting activated subscription:", err)
				return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve activated subscription")
			}
			newSubscription = makeSubscription(activatedSubscriptionResponse)
		} else {
			break
		}
	}

	// get database connection
	db := database.DB.DB

	// check if subscription in db... if no, create it; if so, update it
	query := db.Where("id = ?", req.SubscriptionId).Find(&models.Subscription{})
	if query.RowsAffected == 0 {
		tx := db.Create(newSubscription)
		if tx.Error != nil {
			fmt.Println("db error creating subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "db could not create subscription")
		}
		// if new subscription, send email
		_, _, err := SendNewSubscriptionAlert(c.Context(), newSubscription)
		if err != nil {
			fmt.Println("error sending new subscription alert for subscriptionId:", newSubscription.Id, err)
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
type TenantSubscriptions map[string]map[string]models.Subscription

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
		tenantSubscriptions[user.Tid] = make(map[string]models.Subscription)
		tenantSubscriptions[user.Tid][userSubscription.Id] = *userSubscription
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
				tenantSubscriptions[tid] = make(map[string]models.Subscription)
			}
			tenantSubscriptions[tid][adminSubscription.Id] = adminSubscription
		}
	}

	// create response
	response := &GetSubscriptionsResponse{
		Success:       true,
		Subscriptions: tenantSubscriptions,
	}

	return c.JSON(response)
}

// --------------------------------------------------------------------------------
// Assign User Subscription Request
// --------------------------------------------------------------------------------
type AssignUserSubscriptionRequest struct {
	SubscriptionId string `json:"subscriptionId"`
}

type AssignUserSubscriptionResponse struct {
	Success bool              `json:"success"`
	User    models.TenantUser `json:"user"`
}

func AssignUserSubscription(c *fiber.Ctx) error {
	// check JWT
	claims, err := getClaimsFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "expired jwt")
	}

	// get tenantId, userId from request
	tid := c.Params("tid")
	oid := c.Params("oid")

	// get request body
	req := &AssignUserSubscriptionRequest{}
	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// get database connection
	db := database.DB.DB

	// get user
	user := &models.TenantUser{}
	query := db.Where("oid = ? AND tid = ?", oid, tid).Find(user)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	// get subscriptions
	subscriptions := []models.Subscription{}
	query = db.Where("beneficiary_oid = ? OR purchaser_oid = ?", claims.Oid, claims.Oid).Find(&subscriptions)
	if query.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusNotFound, "user is not admin of any subscriptions")
	}

	// find the subscriptionIds for this tenant
	tenantSubscriptionIds := []string{}
	for _, subscription := range subscriptions {
		if subscription.BeneficiaryTid == tid {
			tenantSubscriptionIds = append(tenantSubscriptionIds, subscription.Id)
		}
	}
	if len(tenantSubscriptionIds) == 0 {
		return fiber.NewError(fiber.StatusNotFound, "user is not admin of any subscriptions for this tenant")
	}

	// are we assigning or removing a subscription?
	var assigning = false
	if req.SubscriptionId != "" {
		assigning = true
	}

	if !assigning {
		// remove subscription
		tx := db.Model(user).Update("subscription_id", "")
		if tx.Error != nil {
			fmt.Println("db error removing subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "could not unassign subscription")
		}
	} else {
		// check if req.SubscriptionId is in tenantSubscriptionIds
		var subscriptionIdFound = false
		for _, subscriptionId := range tenantSubscriptionIds {
			if subscriptionId == req.SubscriptionId {
				subscriptionIdFound = true
				break
			}
		}
		if !subscriptionIdFound {
			return fiber.NewError(fiber.StatusNotFound, "subscription not found")
		}
	}

	// find target subscription
	targetSubscription := &models.Subscription{}
	for _, subscription := range subscriptions {
		if subscription.Id == req.SubscriptionId {
			targetSubscription = &subscription
			break
		}
	}

	// check if subscription is already assigned to user
	isNewSubscription := false
	if targetSubscription.Id != user.SubscriptionId {
		isNewSubscription = true
	}

	// assign subscription
	if isNewSubscription {

		// check if there are seats available
		var count int64 = 0
		tx := db.Model(&models.TenantUser{}).Where("subscription_id = ?", targetSubscription.Id).Count(&count)
		if tx.Error != nil {
			fmt.Println("db error counting users:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "could not assign subscription")
		}
		if count >= int64(targetSubscription.Quantity) && assigning {
			return fiber.NewError(fiber.StatusForbidden, "not enough seats available")
		}

		// update user
		user.SubscriptionId = req.SubscriptionId
		tx = db.Model(user).Update("subscription_id", req.SubscriptionId)
		if tx.Error != nil {
			fmt.Println("db error updating subscription:", tx.Error)
			return fiber.NewError(fiber.StatusInternalServerError, "could not assign subscription")
		}

	}

	// create response
	response := &AssignUserSubscriptionResponse{
		Success: true,
		User:    *user,
	}
	return c.JSON(response)

}

// --------------------------------------------------------------------------------
// Get Users By Tenant
// --------------------------------------------------------------------------------
type GetUsersByTenantResponse struct {
	Success bool                `json:"success"`
	Users   []models.TenantUser `json:"users"`
}

func GetUsersByTenant(c *fiber.Ctx) error {
	// check JWT
	claims, err := getClaimsFromJWT(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "expired jwt")
	}

	// get tenantId from request
	tid := c.Params("tid")

	// get database connection
	db := database.DB.DB

	// check if user has admin access to this tenant or is a user of this tenant
	var hasAccess = false
	if claims.Tid == tid {
		hasAccess = true
	} else {
		// check if user is admin of any subscriptions for this tenant
		subscriptions := []models.Subscription{}
		query := db.Where("beneficiary_oid = ? OR purchaser_oid = ?", claims.Oid, claims.Oid).Find(&subscriptions)
		if query.RowsAffected > 0 {
			for _, subscription := range subscriptions {
				if subscription.BeneficiaryTid == tid {
					hasAccess = true
					break
				}
			}
		}
	}

	if !hasAccess {
		return fiber.NewError(fiber.StatusForbidden, "user does not have access to this tenant")
	}

	// get users
	users := []models.TenantUser{}
	query := db.Where("tid = ?", tid).Find(&users)
	if query.Error != nil {
		fmt.Println("db error getting users:", query.Error)
		return fiber.NewError(fiber.StatusInternalServerError, "could not get users")
	}

	// create response
	response := &GetUsersByTenantResponse{
		Success: true,
		Users:   users,
	}
	return c.JSON(response)
}
