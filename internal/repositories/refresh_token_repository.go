package repositories

import (
	"context"
	"time"

	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	FindByDigest(ctx context.Context, digest string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID, reason models.RevocationReason, at time.Time, replacedBy *uuid.UUID) (bool, error)
	RevokeFamily(ctx context.Context, userID, familyID uuid.UUID, reason models.RevocationReason, at time.Time) (int64, error)
	RevokeUser(ctx context.Context, userID uuid.UUID, reason models.RevocationReason, at time.Time, exceptFamilyID *uuid.UUID) (int64, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.RefreshToken, error)
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// FindByDigest returns the row whatever state it is in. A revoked token being
// presented is the signal rotation exists to catch, so the caller has to see it
// rather than get "not found".
func (r *refreshTokenRepository) FindByDigest(ctx context.Context, digest string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.WithContext(ctx).First(&token, "token_digest = ?", digest).Error; err != nil {
		return nil, translate(err)
	}
	return &token, nil
}

// Revoke ends one token and reports whether this call is what ended it. The
// `revoked_at IS NULL` guard makes it the atomic step rotation turns on: two
// requests presenting the same token race here, and only one wins.
func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID, reason models.RevocationReason, at time.Time, replacedBy *uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(map[string]any{
			"revoked_at":           at,
			"revoked_reason":       reason,
			"replaced_by_token_id": replacedBy,
		})
	return result.RowsAffected == 1, result.Error
}

// RevokeFamily ends a whole session. Scoped by user as well as family so a
// caller cannot end someone else's session by guessing an id.
func (r *refreshTokenRepository) RevokeFamily(ctx context.Context, userID, familyID uuid.UUID, reason models.RevocationReason, at time.Time) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND family_id = ? AND revoked_at IS NULL", userID, familyID).
		Updates(map[string]any{"revoked_at": at, "revoked_reason": reason})
	return result.RowsAffected, result.Error
}

// RevokeUser signs every device out at once, optionally sparing the one the
// request came from - "sign out my other devices" and "the password changed,
// nothing survives" are the same query with and without that exception.
func (r *refreshTokenRepository) RevokeUser(ctx context.Context, userID uuid.UUID, reason models.RevocationReason, at time.Time, exceptFamilyID *uuid.UUID) (int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID)
	if exceptFamilyID != nil {
		query = query.Where("family_id <> ?", *exceptFamilyID)
	}
	result := query.Updates(map[string]any{"revoked_at": at, "revoked_reason": reason})
	return result.RowsAffected, result.Error
}

// ListActiveByUser returns one row per live session: rotation leaves exactly
// one unrevoked token in each family, so this is already the session list.
func (r *refreshTokenRepository) ListActiveByUser(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}
