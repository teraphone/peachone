package routes

import (
	"context"
	"fmt"
	"log"
	"os"
	"peachone/fbadmin"
	"peachone/models"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	lksdk "github.com/livekit/server-sdk-go"

	"github.com/livekit/protocol/auth"

	"github.com/mailgun/mailgun-go/v4"
)

func getIDFromJWT(c *fiber.Ctx) (uint, error) {
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	id := uint(claims["id"].(float64))
	expiration := int64(claims["expiration"].(float64))
	if time.Now().Unix() > expiration {
		return id, fmt.Errorf("token expired")
	}

	return id, nil
}

func createJWTToken(user *models.User) (string, int64, error) {
	expiration := time.Now().Add(time.Hour * 24).Unix()
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	claims["expiration"] = expiration
	SIGNING_KEY := os.Getenv("SIGNING_KEY")
	tokenString, err := token.SignedString([]byte(SIGNING_KEY))
	if err != nil {
		return "", 0, err
	}

	return tokenString, expiration, nil
}

func createFirebaseAuthToken(ctx context.Context, user *models.User) (string, error) {
	uid := strconv.FormatUint(uint64(user.ID), 10)
	token, err := fbadmin.AuthClient.CustomToken(ctx, uid)
	if err != nil {
		log.Printf("error minting custom token: %v\n", err)
		return "", err
	}

	return token, nil
}

func createLiveKitJoinToken(room_user *models.RoomUser, group_id uint, room_id uint, user_id uint) (string, error) {
	LIVEKIT_KEY := os.Getenv("LIVEKIT_KEY")
	LIVEKIT_SECRET := os.Getenv("LIVEKIT_SECRET")
	at := auth.NewAccessToken(LIVEKIT_KEY, LIVEKIT_SECRET)
	grant := &auth.VideoGrant{
		RoomCreate: false,
		RoomList:   false,
		RoomRecord: false,

		RoomAdmin: room_user.RoomRoleID > models.RoomRoleMap["member"],
		RoomJoin:  room_user.CanJoin,
		Room:      EncodeRoomName(group_id, room_id),

		CanPublish:   &room_user.CanJoin,
		CanSubscribe: &room_user.CanJoin,
	}
	at.AddGrant(grant).
		SetIdentity(strconv.Itoa(int(user_id))).
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

func EncodeRoomName(group_id uint, room_id uint) string {
	return strconv.Itoa(int(group_id)) + "/" + strconv.Itoa(int(room_id))
}

func DecodeRoomName(name string) (uint, uint, error) {
	split := strings.Split(name, "/")
	if len(split) != 2 {
		return 0, 0, fmt.Errorf("invalid room name: %s", name)
	}

	group_id, err := strconv.Atoi(split[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid group id: %s", split[0])
	}

	room_id, err := strconv.Atoi(split[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid room id: %s", split[1])
	}

	return uint(group_id), uint(room_id), nil
}

func CreateMailgunClient() *mailgun.MailgunImpl {
	MG_DOMAIN := os.Getenv("MG_DOMAIN")
	MG_API_KEY := os.Getenv("MG_API_KEY")

	mg := mailgun.NewMailgun(MG_DOMAIN, MG_API_KEY)
	return mg
}
