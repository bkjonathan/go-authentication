package dto

import (
	"time"

	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/google/uuid"
)

// ClientInfo is what the request itself says about the device making it. It is
// recorded against every refresh token so a user can recognise their own
// sessions in a list.
type ClientInfo struct {
	IPAddress string
	UserAgent string
}

type SessionResponse struct {
	ID         uuid.UUID `json:"id"`
	Current    bool      `json:"current"`
	UserAgent  *string   `json:"userAgent"`
	IPAddress  *string   `json:"ipAddress"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// NewSessionResponse describes a session by its live refresh token. A session
// is a token family, and rotation leaves exactly one unrevoked token in it, so
// that row is both the session's identity and its most recent activity.
func NewSessionResponse(token *models.RefreshToken, currentSessionID uuid.UUID) SessionResponse {
	return SessionResponse{
		ID:         token.FamilyID,
		Current:    token.FamilyID == currentSessionID,
		UserAgent:  token.UserAgent,
		IPAddress:  token.IPAddress,
		LastUsedAt: token.CreatedAt,
		ExpiresAt:  token.ExpiresAt,
	}
}

type RevokedSessionsResponse struct {
	Revoked int64 `json:"revoked"`
}
