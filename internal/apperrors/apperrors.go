package apperrors

import "fmt"

type InsertConflict struct {
	Code    int
	Message string
}

func (e *InsertConflict) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func NewInsertConflict(code int, message string) error {
	return &InsertConflict{
		Code:    code,
		Message: message,
	}
}
