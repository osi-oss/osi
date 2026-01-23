package dto

// ===== Swagger Response Types =====
// Эти типы используются только для генерации OpenAPI документации

// ErrorResponse представляет стандартный ответ с ошибкой
// @Description Стандартный формат ответа при ошибке
type ErrorResponse struct {
	// Сообщение об ошибке
	// Example: invalid email or code
	Error string `json:"error" example:"invalid email or code"`
}

// MessageResponse представляет ответ с сообщением
// @Description Стандартный формат успешного ответа с сообщением
type MessageResponse struct {
	// Сообщение о результате операции
	// Example: operation completed successfully
	Message string `json:"message" example:"operation completed successfully"`
}

// ===== Auth Responses =====

// ProfileResponse ответ с профилем пользователя
// @Description Ответ с данными профиля пользователя
type ProfileResponse struct {
	// Сообщение о результате
	// Example: Profile retrieved successfully
	Message string `json:"message" example:"Profile retrieved successfully"`
	// Данные пользователя
	User UserResponse `json:"user"`
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
