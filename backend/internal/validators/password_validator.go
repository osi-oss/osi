package validators

import (
	"fmt"
	"unicode"
)

// PasswordValidator валидатор паролей с простыми настройками
type PasswordValidator struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigits    bool
	RequireSpecial   bool
	MaxLength        int
}

// DefaultPasswordValidator создает валидатор с настройками по умолчанию
func DefaultPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:        8,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigits:    true,
		RequireSpecial:   false,
		MaxLength:        128,
	}
}

var Password = DefaultPasswordValidator()

// ValidatePassword проверяет пароль по всем правилам
func (v *PasswordValidator) Validate(password string) error {
	if len(password) < v.MinLength {
		return fmt.Errorf("password must be at least %d characters long", v.MinLength)
	}

	if len(password) > v.MaxLength {
		return fmt.Errorf("password must be no more than %d characters long", v.MaxLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if v.RequireUppercase && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if v.RequireLowercase && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if v.RequireDigits && !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	if v.RequireSpecial && !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
