// Package services holds the business rules. Nothing here knows gin, and
// nothing here writes SQL.
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/bkjonathan/go-authentication/internal/config"
	"github.com/bkjonathan/go-authentication/internal/dto"
	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/bkjonathan/go-authentication/internal/repositories"
	"github.com/bkjonathan/go-authentication/internal/utils"
	"github.com/google/uuid"
)

const tokenTypeBearer = "Bearer"

var (
	ErrInvalidCredentials  = apperror.Unauthorized("invalid_credentials", "Email or password is incorrect.")
	ErrEmailTaken          = apperror.Conflict("email_taken", "That email address is already registered.")
	ErrInvalidRefreshToken = apperror.Unauthorized("invalid_refresh_token", "Your session has ended. Sign in again.")
)

type AuthService struct {
	store      *repositories.Store
	tokens     *utils.TokenIssuer
	hasher     *utils.PasswordHasher
	refreshTTL time.Duration
	lockout    config.AuthConfig
}

func NewAuthService(store *repositories.Store, tokens *utils.TokenIssuer, hasher *utils.PasswordHasher, refreshTTL time.Duration, lockout config.AuthConfig) *AuthService {
	return &AuthService{
		store:      store,
		tokens:     tokens,
		hasher:     hasher,
		refreshTTL: refreshTTL,
		lockout:    lockout,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, client dto.ClientInfo) (*dto.AuthResponse, error) {
	taken, err := s.store.Users.EmailTaken(ctx, req.Email)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	if taken {
		return nil, ErrEmailTaken
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         models.RoleStaff,
		Status:       models.UserStatusActive,
	}

	var pair *dto.TokenPair
	err = s.store.Atomic(ctx, func(tx *repositories.Store) error {
		// The check above races; the unique index is what actually decides, and
		// this is where losing that race surfaces.
		if err := tx.Users.Create(ctx, user); err != nil {
			if errors.Is(err, repositories.ErrDuplicate) {
				return ErrEmailTaken
			}
			return apperror.Internal(err)
		}
		pair, err = s.openSession(ctx, tx, user, uuid.New(), client, uuid.Nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{User: dto.NewUserResponse(user), Tokens: *pair}, nil
}

// Login opens a session without touching any other, which is what lets one
// account be signed in on several devices at once.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, client dto.ClientInfo) (*dto.AuthResponse, error) {
	user, err := s.store.Users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			// Spend what a real comparison costs, so response time does not
			// answer "is this address registered?".
			s.hasher.VerifyDecoy(req.Password)
			return nil, ErrInvalidCredentials
		}
		return nil, apperror.Internal(err)
	}

	// The password is checked before anything else is reported, so a lockout or
	// a suspension is only ever disclosed to whoever already knows the password.
	if !s.hasher.Verify(user.PasswordHash, req.Password) {
		return nil, s.recordFailure(ctx, user)
	}
	if user.IsLocked() {
		return nil, errAccountLocked(*user.LockedUntil)
	}
	if !user.IsActive() {
		return nil, errAccountInactive(user.Status)
	}

	now := time.Now()
	if err := s.store.Users.RecordSuccessfulLogin(ctx, user.ID, now); err != nil {
		return nil, apperror.Internal(err)
	}
	user.LastLoginAt = &now

	var pair *dto.TokenPair
	err = s.store.Atomic(ctx, func(tx *repositories.Store) error {
		pair, err = s.openSession(ctx, tx, user, uuid.New(), client, uuid.Nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{User: dto.NewUserResponse(user), Tokens: *pair}, nil
}

// Refresh rotates: the presented token is spent and a replacement takes its
// place in the same family. A token that has already been spent coming back is
// the one thing rotation is for - it means a copy is in circulation, so the
// whole family goes.
func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest, client dto.ClientInfo) (*dto.AuthResponse, error) {
	now := time.Now()

	token, err := s.store.RefreshTokens.FindByDigest(ctx, utils.DigestToken(req.RefreshToken))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, apperror.Internal(err)
	}

	if token.IsRevoked() {
		if _, err := s.store.RefreshTokens.RevokeFamily(ctx, token.UserID, token.FamilyID, models.RevocationReasonReuseDetected, now); err != nil {
			return nil, apperror.Internal(err)
		}
		return nil, ErrInvalidRefreshToken
	}
	if token.IsExpired() {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.store.Users.FindByID(ctx, token.UserID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, apperror.Internal(err)
	}
	if user.IsLocked() {
		return nil, errAccountLocked(*user.LockedUntil)
	}
	if !user.IsActive() {
		return nil, errAccountInactive(user.Status)
	}

	var pair *dto.TokenPair
	err = s.store.Atomic(ctx, func(tx *repositories.Store) error {
		// Minted before the revoke so the spent token can point at what
		// replaced it, which is what makes the chain readable afterwards.
		replacementID := uuid.New()

		spent, err := tx.RefreshTokens.Revoke(ctx, token.ID, models.RevocationReasonRotated, now, &replacementID)
		if err != nil {
			return apperror.Internal(err)
		}
		// Another request spent it between the read and here. One rotation per
		// token, so this one does not get a replacement.
		if !spent {
			return ErrInvalidRefreshToken
		}

		pair, err = s.openSession(ctx, tx, user, token.FamilyID, client, replacementID)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{User: dto.NewUserResponse(user), Tokens: *pair}, nil
}

// Logout ends the session the token belongs to, and only that one - the other
// devices stay signed in. A token that is already gone is not an error: asking
// to be signed out twice should read as success both times.
func (s *AuthService) Logout(ctx context.Context, req dto.RefreshRequest) error {
	token, err := s.store.RefreshTokens.FindByDigest(ctx, utils.DigestToken(req.RefreshToken))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil
		}
		return apperror.Internal(err)
	}
	if _, err := s.store.RefreshTokens.RevokeFamily(ctx, token.UserID, token.FamilyID, models.RevocationReasonSignedOut, time.Now()); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

// ChangePassword ends every session, including the caller's, then opens one
// fresh session for the device that made the change. Whoever else had a token
// has to sign in again with a password they no longer know.
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req dto.ChangePasswordRequest, client dto.ClientInfo) (*dto.AuthResponse, error) {
	user, err := s.store.Users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, apperror.Internal(err)
	}
	if !s.hasher.Verify(user.PasswordHash, req.CurrentPassword) {
		return nil, ErrInvalidCredentials
	}

	hash, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	now := time.Now()
	var pair *dto.TokenPair
	err = s.store.Atomic(ctx, func(tx *repositories.Store) error {
		if err := tx.Users.UpdatePassword(ctx, user.ID, hash, now); err != nil {
			return apperror.Internal(err)
		}
		if _, err := tx.RefreshTokens.RevokeUser(ctx, user.ID, models.RevocationReasonPasswordChanged, now, nil); err != nil {
			return apperror.Internal(err)
		}
		pair, err = s.openSession(ctx, tx, user, uuid.New(), client, uuid.Nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{User: dto.NewUserResponse(user), Tokens: *pair}, nil
}

func (s *AuthService) Profile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.store.Users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, apperror.NotFound("user_not_found", "That account no longer exists.")
		}
		return nil, apperror.Internal(err)
	}
	return dto.NewUserResponse(user), nil
}

// openSession writes one refresh token into familyID and mints the access token
// that names it. A fresh familyID starts a new session; an existing one
// continues a session that is being rotated. Pass tokenID when the row has to
// have an identity the caller already handed out - rotation links to it.
func (s *AuthService) openSession(ctx context.Context, tx *repositories.Store, user *models.User, familyID uuid.UUID, client dto.ClientInfo, tokenID uuid.UUID) (*dto.TokenPair, error) {
	secret, digest, err := utils.NewOpaqueToken()
	if err != nil {
		return nil, apperror.Internal(err)
	}

	token := &models.RefreshToken{
		UserID:      user.ID,
		FamilyID:    familyID,
		TokenDigest: digest,
		ExpiresAt:   time.Now().Add(s.refreshTTL),
		UserAgent:   trimmedColumn(client.UserAgent, models.UserAgentMaxLength),
		IPAddress:   trimmedColumn(client.IPAddress, models.IPAddressMaxLength),
	}
	token.ID = tokenID

	if err := tx.RefreshTokens.Create(ctx, token); err != nil {
		return nil, apperror.Internal(err)
	}

	access, expiresAt, err := s.tokens.Issue(user.ID, user.Role, familyID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	return &dto.TokenPair{
		TokenType:        tokenTypeBearer,
		AccessToken:      access,
		RefreshToken:     secret,
		ExpiresAt:        expiresAt,
		RefreshExpiresAt: token.ExpiresAt,
	}, nil
}

// recordFailure counts the attempt and closes the account once there have been
// enough. It always reports the same failure the wrong password would have:
// telling the caller an account just locked would let anyone find out which
// addresses are registered.
func (s *AuthService) recordFailure(ctx context.Context, user *models.User) error {
	attempts := user.FailedLoginAttempts + 1

	var lockedUntil *time.Time
	if s.lockout.MaxFailedLoginAttempts > 0 && attempts >= s.lockout.MaxFailedLoginAttempts {
		until := time.Now().Add(s.lockout.LockoutDuration)
		lockedUntil = &until
	}

	if err := s.store.Users.RecordFailedLogin(ctx, user.ID, attempts, lockedUntil); err != nil {
		return apperror.Internal(err)
	}
	return ErrInvalidCredentials
}

func errAccountLocked(until time.Time) *apperror.Error {
	return apperror.Locked("account_locked", "Too many failed attempts. Try again after "+until.UTC().Format(time.RFC3339)+".")
}

func errAccountInactive(status models.UserStatus) *apperror.Error {
	switch status {
	case models.UserStatusPendingVerification:
		return apperror.Forbidden("account_pending", "Verify your email address before signing in.")
	case models.UserStatusSuspended:
		return apperror.Forbidden("account_suspended", "This account has been suspended.")
	default:
		return apperror.Forbidden("account_inactive", "This account is no longer active.")
	}
}

// trimmedColumn fits a header into its column instead of letting the insert
// fail over it, and reads an empty header as "not recorded" rather than "".
func trimmedColumn(value string, max int) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if len(value) > max {
		value = strings.ToValidUTF8(value[:max], "")
	}
	return &value
}
