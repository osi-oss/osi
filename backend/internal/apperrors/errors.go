package apperrors

import (
	"errors"
	"net/http"
)

// AppError представляет ошибку приложения с HTTP кодом
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Is реализует errors.Is для сравнения ошибок по коду и сообщению
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code && e.Message == t.Message
}

// Предопределённые ошибки
var (
	ErrUnauthorized         = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden            = &AppError{Code: http.StatusForbidden, Message: "access denied"}
	ErrNotFound             = &AppError{Code: http.StatusNotFound, Message: "not found"}
	ErrBadRequest           = &AppError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrConflict             = &AppError{Code: http.StatusConflict, Message: "conflict"}
	ErrInternalServer       = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
	ErrUserAlreadyExists    = &AppError{Code: http.StatusConflict, Message: "user with this email already exists"}
	ErrUserNotFound         = &AppError{Code: http.StatusNotFound, Message: "user not found"}
	ErrInvalidCredentials   = &AppError{Code: http.StatusUnauthorized, Message: "invalid email or password"}
	ErrOrganizationNotFound = &AppError{Code: http.StatusNotFound, Message: "organization not found"}
	ErrLocationNotFound     = &AppError{Code: http.StatusNotFound, Message: "location not found"}
	ErrDepartmentNotFound   = &AppError{Code: http.StatusNotFound, Message: "department not found"}
	ErrPositionNotFound     = &AppError{Code: http.StatusNotFound, Message: "position not found"}
	ErrMemberNotFound       = &AppError{Code: http.StatusNotFound, Message: "member not found"}
	ErrEmployeeNotFound     = &AppError{Code: http.StatusNotFound, Message: "employee not found"}
	ErrPermissionNotFound   = &AppError{Code: http.StatusNotFound, Message: "permission not found"}
	ErrAccessDenied         = &AppError{Code: http.StatusForbidden, Message: "access denied"}
	ErrUserAlreadyMember    = &AppError{Code: http.StatusConflict, Message: "user is already a member"}
	ErrMemberNotActive      = &AppError{Code: http.StatusBadRequest, Message: "member is not active"}
	ErrInvalidInput         = &AppError{Code: http.StatusBadRequest, Message: "invalid input data"}
)

// New создаёт новую ошибку с кодом
func New(code int, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap оборачивает ошибку в AppError
func Wrap(err error, code int, message string) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// NotFound создаёт ошибку "не найдено"
func NotFound(entity string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: entity + " not found"}
}

// BadRequest создаёт ошибку "плохой запрос"
func BadRequest(message string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Message: message}
}

// GetHTTPCode возвращает HTTP код для ошибки
func GetHTTPCode(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// GetMessage возвращает сообщение ошибки
func GetMessage(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}
