package models

import "slices"

type RevocationReason string

const (
	RevocationReasonRotated         RevocationReason = "rotated"
	RevocationReasonSignedOut       RevocationReason = "signed_out"
	RevocationReasonPasswordChanged RevocationReason = "password_changed"
	RevocationReasonReuseDetected   RevocationReason = "reuse_detected"
	RevocationReasonAdminRevoked    RevocationReason = "admin_revoked"
)

var AllRevocationReasons = []RevocationReason{
	RevocationReasonRotated,
	RevocationReasonSignedOut,
	RevocationReasonPasswordChanged,
	RevocationReasonReuseDetected,
	RevocationReasonAdminRevoked,
}

func (r RevocationReason) String() string { return string(r) }

// IsValid reports whether a stored string is a reason this build knows about.
func (r RevocationReason) IsValid() bool { return slices.Contains(AllRevocationReasons, r) }
