package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"peachone/fbadmin"
	"peachone/models"
	"strings"
	"text/template"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	lksdk "github.com/livekit/server-sdk-go"

	"github.com/livekit/protocol/auth"

	"github.com/mailgun/mailgun-go/v4"
)

type TokenClaims struct {
	Oid        string `json:"oid"`
	Tid        string `json:"tid"`
	Expiration int64  `json:"expiration"`
}

func getClaimsFromJWT(c *fiber.Ctx) (*TokenClaims, error) {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	oid := claims["oid"].(string)
	tid := claims["tid"].(string)
	expiration := int64(claims["expiration"].(float64))
	tokenClaims := &TokenClaims{
		Oid:        oid,
		Tid:        tid,
		Expiration: expiration,
	}

	if time.Now().Unix() > expiration {
		return tokenClaims, fmt.Errorf("token expired")
	}

	return tokenClaims, nil
}

func createAccessToken(user *models.TenantUser) (string, int64, error) {
	expiration := time.Now().Add(time.Hour * 24).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["oid"] = user.Oid
	claims["tid"] = user.Tid
	claims["expiration"] = expiration
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	tokenString, err := token.SignedString([]byte(SIGNING_KEY))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiration, nil
}

func createRefreshToken(user *models.TenantUser) (string, int64, error) {
	expiration := time.Now().Add(time.Hour * 24 * 30).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["oid"] = user.Oid
	claims["tid"] = user.Tid
	claims["expiration"] = expiration
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	tokenString, err := token.SignedString([]byte(SIGNING_KEY))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiration, nil
}

func validateToken(tokenString string) (*TokenClaims, error) {
	getKey := func(_ *jwt.Token) (interface{}, error) {
		SIGNING_KEY := os.Getenv("SIGNING_KEY")
		return []byte(SIGNING_KEY), nil
	}

	token, err := jwt.Parse(tokenString, getKey)
	if err != nil {
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	oid := claims["oid"].(string)
	tid := claims["tid"].(string)
	expiration := int64(claims["expiration"].(float64))
	tokenClaims := &TokenClaims{
		Oid:        oid,
		Tid:        tid,
		Expiration: expiration,
	}

	if time.Now().Unix() > expiration {
		return nil, errors.New("token expired")
	}

	return tokenClaims, nil
}

func createFirebaseAuthToken(ctx context.Context, userId string) (string, error) {
	token, err := fbadmin.AuthClient.CustomToken(ctx, userId)
	if err != nil {
		log.Printf("error minting custom token: %v\n", err)
		return "", err
	}

	return token, nil
}

func createLiveKitJoinToken(teamId, roomId, userId string) (string, error) {
	LIVEKIT_KEY := os.Getenv("LIVEKIT_KEY")
	LIVEKIT_SECRET := os.Getenv("LIVEKIT_SECRET")
	at := auth.NewAccessToken(LIVEKIT_KEY, LIVEKIT_SECRET)
	canPublish := true
	canSubscribe := true
	grant := &auth.VideoGrant{
		RoomCreate: false,
		RoomList:   false,
		RoomRecord: false,

		RoomAdmin: false,
		RoomJoin:  true,
		Room:      EncodeRoomName(teamId, roomId),

		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}
	at.AddGrant(grant).
		SetIdentity(userId).
		SetValidFor(730 * time.Hour)

	token, err := at.ToJWT()

	return token, err
}

func CreateRoomServiceClient() *lksdk.RoomServiceClient {
	LIVEKIT_KEY := os.Getenv("LIVEKIT_KEY")
	LIVEKIT_SECRET := os.Getenv("LIVEKIT_SECRET")
	LIVEKIT_HOST := os.Getenv("LIVEKIT_HOST")

	client := lksdk.NewRoomServiceClient(LIVEKIT_HOST, LIVEKIT_KEY, LIVEKIT_SECRET)

	return client
}

func EncodeRoomName(teamId string, roomId string) string {
	return teamId + "/" + roomId
}

func DecodeRoomName(name string) (string, string, error) {
	split := strings.Split(name, "/")
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid room name: %s", name)
	}

	teamId := split[0]
	roomId := split[1]

	return teamId, roomId, nil
}

func CreateMailgunClient() *mailgun.MailgunImpl {
	MG_DOMAIN := os.Getenv("MG_DOMAIN")
	MG_API_KEY := os.Getenv("MG_API_KEY")

	mg := mailgun.NewMailgun(MG_DOMAIN, MG_API_KEY)
	return mg
}

type AccountVerificationVars struct {
	SenderEmail    string
	Subject        string
	RecipientEmail string
	TemplateVars   *AccountVerificationTemplateVars
}

type AccountVerificationTemplateVars struct {
	Name       string
	Code       string
	SenderName string
}

