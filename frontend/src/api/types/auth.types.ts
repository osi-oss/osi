// Auth Types (базовые, позже автогенерируем из Swagger)

export interface RequestCodeRequest {
    email: string
}

export interface RequestCodeResponse {
    message: string
    expires_in: number
    is_new_user: boolean
}

export interface VerifyCodeRequest {
    email: string
    code: string
}

export interface AuthResponse {
    token: string
    expires_in: number
    next_step?: 'complete_profile' | ''
    user: UserResponse
}

export interface PasswordLoginRequest {
    email: string
    password: string
}

export interface CompleteProfileRequest {
    first_name: string
    last_name: string
    middle_name?: string
}

export interface UserResponse {
    id: number
    email: string
    first_name?: string
    last_name?: string
    middle_name?: string
    email_verified: boolean
    has_password: boolean
    status: 'pending_profile' | 'active'
    created_at: string
}

export interface ErrorResponse {
    error: string
}
