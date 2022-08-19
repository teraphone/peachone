package saasapi

import (
	"peachone/meta"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

func NewDefaultPipeline() runtime.Pipeline {
	po := runtime.PipelineOptions{}
	co := policy.ClientOptions{
		Telemetry: policy.TelemetryOptions{
			ApplicationID: "9ef60b2f-3246-4390-8e17-a57478e7ec45",
		},
		// Transport: http.DefaultClient,
	}
	pl := runtime.NewPipeline("saasapi", meta.Version, po, &co)

	return pl
}

func NewDefaultFulfillmentOperationsClient() *FulfillmentOperationsClient {
	pl := NewDefaultPipeline()
	return NewFulfillmentOperationsClient(pl)
}

// todo:
// - how to add authorization header to pipeline?
// -- use azidentity.NewClientSecretCredential to get access token?
