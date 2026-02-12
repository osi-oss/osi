import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores'

const routes: RouteRecordRaw[] = [
    {
        path: '/',
        name: 'Splash',
        component: () => import('@/views/auth/SplashView.vue'),
        meta: { requiresGuest: true }
    },
    {
        path: '/login',
        name: 'Login',
        component: () => import('@/views/auth/LoginView.vue'),
        meta: { requiresGuest: true }
    },
    {
        path: '/verify',
        name: 'Verify',
        component: () => import('@/views/auth/VerifyCodeView.vue'),
        meta: { requiresGuest: true }
    },
    {
        path: '/complete-profile',
        name: 'CompleteProfile',
        component: () => import('@/views/auth/CompleteProfileView.vue'),
        meta: { requiresAuth: true }
    },
    {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('@/views/DashboardView.vue'),
        meta: { requiresAuth: true }
    }
]

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes
})

// Navigation guards
router.beforeEach((to, _from, next) => {
    const authStore = useAuthStore()

    // Проверяем авторизацию
    if (to.meta.requiresAuth) {
        if (!authStore.isAuthenticated) {
            return next({ name: 'Login', query: { redirect: to.fullPath } })
        }

        // Проверяем, заполнен ли профиль
        if (authStore.needsProfileCompletion && to.name !== 'CompleteProfile') {
            return next({ name: 'CompleteProfile' })
        }
    }

    // Если гость пытается зайти на гостевые страницы
    if (to.meta.requiresGuest && authStore.isAuthenticated) {
        if (authStore.needsProfileCompletion) {
            return next({ name: 'CompleteProfile' })
        }
        return next({ name: 'Dashboard' })
    }

    next()
})

export default router
