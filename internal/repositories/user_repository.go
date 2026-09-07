package repositories

import (
	"context"
	"time"

	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	EmailTaken(ctx context.Context, email string) (bool, error)
	RecordFailedLogin(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error
	RecordSuccessfulLogin(ctx context.Context, id uuid.UUID, at time.Time) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string, changedAt time.Time) error
}

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return translate(r.db.WithContext(ctx).Create(user).Error)
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}
	return &user, nil
}

// FindByEmail normalises the address the same way the model does on write, so
// a lookup can never miss a row over spelling.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", models.NormalizeEmail(email)).Error; err != nil {
		return nil, translate(err)
	}
	return &user, nil
}

// EmailTaken is advisory: the unique index is what actually decides, and a
// concurrent registration is caught there.
func (r *userRepository) EmailTaken(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", models.NormalizeEmail(email)).
		Count(&count).Error
	return count > 0, err
}

// RecordFailedLogin writes the counter the caller worked out rather than
// incrementing in SQL, because the lockout decision needs the resulting value
// anyway.
func (r *userRepository) RecordFailedLogin(ctx context.Context, id uuid.UUID, attempts int, lockedUntil *time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"failed_login_attempts": attempts,
			"locked_until":          lockedUntil,
		}).Error
}

func (r *userRepository) RecordSuccessfulLogin(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_login_at":         at,
			"failed_login_attempts": 0,
			"locked_until":          nil,
		}).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string, changedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password_hash":       passwordHash,
			"password_changed_at": changedAt,
		}).Error
}
