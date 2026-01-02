package models

// User представляет пользователя системы
type User struct {
	BaseModel

	Email        *string `gorm:"uniqueIndex"`
	Phone        *string `gorm:"uniqueIndex"`
	PasswordHash string  `gorm:"not null"`

	FirstName  string `gorm:"not null"`
	LastName   string `gorm:"not null"`
	MiddleName *string

	IsEmailVerified bool
	IsPhoneVerified bool
}
