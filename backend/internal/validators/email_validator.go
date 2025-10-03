package validators

import (
	"fmt"
	"regexp"
)

// EmailValidator валидатор email адресов
type EmailValidator struct {
	MaxLength int
}

// DefaultEmailValidator создает валидатор с настройками по умолчанию
func DefaultEmailValidator() *EmailValidator {
	return &EmailValidator{
		MaxLength: 254,
	}
}

var Email = DefaultEmailValidator()

// ValidateEmail проверяет корректность email адреса
func (v *EmailValidator) Validate(email string) error {
	if len(email) == 0 {
		return fmt.Errorf("email is required")
	}

	if len(email) > v.MaxLength {
		return fmt.Errorf("email is too long (max %d characters)", v.MaxLength)
	}

	// Базовая регулярка для email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
