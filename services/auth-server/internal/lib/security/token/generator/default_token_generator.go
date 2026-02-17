package generator

import (
	"context"
	"errors"
	"fmt"
	"sso/internal/domain"
	"sso/internal/dto"
	"sso/internal/lib/security/token/claims"
	"sso/internal/service"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrSignToken = errors.New("sign token error")
)

type DefaultTokenGenerator struct {
	signer Signer
	issuer string
	ttl    time.Duration
}

type Signer interface {
	Sign(ctx context.Context, claims jwt.Claims) (string, error)
	Verify(ctx context.Context, token string, claims jwt.Claims) error
}

func NewDefaultTokenGenerator(
	signer Signer,
	issuer string,
	ttl time.Duration,
) service.TokenGenerator {
	return &DefaultTokenGenerator{
		signer: signer,
		issuer: issuer,
		ttl:    ttl,
	}
}

func (d *DefaultTokenGenerator) GenerateToken(
	ctx context.Context,
	userDetails domain.UserDetails,
	appID string,
) (*dto.TokenGenerateResponse, error) {
	claims := claims.AccessClaims{
		UserID: userDetails.ID,
		Email:  userDetails.Email,
		Role:   userDetails.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    d.issuer,
			Subject:   strconv.Itoa(int(userDetails.ID)),
			Audience:  jwt.ClaimStrings{appID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(d.ttl)),
		},
	}
	token, err := d.signer.Sign(ctx, claims)
	if err != nil {
		return &dto.TokenGenerateResponse{}, fmt.Errorf("error signing token: %w", err)
	}
	return dto.NewTokenGenerateResponse(token), nil
}
