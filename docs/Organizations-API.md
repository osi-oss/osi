# Organizations API

API для управления организациями. Все эндпоинты требуют авторизации.

## Endpoints

### 1. Создать организацию
**POST** `/api/organizations`

Создает новую организацию. Авторизованный пользователь автоматически становится главным основателем (founder).

**Headers:**
```
Authorization: Bearer <token>
```
или
```
Cookie: auth_token=<token>
```

**Request Body:**
```json
{
  "name": "Название организации",          // required
  "legal_name": "Юридическое название",     // optional
  "inn": "1234567890",                      // optional
  "ogrn": "1234567890123",                  // optional
  "kpp": "123456789",                       // optional
  "legal_address": "Юридический адрес",     // optional
  "share_percent": 100.0                    // optional, доля в процентах
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "Название организации",
  "legal_name": "Юридическое название",
  "inn": "1234567890",
  "ogrn": "1234567890123",
  "kpp": "123456789",
  "legal_address": "Юридический адрес",
  "status": "draft",
  "created_at": "2026-01-02T10:00:00Z"
}
```

**Errors:**
- `400 Bad Request` - неверные данные
- `401 Unauthorized` - требуется авторизация
- `500 Internal Server Error` - ошибка сервера

---

### 2. Получить список организаций пользователя
**GET** `/api/organizations`

Возвращает список всех организаций, в которых пользователь является основателем.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "organizations": [
    {
      "id": 1,
      "name": "Организация 1",
      "legal_name": "ООО Организация 1",
      "inn": "1234567890",
      "ogrn": "1234567890123",
      "kpp": "123456789",
      "legal_address": "Адрес 1",
      "status": "draft",
      "created_at": "2026-01-02T10:00:00Z",
      "updated_at": "2026-01-02T10:00:00Z"
    },
    {
      "id": 2,
      "name": "Организация 2",
      "status": "approved",
      "created_at": "2026-01-01T09:00:00Z",
      "updated_at": "2026-01-01T09:30:00Z"
    }
  ]
}
```

**Errors:**
- `401 Unauthorized` - требуется авторизация
- `500 Internal Server Error` - ошибка сервера

---

### 3. Получить организацию по ID
**GET** `/api/organizations/:id`

Возвращает детальную информацию об организации.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Название организации",
  "legal_name": "Юридическое название",
  "inn": "1234567890",
  "ogrn": "1234567890123",
  "kpp": "123456789",
  "legal_address": "Юридический адрес",
  "status": "draft",
  "created_at": "2026-01-02T10:00:00Z",
  "updated_at": "2026-01-02T10:00:00Z"
}
```

**Errors:**
- `400 Bad Request` - неверный ID
- `401 Unauthorized` - требуется авторизация
- `403 Forbidden` - нет доступа к организации
- `404 Not Found` - организация не найдена
- `500 Internal Server Error` - ошибка сервера

---

### 4. Обновить организацию
**PUT** `/api/organizations/:id`

Обновляет данные организации. Доступно только основателям организации.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "name": "Новое название",
  "legal_name": "Новое юридическое название",
  "inn": "9876543210",
  "ogrn": "9876543210987",
  "kpp": "987654321",
  "legal_address": "Новый адрес"
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Новое название",
  "legal_name": "Новое юридическое название",
  "inn": "9876543210",
  "ogrn": "9876543210987",
  "kpp": "987654321",
  "legal_address": "Новый адрес",
  "status": "draft",
  "created_at": "2026-01-02T10:00:00Z",
  "updated_at": "2026-01-02T11:00:00Z"
}
```

**Errors:**
- `400 Bad Request` - неверные данные
- `401 Unauthorized` - требуется авторизация
- `403 Forbidden` - нет доступа к организации
- `404 Not Found` - организация не найдена
- `500 Internal Server Error` - ошибка сервера

---

### 5. Удалить организацию
**DELETE** `/api/organizations/:id`

Удаляет организацию. Доступно только главному основателю (IsMain=true).

**Headers:**
```
Authorization: Bearer <token>
```

**Response:** `200 OK`
```json
{
  "message": "organization deleted successfully"
}
```

**Errors:**
- `400 Bad Request` - неверный ID
- `401 Unauthorized` - требуется авторизация
- `403 Forbidden` - только главный основатель может удалить организацию
- `404 Not Found` - организация не найдена
- `500 Internal Server Error` - ошибка сервера

---

## Статусы организаций

- `draft` - черновик (по умолчанию при создании)
- `pending` - на рассмотрении
- `approved` - одобрена
- `rejected` - отклонена

---

## Примеры использования

### Создание организации (с curl)

```bash
curl -X POST http://localhost:8080/api/organizations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Моя компания",
    "legal_name": "ООО Моя компания",
    "inn": "1234567890",
    "share_percent": 100.0
  }'
```

### Получение списка организаций

```bash
curl -X GET http://localhost:8080/api/organizations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Обновление организации

```bash
curl -X PUT http://localhost:8080/api/organizations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Обновленное название",
    "legal_name": "ООО Обновленное название"
  }'
```

### Удаление организации

```bash
curl -X DELETE http://localhost:8080/api/organizations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Бизнес-логика

1. При создании организации автоматически создается запись в `organization_founders` с:
   - `user_id` - ID текущего пользователя
   - `is_main` = `true` - главный основатель
   - `share_percent` - доля в процентах (опционально)

2. Доступ к организации имеют только основатели (founders)

3. Удалить организацию может только главный основатель (`is_main=true`)

4. Обновлять организацию могут все основатели

5. Статус организации по умолчанию - `draft`
