package token

import "time"

type Maker[PT any] interface {
	CreateToken(username string, duration time.Duration) (string, PT, error)
	// verifyToken checks if the token is valid or not
	VerifyToken(token string) (PT, error)
}
