package models

import (
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
)

const EmailMaxLength = 320
const PhotoKeyMaxLength = 255

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

type User struct {
	Base

	Email               string     `gorm:"type:varchar(320);not null;index:idx_users_email,unique,where:deleted_at IS NULL" json:"email"`
	PasswordHash        string     `gorm:"type:varchar(255);not null" json:"-"`
	FirstName           string     `gorm:"type:varchar(100);not null" json:"firstName"`
	LastName            string     `gorm:"type:varchar(100);not null" json:"lastName"`
	PhotoKey            *string    `gorm:"type:varchar(255)" json:"photoKey"`
	Role                Role       `gorm:"type:varchar(32);not null;default:staff;check:chk_users_role,role IN ('owner','admin','manager','staff')" json:"role"`
	Status              UserStatus `gorm:"type:varchar(32);not null;default:active;check:chk_users_status,status IN ('active','pending_verification','suspended','deactivated')" json:"status"`
	FailedLoginAttempts int        `gorm:"type:integer;not null;default:0" json:"-"`
	LockedUntil         *time.Time `gorm:"type:timestamptz" json:"-"`
	LastLoginAt         *time.Time `gorm:"type:timestamptz" json:"lastLoginAt"`
	PasswordChangedAt   *time.Time `gorm:"type:timestamptz" json:"-"`
}

func (User) TableName() string { return "users" }

// BeforeSave applies the normalisation every lookup also passes through, so a
// row can never be written in a spelling a lookup would miss.
func (u *User) BeforeSave(*gorm.DB) error {
	u.Email = NormalizeEmail(u.Email)
	return nil
}

func (u *User) FullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

// IsLocked reports whether a lockout is currently in force. A lapsed one reads
// as unlocked.
func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && u.LockedUntil.After(time.Now())
}

// IsActive reports whether the account's status permits signing in. It says
// nothing about a lockout - check IsLocked too.
func (u *User) IsActive() bool {
	return u.Status.AllowsSignIn()
}

// HasRole reports whether the account holds any of the roles given, which is
// the shape a route guard asks the question in: "owner or admin may do this".
func (u *User) HasRole(roles ...Role) bool {
	return slices.Contains(roles, u.Role)
}
