package utils

import (
	"fmt"
	"time"

	"github.com/bkjonathan/go-authentication/internal/apperror"
	"github.com/bkjonathan/go-authentication/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	Role      models.Role `json:"role"`
	SessionID uuid.UUID   `json:"sid"`
}

type TokenIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{secret: []byte(secret), ttl: ttl}
}

func (t *TokenIssuer) TTL() time.Duration { return t.ttl }

// Issue mints an access token. It carries the session it was issued under so a
// request can tell which of the caller's devices it came from - the sessions
// list needs that to mark one entry "this device".
func (t *TokenIssuer) Issue(userID uuid.UUID, role models.Role, sessionID uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(t.ttl)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.NewString(),
		},
		Role:      role,
		SessionID: sessionID,
	})

	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

// Parse verifies and decodes an access token. Every rejection reads the same to
// the caller: a token that is expired, forged or malformed is equally not proof
// of anything.
func (t *TokenIssuer) Parse(raw string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %q", token.Header["alg"])
		}
		return t.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, ErrInvalidAccessToken.Wrap(err)
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, ErrInvalidAccessToken.Wrap(err)
	}
	return claims, nil
}

var ErrInvalidAccessToken = apperror.Unauthorized("invalid_token", "Your session has expired. Sign in again.")

// UserID is the subject, already known to parse - Parse rejects a token whose
// subject is not a UUID.
func (c *AccessClaims) UserID() uuid.UUID {
	id, _ := uuid.Parse(c.Subject)
	return id
}
