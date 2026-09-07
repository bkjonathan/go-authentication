package models

import "slices"

type Role string

const (
	RoleOwner   Role = "owner"
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleStaff   Role = "staff"
)

// AllRoles is every role this build declares, most privileged first.
var AllRoles = []Role{
	RoleOwner,
	RoleAdmin,
	RoleManager,
	RoleStaff,
}

func (r Role) String() string { return string(r) }

// IsValid reports whether a stored string is a role this build knows about.
func (r Role) IsValid() bool { return slices.Contains(AllRoles, r) }
