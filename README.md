# OSI Web Service

MVP веб-сервиса с Go backend и React frontend

## Структура проекта

```
osi/
├── backend/           # Go REST API
│   ├── cmd/
│   ├── internal/
│   ├── pkg/
│   ├── go.mod
│   └── Dockerfile
├── frontend/          # React приложение
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml # Локальная разработка
├── .github/
│   └── workflows/     # CI/CD
└── deployments/       # Конфигурация деплоя
```

## Workflow разработки

- `main` - продакшн ветка
- `develop` - основная ветка разработки
- `feature/*` - ветки для фич
- `hotfix/*` - срочные исправления

## Локальная разработка

```bash
# Запуск всего стека
docker-compose up

# Backend: http://localhost:8080
# Frontend: http://localhost:3000
```
