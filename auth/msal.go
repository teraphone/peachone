package auth

import "os"

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
	Scopes:       []string{"User.Read", "openid", "profile", "email", "offline_access"},
	RedirectURI:  "http://localhost:8080",
	ClientSecret: os.Getenv("MSAL_CLIENT_SECRET"),
}
