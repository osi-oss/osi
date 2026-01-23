package models

// User представляет пользователя системы
type User struct {
	BaseModel

	Email        string  `gorm:"uniqueIndex;not null"`
	Phone        *string `gorm:"uniqueIndex"`
	PasswordHash *string // nullable - пароль опционален

	FirstName  *string // nullable - заполняется после подтверждения email
	LastName   *string // nullable - заполняется после подтверждения email
	MiddleName *string

	Status          UserStatus `gorm:"type:user_status;default:'pending_email'"`
	IsEmailVerified bool       `gorm:"default:false"`
	IsPhoneVerified bool       `gorm:"default:false"`
}

// HasPassword проверяет, установлен ли пароль
func (u *User) HasPassword() bool {
	return u.PasswordHash != nil && *u.PasswordHash != ""
}

// IsProfileComplete проверяет, заполнен ли профиль
func (u *User) IsProfileComplete() bool {
	return u.FirstName != nil && *u.FirstName != "" &&
		u.LastName != nil && *u.LastName != ""
}

// CanAccessSystem проверяет, имеет ли пользователь полный доступ
func (u *User) CanAccessSystem() bool {
	return u.Status == UserStatusActive
}
