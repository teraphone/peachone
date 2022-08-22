//go:build go1.18
// +build go1.18

// Code generated by Microsoft (R) AutoRest Code Generator (autorest: 3.9.1, generator: @autorest/go@4.0.0-preview.43)
// Changes may cause incorrect behavior and will be lost if the code is regenerated.
// DO NOT EDIT.

package saasapi

const host = "https://marketplaceapi.microsoft.com/api"

// APIVersion - The request must send the following parameters as a URL Encoded form; granttype - clientcredentials; resource
// - 20e940b3-4c77-4b0b-9a53-9e16a1b010a7; clientid - AAD Registered App Client ID; client
// secret - AAD Registered App Client Secret
type APIVersion string

const (
	APIVersionTwoThousandEighteen0831 APIVersion = "2018-08-31"
)

// PossibleAPIVersionValues returns the possible values for the APIVersion const type.
func PossibleAPIVersionValues() []APIVersion {
	return []APIVersion{	
		APIVersionTwoThousandEighteen0831,
	}
}

type AllowedCustomerOperationsEnum string

const (
	AllowedCustomerOperationsEnumRead AllowedCustomerOperationsEnum = "Read"
	AllowedCustomerOperationsEnumUpdate AllowedCustomerOperationsEnum = "Update"
	AllowedCustomerOperationsEnumDelete AllowedCustomerOperationsEnum = "Delete"
)

// PossibleAllowedCustomerOperationsEnumValues returns the possible values for the AllowedCustomerOperationsEnum const type.
func PossibleAllowedCustomerOperationsEnumValues() []AllowedCustomerOperationsEnum {
	return []AllowedCustomerOperationsEnum{	
		AllowedCustomerOperationsEnumRead,
		AllowedCustomerOperationsEnumUpdate,
		AllowedCustomerOperationsEnumDelete,
	}
}

type OperationActionEnum string

const (
	OperationActionEnumUnsubscribe OperationActionEnum = "Unsubscribe"
	OperationActionEnumChangePlan OperationActionEnum = "ChangePlan"
	OperationActionEnumChangeQuantity OperationActionEnum = "ChangeQuantity"
	OperationActionEnumSuspend OperationActionEnum = "Suspend"
	OperationActionEnumReinstate OperationActionEnum = "Reinstate"
)

// PossibleOperationActionEnumValues returns the possible values for the OperationActionEnum const type.
func PossibleOperationActionEnumValues() []OperationActionEnum {
	return []OperationActionEnum{	
		OperationActionEnumUnsubscribe,
		OperationActionEnumChangePlan,
		OperationActionEnumChangeQuantity,
		OperationActionEnumSuspend,
		OperationActionEnumReinstate,
	}
}

type OperationStatusEnum string

const (
	OperationStatusEnumNotStarted OperationStatusEnum = "NotStarted"
	OperationStatusEnumInProgress OperationStatusEnum = "InProgress"
	OperationStatusEnumSucceeded OperationStatusEnum = "Succeeded"
	OperationStatusEnumFailed OperationStatusEnum = "Failed"
	OperationStatusEnumConflict OperationStatusEnum = "Conflict"
)

// PossibleOperationStatusEnumValues returns the possible values for the OperationStatusEnum const type.
func PossibleOperationStatusEnumValues() []OperationStatusEnum {
	return []OperationStatusEnum{	
		OperationStatusEnumNotStarted,
		OperationStatusEnumInProgress,
		OperationStatusEnumSucceeded,
		OperationStatusEnumFailed,
		OperationStatusEnumConflict,
	}
}

// SandboxTypeEnum - Possible Values are None, Csp (Csp sandbox purchase)
type SandboxTypeEnum string

const (
	SandboxTypeEnumNone SandboxTypeEnum = "None"
	SandboxTypeEnumCsp SandboxTypeEnum = "Csp"
)

// PossibleSandboxTypeEnumValues returns the possible values for the SandboxTypeEnum const type.
func PossibleSandboxTypeEnumValues() []SandboxTypeEnum {
	return []SandboxTypeEnum{	
		SandboxTypeEnumNone,
		SandboxTypeEnumCsp,
	}
}

// SessionModeEnum - Dry Run indicates all transactions run as Test-Mode in the commerce stack
type SessionModeEnum string

const (
	SessionModeEnumNone SessionModeEnum = "None"
	SessionModeEnumDryRun SessionModeEnum = "DryRun"
)

// PossibleSessionModeEnumValues returns the possible values for the SessionModeEnum const type.
func PossibleSessionModeEnumValues() []SessionModeEnum {
	return []SessionModeEnum{	
		SessionModeEnumNone,
		SessionModeEnumDryRun,
	}
}

// SubscriptionStatusEnum - Indicates the status of the operation.
type SubscriptionStatusEnum string

const (
	SubscriptionStatusEnumNotStarted SubscriptionStatusEnum = "NotStarted"
	SubscriptionStatusEnumPendingFulfillmentStart SubscriptionStatusEnum = "PendingFulfillmentStart"
	SubscriptionStatusEnumSubscribed SubscriptionStatusEnum = "Subscribed"
	SubscriptionStatusEnumSuspended SubscriptionStatusEnum = "Suspended"
	SubscriptionStatusEnumUnsubscribed SubscriptionStatusEnum = "Unsubscribed"
)

// PossibleSubscriptionStatusEnumValues returns the possible values for the SubscriptionStatusEnum const type.
func PossibleSubscriptionStatusEnumValues() []SubscriptionStatusEnum {
	return []SubscriptionStatusEnum{	
		SubscriptionStatusEnumNotStarted,
		SubscriptionStatusEnumPendingFulfillmentStart,
		SubscriptionStatusEnumSubscribed,
		SubscriptionStatusEnumSuspended,
		SubscriptionStatusEnumUnsubscribed,
	}
}

type TermUnitEnum string

const (
	TermUnitEnumP1M TermUnitEnum = "P1M"
	TermUnitEnumP1Y TermUnitEnum = "P1Y"
	TermUnitEnumP2Y TermUnitEnum = "P2Y"
	TermUnitEnumP3Y TermUnitEnum = "P3Y"
)

// PossibleTermUnitEnumValues returns the possible values for the TermUnitEnum const type.
func PossibleTermUnitEnumValues() []TermUnitEnum {
	return []TermUnitEnum{	
		TermUnitEnumP1M,
		TermUnitEnumP1Y,
		TermUnitEnumP2Y,
		TermUnitEnumP3Y,
	}
}

type UpdateOperationStatusEnum string

const (
	UpdateOperationStatusEnumSuccess UpdateOperationStatusEnum = "Success"
	UpdateOperationStatusEnumFailure UpdateOperationStatusEnum = "Failure"
)

// PossibleUpdateOperationStatusEnumValues returns the possible values for the UpdateOperationStatusEnum const type.
func PossibleUpdateOperationStatusEnumValues() []UpdateOperationStatusEnum {
	return []UpdateOperationStatusEnum{	
		UpdateOperationStatusEnumSuccess,
		UpdateOperationStatusEnumFailure,
	}
}
