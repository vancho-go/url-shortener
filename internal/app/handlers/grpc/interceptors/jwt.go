package interceptors

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

type key int

// CookieKey - параметр, необходимый для передачи токена через context.
const (
	CookieKey key = iota
)

// tokenExp - время действия сгенерированного токена (cookie).
const tokenExp = time.Hour * 24

// secretKey - секретный ключ для генерации токена (bad practice to store it here, just for study case).
const secretKey = "temp_secret_key"

// Claims - данные, которые в себе будет содержать генерируемый токен.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// GenerateUserID генерирует рандомный UUID.
func GenerateUserID() string {
	return uuid.New().String()
}

// generateJWTToken токен для пользователя, чей userID передан в качестве параметра.
func generateJWTToken(userID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	expirationTime := time.Now().Add(tokenExp)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "",                             // не используется
			Subject:   "",                             // не используется
			Audience:  []string{},                     // не используется
			NotBefore: jwt.NewNumericDate(time.Now()), // не используется
			IssuedAt:  jwt.NewNumericDate(time.Now()), // не используется
			ID:        "",                             // не используется
		},
		// собственное утверждение
		UserID: userID,
	})
	return token.SignedString([]byte(secretKey))
}

// isTokenValid проверяет токен на валидность.
func isTokenValid(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return false
	}
	return token.Valid
}

// GetUserID извлекает userID из валидного токена.
func GetUserID(tokenString string) (string, error) {
	if !isTokenValid(tokenString) {
		return "", fmt.Errorf("token is not valid")
	}
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "",                             // не используется
			Subject:   "",                             // не используется
			Audience:  []string{},                     // не используется
			NotBefore: jwt.NewNumericDate(time.Now()), // не используется
			IssuedAt:  jwt.NewNumericDate(time.Now()), // не используется
			ID:        "",                             // не используется
			ExpiresAt: nil,                            // не используется
		},
		UserID: "",
	}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return "", fmt.Errorf("error parsing token")
	}
	return claims.UserID, nil
}

func JWTInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var jwtToken string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("AuthToken")
		if len(values) > 0 {
			jwtToken = values[0]
		}
	}

	if !isTokenValid(jwtToken) {
		userID := GenerateUserID()
		newJWTToken, err := generateJWTToken(userID)
		if err != nil {
			return nil, status.Error(codes.Internal, "missing jwtToken")
		}
		jwtToken = newJWTToken

		// Создание новых метаданных
		md := metadata.Pairs("AuthToken", jwtToken)
		// Установка метаданных для отправки с ответом
		grpc.SetHeader(ctx, md)
	}

	ctxWV := context.WithValue(ctx, CookieKey, jwtToken)
	resp, err := handler(ctxWV, req)

	return resp, err
}
