package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type Auth struct {
	cfg          config.Config
	log          *zap.Logger
	usertService *db.UsertService
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func NewAuthMiddleware(newConfig config.Config, newLogger *zap.Logger, newUsertService *db.UsertService) *Auth {
	var newAuth = &Auth{
		cfg:          newConfig,
		log:          newLogger,
		usertService: newUsertService,
	}

	newLogger.Info("NewAuth created")

	return newAuth
}

func (ref *Auth) Work(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("run userauthmiddlerware")

		// Логируем все заголовки запроса.
		ref.log.Info("Request Headers", zap.Any("headers", r.Header))

		var tokenString string
		cookies := r.Cookies()
		if len(cookies) > 0 {
			cookieMap := make(map[string]string)
			for _, cookie := range cookies {
				cookieMap[cookie.Name] = cookie.Value
			}
			ref.log.Info("Request Cookies", zap.Any("cookies", cookieMap))
			cookie, _ := r.Cookie("myJWTtoken")
			tokenString = cookie.Value
		} else {
			ref.log.Info("No cookies in request")
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Валидируем токен.
		userID := ref.ValidateToken(tokenString)
		ctx := context.WithValue(r.Context(), common.UserIDKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (ref *Auth) ValidateToken(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(ref.cfg.GetConfig().TokenSecretKey), nil
		})
	if err != nil {
		return -1
	}

	if !token.Valid {
		ref.log.Error("Token is not valid")
		return -1
	}

	ref.log.Info("Token is valid")
	return claims.UserID
}
