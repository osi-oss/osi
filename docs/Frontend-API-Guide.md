# Инструкция для фронтенд разработчика: Работа с API через Docker Compose

## 📋 Общая информация

Этот проект использует Docker Compose для запуска backend API и PostgreSQL базы данных. Фронтенд может работать с API локально или через Docker.

## 🚀 Быстрый старт

### 1. Настройка окружения

```bash
# Клонируйте репозиторий
git clone <repository-url>
cd osi

# Скопируйте примеры конфигурации
cp .env.example .env
cp backend/.env.example backend/.env

# Запустите API и базу данных
docker-compose up -d
```

### 2. Проверка работы API

```bash
# Проверьте статус сервисов
docker-compose ps

# Проверьте логи API
docker-compose logs backend

# Проверьте доступность API
curl http://localhost:8080/health
```

## 🌐 Доступ к API

### Базовые настройки
- **API URL**: `http://localhost:8080`
- **База данных**: PostgreSQL на порту `5432` (только для backend)

### Переменные окружения (из .env)
```properties
# Порт API
BACKEND_PORT=8080

# PostgreSQL (автоматически настраивается)
POSTGRES_DB=osi_db
POSTGRES_USER=osi_user
POSTGRES_PASSWORD=osi_password
POSTGRES_PORT=5432
```

## 🔧 Команды для разработки

### Управление сервисами

```bash
# Запуск всех сервисов
docker-compose up -d

# Запуск только API (если база уже работает)
docker-compose up backend -d

# Запуск только базы данных
docker-compose up postgres -d

# Остановка всех сервисов
docker-compose down

# Перезапуск API после изменений
docker-compose restart backend

# Пересборка API после изменений в коде
docker-compose up --build backend
```

### Мониторинг и отладка

```bash
# Просмотр логов всех сервисов
docker-compose logs -f

# Просмотр логов только API
docker-compose logs -f backend

# Просмотр логов только базы данных
docker-compose logs -f postgres

# Проверка состояния контейнеров
docker-compose ps

# Вход в контейнер API для отладки
docker-compose exec backend sh
```

## 📡 Примеры работы с API

### Проверка здоровья API
```javascript
// GET запрос для проверки доступности
fetch('http://localhost:8080/health')
  .then(response => response.json())
  .then(data => console.log('API работает:', data));
```

### Базовая конфигурация для фронтенда

#### React/Vue/Angular
```javascript
// config/api.js
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

export const apiClient = {
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  }
};

// Пример запроса
export const getUsers = async () => {
  const response = await fetch(`${API_BASE_URL}/api/users`);
  return response.json();
};
```

#### Axios конфигурация
```javascript
// services/api.js
import axios from 'axios';

const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

export default api;
```

### CORS настройки
API настроен на прием запросов с любых доменов в режиме разработки. В production нужно будет ограничить CORS настройки.

## 🔄 Workflow для разработки

### При работе с новой веткой

```bash
# 1. Переключитесь на новую ветку
git checkout -b feature/my-feature

# 2. Убедитесь что .env файлы настроены
ls -la .env backend/.env

# 3. Запустите API
docker-compose up -d

# 4. Разрабатывайте фронтенд, используя http://localhost:8080
```

### При изменениях в API

```bash
# Если изменился только код API
docker-compose restart backend

# Если изменились зависимости или Dockerfile
docker-compose up --build backend

# Если нужно полностью пересоздать
docker-compose down
docker-compose up --build
```

## 🐛 Отладка проблем

### API недоступен
```bash
# Проверьте статус контейнеров
docker-compose ps

# Проверьте логи API
docker-compose logs backend

# Проверьте сетевое подключение
curl -v http://localhost:8080/health
```

### База данных недоступна
```bash
# Проверьте статус PostgreSQL
docker-compose logs postgres

# Подключитесь к базе напрямую
docker-compose exec postgres psql -U osi_user -d osi_db
```

### Проблемы с портами
```bash
# Проверьте занятые порты
lsof -i :8080
lsof -i :5432

# Измените порты в .env если нужно
BACKEND_PORT=8081
POSTGRES_PORT=5433
```

## 📝 Полезные ссылки

- **API документация**: `http://localhost:8080/docs` (если настроена)
- **Админ панель PostgreSQL**: Можно настроить pgAdmin через docker-compose
- **Мониторинг**: `docker-compose logs -f` для отслеживания в реальном времени

## 🎯 Рекомендации

1. **Всегда проверяйте статус API** перед началом работы: `docker-compose ps`
2. **Следите за логами** при возникновении ошибок: `docker-compose logs backend`
3. **Используйте переменные окружения** для URL API в фронтенде
4. **Не коммитьте .env файлы** - они в .gitignore
5. **При проблемах сначала перезапустите** API: `docker-compose restart backend`

## 💡 Продвинутые возможности

### Подключение к базе данных для отладки
```bash
# Через Docker
docker-compose exec postgres psql -U osi_user -d osi_db

# Через локальный клиент (если установлен psql)
psql -h localhost -p 5432 -U osi_user -d osi_db
```

