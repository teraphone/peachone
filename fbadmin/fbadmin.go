package fbadmin

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/db"
	"google.golang.org/api/option"
)

var App *firebase.App

// For creating custom tokens: https://firebase.google.com/docs/auth/admin/create-custom-tokens
var AuthClient *auth.Client

var DBClient *db.Client

func InitFirebaseApp(ctx context.Context) {
	conf := &firebase.Config{
		DatabaseURL: "https://dally-arty.firebaseio.com",
	}

	SERVICE_ACCOUNT_JSON := os.Getenv("SERVICE_ACCOUNT_JSON")
	if SERVICE_ACCOUNT_JSON == "" {
		conf.ServiceAccountID = "firebase-adminsdk-7dp4y@livekit-demo.iam.gserviceaccount.com"
		app, err := firebase.NewApp(ctx, conf)
		if err != nil {
			log.Fatalf("error initializing app: %v\n", err)
		}
		App = app
	} else {
		opt := option.WithCredentialsFile(SERVICE_ACCOUNT_JSON)
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Fatalf("error initializing app: %v\n", err)
		}
		App = app
	}

	client, err := App.Database(ctx)
	if err != nil {
		log.Fatal("Error initializing database client:", err)
	}
	DBClient = client
}

func InitFirebaseAuthClient(ctx context.Context) {
	client, err := App.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	AuthClient = client
}
