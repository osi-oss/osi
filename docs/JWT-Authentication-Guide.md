# 🔐 JWT Аутентификация через куки

## 📋 Доступные endpoints

### Регистрация пользователя
```bash
POST /api/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "message": "User created successfully",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "verified": false,
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### Вход в систему
```bash
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Куки:** Автоматически устанавливается `auth_token` cookie с JWT токеном

### Получение профиля (защищенный)
```bash
GET /api/profile
Authorization: Bearer <token>  # или через куки
```

**Ответ:**
```json
{
  "message": "Profile retrieved successfully",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "",
    "last_name": "",
    "verified": false,
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### Выход из системы
```bash
POST /api/logout
Authorization: Bearer <token>  # или через куки
```

**Ответ:**
```json
{
  "message": "Logged out successfully"
}
```

## 🔧 Middleware

### AuthRequired
- Проверяет JWT токен из куки `auth_token` или `Authorization` заголовка
- Автоматически добавляет `userID` и `email` в gin.Context
- Возвращает 401 при отсутствии или невалидном токене

### CORS
- Разрешает cross-origin запросы
- Поддерживает куки (`Access-Control-Allow-Credentials: true`)
- Обрабатывает OPTIONS preflight запросы

### ValidateJSON
- Проверяет `Content-Type: application/json` для POST/PUT запросов

## 🍪 Работа с куками

### Настройки куки
```go
c.SetCookie(
    "auth_token",  // name
    token,         // value
    3600*24,       // maxAge (24 часа)
    "/",           // path
    "",            // domain
    true,          // secure (только HTTPS)
    true,          // httpOnly (недоступно из JS)
)
```

### JavaScript пример
```javascript
// Вход через fetch (куки устанавливаются автоматически)
fetch('/api/login', {
  method: 'POST',
  credentials: 'include', // Важно для куки!
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
});

// Защищенный запрос (куки отправляются автоматически)
fetch('/api/profile', {
  method: 'GET',
  credentials: 'include' // Важно для куки!
});
```

## ⚠️ Безопасность

1. **HTTPS обязательно в production** для secure куки
2. **HttpOnly куки** защищают от XSS атак
3. **SameSite** политика для защиты от CSRF (можно добавить)
4. **Короткий срок жизни токенов** (24 часа по умолчанию)

## 🐛 Отладка

### Проверка токена в куки
```bash
# Chrome DevTools -> Application -> Cookies
# Или через curl
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

curl -b cookies.txt http://localhost:8080/api/profile
```

### Возможные ошибки
- `Authorization token required` - нет токена в куки или заголовках
- `Invalid or expired token` - токен поврежден или истек срок
- `Invalid token claims` - неверная структура токена
- `Content-Type must be application/json` - неверный Content-Type

## 📝 Примеры cURL

```bash
# Регистрация
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Вход (сохраняем куки)
curl -c cookies.txt -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Профиль (используем куки)
curl -b cookies.txt http://localhost:8080/api/profile

# Выход
curl -b cookies.txt -X POST http://localhost:8080/api/logout
```