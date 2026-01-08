# Locations API

API для управления филиалами/локациями организаций.

## Endpoints

### 1. Создать локацию

**POST** `/api/organizations/:id/locations`

Создает новую локацию/филиал для организации. Только создатель организации (founder) может создавать локации.

**Headers:**
- `Authorization: Bearer <token>`

**URL Parameters:**
- `id` - ID организации

**Request Body:**
```json
{
  "name": "Main Office",
  "address": "123 Main St, Moscow",
  "source": "manual",
  "is_verified": false
}
```

**Fields:**
- `name` (string, required) - Название локации/филиала
- `address` (string, optional) - Адрес локации
- `source` (string, required) - Источник данных: `"manual"` или `"registry"`
- `is_verified` (boolean, optional) - Статус верификации (по умолчанию `false`)

**Response:** `201 Created`
```json
{
  "id": 1,
  "organization_id": 5,
  "name": "Main Office",
  "address": "123 Main St, Moscow",
  "source": "manual",
  "is_verified": false,
  "is_active": true,
  "created_at": "2026-01-08T12:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Невалидные данные
- `401 Unauthorized` - Пользователь не авторизован
- `403 Forbidden` - Пользователь не является создателем организации
- `404 Not Found` - Организация не найдена

---

### 2. Получить все локации организации

**GET** `/api/organizations/:id/locations`

Получает список всех локаций организации.

**Headers:**
- `Authorization: Bearer <token>`

**URL Parameters:**
- `id` - ID организации

**Response:** `200 OK`
```json
{
  "locations": [
    {
      "id": 1,
      "organization_id": 5,
      "name": "Main Office",
      "address": "123 Main St, Moscow",
      "source": "manual",
      "is_verified": false,
      "is_active": true,
      "created_at": "2026-01-08T12:00:00Z",
      "updated_at": "2026-01-08T12:00:00Z"
    },
    {
      "id": 2,
      "organization_id": 5,
      "name": "Branch Office",
      "address": "456 Side St, St. Petersburg",
      "source": "registry",
      "is_verified": true,
      "is_active": true,
      "created_at": "2026-01-08T13:00:00Z",
      "updated_at": "2026-01-08T13:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `401 Unauthorized` - Пользователь не авторизован
- `403 Forbidden` - Нет доступа к организации
- `404 Not Found` - Организация не найдена

---

### 3. Получить конкретную локацию

**GET** `/api/locations/:id`

Получает информацию о конкретной локации по её ID.

**Headers:**
- `Authorization: Bearer <token>`

**URL Parameters:**
- `id` - ID локации

**Response:** `200 OK`
```json
{
  "id": 1,
  "organization_id": 5,
  "name": "Main Office",
  "address": "123 Main St, Moscow",
  "source": "manual",
  "is_verified": false,
  "is_active": true,
  "created_at": "2026-01-08T12:00:00Z",
  "updated_at": "2026-01-08T12:00:00Z"
}
```

**Error Responses:**
- `401 Unauthorized` - Пользователь не авторизован
- `403 Forbidden` - Нет доступа к организации локации
- `404 Not Found` - Локация не найдена

---

### 4. Обновить локацию

**PUT** `/api/locations/:id`

Обновляет информацию о локации. Только создатель организации может обновлять локации.

**Headers:**
- `Authorization: Bearer <token>`

**URL Parameters:**
- `id` - ID локации

**Request Body:**
```json
{
  "name": "Updated Office",
  "address": "789 New St, Moscow",
  "source": "registry",
  "is_verified": true,
  "is_active": false
}
```

**Fields (все опциональные):**
- `name` (string) - Новое название локации
- `address` (string) - Новый адрес
- `source` (string) - Источник данных
- `is_verified` (boolean) - Статус верификации
- `is_active` (boolean) - Статус активности

**Response:** `200 OK`
```json
{
  "id": 1,
  "organization_id": 5,
  "name": "Updated Office",
  "address": "789 New St, Moscow",
  "source": "registry",
  "is_verified": true,
  "is_active": false,
  "created_at": "2026-01-08T12:00:00Z",
  "updated_at": "2026-01-08T15:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request` - Невалидные данные
- `401 Unauthorized` - Пользователь не авторизован
- `403 Forbidden` - Нет прав на обновление
- `404 Not Found` - Локация не найдена

---

### 5. Удалить локацию

**DELETE** `/api/locations/:id`

Удаляет локацию. Только создатель организации может удалять локации.

**Headers:**
- `Authorization: Bearer <token>`

**URL Parameters:**
- `id` - ID локации

**Response:** `200 OK`
```json
{
  "message": "location deleted successfully"
}
```

**Error Responses:**
- `401 Unauthorized` - Пользователь не авторизован
- `403 Forbidden` - Нет прав на удаление
- `404 Not Found` - Локация не найдена

---

## Примеры использования

### cURL

#### Создание локации
```bash
curl -X POST http://localhost:8080/api/organizations/1/locations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Main Office",
    "address": "123 Main St, Moscow",
    "source": "manual",
    "is_verified": false
  }'
```

#### Получение всех локаций организации
```bash
curl -X GET http://localhost:8080/api/organizations/1/locations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Получение конкретной локации
```bash
curl -X GET http://localhost:8080/api/locations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Обновление локации
```bash
curl -X PUT http://localhost:8080/api/locations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Office",
    "is_verified": true
  }'
```

#### Удаление локации
```bash
curl -X DELETE http://localhost:8080/api/locations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Права доступа

- **Создание локации**: только создатель (founder) организации
- **Просмотр локаций**: любой участник организации
- **Обновление локации**: только создатель организации
- **Удаление локации**: только создатель организации

---

## Модель данных

```go
type Location struct {
    ID             int64
    OrganizationID int64
    Name           string
    Address        *string
    Source         string  // 'registry' | 'manual'
    IsVerified     bool
    IsActive       bool
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```
