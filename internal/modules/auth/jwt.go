package auth

import (
	"time"

	"aipivot/internal/config"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID   int64  `json:"userId"`
	TenantID int64  `json:"tenantId"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(conf config.AuthConf, userID, tenantID int64, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    conf.Issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(conf.AccessExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(conf.AccessSecret))
}

func ParseToken(conf config.AuthConf, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(conf.AccessSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
