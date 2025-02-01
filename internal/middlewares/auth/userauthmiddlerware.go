package auth

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
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

const TOKEN_EXP = time.Hour * 3
const SECRET_KEY = "supersecretkey"

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

		// Логируем все заголовки запроса
		ref.log.Info("Request Headers", zap.Any("headers", r.Header))

		var hasCookie bool = false

		var tokenString string
		cookies := r.Cookies()
		if cookies != nil && len(cookies) > 0 {
			cookieMap := make(map[string]string)
			for _, cookie := range cookies {
				cookieMap[cookie.Name] = cookie.Value
			}
			ref.log.Info("Request Cookies", zap.Any("cookies", cookieMap))
			hasCookie = true
			cookie, _ := r.Cookie("myJWTtoken")
			tokenString = cookie.Value
		} else {
			ref.log.Info("No cookies in request")
		}

		// Сначала пытаемся получить токен из заголовка Authorization
		var hasHeader bool = false

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			hasHeader = true
		}

		ref.log.Info("Work()", zap.String("hasCookie", strconv.FormatBool(hasCookie)))
		ref.log.Info("Work()", zap.String("hasHeader", strconv.FormatBool(hasHeader)))

		// Валидируем токен
		userID := ref.ValidateToken(tokenString)
		if userID > -1 {
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		next.ServeHTTP(w, r)
	})
}

func (ref *Auth) ValidateToken(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
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
