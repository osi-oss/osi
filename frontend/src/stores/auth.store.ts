import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService } from '@/api/services/auth.service'
import type {
    RequestCodeRequest,
    VerifyCodeRequest,
    PasswordLoginRequest,
    CompleteProfileRequest,
    UserResponse
} from '@/api/types/auth.types'

export const useAuthStore = defineStore('auth', () => {
    // State
    const token = ref<string | null>(localStorage.getItem('auth_token'))
    const expiresAt = ref<number | null>(null)
    const isLoading = ref(false)
    const error = ref<string | null>(null)
    const user = ref<UserResponse | null>(null)

    // Getters
    const isAuthenticated = computed(() => !!token.value)
    const needsProfileCompletion = computed(() => user.value?.status === 'pending_profile')

    // Actions
    const requestCode = async (payload: RequestCodeRequest) => {
        try {
            isLoading.value = true
            error.value = null
            const response = await authService.requestCode(payload)
            return response
        } catch (e: any) {
            error.value = e.response?.data?.error || 'Ошибка отправки кода'
            throw e
        } finally {
            isLoading.value = false
        }
    }

    const verifyCode = async (payload: VerifyCodeRequest) => {
        try {
            isLoading.value = true
            error.value = null
            const response = await authService.verifyCode(payload)

            // Сохраняем токен
            token.value = response.token
            localStorage.setItem('auth_token', response.token)

            // Сохраняем время истечения
            expiresAt.value = Date.now() + response.expires_in * 1000

            // Сохраняем пользователя
            user.value = response.user

            return response
        } catch (e: any) {
            error.value = e.response?.data?.error || 'Неверный код'
            throw e
        } finally {
            isLoading.value = false
        }
    }

    const loginWithPassword = async (payload: PasswordLoginRequest) => {
        try {
            isLoading.value = true
            error.value = null
            const response = await authService.loginWithPassword(payload)

            token.value = response.token
            localStorage.setItem('auth_token', response.token)
            expiresAt.value = Date.now() + response.expires_in * 1000
            user.value = response.user

            return response
        } catch (e: any) {
            error.value = e.response?.data?.error || 'Неверный email или пароль'
            throw e
        } finally {
            isLoading.value = false
        }
    }

    const completeProfile = async (payload: CompleteProfileRequest) => {
        try {
            isLoading.value = true
            error.value = null
            const response = await authService.completeProfile(payload)

            // Обновляем пользователя
            user.value = response.user

            return response
        } catch (e: any) {
            error.value = e.response?.data?.error || 'Ошибка сохранения профиля'
            throw e
        } finally {
            isLoading.value = false
        }
    }

    const logout = async () => {
        try {
            await authService.logout()
        } catch (e) {
            console.error('Logout error:', e)
        } finally {
            // Очищаем состояние
            token.value = null
            expiresAt.value = null
            user.value = null
            localStorage.removeItem('auth_token')
        }
    }

    const checkAuth = () => {
        // Проверяем, есть ли токен и не истёк ли он
        if (token.value && expiresAt.value && Date.now() < expiresAt.value) {
            return true
        }

        // Если токен истёк, очищаем
        if (expiresAt.value && Date.now() >= expiresAt.value) {
            logout()
        }

        return false
    }

    return {
        // State
        token,
        expiresAt,
        isLoading,
        error,
        user,

        // Getters
        isAuthenticated,
        needsProfileCompletion,

        // Actions
        requestCode,
        verifyCode,
        loginWithPassword,
        completeProfile,
        logout,
        checkAuth
    }
})
