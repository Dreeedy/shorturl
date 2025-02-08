package authservice

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
	"github.com/Dreeedy/shorturl/internal/storages"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type AuthService interface {
	Auth(w http.ResponseWriter, userID int) int
	CreateCookie(tokenString string) *http.Cookie
	BuildJWTString(useDefaultUser bool) (string, error)
	ValidateToken(tokenString string) int
}

type AuthServiceImpl struct {
	cfg          config.Config
	log          *zap.Logger
	usertService *db.UsertService
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const (
	buildJWTStringText = "BuildJWTString"
)

func NewAuthService(newConfig config.Config, newLogger *zap.Logger, newUsertService *db.UsertService) AuthService {
	var newAuthService = &AuthServiceImpl{
		cfg:          newConfig,
		log:          newLogger,
		usertService: newUsertService,
	}

	return newAuthService
}

func (ref *AuthServiceImpl) Auth(w http.ResponseWriter, userID int) int {
	storageType := storages.GetStorageType(ref.cfg, ref.log)
	if userID > 0 || storageType != "db" {
		return userID
	}

	tokenString, _ := ref.BuildJWTString(false)
	newCookie := ref.CreateCookie(tokenString)
	http.SetCookie(w, newCookie)

	w.Header().Set("Authorization", "Bearer "+tokenString)

	newUserID := ref.ValidateToken(tokenString)

	return newUserID
}

func (ref *AuthServiceImpl) CreateCookie(tokenString string) *http.Cookie {
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

func (ref *AuthServiceImpl) BuildJWTString(useDefaultUser bool) (string, error) {
	cfg := ref.cfg.GetConfig()

	expiresAt := time.Now().Add(time.Hour * time.Duration(cfg.TokenExpHours))

	var userID = 0
	if !useDefaultUser {
		userID, _ = ref.usertService.CreateUsert(expiresAt)
	}

	ref.log.Info("useDefaultUser", zap.String("useDefaultUser", strconv.FormatBool(useDefaultUser)))
	ref.log.Info(buildJWTStringText, zap.String("new expiresAt", expiresAt.Format(time.RFC3339)))
	ref.log.Info(buildJWTStringText, zap.String("new userID", strconv.Itoa(userID)))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(cfg.TokenSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to get SignedString: %w", err)
	}

	return tokenString, nil
}

func (ref *AuthServiceImpl) ValidateToken(tokenString string) int {
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
	ref.log.Info("ValidateToken()", zap.String("claims.UserID", strconv.Itoa(claims.UserID)))

	return claims.UserID
}
