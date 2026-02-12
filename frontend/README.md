# ОСИ Frontend

PWA приложение для управления организациями, построенное на Vue 3 + TypeScript.

## 🚀 Стек технологий

- **Vue 3** (Composition API)
- **TypeScript** - строгая типизация
- **Vite** - сборщик
- **Vue Router** - навигация
- **Pinia** - state management
- **Tailwind CSS** - стилизация
- **Axios** - HTTP client
- **Vee-Validate + Zod** - валидация форм
- **vite-plugin-pwa** - PWA поддержка

## 📦 Установка

```bash
npm install
```

## 🛠️ Разработка

```bash
npm run dev
```

Приложение будет доступно по адресу: http://localhost:5173

## 🏗️ Сборка

```bash
npm run build
```

## 📁 Структура проекта

```
src/
├── api/              # API клиент и типы
├── stores/           # Pinia stores
├── components/       # Переиспользуемые компоненты
│   ├── ui/          # Базовые UI компоненты
│   ├── forms/       # Компоненты форм
│   └── layout/      # Layout компоненты
├── views/            # Страницы приложения
├── router/           # Конфигурация роутера
├── utils/            # Утилиты
├── assets/           # Статические ресурсы
└── types/            # TypeScript типы
```

## 🎨 Дизайн система

### Цвета
- **Primary**: `#000000` (черный)
- **Secondary**: `#9CA3AF` (серый)
- **Background**: `#FFFFFF` (белый)

### Радиусы
- **Button**: `12px`
- **Input**: `12px`
- **Card**: `16px`

## 🔐 Аутентификация

Приложение поддерживает два способа входа:
1. **Passwordless** (основной): email → код → вход
2. **Password** (опционально): email + пароль

## 🌐 API

Backend API: http://localhost:8080/api
Swagger: http://localhost:8080/swagger/index.html
