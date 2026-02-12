import apiClient from '../client'
import type {
    RequestCodeRequest,
    RequestCodeResponse,
    VerifyCodeRequest,
    AuthResponse,
    PasswordLoginRequest,
    CompleteProfileRequest
} from '../types/auth.types'

export const authService = {
    /**
     * Отправить код на email
     */
    async requestCode(payload: RequestCodeRequest): Promise<RequestCodeResponse> {
        const { data } = await apiClient.post<RequestCodeResponse>('/auth/request-code', payload)
        return data
    },

    /**
     * Проверить код и получить токен
     */
    async verifyCode(payload: VerifyCodeRequest): Promise<AuthResponse> {
        const { data } = await apiClient.post<AuthResponse>('/auth/verify-code', payload)
        return data
    },

    /**
     * Вход по паролю
     */
    async loginWithPassword(payload: PasswordLoginRequest): Promise<AuthResponse> {
        const { data } = await apiClient.post<AuthResponse>('/auth/login-password', payload)
        return data
    },

    /**
     * Завершить профиль после регистрации
     */
    async completeProfile(payload: CompleteProfileRequest): Promise<AuthResponse> {
        const { data } = await apiClient.post<AuthResponse>('/auth/complete-profile', payload)
        return data
    },

    /**
     * Выход
     */
    async logout(): Promise<void> {
        await apiClient.post('/auth/logout')
    }
}