### Настройка окружений
Можно создать разные .env файлы для разных окружений:
- `.env.development`
- `.env.staging`
- `.env.production`

---

**Нужна помощь?** Обращайтесь к backend разработчикам или смотрите логи: `docker-compose logs backend`

---

## 🎟️ Приглашения в организацию (Invites)

### Концепция

Система приглашений позволяет добавлять пользователей в организацию на конкретные должности по email.

**Ключевые особенности:**
- ✅ Приглашение может быть создано только основателем организации
- ✅ Email используется как адрес уведомления
- ✅ Приглашение принимается только залогиненным пользователем
- ✅ Нет необходимости в специальных токенах (без invite-token'ов)

### Workflow приглашения

#### 1️⃣ Создание приглашения (Administrator)

```http
POST /api/organizations/{orgId}/invites
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "email": "john.doe@company.com",
  "position_id": 42
}
```

**Ответ (201 Created):**
```json
{
  "id": 1,
  "organization_id": 5,
  "position_id": 42,
  "invited_email": "john.doe@company.com",
  "invited_user_id": 10,
  "invited_by_user_id": 1,
  "status": "pending",
  "created_at": "2024-01-24T10:30:00Z",
  "accepted_at": null,
  "declined_at": null
}
```

**Что происходит:**
- Проверяются права доступа (только основатель организации)
- Проверяется валидность должности (принадлежит организации)
- Если пользователя с таким email нет, он автоматически создаётся со статусом `pending_email`
- Отправляется email с приглашением (содержит ссылку для UX редиректа)

#### 2️⃣ Пользователь получает письмо

Письмо содержит:
- Название организации и должности
- Инструкции как присоединиться
- Ссылка для UX редиректа (опционально): `https://app.com/login?invite_org_id=5&invite_position_id=42`

#### 3️⃣ Пользователь логинится

Обычный flow аутентификации:
```http
POST /api/auth/request-code
Content-Type: application/json

{"email": "john.doe@company.com"}
```

После получения кода:
```http
POST /api/auth/verify-code
Content-Type: application/json

{
  "email": "john.doe@company.com",
  "code": "A3K7"
}
```

**Получает JWT токен**

#### 4️⃣ Принятие приглашения (User)

```http
POST /api/invites/{inviteId}/accept
Authorization: Bearer <JWT_TOKEN>
```

**Ответ (200 OK):**
```json
{
  "message": "Invite accepted successfully"
}
```

**Что происходит:**
- Проверяется, что email залогиненного пользователя совпадает с invited_email
- Статус приглашения меняется на `accepted`
- Пользователь добавляется в организацию как активный члена
- Автоматически создаётся запись сотрудника на этой должности

#### 5️⃣ Отклонение приглашения (User)

```http
POST /api/invites/{inviteId}/decline
Authorization: Bearer <JWT_TOKEN>
```

**Ответ (200 OK):**
```json
{
  "message": "Invite declined successfully"
}
```

### Endpoints для управления приглашениями

| Метод | Путь | Действие | Права |
|-------|------|----------|-------|
| `POST` | `/api/organizations/{orgId}/invites` | Создать приглашение | Основатель организации |
| `GET` | `/api/invites/my` | Мои приглашения | Залогиненный пользователь |
| `POST` | `/api/invites/{inviteId}/accept` | Принять приглашение | Должен быть email == invited_email |
| `POST` | `/api/invites/{inviteId}/decline` | Отклонить приглашение | Должен быть email == invited_email |
| `DELETE` | `/api/invites/{inviteId}` | Отменить приглашение | Создатель или основатель организации |
| `GET` | `/api/organizations/{orgId}/invites` | Все приглашения организации | Основатель организации |

### Статусы приглашения

- `pending` — ожидает принятия/отклонения
- `accepted` — принято, пользователь добавлен в организацию
- `declined` — отклонено пользователем

### Обработка ошибок

```json
// Email не валиден
{
  "error": "invited email is required"
}

// Пользователь уже член организации
{
  "error": "user is already an active member of this organization"
}

// Нет pending приглашения
{
  "error": "invite is not pending"
}

// Нет прав на операцию
{
  "error": "forbidden"
}

// Позиция не принадлежит организации
{
  "error": "position does not belong to this organization"
}
```

### Рекомендации для фронтенда

1. **При создании приглашения:**
   - Валидируйте email на фронтенде
   - Проверяйте ошибки и показывайте пользователю понятные сообщения
   - Подтвердите действие перед отправкой

2. **При логине пользователя, приглашённого по email:**
   - Проверяйте есть ли pending приглашения в `/api/invites/my` после успешного логина
   - Если есть приглашение, покажите UI для его принятия
   - Используйте `?invite_org_id` и `?invite_position_id` из ссылки письма для автоматического выделения приглашения

3. **В интерфейсе приглашений:**
   - Показывайте статус приглашения
   - Позволяйте отклонить или принять
   - Показывайте дату создания и статус обработки

```
