# Тестирование API организаций

## Предварительные требования

1. Запустите сервер
2. Зарегистрируйтесь и залогиньтесь для получения JWT токена

## Пример использования

### 1. Регистрация и вход

```bash
# Регистрация
curl -X POST http://localhost:8080/api/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'

# Вход (получите токен)
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}' \
  -c cookies.txt
```

### 2. Создание организации

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

Или с использованием cookie:

```bash
curl -X POST http://localhost:8080/api/organizations \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Моя компания",
    "legal_name": "ООО Моя компания",
    "inn": "1234567890"
  }'
```

### 3. Получение списка организаций

```bash
curl -X GET http://localhost:8080/api/organizations \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 4. Получение конкретной организации

```bash
curl -X GET http://localhost:8080/api/organizations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 5. Обновление организации

```bash
curl -X PUT http://localhost:8080/api/organizations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Обновленное название",
    "legal_name": "ООО Обновленное название"
  }'
```

### 6. Удаление организации

```bash
curl -X DELETE http://localhost:8080/api/organizations/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Проверка в базе данных

```sql
-- Просмотр организаций
SELECT * FROM organizations;

-- Просмотр основателей
SELECT 
  of.id,
  of.organization_id,
  of.user_id,
  of.share_percent,
  of.is_main,
  u.email as user_email,
  o.name as org_name
FROM organization_founders of
JOIN users u ON u.id = of.user_id
JOIN organizations o ON o.id = of.organization_id;
```
