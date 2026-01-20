package dto

import "time"

// ===== User DTOs =====

// SignUpRequest запрос на регистрацию нового пользователя
// @Description Данные для регистрации нового пользователя в системе
type SignUpRequest struct {
	// Email пользователя (должен быть уникальным)
	// Example: user@example.com
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// Пароль (минимум 6 символов)
	// Example: SecurePass123
	Password string `json:"password" binding:"required,min=6" example:"SecurePass123"`
}

// LoginRequest запрос на авторизацию
// @Description Учётные данные для входа в систему
type LoginRequest struct {
	// Email пользователя
	// Example: user@example.com
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
	// Пароль пользователя
	// Example: SecurePass123
	Password string `json:"password" binding:"required,min=6" example:"SecurePass123"`
}

// PasswordResetRequest запрос на сброс пароля
// @Description Email для отправки инструкций по сбросу пароля
type PasswordResetRequest struct {
	// Email пользователя для восстановления доступа
	// Example: user@example.com
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ResetPasswordRequest запрос на установку нового пароля
// @Description Токен сброса и новый пароль
type ResetPasswordRequest struct {
	// Токен из письма для сброса пароля
	// Example: a1b2c3d4e5f6g7h8i9j0
	Token string `json:"token" binding:"required" example:"a1b2c3d4e5f6g7h8i9j0"`
	// Новый пароль (минимум 6 символов)
	// Example: NewSecurePass456
	NewPassword string `json:"new_password" binding:"required,min=6" example:"NewSecurePass456"`
}

// UserResponse данные пользователя
// @Description Публичные данные пользователя
type UserResponse struct {
	// Уникальный идентификатор пользователя
	// Example: 1
	ID int64 `json:"id" example:"1"`
	// Email пользователя
	// Example: user@example.com
	Email *string `json:"email" example:"user@example.com"`
	// Имя пользователя
	// Example: Иван
	FirstName string `json:"first_name" example:"Иван"`
	// Фамилия пользователя
	// Example: Петров
	LastName string `json:"last_name" example:"Петров"`
	// Подтверждён ли email
	// Example: true
	EmailVerified bool `json:"email_verified" example:"true"`
	// Дата регистрации
	// Example: 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// ===== Organization DTOs =====

// CreateOrganizationRequest запрос на создание организации
// @Description Данные для создания новой организации
type CreateOrganizationRequest struct {
	// Название организации (обязательное)
	// Example: ООО "Ромашка"
	Name string `json:"name" binding:"required" example:"ООО Ромашка"`
	// Полное юридическое название
	// Example: Общество с ограниченной ответственностью "Ромашка"
	LegalName *string `json:"legal_name" example:"Общество с ограниченной ответственностью Ромашка"`
	// ИНН организации (10 или 12 цифр)
	// Example: 7707083893
	INN *string `json:"inn" example:"7707083893"`
	// ОГРН организации (13 цифр)
	// Example: 1027700132195
	OGRN *string `json:"ogrn" example:"1027700132195"`
	// КПП организации (9 цифр)
	// Example: 770701001
	KPP *string `json:"kpp" example:"770701001"`
	// Юридический адрес организации
	// Example: г. Москва, ул. Тверская, д. 1
	LegalAddress *string `json:"legal_address" example:"г. Москва, ул. Тверская, д. 1"`
	// Доля владения (в процентах, от 0 до 100)
	// Example: 100.0
	SharePercent *float64 `json:"share_percent" example:"100.0"`
}

// UpdateOrganizationRequest запрос на обновление организации
// @Description Данные для обновления организации (передаются только изменяемые поля)
type UpdateOrganizationRequest struct {
	// Новое название организации
	// Example: ООО "Ромашка Плюс"
	Name string `json:"name" example:"ООО Ромашка Плюс"`
	// Полное юридическое название
	// Example: Общество с ограниченной ответственностью "Ромашка Плюс"
	LegalName *string `json:"legal_name" example:"Общество с ограниченной ответственностью Ромашка Плюс"`
	// ИНН организации
	// Example: 7707083893
	INN *string `json:"inn" example:"7707083893"`
	// ОГРН организации
	// Example: 1027700132195
	OGRN *string `json:"ogrn" example:"1027700132195"`
	// КПП организации
	// Example: 770701001
	KPP *string `json:"kpp" example:"770701001"`
	// Юридический адрес
	// Example: г. Москва, ул. Тверская, д. 2
	LegalAddress *string `json:"legal_address" example:"г. Москва, ул. Тверская, д. 2"`
}

// OrganizationResponse данные организации
// @Description Полная информация об организации
type OrganizationResponse struct {
	// Уникальный идентификатор организации
	// Example: 1
	ID int64 `json:"id" example:"1"`
	// Название организации
	// Example: ООО "Ромашка"
	Name string `json:"name" example:"ООО Ромашка"`
	// Полное юридическое название
	// Example: Общество с ограниченной ответственностью "Ромашка"
	LegalName *string `json:"legal_name" example:"Общество с ограниченной ответственностью Ромашка"`
	// ИНН организации
	// Example: 7707083893
	INN *string `json:"inn" example:"7707083893"`
	// ОГРН организации
	// Example: 1027700132195
	OGRN *string `json:"ogrn" example:"1027700132195"`
	// КПП организации
	// Example: 770701001
	KPP *string `json:"kpp" example:"770701001"`
	// Юридический адрес
	// Example: г. Москва, ул. Тверская, д. 1
	LegalAddress *string `json:"legal_address" example:"г. Москва, ул. Тверская, д. 1"`
	// Статус организации: draft, pending, approved, rejected
	// Example: approved
	Status string `json:"status" example:"approved" enums:"draft,pending,approved,rejected"`
	// Дата создания
	// Example: 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Дата последнего обновления
	// Example: 2024-01-20T15:45:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-20T15:45:00Z"`
}

// ===== Location DTOs =====

// CreateLocationRequest запрос на создание локации
// @Description Данные для создания новой локации (офиса, филиала, склада и т.д.)
type CreateLocationRequest struct {
	// Название локации (обязательное)
	// Example: Главный офис
	Name string `json:"name" binding:"required" example:"Главный офис"`
	// Физический адрес локации
	// Example: г. Москва, ул. Ленина, д. 10
	Address *string `json:"address" example:"г. Москва, ул. Ленина, д. 10"`
	// Источник данных: manual, egrul, api
	// Example: manual
	Source string `json:"source" binding:"required" example:"manual" enums:"manual,egrul,api"`
	// Подтверждена ли локация
	// Example: false
	IsVerified bool `json:"is_verified" example:"false"`
}

// UpdateLocationRequest запрос на обновление локации
// @Description Данные для обновления локации
type UpdateLocationRequest struct {
	// Новое название локации
	// Example: Центральный офис
	Name string `json:"name" example:"Центральный офис"`
	// Новый адрес
	// Example: г. Москва, ул. Пушкина, д. 5
	Address *string `json:"address" example:"г. Москва, ул. Пушкина, д. 5"`
	// Источник данных
	// Example: manual
	Source string `json:"source" example:"manual" enums:"manual,egrul,api"`
	// Статус верификации
	// Example: true
	IsVerified *bool `json:"is_verified" example:"true"`
	// Активна ли локация
	// Example: true
	IsActive *bool `json:"is_active" example:"true"`
}

// LocationResponse данные локации
// @Description Полная информация о локации
type LocationResponse struct {
	// Уникальный идентификатор локации
	// Example: 1
	ID int64 `json:"id" example:"1"`
	// ID организации, которой принадлежит локация
	// Example: 1
	OrganizationID int64 `json:"organization_id" example:"1"`
	// Название локации
	// Example: Главный офис
	Name string `json:"name" example:"Главный офис"`
	// Физический адрес
	// Example: г. Москва, ул. Ленина, д. 10
	Address *string `json:"address" example:"г. Москва, ул. Ленина, д. 10"`
	// Источник данных
	// Example: manual
	Source string `json:"source" example:"manual" enums:"manual,egrul,api"`
	// Подтверждена ли локация
	// Example: true
	IsVerified bool `json:"is_verified" example:"true"`
	// Активна ли локация
	// Example: true
	IsActive bool `json:"is_active" example:"true"`
	// Дата создания
	// Example: 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Дата последнего обновления
	// Example: 2024-01-20T15:45:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-20T15:45:00Z"`
}

// ===== Department DTOs =====

// CreateDepartmentRequest запрос на создание отдела
// @Description Данные для создания нового отдела в локации
type CreateDepartmentRequest struct {
	// Название отдела (обязательное)
	// Example: Отдел продаж
	Name string `json:"name" binding:"required" example:"Отдел продаж"`
	// ID родительского отдела (для создания иерархии)
	// Example: 1
	ParentID *int64 `json:"parent_id" example:"1"`
	// Описание отдела
	// Example: Отдел занимается продажами и работой с клиентами
	Description *string `json:"description" example:"Отдел занимается продажами и работой с клиентами"`
}

// UpdateDepartmentRequest запрос на обновление отдела
// @Description Данные для обновления отдела
type UpdateDepartmentRequest struct {
	// Новое название отдела
	// Example: Отдел корпоративных продаж
	Name string `json:"name" example:"Отдел корпоративных продаж"`
	// Новый родительский отдел
	// Example: 2
	ParentID *int64 `json:"parent_id" example:"2"`
	// Новое описание
	// Example: Отдел работает с корпоративными клиентами
	Description *string `json:"description" example:"Отдел работает с корпоративными клиентами"`
}

// DepartmentResponse данные отдела
// @Description Полная информация об отделе
type DepartmentResponse struct {
	// Уникальный идентификатор отдела
	// Example: 1
	ID int64 `json:"id" example:"1"`
	// ID локации, в которой находится отдел
	// Example: 1
	LocationID int64 `json:"location_id" example:"1"`
	// ID родительского отдела (может быть пустым для корневых отделов)
	ParentID *int64 `json:"parent_id"`
	// Название отдела
	// Example: Отдел продаж
	Name string `json:"name" example:"Отдел продаж"`
	// Описание отдела
	// Example: Отдел занимается продажами
	Description *string `json:"description" example:"Отдел занимается продажами"`
	// Дата создания
	// Example: 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Дата последнего обновления
	// Example: 2024-01-20T15:45:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-20T15:45:00Z"`
}

// ===== Position DTOs =====

// CreatePositionRequest запрос на создание позиции
// @Description Данные для создания новой должности в организации
type CreatePositionRequest struct {
	// ID отдела для привязки (опционально)
	// Example: 1
	DepartmentID *int64 `json:"department_id" example:"1"`
	// Название должности (обязательное)
	// Example: Менеджер по продажам
	Name string `json:"name" binding:"required" example:"Менеджер по продажам"`
	// Является ли позиция административной (имеет расширенные права)
	// Example: false
	IsAdmin bool `json:"is_admin" example:"false"`
	// Описание должностных обязанностей
	// Example: Работа с клиентами, заключение договоров
	Description *string `json:"description" example:"Работа с клиентами, заключение договоров"`
}

// UpdatePositionRequest запрос на обновление позиции
// @Description Данные для обновления должности
type UpdatePositionRequest struct {
	// Новый ID отдела
	// Example: 2
	DepartmentID *int64 `json:"department_id" example:"2"`
	// Новое название должности
	// Example: Старший менеджер по продажам
	Name string `json:"name" example:"Старший менеджер по продажам"`
	// Административная позиция
	// Example: true
	IsAdmin *bool `json:"is_admin" example:"true"`
	// Новое описание
	// Example: Руководство отделом продаж
	Description *string `json:"description" example:"Руководство отделом продаж"`
}

// PositionResponse данные позиции
// @Description Полная информация о должности
type PositionResponse struct {
	// Уникальный идентификатор позиции
	// Example: 1
	ID int64 `json:"id" example:"1"`
	// ID организации
	// Example: 1
	OrganizationID int64 `json:"organization_id" example:"1"`
	// ID отдела (null если не привязана к отделу)
	// Example: 1
	DepartmentID *int64 `json:"department_id" example:"1"`
	// Название должности
	// Example: Менеджер по продажам
	Name string `json:"name" example:"Менеджер по продажам"`
	// Является ли административной
	// Example: false
	IsAdmin bool `json:"is_admin" example:"false"`
	// Описание должности
	// Example: Работа с клиентами
	Description *string `json:"description" example:"Работа с клиентами"`
	// Дата создания
	// Example: 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	// Дата последнего обновления
	// Example: 2024-01-20T15:45:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-20T15:45:00Z"`
}

// ===== Member DTOs =====

// InviteMemberRequest запрос на приглашение участника
// @Description Данные для приглашения нового участника в организацию
type InviteMemberRequest struct {
	// ID организации для приглашения
	// Example: 1
	OrganizationID int64 `json:"organization_id" binding:"required" example:"1"`
	// Email приглашаемого пользователя
	// Example: newmember@example.com
	Email string `json:"email" binding:"required,email" example:"newmember@example.com"`
}

// MemberResponse данные участника организации
// @Description Информация об участнике организации
type MemberResponse struct {
	// Уникальный идентификатор участника
	// Example: 1
	ID int64 `json:"id" example:"1"`
	// ID организации
	// Example: 1
	OrganizationID int64 `json:"organization_id" example:"1"`
	// ID пользователя
	// Example: 5
	UserID int64 `json:"user_id" example:"5"`
	// Статус участника: invited, active, blocked
	// Example: active
	Status string `json:"status" example:"active" enums:"invited,active,blocked"`
	// Дата вступления в организацию
	// Example: 2024-01-15T10:30:00Z
	JoinedAt *time.Time `json:"joined_at" example:"2024-01-15T10:30:00Z"`
	// Данные пользователя (опционально)
	User UserResponse `json:"user,omitempty"`
}

// ===== Employee DTOs =====

// AssignPositionRequest запрос на назначение на должность
// @Description Данные для назначения участника на должность
type AssignPositionRequest struct {
	// ID участника организации
	// Example: 1
	MemberID int64 `json:"member_id" binding:"required" example:"1"`
	// ID должности
	// Example: 3
	PositionID int64 `json:"position_id" binding:"required" example:"3"`
	// Дата начала работы (по умолчанию - текущая дата)
	// Example: 2024-02-01T00:00:00Z
	StartDate *time.Time `json:"start_date" example:"2024-02-01T00:00:00Z"`
	// Является ли стажёром
	// Example: false
	IsIntern bool `json:"is_intern" example:"false"`
}

// UpdateEmployeeRequest запрос на обновление данных сотрудника
// @Description Данные для обновления информации о сотруднике
type UpdateEmployeeRequest struct {
	// Дата окончания работы на должности
	// Example: 2024-12-31T23:59:59Z
	EndDate *time.Time `json:"end_date" example:"2024-12-31T23:59:59Z"`
	// Статус стажёра
	// Example: false
	IsIntern *bool `json:"is_intern" example:"false"`
}

// EmployeeResponse данные сотрудника
// @Description Информация о сотруднике на должности
type EmployeeResponse struct {
	// Уникальный идентификатор записи о сотруднике
	// Example: 1
	ID int64 `json:"id" example:"1"`
	// ID участника организации
	// Example: 1
	MemberID int64 `json:"member_id" example:"1"`
	// ID должности
	// Example: 3
	PositionID int64 `json:"position_id" example:"3"`
	// Является ли стажёром
	// Example: false
	IsIntern bool `json:"is_intern" example:"false"`
	// Дата начала работы на должности
	// Example: 2024-02-01T00:00:00Z
	StartDate *time.Time `json:"start_date" example:"2024-02-01T00:00:00Z"`
	// Дата окончания работы на должности (может быть пустым если работает)
	EndDate *time.Time `json:"end_date"`
	// Информация о должности (опционально)
	Position PositionResponse `json:"position,omitempty"`
}
