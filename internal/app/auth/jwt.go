// Модуль auth реализует логику аутентификации пользователя.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/vancho-go/url-shortener/internal/app/logger"
)

type key int

// CookieKey - параметр, необходимый для передачи токена через context.
const (
	CookieKey key = iota
)

// TokenExp - время действия сгенерированного токена (cookie).
const TokenExp = time.Hour * 24

// SecretKey - секретный ключ для генерации токена (bad practice to store it here, just for study case).
const SecretKey = "temp_secret_key"

// Claims - данные, которые в себе будет содержать генерируемый токен.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// JWTMiddleware выполняет роль middleware, которая проверяет наличие токена аутентификации.
// Если токен присутствует и он валидный, middleware передает запрос следующему обработчику.
// Если токена нет, генерируется новый токен, который передается следующему обработчику через context.
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
		logger.Log.Debug(fmt.Sprintf("generated new jwt token for user %s", userID))
		cookieNew := &http.Cookie{
			Name:     "AuthToken",
			Value:    jwtToken,
			Expires:  time.Now().Add(TokenExp),
			HttpOnly: true,
			Path:     "/",
		}

		http.SetCookie(res, cookieNew)

		ctx := context.WithValue(req.Context(), CookieKey, cookieNew)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// generateUserID генерирует рандомный UUID.
func generateUserID() string {
	return uuid.New().String()
}

// generateJWTToken токен для пользователя, чей userID передан в качестве параметра.
func generateJWTToken(userID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	expirationTime := time.Now().Add(TokenExp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		// собственное утверждение
		UserID: userID,
	})
	return token.SignedString([]byte(SecretKey))
}

// IsTokenValid проверяет токен на валидность.
func IsTokenValid(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(SecretKey), nil
	})
	if err != nil {
		return false
	}
	return token.Valid
}

// GetUserID извлекает userID из валидного токена.
func GetUserID(tokenString string) (string, error) {
	if !IsTokenValid(tokenString) {
		return "", fmt.Errorf("token is not valid")
	}
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("error parsing token")
	}
	return claims.UserID, nil
}
