package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret []byte
	issuer string
	expiry time.Duration
}

func NewJWTManager(secret, issuer string, expiry time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), issuer: issuer, expiry: expiry}
}

type Claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (j *JWTManager) Generate(userID int64, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: j.issuer,
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTManager) Parse(tokenStr string) (*Claims, error) {
	tkn, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 { return nil, errors.New("unexpected signing method") }
		return j.secret, nil
	})
	if err != nil { return nil, err }
	claims, ok := tkn.Claims.(*Claims)
	if !ok || !tkn.Valid { return nil, errors.New("invalid token") }
	return claims, nil
}
