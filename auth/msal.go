package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type MSALConfig struct {
	ClientID            string   `json:"clientId"`
	Authority           string   `json:"authority"`
	Scopes              []string `json:"scopes"`
	Username            string   `json:"username"`
	Password            string   `json:"password"`
	RedirectURI         string   `json:"redirectUri"`
	CodeChallenge       string   `json:"codeChallenge"`
	CodeChallengeMethod string   `json:"codeChallengeMethod"`
	State               string   `json:"state"`
	ClientSecret        string   `json:"clientSecret"`
	Thumbprint          string   `json:"thumbprint"`
	PemData             string   `json:"pemFile"`
}

var Config = &MSALConfig{
	ClientID:     "9ef60b2f-3246-4390-8e17-a57478e7ec45",
	Authority:    "https://login.microsoftonline.com/common",
	Scopes:       []string{"User.Read", "openid", "profile", "email"},
	RedirectURI:  "http://localhost:8080",
	ClientSecret: os.Getenv("MSAL_CLIENT_SECRET"),
}

type TokenCredentialHelper struct {
	app             *confidential.Client
	userAccessToken string
	authResult      *confidential.AuthResult
}

// implements azcore.TokenCredential interface
func (helper *TokenCredentialHelper) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	authResult, err := helper.app.AcquireTokenOnBehalfOf(ctx, helper.userAccessToken, Config.Scopes)
	if err != nil {
		fmt.Println("Error acquiring token on-behalf-of user:", err)
		return azcore.AccessToken{}, err
	}
	helper.authResult = &authResult

	accessToken := azcore.AccessToken{
		Token:     authResult.AccessToken,
		ExpiresOn: authResult.ExpiresOn,
	}

	return accessToken, nil

}

func NewTokenCredentialHelper(userAccessToken string) (*TokenCredentialHelper, error) {
	cred, err := confidential.NewCredFromSecret(Config.ClientSecret)
	if err != nil {
		fmt.Println("Error creating credential:", err)
		return nil, err
	}

	app, err := confidential.New(
		Config.ClientID, cred,
		confidential.WithAuthority(Config.Authority),
	)
	if err != nil {
		fmt.Println("Error creating auth client:", err)
		return nil, err
	}

	return &TokenCredentialHelper{
		app:             &app,
		userAccessToken: userAccessToken,
	}, nil
}
