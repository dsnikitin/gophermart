package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

type Config struct {
	SecretKey  string        `env:"SECRET_KEY,required"`
	TokenExp   time.Duration `env:"TOKEN_TTL"`
	CookieName string        `env:"COOKIE_NAME"`
}

type Claims struct {
	jwt.RegisteredClaims
	Login string
}

func CreateToken(cfg *Config, login string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(cfg.TokenExp)),
		},
		Login: login,
	})

	tokenStr, err := token.SignedString([]byte(cfg.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func ParseToken(cfg *Config, tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("token is empty")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.Errorf("unexpected signing method: %v", t.Header["alg"])
			}

			return []byte(cfg.SecretKey), nil
		})
	if err != nil {
		return nil, errors.Wrap(err, "parse jwt with claims")
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	if !claims.VerifyExpiresAt(time.Now().UTC(), true) {
		return nil, errors.New("token is expired")
	}

	return claims, nil
}
