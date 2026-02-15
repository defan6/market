package signer

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type HMACSigner struct {
	secret []byte
}

func (h *HMACSigner) Sign(
	ctx context.Context,
	claims jwt.Claims,
) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(h.secret)
}

func (h *HMACSigner) Verify(
	ctx context.Context,
	token string,
	claims jwt.Claims,
) error {
	return nil
}

func NewHMACSigner(secret []byte) *HMACSigner {
	return &HMACSigner{
		secret: secret,
	}
}
