// src/api/auth.ts
import apiClient from './client';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { STORAGE_KEYS } from '@/utils/config';

export interface UserResponse {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  status: 'pending_profile' | 'active';
}

export interface AuthResponse {
  token: string;
  user: UserResponse;
}

export interface RequestCodeResponse {
  message: string;
  expires_in: number;
  is_new_user: boolean;
}

export const authAPI = {
  // 1. Запрос кода на email
  requestCode: (email: string): Promise<RequestCodeResponse> => {
    return apiClient.post('/auth/request-code', { email });
  },

  // 2. Верификация кода
  verifyCode: (email: string, code: string): Promise<AuthResponse> => {
    return apiClient.post('/auth/verify-code', { email, code });
  },

  // 3. Заполнение профиля
  completeProfile: (data: { first_name: string; last_name: string }): Promise<{ message: string; user: UserResponse }> => {
    return apiClient.post('/auth/complete-profile', data);
  },

  // 4. Получение токена из хранилища
  getToken: async (): Promise<string | null> => {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        // Для веба
        return localStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
      } else {
        // Для мобильных устройств
        const token = await AsyncStorage.getItem(STORAGE_KEYS.AUTH_TOKEN);
        return token;
      }
    } catch (error) {
      console.error('Error getting token:', error);
      return null;
    }
  },

  // 5. Получение пользователя из хранилища
  getUser: async (): Promise<UserResponse | null> => {
    try {
      let userStr: string | null = null;
      
      if (typeof window !== 'undefined' && window.localStorage) {
        // Для веба
        userStr = localStorage.getItem(STORAGE_KEYS.USER_DATA);
      } else {
        // Для мобильных устройств
        userStr = await AsyncStorage.getItem(STORAGE_KEYS.USER_DATA);
      }
      
      return userStr ? JSON.parse(userStr) : null;
    } catch (error) {
      console.error('Error getting user:', error);
      return null;
    }
  },

  // 6. Сохранение токена
  saveToken: async (token: string): Promise<void> => {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        localStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
      } else {
        await AsyncStorage.setItem(STORAGE_KEYS.AUTH_TOKEN, token);
      }
    } catch (error) {
      console.error('Error saving token:', error);
      throw error;
    }
  },

  // 7. Сохранение пользователя
  saveUser: async (user: UserResponse): Promise<void> => {
    try {
      const userStr = JSON.stringify(user);
      
      if (typeof window !== 'undefined' && window.localStorage) {
        localStorage.setItem(STORAGE_KEYS.USER_DATA, userStr);
      } else {
        await AsyncStorage.setItem(STORAGE_KEYS.USER_DATA, userStr);
      }
    } catch (error) {
      console.error('Error saving user:', error);
      throw error;
    }
  },

  // 8. Удаление токена
  removeToken: async (): Promise<void> => {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
      } else {
        await AsyncStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
      }
    } catch (error) {
      console.error('Error removing token:', error);
      throw error;
    }
  },

  // 9. Удаление пользователя
  removeUser: async (): Promise<void> => {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        localStorage.removeItem(STORAGE_KEYS.USER_DATA);
      } else {
        await AsyncStorage.removeItem(STORAGE_KEYS.USER_DATA);
      }
    } catch (error) {
      console.error('Error removing user:', error);
      throw error;
    }
  },

  // 10. Очистка всех данных авторизации
  clearAuth: async (): Promise<void> => {
    try {
      if (typeof window !== 'undefined' && window.localStorage) {
        localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
        localStorage.removeItem(STORAGE_KEYS.USER_DATA);
      } else {
        await AsyncStorage.multiRemove([
          STORAGE_KEYS.AUTH_TOKEN,
          STORAGE_KEYS.USER_DATA,
        ]);
      }
    } catch (error) {
      console.error('Error clearing auth:', error);
      throw error;
    }
  },
};