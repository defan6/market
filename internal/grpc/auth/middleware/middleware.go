package middleware

import (
	"context"
	"sso/internal/lib/security/token/claims"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Signer interface {
	Verify(ctx context.Context, token string, claims jwt.Claims) error
}

func AuthInterceptor(signer Signer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		switch info.FullMethod {
		case "/auth.Auth/Register":
			return handler(ctx, req)
		case "/auth.Auth/Login":
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata missing")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "token missing")
		}

		token := strings.TrimPrefix(values[0], "Bearer ")
		var claims claims.AccessClaims
		if err := signer.Verify(ctx, token, &claims); err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)

		return handler(ctx, req)
	}
}
