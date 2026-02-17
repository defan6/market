package claims

import "github.com/golang-jwt/jwt/v5"

type AccessClaims struct {
	UserID int64
	Email  string
	Role   string
	jwt.RegisteredClaims
}
