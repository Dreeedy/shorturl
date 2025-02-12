package db

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Dreeedy/shorturl/internal/config"
	"github.com/Dreeedy/shorturl/internal/storages/common"
	"go.uber.org/zap"
)

type Usert struct {
	TokenExpirationDate time.Time
	TokenString         string
	ID                  int
}

type UsertService struct {
	log *zap.Logger
	cfg config.Config
	db  DB
}

func NewUsertService(newConfig config.Config, newLogger *zap.Logger, newDB DB) *UsertService {
	return &UsertService{
		log: newLogger,
		cfg: newConfig,
		db:  newDB,
	}
}

// CreateUsert create new usert.
func (ref *UsertService) CreateUsert(tokenExpirationDate time.Time) (int, error) {
	var id int
	query := `
        INSERT INTO usert (token_expiration_date)
        VALUES ($1)
        RETURNING user_id
    `
	err := ref.db.GetConnPool().QueryRow(query, tokenExpirationDate).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to scan row: %w", err)
	}
	return id, nil
}

// UpdateUsert updates user by ID.
func (ref *UsertService) UpdateUsert(id int, name, email string) error {
	_, err := ref.db.GetConnPool().Exec("UPDATE usert SET name=$1, email=$2 WHERE id=$3", name, email, id)
	if err != nil {
		return fmt.Errorf("failed to update usert table: %w", err)
	}
	return nil
}

func GetUsertIDFromContext(req *http.Request, log *zap.Logger) int {
	ctx := req.Context()
	userID, ok := ctx.Value(common.UserIDKey).(int)
	if !ok {
		log.Info("userID not found in context")
		userID = -1
	}
	log.Info("GetUsertIDFromContext()", zap.String("Read userID", strconv.Itoa(userID)))

	return userID
}
