package models

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	Base

	UserID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_password_reset_tokens_user_id_used_at,priority:1" json:"userId"`
	TokenDigest     string     `gorm:"type:char(64);not null;uniqueIndex:idx_password_reset_tokens_token_digest" json:"-"`
	ExpiresAt       time.Time  `gorm:"type:timestamptz;not null;index:idx_password_reset_tokens_expires_at" json:"expiresAt"`
	UsedAt          *time.Time `gorm:"type:timestamptz;index:idx_password_reset_tokens_user_id_used_at,priority:2" json:"usedAt"`
	RequestedFromIP *string    `gorm:"type:varchar(45)" json:"requestedFromIp"`

	// Relationship
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }

func (t *PasswordResetToken) IsUsed() bool { return t.UsedAt != nil }

func (t *PasswordResetToken) IsExpired() bool { return !t.ExpiresAt.After(time.Now()) }

// IsRedeemable is advisory only. Redemption itself must be an atomic
// UPDATE ... WHERE used_at IS NULL, or two clicks on the same link both win.
func (t *PasswordResetToken) IsRedeemable() bool { return !t.IsUsed() && !t.IsExpired() }
