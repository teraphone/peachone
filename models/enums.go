package models

type LicenseStatus int

const (
	Inactive LicenseStatus = iota
	Suspended
	Pending
	Active
)

func (s LicenseStatus) String() string {
	switch s {
	case Inactive:
		return "inactive"
	case Suspended:
		return "suspended"
	case Pending:
		return "pending"
	case Active:
		return "active"
	default:
		return "unknown"
	}
}

type LicensePlan int

const (
	None LicensePlan = iota
	Standard
	Professional
)

func (s LicensePlan) String() string {
	switch s {
	case Standard:
		return "standard"
	case Professional:
		return "professional"
	default:
		return "unknown"
	}
}

type DeploymentZone int

const (
	USWest1B DeploymentZone = iota
)

func (s DeploymentZone) String() string {
	switch s {
	case USWest1B:
		return "us-west-1b"
	default:
		return "unknown"
	}
}

type RoomType int

const (
	Public RoomType = iota
	Private
	Secret
)

func (s RoomType) String() string {
	switch s {
	case Public:
		return "public"
	case Private:
		return "private"
	case Secret:
		return "secret"
	default:
		return "unknown"
	}
}
