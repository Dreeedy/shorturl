package authservice

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/db"
	"github.com/Dreeedy/shorturl/internal/storages"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// принимает ид пользователя
// если ид есть то мы ничего не делаем
// если ид нет, то создаем нового пользователя, создаем куку и тд

type Authservice struct {
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

func NewAuthservice(newConfig config.Config, newLogger *zap.Logger, newUsertService *db.UsertService) *Authservice {
	var newAuthservice = &Authservice{
		cfg:          newConfig,
		log:          newLogger,
		usertService: newUsertService,
	}

	return newAuthservice
}

func (ref *Authservice) Auth(w http.ResponseWriter, userID int) int {
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

func (ref *Authservice) CreateCookie(tokenString string) *http.Cookie {
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

func (ref *Authservice) BuildJWTString(useDefaultUser bool) (string, error) {
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

func (ref *Authservice) ValidateToken(tokenString string) int {
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
	ref.log.Info("ValidateToken()", zap.String("claims.UserID", strconv.Itoa(claims.UserID)))

	return claims.UserID
}
