# 🚀 Шпаргалка: API для фронтенда

## Быстрый старт
```bash
# Запуск API и БД
docker-compose up -d

# Проверка работы
curl http://localhost:8080/health

# Остановка
docker-compose down
```

## API настройки
- **URL**: `http://localhost:8080`
- **Порт можно изменить в**: `.env` → `BACKEND_PORT=8080`

## Основные команды

| Команда | Описание |
|---------|----------|
| `docker-compose up -d` | Запуск всех сервисов |
| `docker-compose logs backend` | Логи API |
| `docker-compose restart backend` | Перезапуск API |
| `docker-compose ps` | Статус контейнеров |
| `docker-compose down` | Остановка всех сервисов |

## Отладка

### API не отвечает
```bash
docker-compose logs backend
docker-compose restart backend
```

### Порт занят
```bash
# Проверить процессы на порту 8080
lsof -i :8080

# Изменить порт в .env
BACKEND_PORT=8081
```

### База недоступна
```bash
docker-compose logs postgres
docker-compose restart postgres
```

## Конфигурация фронтенда

### React/Vue/Angular
```javascript
const API_URL = 'http://localhost:8080';

// Проверка API
fetch(`${API_URL}/health`)
  .then(r => r.json())
  .then(console.log);
```

### Axios
```javascript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080'
});
```

## Проблемы?
1. `docker-compose ps` - проверить статус
2. `docker-compose logs backend` - посмотреть ошибки  
3. `docker-compose restart backend` - перезапустить
4. Спросить backend команду 😊
