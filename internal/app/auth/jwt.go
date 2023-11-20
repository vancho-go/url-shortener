package auth

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const TOKEN_EXP = time.Hour * 24
const SECRET_KEY = "temp_secret_key"

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("AuthToken")
		if err == nil && IsTokenValid(cookie.Value) {
			next.ServeHTTP(res, req)
			return
		}

		userID := generateUserID()
		jwtToken, err := generateJWTToken(userID)
		if err != nil {
			logger.Log.Error("error building new token", zap.Error(err))
			return
		}
		logger.Log.Debug(fmt.Sprintf("generated new jwt token for user %d", userID))
		cookieNew := &http.Cookie{
			Name:     "AuthToken",
			Value:    jwtToken,
			Expires:  time.Now().Add(TOKEN_EXP),
			HttpOnly: true,
			Path:     "/",
		}

		http.SetCookie(res, cookieNew)

		ctx := context.WithValue(req.Context(), "cookie", cookieNew)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

func generateUserID() string {
	return uuid.New().String()
}

func generateJWTToken(userID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	expirationTime := time.Now().Add(TOKEN_EXP)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		// собственное утверждение
		UserID: userID,
	})
	return token.SignedString([]byte(SECRET_KEY))
}

func IsTokenValid(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return false
	}
	return token.Valid
}

func GetUserId(tokenString string) (string, error) {
	if !IsTokenValid(tokenString) {
		return "", fmt.Errorf("token is not valid")
	}
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return "", fmt.Errorf("error parsing token")
	}
	return claims.UserID, nil
}
