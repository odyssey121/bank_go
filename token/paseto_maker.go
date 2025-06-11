package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

var (
	ErrPasetoExpiredToken = errors.New("paseto token is expired")
)

type PasetoMaker struct {
	symetricKey []byte
	paseto      *paseto.V2
}

func NewPasetoMaker(symetricKey string) (Maker[*PasetoPayload], error) {
	if len(symetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("secret key len must be equals %d characters", chacha20poly1305.KeySize)
	}
	maker := &PasetoMaker{
		symetricKey: []byte(symetricKey),
		paseto:      paseto.NewV2(),
	}

	return maker, nil
}

func (m *PasetoMaker) CreateToken(username string, duration time.Duration) (string, *PasetoPayload, error) {
	payload, err := NewPasetoPayload(username, duration)
	if err != nil {
		return "", nil, err
	}

	s, err := m.paseto.Encrypt(m.symetricKey, payload, nil)
	if err != nil {
		return "", nil, err
	}
	return s, payload, nil

}
func (m *PasetoMaker) VerifyToken(tokenStr string) (*PasetoPayload, error) {
	var varifyPayload PasetoPayload

	err := m.paseto.Decrypt(tokenStr, m.symetricKey, &varifyPayload, nil)
	if err != nil {
		return nil, err
	}

	err = varifyPayload.Valid()
	if err != nil {
		return nil, err
	}

	return &varifyPayload, nil
}
