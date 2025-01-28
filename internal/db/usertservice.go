package db

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"go.uber.org/zap"
)

type Usert struct {
	ID                  int
	TokenString         string
	TokenExpirationDate time.Time
}

type UsertService struct {
	db  *DB
	log *zap.Logger
	cfg config.Config
}

func NewUsertService(newConfig config.Config, newLogger *zap.Logger, newDB *DB) *UsertService {
	return &UsertService{
		db:  newDB,
		log: newLogger,
		cfg: newConfig,
	}
}

// CreateUsert создает нового пользователя
func (ref *UsertService) CreateUsert(tokenExpirationDate time.Time) (int, error) {
	var id int
	query := `
        INSERT INTO usert (token_expiration_date)
        VALUES ($1)
        RETURNING user_id
    `
	err := ref.db.pool.QueryRow(query, tokenExpirationDate).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUser получает пользователя по ID
// func (ref *UsertService) GetUsert(id int) (Usert, error) {
// 	var usert Usert
// 	err := ref.db.pool.QueryRow(
// 		"SELECT id, name, email FROM usert WHERE id=$1", id).Scan(&usert.ID, &usert.Name, &usert.Email)
// 	if err != nil {
// 		return Usert{}, err
// 	}
// 	return usert, nil
// }

// UpdateUser обновляет пользователя по ID
func (ref *UsertService) UpdateUsert(id int, name, email string) error {
	_, err := ref.db.pool.Exec(
		"UPDATE usert SET name=$1, email=$2 WHERE id=$3", name, email, id)
	return err
}

func GetUsertIDFromContext(req *http.Request, log *zap.Logger) int {
	ctx := req.Context()
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		log.Error("userID not found in context")
		userID = -1
	}
	log.Info("GetUsertIDFromContext()", zap.String("Read userID", strconv.Itoa(userID)))

	return userID
}
