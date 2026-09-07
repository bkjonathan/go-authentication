package models

import "slices"

type UserStatus string

const (
	UserStatusActive              UserStatus = "active"
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusDeactivated         UserStatus = "deactivated"
)

var SignInAllowedStatuses = []UserStatus{UserStatusActive}

// AllUserStatuses is every status this build declares.
var AllUserStatuses = []UserStatus{
	UserStatusActive,
	UserStatusPendingVerification,
	UserStatusSuspended,
	UserStatusDeactivated,
}

func (s UserStatus) String() string { return string(s) }

// IsValid reports whether a stored string is a status this build knows about.
func (s UserStatus) IsValid() bool { return slices.Contains(AllUserStatuses, s) }

// AllowsSignIn reports whether an account in this state may authenticate.
func (s UserStatus) AllowsSignIn() bool { return slices.Contains(SignInAllowedStatuses, s) }
