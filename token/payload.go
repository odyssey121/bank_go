package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtPayload struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	*jwt.RegisteredClaims
}

type PasetoPayload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time
	ExpiredAt time.Time
}

func (payload *PasetoPayload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrPasetoExpiredToken
	}
	return nil
}

func NewJwtPayload(username string, duration time.Duration) (*JwtPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := &JwtPayload{
		ID:       tokenID,
		Username: username,
		RegisteredClaims: &jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
	return payload, nil
}

func NewPasetoPayload(username string, duration time.Duration) (*PasetoPayload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := &PasetoPayload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}
