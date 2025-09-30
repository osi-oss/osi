# 📧 Настройка Email для восстановления пароля

## 🔧 Настройка SMTP

### Gmail (рекомендуется для разработки)

1. **Включить 2FA** в вашем Google аккаунте
2. **Создать App Password:**
   - Перейти в Google Account settings
   - Security → 2-Step Verification → App passwords
   - Создать пароль для "Mail"
   - Использовать этот пароль в `SMTP_PASSWORD`

3. **Настроить переменные окружения:**
```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-16-char-app-password
FROM_EMAIL=your-email@gmail.com
FROM_NAME=OSI Team
BASE_URL=http://localhost:3000
```

### Mailgun (для production)

```bash
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USER=postmaster@your-domain.mailgun.org
SMTP_PASSWORD=your-mailgun-password
FROM_EMAIL=noreply@your-domain.com
FROM_NAME=Your App Name
BASE_URL=https://your-app.com
```

### SendGrid

```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=your-sendgrid-api-key
FROM_EMAIL=noreply@your-domain.com
FROM_NAME=Your App Name
BASE_URL=https://your-app.com
```

## 📋 API Endpoints

### Запрос сброса пароля
```bash
POST /api/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Ответ:**
```json
{
  "message": "If email exists, password reset instructions have been sent"
}
```

### Сброс пароля
```bash
POST /api/reset-password
Content-Type: application/json

{
  "token": "abc123...",
  "new_password": "newPassword123"
}
```

**Ответ:**
```json
{
  "message": "Password has been reset successfully"
}
```

## 🔄 Процесс восстановления

1. **Пользователь запрашивает сброс** → `POST /api/forgot-password`
2. **Система генерирует токен** и сохраняет в БД (срок жизни: 1 час)
3. **Отправляется email** с ссылкой: `https://your-app.com/reset-password?token=abc123`
4. **Пользователь переходит по ссылке** → фронтенд показывает форму
5. **Пользователь вводит новый пароль** → `POST /api/reset-password`
6. **Токен проверяется** и пароль обновляется

## 🛡️ Безопасность

### Особенности реализации:
- **Токены одноразовые** - помечаются как использованные
- **Ограниченный срок жизни** - 1 час
- **Случайная генерация** - 32 байта криптографически стойкий
- **Не раскрываем информацию** - одинаковый ответ для существующих и несуществующих email
- **Очистка старых токенов** - автоматическое удаление просроченных

### Рекомендации:
- **Rate limiting** для `/forgot-password` (максимум 5 запросов в час)
- **HTTPS обязательно** для production
- **Логирование** всех попыток сброса пароля
- **Уведомления** пользователя о смене пароля

## 🧪 Тестирование

### Локальное тестирование с MailHog

1. **Запустить MailHog:**
```bash
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
```

2. **Настроить переменные:**
```bash
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
FROM_EMAIL=test@localhost
FROM_NAME=Test Team
BASE_URL=http://localhost:3000
```

3. **Проверить письма:** http://localhost:8025

### Тестирование API

```bash
# Запрос сброса
curl -X POST http://localhost:8080/api/forgot-password \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# Сброс пароля (получить токен из письма)
curl -X POST http://localhost:8080/api/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token":"your-token-here","new_password":"newpass123"}'
```

## 🐛 Отладка

### Частые проблемы:

1. **"Authentication failed"** - проверьте App Password для Gmail
2. **"Connection timeout"** - проверьте SMTP_HOST и SMTP_PORT
3. **"Invalid token"** - токен истек или уже использован
4. **Письма не приходят** - проверьте spam папку

### Логирование:
```go
log.Printf("Sending password reset email to: %s", email)
log.Printf("Generated reset token for user ID: %d", userID)
log.Printf("Reset token used for user ID: %d", userID)
```

## 📝 Frontend интеграция

### React пример:

```javascript
// Запрос сброса
const requestReset = async (email) => {
  const response = await fetch('/api/forgot-password', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email })
  });
  const data = await response.json();
  alert(data.message);
};

// Сброс пароля
const resetPassword = async (token, newPassword) => {
  const response = await fetch('/api/reset-password', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, new_password: newPassword })
  });
  
  if (response.ok) {
    alert('Password reset successfully!');
    // Redirect to login page
  } else {
    const error = await response.json();
    alert(error.error);
  }
};
```