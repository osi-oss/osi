// src/utils/config.ts
export const API_BASE_URL = 'http://localhost:8080/api';

export const API_TIMEOUT = 30000;

export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  USER_DATA: 'user_data',
  THEME: 'app_theme',
  LANGUAGE: 'app_language',
};

// Проверка платформы
export const IS_WEB = typeof window !== 'undefined' && window.localStorage;