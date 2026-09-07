package handlers

import (
	"github.com/bkjonathan/go-authentication/internal/dto"
	"github.com/bkjonathan/go-authentication/internal/middleware"
	"github.com/bkjonathan/go-authentication/internal/services"
	"github.com/bkjonathan/go-authentication/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(ctx *gin.Context) error {
	var req dto.RegisterRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return err
	}

	response, err := h.auth.Register(ctx.Request.Context(), req, clientInfo(ctx))
	if err != nil {
		return err
	}

	utils.Created(ctx, response)
	return nil
}

func (h *AuthHandler) Login(ctx *gin.Context) error {
	var req dto.LoginRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return err
	}

	response, err := h.auth.Login(ctx.Request.Context(), req, clientInfo(ctx))
	if err != nil {
		return err
	}

	utils.OK(ctx, response)
	return nil
}

func (h *AuthHandler) Refresh(ctx *gin.Context) error {
	var req dto.RefreshRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return err
	}

	response, err := h.auth.Refresh(ctx.Request.Context(), req, clientInfo(ctx))
	if err != nil {
		return err
	}

	utils.OK(ctx, response)
	return nil
}

// Logout takes the refresh token rather than the access token, because the
// refresh token is what names the session being ended - and it works whether or
// not the access token has already expired.
func (h *AuthHandler) Logout(ctx *gin.Context) error {
	var req dto.RefreshRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return err
	}

	if err := h.auth.Logout(ctx.Request.Context(), req); err != nil {
		return err
	}

	utils.NoContent(ctx)
	return nil
}

func (h *AuthHandler) Me(ctx *gin.Context) error {
	principal, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return err
	}

	user, err := h.auth.Profile(ctx.Request.Context(), principal.UserID)
	if err != nil {
		return err
	}

	utils.OK(ctx, user)
	return nil
}

func (h *AuthHandler) ChangePassword(ctx *gin.Context) error {
	principal, err := middleware.CurrentPrincipal(ctx)
	if err != nil {
		return err
	}

	var req dto.ChangePasswordRequest
	if err := utils.BindJSON(ctx, &req); err != nil {
		return err
	}

	response, err := h.auth.ChangePassword(ctx.Request.Context(), principal.UserID, req, clientInfo(ctx))
	if err != nil {
		return err
	}

	utils.OK(ctx, response)
	return nil
}
