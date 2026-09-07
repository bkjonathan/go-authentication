package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// BcryptMaxPasswordBytes is bcrypt's own limit: it ignores anything past 72
// bytes, and Go's implementation refuses rather than silently truncating.
const BcryptMaxPasswordBytes = 72

// opaqueTokenBytes is the entropy in a refresh token. 256 bits, so the digest
// column is the only thing that ever has to be unique.
const opaqueTokenBytes = 32

type PasswordHasher struct {
	cost  int
	decoy []byte
}

// NewPasswordHasher precomputes a hash of a value nobody knows, so a sign-in
// attempt for an address that has no account can still spend the time a real
// comparison would. Without it, response latency answers "does this email have
// an account?" for anyone who asks.
func NewPasswordHasher(cost int) (*PasswordHasher, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	decoy, err := bcrypt.GenerateFromPassword([]byte("no-account-with-this-address"), cost)
	if err != nil {
		return nil, fmt.Errorf("prepare password hasher: %w", err)
	}
	return &PasswordHasher{cost: cost, decoy: decoy}, nil
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (h *PasswordHasher) Verify(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// VerifyDecoy burns what a real comparison costs and always fails.
func (h *PasswordHasher) VerifyDecoy(password string) {
	_ = bcrypt.CompareHashAndPassword(h.decoy, []byte(password))
}

// NewOpaqueToken returns the secret to hand the client and the digest to store.
// The database never holds anything a stolen backup could be replayed with.
func NewOpaqueToken() (secret, digest string, err error) {
	raw := make([]byte, opaqueTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate token: %w", err)
	}
	secret = base64.RawURLEncoding.EncodeToString(raw)
	return secret, DigestToken(secret), nil
}

func DigestToken(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}
