package auth

import (
	"context"
	"net/http"
	"strconv"
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

// Claims — структура утверждений, которая включает стандартные утверждения и
// одно пользовательское UserID
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

		ref.log.Info("run userauthmiddlerware")

		// Get Cookie
		cookie, err := r.Cookie("myJWTtoken")
		if err != nil {
			if err == http.ErrNoCookie {
				ref.log.Error("Cookie not found", zap.Error(err))

				// Set the cookie in the response
				cookie = ref.CreateCookie()
				http.SetCookie(w, cookie)

				ref.log.Info("Set myJWTtoken cookie")
			} else {
				ref.log.Error("Error reading cookie", zap.Error(err))
			}
		}

		ref.log.Info("cookie.Name", zap.String("cookie.Name", cookie.Name))
		ref.log.Info("cookie.Value", zap.String("cookie.Value", cookie.Value))

		userID := ref.ValidateToken(cookie.Value)
		if userID > -1 {
			ref.log.Info("Work()", zap.String("userID", strconv.Itoa(userID)))
			// Получается и кука и юзер есть.
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			ref.log.Error("cookies do not contain the userID")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ref *Auth) CreateCookie() *http.Cookie {
	// Create new User

	myJWTtoken, _ := ref.BuildJWTString()

	ref.log.Info("CreateCookie", zap.String("myJWTtoken", myJWTtoken))

	// Create a new cookie
	newCookie := &http.Cookie{
		Name:  "myJWTtoken",
		Value: myJWTtoken,
		//Path:     "/",
		//Domain:   "example.com",
		Expires:  time.Now().Add(365 * 24 * time.Hour), // 1 year
		Secure:   true,
		HttpOnly: true,
		//SameSite: http.SameSiteLaxMode,
	}

	return newCookie
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (ref *Auth) BuildJWTString() (string, error) {

	expiresAt := time.Now().Add(TOKEN_EXP)
	userID, _ := ref.usertService.CreateUsert(expiresAt)

	ref.log.Info("BuildJWTString", zap.String("new expiresAt", expiresAt.Format("2006-01-02 15:04:05")))
	ref.log.Info("BuildJWTString", zap.String("new userID", strconv.Itoa(userID)))

	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (ref *Auth) ValidateToken(tokenString string) int {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
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
	// возвращаем ID пользователя в читаемом виде
	return claims.UserID
}

// читаем куку юзера, которая должна содержать JWT
// Если куки нет, то создаем user_id и JWT
// Если кука есть, но протухла, то создаем новую?
// Если есть актуальная кука, то просто сохраняем в контекст и пропускаем
