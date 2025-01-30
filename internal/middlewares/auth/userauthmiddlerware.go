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
		} else {
			if tokenString == "" {
				tokenString, _ = ref.BuildJWTString(false)
			}
		}

		ref.log.Info("hasCookie", zap.String("hasCookie", strconv.FormatBool(hasCookie)))
		ref.log.Info("hasHeader", zap.String("hasHeader", strconv.FormatBool(hasHeader)))

		// Валидируем токен
		userID := ref.ValidateToken(tokenString)
		if userID > -1 {
			ctx := context.WithValue(r.Context(), "userID", userID)
			// Устанавливаем заголовок Authorization в ответе
			w.Header().Set("Authorization", "Bearer "+tokenString)

			// Создаем новый токен и куку
			newCookie := ref.CreateCookie(tokenString)
			http.SetCookie(w, newCookie)

			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			ref.log.Error("Invalid token")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	})
}

func (ref *Auth) CreateCookie(tokenString string) *http.Cookie {
	newCookie := &http.Cookie{
		Name:     "myJWTtoken",
		Value:    tokenString,
		Path:     "/",
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		Secure:   false,
		HttpOnly: true,
	}
	return newCookie
}

func (ref *Auth) BuildJWTString(useDefaultUser bool) (string, error) {
	expiresAt := time.Now().Add(TOKEN_EXP)

	var userID = 0
	if !useDefaultUser {
		userID, _ = ref.usertService.CreateUsert(expiresAt)
	}

	ref.log.Info("useDefaultUser", zap.String("useDefaultUser", strconv.FormatBool(useDefaultUser)))

	ref.log.Info("BuildJWTString", zap.String("new expiresAt", expiresAt.Format("2006-01-02 15:04:05")))
	ref.log.Info("BuildJWTString", zap.String("new userID", strconv.Itoa(userID)))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	return tokenString, nil
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
