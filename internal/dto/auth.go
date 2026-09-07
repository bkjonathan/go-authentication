package dto

import (
	"strings"
	"time"

	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email,max=320"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
	FirstName string `json:"firstName" binding:"required,max=100"`
	LastName  string `json:"lastName" binding:"required,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=320"`
	Password string `json:"password" binding:"required,max=72"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required,max=72"`
	NewPassword     string `json:"newPassword" binding:"required,min=8,max=72,nefield=CurrentPassword"`
}

type UserResponse struct {
	ID          uuid.UUID         `json:"id"`
	Email       string            `json:"email"`
	FirstName   string            `json:"firstName"`
	LastName    string            `json:"lastName"`
	FullName    string            `json:"fullName"`
	PhotoKey    *string           `json:"photoKey"`
	Role        models.Role       `json:"role"`
	Status      models.UserStatus `json:"status"`
	LastLoginAt *time.Time        `json:"lastLoginAt"`
	CreatedAt   time.Time         `json:"createdAt"`
}

type TokenPair struct {
	TokenType        string    `json:"tokenType"`
	AccessToken      string    `json:"accessToken"`
	RefreshToken     string    `json:"refreshToken"`
	ExpiresAt        time.Time `json:"expiresAt"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}

type AuthResponse struct {
	User   *UserResponse `json:"user"`
	Tokens TokenPair     `json:"tokens"`
}

func NewUserResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		FullName:    user.FullName(),
		PhotoKey:    user.PhotoKey,
		Role:        user.Role,
		Status:      user.Status,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

func (r *RegisterRequest) Normalize() {
	r.Email = models.NormalizeEmail(r.Email)
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
}

func (r *LoginRequest) Normalize() {
	r.Email = models.NormalizeEmail(r.Email)
}
