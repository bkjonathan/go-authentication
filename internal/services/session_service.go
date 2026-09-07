package services

import (
	"context"
	"time"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/bkjonathan/go-authentication/internal/dto"
	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/bkjonathan/go-authentication/internal/repositories"
	"github.com/google/uuid"
)

var ErrSessionNotFound = apperror.NotFound("session_not_found", "That session is not signed in.")

// SessionService answers "where am I signed in, and sign that one out". A
// session is a refresh token family: every device gets its own family at
// sign-in, rotation keeps exactly one live token in it, and revoking the family
// is what ends the session.
type SessionService struct {
	store *repositories.Store
}

func NewSessionService(store *repositories.Store) *SessionService {
	return &SessionService{store: store}
}

func (s *SessionService) List(ctx context.Context, userID, currentSessionID uuid.UUID) ([]dto.SessionResponse, error) {
	tokens, err := s.store.RefreshTokens.ListActiveByUser(ctx, userID, time.Now())
	if err != nil {
		return nil, apperror.Internal(err)
	}

	sessions := make([]dto.SessionResponse, 0, len(tokens))
	for i := range tokens {
		sessions = append(sessions, dto.NewSessionResponse(&tokens[i], currentSessionID))
	}
	return sessions, nil
}

// Revoke signs one device out. The query is scoped to the caller, so a session
// id belonging to someone else is indistinguishable from one that does not
// exist.
func (s *SessionService) Revoke(ctx context.Context, userID, sessionID uuid.UUID) error {
	revoked, err := s.store.RefreshTokens.RevokeFamily(ctx, userID, sessionID, models.RevocationReasonSignedOut, time.Now())
	if err != nil {
		return apperror.Internal(err)
	}
	if revoked == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// RevokeOthers signs out every device except the one asking.
func (s *SessionService) RevokeOthers(ctx context.Context, userID, currentSessionID uuid.UUID) (int64, error) {
	revoked, err := s.store.RefreshTokens.RevokeUser(ctx, userID, models.RevocationReasonSignedOut, time.Now(), &currentSessionID)
	if err != nil {
		return 0, apperror.Internal(err)
	}
	return revoked, nil
}
