package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned for any token that cannot be trusted.
var ErrInvalidToken = errors.New("invalid token")

// Claims is what the API trusts about the caller after a token is verified.
// StaffID lives here, not in the request body, so a caller cannot ask for
// someone else's data by changing the request body. HospitalID is also here to
type Claims struct {
	StaffID    int64 `json:"staff_id"`
	HospitalID int64 `json:"hospital_id"`
	jwt.RegisteredClaims
}

// Token issues and verifies access tokens.
type Token struct {
	secret []byte
	ttl    time.Duration
}

// NewToken returns a Token signing with secret for ttl.
func NewToken(secret string, ttl time.Duration) *Token {
	return &Token{secret: []byte(secret), ttl: ttl}
}

// Issue returns a signed token describing account.
func (t *Token) Issue(staffID, hospitalID int64, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		StaffID:    staffID,
		HospitalID: hospitalID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// Parse verifies a token and returns its claims.
func (t *Token) Parse(raw string) (Claims, error) {
	var claims Claims

	_, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		// Reject a token that asks to be verified with a different algorithm,
		// which is how the "alg: none" forgery works.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}
		return t.secret, nil
	})
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}
