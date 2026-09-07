// Package repositories holds every query in the application. Nothing above it
// knows gorm, and nothing in it knows HTTP.
package repositories

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// The two outcomes a service reacts to differently from any other failure,
// named here so nothing above this package has to import gorm to recognise them.
var (
	ErrNotFound  = errors.New("record not found")
	ErrDuplicate = errors.New("record already exists")
)

type Store struct {
	db            *gorm.DB
	Users         UserRepository
	RefreshTokens RefreshTokenRepository
}

func NewStore(db *gorm.DB) *Store {
	return &Store{
		db:            db,
		Users:         &userRepository{db: db},
		RefreshTokens: &refreshTokenRepository{db: db},
	}
}

// Atomic runs fn against a store bound to one transaction, so every repository
// call inside it commits or rolls back together. Returning an error from fn
// rolls back.
func (s *Store) Atomic(ctx context.Context, fn func(tx *Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewStore(tx))
	})
}

// translate turns gorm's sentinel into the package's own, and leaves every
// other error alone.
func translate(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrDuplicate
	default:
		return err
	}
}
