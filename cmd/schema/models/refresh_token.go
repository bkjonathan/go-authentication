package models

import (
	"time"

	"github.com/google/uuid"
)

// SHA-256 rendered as hex is exactly 64 characters, always.
const TokenDigestLength = 64

// IPv6 in its longest textual form, including an IPv4 tail.
const IPAddressMaxLength = 45

type RefreshToken struct {
	Base

	UserID            uuid.UUID         `gorm:"type:uuid;not null;index:idx_refresh_tokens_user_id_revoked_at,priority:1" json:"userId"`
	FamilyID          uuid.UUID         `gorm:"type:uuid;not null;index:idx_refresh_tokens_family_id_revoked_at,priority:1" json:"familyId"`
	TokenDigest       string            `gorm:"type:char(64);not null" json:"-"`
	ExpiresAt         time.Time         `gorm:"type:timestamptz;not null;index:idx_refresh_tokens_expires_at" json:"expiresAt"`
	RevokedAt         *time.Time        `gorm:"type:timestamptz;index:idx_refresh_tokens_user_id_revoked_at,priority:2;index:idx_refresh_tokens_family_id_revoked_at,priority:2" json:"revokedAt"`
	RevokedReason     *RevocationReason `gorm:"type:varchar(32);check:chk_refresh_tokens_revoked_reason,revoked_reason IS NULL OR revoked_reason IN ('rotated','signed_out','password_changed','reuse_detected','admin_revoked')" json:"revokedReason"`
	ReplacedByTokenID *uuid.UUID        `gorm:"type:uuid" json:"replacedByTokenId"`
	UserAgent         *string           `gorm:"type:varchar(255)" json:"userAgent"`
	IPAddress         *string           `gorm:"type:varchar(45)" json:"ipAddress"`

	// Relationship
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

func (t *RefreshToken) IsRevoked() bool { return t.RevokedAt != nil }

func (t *RefreshToken) IsExpired() bool { return !t.ExpiresAt.After(time.Now()) }

// IsActive reports whether the token may still be presented for a refresh.
func (t *RefreshToken) IsActive() bool { return !t.IsRevoked() && !t.IsExpired() }

// Revoke ends this link in the chain. Set once - a second call leaves the
// original reason and timestamp in place, so the first cause is what survives
// in the audit trail.
func (t *RefreshToken) Revoke(reason RevocationReason, at time.Time) {
	if t.IsRevoked() {
		return
	}
	t.RevokedAt = &at
	t.RevokedReason = &reason
}