func SendAccountVerificationEmail(ctx context.Context, vars *AccountVerificationVars) (mes string, id string, err error) {
	// email template
	htmlAccountVerificationTemplate := `
<html>
	<body>
		<p>Hi {{.Name}},</p>
		<p>Welcome to Teraphone! To verify your account, please click the link below:</p>
		<p><a href="https://teraphone.app/email-verification?code={{.Code}}">https://teraphone.app/email-verification?code={{.Code}}</a></p>
		<p>If you did not sign up for a Teraphone account, you can simply disregard this email.</p>
		<p>Thanks,</p>
		<p>{{.SenderName}}</p>
	</body>
</html>
`
	// create email message
	mg := CreateMailgunClient()
	message := mg.NewMessage(vars.SenderEmail, vars.Subject, "", vars.RecipientEmail)
	parsedHtmlTemplate, err := template.New("body").Parse(htmlAccountVerificationTemplate)
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	var htmlBuffer bytes.Buffer
	if err := parsedHtmlTemplate.Execute(&htmlBuffer, vars.TemplateVars); err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	message.SetHtml(htmlBuffer.String())

	// send message with 10 second timeout
	log.Printf("Sending password reset email to %s", vars.RecipientEmail)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	resp, id, err := mg.Send(ctxWithTimeout, message)
	if err != nil {
		fmt.Println(err.Error())
		return resp, id, err
	}
	log.Printf("ID: %s Resp: %s", id, resp)

	return resp, id, nil
}

type EmailSignupAlertVars struct {
	SenderEmail     string
	Subject         string
	RecipientEmails []string
	TemplateVars    *EmailSignupAlertTemplateVars
}

type EmailSignupAlertTemplateVars struct {
	Email string
}

func SendEmailSignupAlert(ctx context.Context, vars *EmailSignupAlertVars) (mes string, id string, err error) {
	// email template
	htmlEmailSignupAlertTemplate := `
<html>
	<body>
		<p>{{.Email}}</p>
	</body>
</html>
`
	// create email message
	mg := CreateMailgunClient()
	message := mg.NewMessage(vars.SenderEmail, vars.Subject, "", vars.RecipientEmails...)
	parsedHtmlTemplate, err := template.New("body").Parse(htmlEmailSignupAlertTemplate)
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	var htmlBuffer bytes.Buffer
	if err := parsedHtmlTemplate.Execute(&htmlBuffer, vars.TemplateVars); err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	message.SetHtml(htmlBuffer.String())

	// send message with 10 second timeout
	log.Printf("Sending password reset email to %v", vars.RecipientEmails)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	resp, id, err := mg.Send(ctxWithTimeout, message)
	if err != nil {
		fmt.Println(err.Error())
		return resp, id, err
	}
	log.Printf("ID: %s Resp: %s", id, resp)

	return resp, id, nil
}

func ReadString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func SendSubscriptionDowngradeAlert(ctx context.Context, newSub *models.Subscription, oldSub *models.Subscription) (mes string, id string, err error) {
	// email template
	htmlSubscriptionDowngradeAlertTemplate := `
<html>
	<body>
		<p>Subscription Downgrade Alert</p>
		<p>Old Subscription:</p>
		<p>{{.OldSubJSON}}</p>
		<p>New Subscription:</p>
		<p>{{.NewSubJSON}}</p>
	</body>
</html>
`
	type TemplateVars struct {
		OldSubJSON string
		NewSubJSON string
	}

	newSubJSON, err := json.MarshalIndent(*newSub, "", "  ")
	if err != nil {
		return "", "", err
	}

	oldSubJSON, err := json.MarshalIndent(*oldSub, "", "  ")
	if err != nil {
		return "", "", err
	}

	templateVars := &TemplateVars{
		OldSubJSON: string(oldSubJSON),
		NewSubJSON: string(newSubJSON),
	}

	// create email message
	mg := CreateMailgunClient()
	message := mg.NewMessage("alerts@teraphone.app", "Subscription Downgrade", "", "help@teraphone.app")
	parsedHtmlTemplate, err := template.New("body").Parse(htmlSubscriptionDowngradeAlertTemplate)
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	var htmlBuffer bytes.Buffer
	if err := parsedHtmlTemplate.Execute(&htmlBuffer, templateVars); err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}
	message.SetHtml(htmlBuffer.String())

	// send message with 10 second timeout
	log.Printf("Sending alert to help@teraphone.app")
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	resp, id, err := mg.Send(ctxWithTimeout, message)
	if err != nil {
		fmt.Println(err.Error())
		return resp, id, err
	}
	log.Printf("ID: %s Resp: %s", id, resp)

	return resp, id, nil
}
