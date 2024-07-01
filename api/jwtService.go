package api

import (
	"chirpy/internal/model"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IdContext string

const ContextUserKey IdContext = "user"

type JwtService struct {
	JwtSecret string
}

func (jw *JwtService) ValidateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetJwtToken(r.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "UNAUTHORIZED"}`))
			return
		}
		if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
			id := claims.Subject
			if id == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"error": "UNAUTHORIZED"}`))
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"error": "UNAUTHORIZED"}`))
			return
		}
	})

}

func ExtractTokenString(word, header string) (string, error) {

	tokenString := strings.TrimSpace(strings.TrimPrefix(header, word))
	if tokenString == "" {
		return "", errors.New("the token is empty")
	}
	return tokenString, nil
}

func GetJwtToken(rawToken string) (*jwt.Token, error) {
	tokenString, err := ExtractTokenString("Bearer", rawToken)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func (jw *JwtService) CreateJwtTokenForUser(user model.User) (string, error) {
	var duration time.Time
	if user.Expires_in_seconds != nil {

		duration = time.Now().UTC().Add(time.Duration(*user.Expires_in_seconds) * time.Second)
	} else {
		duration = time.Now().UTC().Add(time.Hour * 1)
	}
	claimsMap := jwt.RegisteredClaims{
		Subject:   user.Email,
		Issuer:    "Chirp",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(duration),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsMap)
	s, err := t.SignedString([]byte(jw.JwtSecret))
	if err != nil {
		return "", err
	}
	return s, nil
}

func (jw *JwtService) CreateRefreshRoken(user model.User) (*model.RefreshToken, error) {
	duration := time.Hour * 24 * 60
	byteSlice := make([]byte, 30)
	_, err := rand.Read(byteSlice)
	if err != nil {
		return nil, err
	}
	tokenString := hex.EncodeToString(byteSlice)
	expirationTime := time.Now().Add(duration)

	return &model.RefreshToken{
		UserEmail:      user.Email,
		Token:          tokenString,
		ExpirationTime: expirationTime,
	}, nil

}
