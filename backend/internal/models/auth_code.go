package models

import "time"

// AuthCode код аутентификации для входа/регистрации по email
type AuthCode struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"not null;index"`
	Code      string    `gorm:"size:4;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	Attempts  int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	// Связь с пользователем
	User User `gorm:"foreignKey:UserID"`
}

// IsExpired проверяет, истёк ли код
func (c *AuthCode) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsValid проверяет валидность кода
func (c *AuthCode) IsValid() bool {
	return !c.Used && !c.IsExpired() && c.Attempts < 5
}

// MaxAttempts максимальное количество попыток ввода кода
const MaxAuthCodeAttempts = 5

// AuthCodeValidityMinutes время действия кода в минутах
const AuthCodeValidityMinutes = 20
