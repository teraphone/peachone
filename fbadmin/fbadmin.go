package fbadmin

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

var App *firebase.App

// For creating custom tokens: https://firebase.google.com/docs/auth/admin/create-custom-tokens
var AuthClient *auth.Client

func InitFirebaseApp(ctx context.Context) {
	conf := &firebase.Config{
		ServiceAccountID: "firebase-adminsdk-7dp4y@livekit-demo.iam.gserviceaccount.com",
	}

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	App = app
}

func InitFirebaseAuthClient(ctx context.Context) {
	client, err := App.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	AuthClient = client
}
