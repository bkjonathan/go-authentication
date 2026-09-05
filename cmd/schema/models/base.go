package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	CreatedAt time.Time      `gorm:"type:timestamptz;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"type:timestamptz;index" json:"-"`
	Version   int            `gorm:"type:integer;not null;default:1" json:"version"`
}

// BeforeCreate mints the identifier when the caller has not already chosen one,
// which is the whole point of a UUID key - a row can be referenced before it
// exists.
func (b *Base) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	if b.Version == 0 {
		b.Version = 1
	}
	return nil
}

// BeforeUpdate bumps the version in SQL rather than from the in-memory value,
// so a stale struct cannot roll the counter backwards.
func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("version", gorm.Expr("version + 1"))
	return nil
}
