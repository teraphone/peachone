package saasapi

import (
	"os"
	"peachone/meta"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	AppId    = "9ef60b2f-3246-4390-8e17-a57478e7ec45"
	TenantId = "a6db0c33-ff9b-49f7-be5a-a5c50ee313cd"
)

func NewDefaultPipeline() (*runtime.Pipeline, error) {
	co := policy.ClientOptions{
		Telemetry: policy.TelemetryOptions{
			ApplicationID: AppId,
		},
	}

	cred, err := azidentity.NewClientSecretCredential(
		TenantId,
		AppId,
		os.Getenv("MSAL_CLIENT_SECRET"),
		nil,
	)
	if err != nil {
		return nil, err
	}

	scopes := []string{"20e940b3-4c77-4b0b-9a53-9e16a1b010a7/.default"}
	tokenPolicy := runtime.NewBearerTokenPolicy(cred, scopes, nil)
	po := runtime.PipelineOptions{
		PerRetry: []policy.Policy{tokenPolicy},
	}
	pl := runtime.NewPipeline("saasapi", meta.Version, po, &co)

	return &pl, nil
}

func NewDefaultFulfillmentOperationsClient() (*FulfillmentOperationsClient, error) {
	pl, err := NewDefaultPipeline()
	if err != nil {
		return nil, err
	}
	return NewFulfillmentOperationsClient(*pl), nil
}
