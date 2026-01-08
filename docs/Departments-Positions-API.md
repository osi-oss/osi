# Departments & Positions API

API для управления отделами и позициями в организациях.

## Departments (Отделы)

### Endpoints

#### 1. Создать отдел в локации
**POST** `/api/locations/:id/departments`

```json
{
  "name": "IT Department",
  "parent_id": null,
  "description": "Information Technology Department"
}
```

**Response:** `201 Created`

#### 2. Получить все отделы локации
**GET** `/api/locations/:id/departments`

**Response:** `200 OK`
```json
{
  "departments": [
    {
      "id": 1,
      "location_id": 5,
      "parent_id": null,
      "name": "IT Department",
      "description": "Information Technology",
      "created_at": "2026-01-08T12:00:00Z",
      "updated_at": "2026-01-08T12:00:00Z"
    }
  ]
}
```

#### 3. Получить отдел
**GET** `/api/departments/:id`

#### 4. Обновить отдел
**PUT** `/api/departments/:id`

```json
{
  "name": "Updated Department Name",
  "parent_id": 2,
  "description": "Updated description"
}
```

#### 5. Удалить отдел
**DELETE** `/api/departments/:id`

**Note:** Нельзя удалить отдел, у которого есть дочерние отделы.

---

## Positions (Позиции)

### Endpoints

#### 1. Создать позицию в организации
**POST** `/api/organizations/:id/positions`

```json
{
  "name": "Software Engineer",
  "department_id": 1,
  "is_admin": false,
  "description": "Develops software applications"
}
```

**Response:** `201 Created`

#### 2. Получить все позиции организации
**GET** `/api/organizations/:id/positions`

**Response:** `200 OK`
```json
{
  "positions": [
    {
      "id": 1,
      "organization_id": 5,
      "department_id": 1,
      "name": "Software Engineer",
      "is_admin": false,
      "description": "Develops software",
      "created_at": "2026-01-08T12:00:00Z",
      "updated_at": "2026-01-08T12:00:00Z"
    }
  ]
}
```

#### 3. Получить позиции отдела
**GET** `/api/departments/:id/positions`

#### 4. Получить позицию
**GET** `/api/positions/:id`

#### 5. Обновить позицию
**PUT** `/api/positions/:id`

```json
{
  "name": "Senior Software Engineer",
  "department_id": 2,
  "is_admin": false,
  "description": "Senior level developer"
}
```

#### 6. Удалить позицию
**DELETE** `/api/positions/:id`

---

## Примеры использования

### Создание структуры отдела с позициями

```bash
# 1. Создаем главный отдел
curl -X POST http://localhost:8080/api/locations/1/departments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "IT Department",
    "description": "Information Technology"
  }'

# 2. Создаем подотдел
curl -X POST http://localhost:8080/api/locations/1/departments \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Development Team",
    "parent_id": 1,
    "description": "Software Development"
  }'

# 3. Создаем позиции в отделе
curl -X POST http://localhost:8080/api/organizations/1/positions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Junior Developer",
    "department_id": 2,
    "is_admin": false
  }'

curl -X POST http://localhost:8080/api/organizations/1/positions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tech Lead",
    "department_id": 2,
    "is_admin": true
  }'
```

---

## Права доступа

- **Создание:** только создатель (founder) организации
- **Просмотр:** любой участник организации  
- **Обновление:** только создатель организации
- **Удаление:** только создатель организации

---

## Модели данных

### Department
```go
type Department struct {
    ID          int64
    LocationID  int64
    ParentID    *int64
    Name        string
    Description *string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Position
```go
type Position struct {
    ID             int64
    OrganizationID int64
    DepartmentID   *int64
    Name           string
    IsAdmin        bool
    Description    *string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

---

## Иерархия структуры

```
Organization
  └── Location (Филиал)
      └── Department (Отдел)
          ├── Department (Подотдел)
          └── Position (Позиция в отделе)
      └── Position (Позиция без отдела)
```

Позиции могут быть созданы как внутри отдела (с `department_id`), так и напрямую в организации (без `department_id`).
