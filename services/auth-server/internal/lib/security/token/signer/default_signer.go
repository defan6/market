package signer

import (
	"context"
	"fmt"

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
	tokenString string,
	claims jwt.Claims,
) error {
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", token.Header["alg"])
		}

		return h.secret, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func NewHMACSigner(secret []byte) *HMACSigner {
	return &HMACSigner{
		secret: secret,
	}
}
