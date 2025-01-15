package apperrors

import "fmt"

// InsertConflictError represents an error that occurs during an insert conflict.
type InsertConflictError struct {
	Message string
	Code    int
}

func (e *InsertConflictError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func NewInsertConflict(code int, message string) error {
	return &InsertConflictError{
		Code:    code,
		Message: message,
	}
}
