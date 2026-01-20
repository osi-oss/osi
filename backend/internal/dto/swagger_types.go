package dto

// ===== Swagger Response Types =====
// Эти типы используются только для генерации OpenAPI документации

// ErrorResponse представляет стандартный ответ с ошибкой
// @Description Стандартный формат ответа при ошибке
type ErrorResponse struct {
	// Сообщение об ошибке
	// Example: invalid email or password
	Error string `json:"error" example:"invalid email or password"`
}

// MessageResponse представляет ответ с сообщением
// @Description Стандартный формат успешного ответа с сообщением
type MessageResponse struct {
	// Сообщение о результате операции
	// Example: operation completed successfully
	Message string `json:"message" example:"operation completed successfully"`
}

// ===== Auth Responses =====

// SignUpResponse ответ при успешной регистрации
// @Description Ответ при успешной регистрации пользователя
type SignUpResponse struct {
	// Сообщение о результате
	// Example: User created successfully
	Message string `json:"message" example:"User created successfully"`
	// Данные созданного пользователя
	User UserResponse `json:"user"`
}

// LoginResponse ответ при успешной авторизации
// @Description Ответ при успешной авторизации с JWT токеном
type LoginResponse struct {
	// Сообщение о результате
	// Example: Login successful
	Message string `json:"message" example:"Login successful"`
	// JWT токен для авторизации запросов
	// Example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ"`
}

// ProfileResponse ответ с профилем пользователя
// @Description Ответ с данными профиля пользователя
type ProfileResponse struct {
	// Сообщение о результате
	// Example: Profile retrieved successfully
	Message string `json:"message" example:"Profile retrieved successfully"`
	// Данные пользователя
	User UserResponse `json:"user"`
}

// TokenValidationResponse ответ проверки токена
// @Description Результат проверки токена сброса пароля
type TokenValidationResponse struct {
	// Валиден ли токен
	// Example: true
	Valid bool `json:"valid" example:"true"`
	// Сообщение о результате
	// Example: Token is valid
	Message string `json:"message" example:"Token is valid"`
}

// ===== Organization Responses =====

// OrganizationsListResponse список организаций
// @Description Список организаций пользователя
type OrganizationsListResponse struct {
	// Массив организаций
	Organizations []OrganizationResponse `json:"organizations"`
}

// ===== Location Responses =====

// LocationsListResponse список локаций
// @Description Список локаций организации
type LocationsListResponse struct {
	// Массив локаций
	Locations []LocationResponse `json:"locations"`
}

// ===== Department Responses =====

// DepartmentsListResponse список отделов
// @Description Список отделов локации
type DepartmentsListResponse struct {
	// Массив отделов
	Departments []DepartmentResponse `json:"departments"`
}

// ===== Position Responses =====

// PositionsListResponse список позиций
// @Description Список позиций организации или отдела
type PositionsListResponse struct {
	// Массив позиций
	Positions []PositionResponse `json:"positions"`
}
