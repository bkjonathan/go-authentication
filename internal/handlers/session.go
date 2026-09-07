package handlers

import (
	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/bkjonathan/go-authentication/internal/dto"
	"github.com/bkjonathan/go-authentication/internal/middleware"
	"github.com/bkjonathan/go-authentication/internal/services"
	"github.com/bkjonathan/go-authentication/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SessionHandler struct {
	sessions *services.SessionService
}

func NewSessionHandler(sessions *services.SessionService) *SessionHandler {
	return &SessionHandler{sessions: sessions}
}

// List answers "where am I signed in?", with the entry the request came from
// marked so a client can label it and think twice about ending it.
func (h *SessionHandler) List(ctx *gin.Context) error {
	principal, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return err
	}

	sessions, err := h.sessions.List(ctx.Request.Context(), principal.UserID, principal.SessionID)
	if err != nil {
		return err
	}

	utils.OK(ctx, sessions)
	return nil
}

func (h *SessionHandler) Revoke(ctx *gin.Context) error {
	principal, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return err
	}

	sessionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return apperror.BadRequest("invalid_session_id", "That is not a valid session id.")
	}

	if err := h.sessions.Revoke(ctx.Request.Context(), principal.UserID, sessionID); err != nil {
		return err
	}

	utils.NoContent(ctx)
	return nil
}

func (h *SessionHandler) RevokeOthers(ctx *gin.Context) error {
	principal, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return err
	}

	revoked, err := h.sessions.RevokeOthers(ctx.Request.Context(), principal.UserID, principal.SessionID)
	if err != nil {
		return err
	}

	utils.OK(ctx, dto.RevokedSessionsResponse{Revoked: revoked})
	return nil
}
