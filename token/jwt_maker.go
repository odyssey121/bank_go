package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrUnknownCalmsType     = errors.New("unknown claims type, cannot proceed")
	ErrTokenExpired         = errors.New("token has invalid claims: token is expired")
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (Maker[*JwtPayload], error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("secret key len must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey: secretKey}, nil
}

func (m *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewJwtPayload(username, duration)

	if err != nil {
		return "", err
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	s, err := t.SignedString([]byte(m.secretKey))
	if err != nil {
		return "", err
	}
	return s, nil

}
func (m *JWTMaker) VerifyToken(tokenStr string) (*JwtPayload, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JwtPayload{}, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidSigningMethod
		}
		s := []byte(m.secretKey)
		return s, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	} else if claims, ok := token.Claims.(*JwtPayload); ok {
		return claims, nil
	} else {
		return nil, ErrUnknownCalmsType
	}
}
