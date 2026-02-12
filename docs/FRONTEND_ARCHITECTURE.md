# Архитектура Frontend приложения OSI

## 📋 Анализ API

### Основные доменные области

#### 1. **Аутентификация & Авторизация**
- **Passwordless flow**: request-code → verify-code → complete-profile
- **Password flow**: login-password (опционально)
- JWT токены с expires_in
- Управление паролями (set/change/remove)

#### 2. **Профиль пользователя**
- Статусы: `pending_profile`, `active`
- Обязательные поля: first_name, last_name
- Опциональные: middle_name, password
- Флаги: email_verified, has_password

#### 3. **Организации**
- Статусы: `draft`, `pending`, `approved`, `rejected`
- Иерархия: Organization → Location → Department → Position → Employee
- Роли: Founder (владелец), Employee (сотрудник)
- Реквизиты: ИНН, КПП, ОГРН, legal_name, legal_address

#### 4. **Локации (филиалы, офисы)**
- Источники данных: `manual`, `egrul`, `api`
- Флаги: is_active, is_verified
- Привязка к организации

#### 5. **Отделы (departments)**
- Иерархическая структура (parent_id)
- Привязка к локации
- Описание и метаданные

#### 6. **Должности (positions)**
- Привязка к отделу (опционально) или прямо к организации
- Флаг is_admin (административная позиция)
- Описание обязанностей

#### 7. **Сотрудники (employees)**
- Статусы: `active`, `inactive`, `invited`
- Привязка к позиции
- Даты: start_date, end_date, joined_at
- Флаг is_intern (стажёр)

#### 8. **Приглашения (invites)**
- Статусы: `pending`, `accepted`, `declined`
- Workflow: create → accept/decline
- Отправка по email (с автосоздание pending_email user)

#### 9. **Права доступа (permissions)**
- Коды прав: `members.view`, `locations.create`, и т.д.
- Scope types: `organization`, `location`, `department`, `position`
- Grant/Revoke для Position или Employee
- Иерархическая проверка прав

---

## 🏗️ Архитектура Frontend

### Стек технологий

```
Vue 3 (Composition API) + TypeScript
├── Vite (build tool)
├── Vue Router (навигация)
├── Pinia (state management)
├── Axios (HTTP client)
├── Vee-Validate + Zod (валидация форм)
├── VueUse (utilities)
└── vite-plugin-pwa (PWA/Service Worker)

UI Kit:
├── Tailwind CSS (styling)
└── Собственная дизайн-система
```

### Структура проекта

```
frontend/
├── public/
│   ├── manifest.json          # PWA manifest
│   ├── icons/                 # PWA icons
│   └── robots.txt
│
├── src/
│   ├── main.ts               # Entry point
│   ├── App.vue               # Root component
│   │
│   ├── api/                  # API Layer (generated from Swagger)
│   │   ├── client.ts         # Axios instance + interceptors
│   │   ├── types/            # TypeScript types (auto-generated)
│   │   │   ├── auth.types.ts
│   │   │   ├── organization.types.ts
│   │   │   ├── employee.types.ts
│   │   │   └── ...
│   │   ├── services/         # API service methods
│   │   │   ├── auth.service.ts
│   │   │   ├── organization.service.ts
│   │   │   ├── location.service.ts
│   │   │   ├── department.service.ts
│   │   │   ├── position.service.ts
│   │   │   ├── employee.service.ts
│   │   │   ├── invite.service.ts
│   │   │   └── permission.service.ts
│   │   └── index.ts
│   │
│   ├── stores/               # Pinia stores (state management)
│   │   ├── auth.store.ts     # Auth state + JWT
│   │   ├── profile.store.ts  # User profile
│   │   ├── organization.store.ts  # Active org + list
│   │   ├── hierarchy.store.ts     # Org structure cache
│   │   ├── invite.store.ts   # Invites management
│   │   ├── ui.store.ts       # UI state (sidebar, modals)
│   │   └── index.ts
│   │
│   ├── composables/          # Reusable composition functions
│   │   ├── useAuth.ts        # Auth logic
│   │   ├── useForm.ts        # Form utilities
│   │   ├── useAutoSave.ts    # Auto-save logic
│   │   ├── usePermissions.ts # Permission checks
│   │   ├── useOrganization.ts # Org context
│   │   ├── useOffline.ts     # Offline detection
│   │   └── usePWA.ts         # PWA install prompt
│   │
│   ├── components/           # Reusable components
│   │   ├── ui/               # Base UI components
│   │   │   ├── Button.vue
│   │   │   ├── Input.vue
│   │   │   ├── Select.vue
│   │   │   ├── Modal.vue
│   │   │   ├── Toast.vue
│   │   │   ├── Card.vue
│   │   │   ├── Badge.vue
│   │   │   ├── Spinner.vue
│   │   │   └── ...
│   │   ├── forms/            # Form components
│   │   │   ├── FormField.vue
│   │   │   ├── FormStep.vue
│   │   │   ├── FormWizard.vue
│   │   │   └── AutoSaveIndicator.vue
│   │   ├── layout/           # Layout components
│   │   │   ├── AppHeader.vue
│   │   │   ├── AppSidebar.vue
│   │   │   ├── AppFooter.vue
│   │   │   └── Breadcrumbs.vue
│   │   ├── organization/     # Organization-specific
│   │   │   ├── OrganizationCard.vue
│   │   │   ├── OrganizationSelector.vue
│   │   │   └── HierarchyTree.vue
│   │   ├── employee/
│   │   │   ├── EmployeeCard.vue
│   │   │   ├── EmployeeList.vue
│   │   │   └── EmployeeAvatar.vue
│   │   └── invite/
│   │       ├── InviteCard.vue
│   │       └── InviteBadge.vue
│   │
│   ├── views/                # Page components (routes)
│   │   ├── auth/
│   │   │   ├── LoginView.vue          # Step 1: email input
│   │   │   ├── VerifyCodeView.vue     # Step 2: code verify
│   │   │   ├── CompleteProfileView.vue # Step 3: profile
│   │   │   └── LoginPasswordView.vue  # Alternative: password
│   │   │
│   │   ├── profile/
│   │   │   ├── ProfileView.vue        # View/edit profile
│   │   │   └── SecurityView.vue       # Password management
│   │   │
│   │   ├── organizations/
│   │   │   ├── OrganizationListView.vue      # My orgs
│   │   │   ├── OrganizationCreateView.vue    # Multi-step wizard
│   │   │   ├── OrganizationDetailView.vue    # Org dashboard
│   │   │   ├── OrganizationEditView.vue      # Edit org
│   │   │   └── OrganizationHierarchyView.vue # Full structure
│   │   │
│   │   ├── locations/
│   │   │   ├── LocationListView.vue
│   │   │   ├── LocationCreateView.vue
│   │   │   └── LocationDetailView.vue
│   │   │
│   │   ├── departments/
│   │   │   ├── DepartmentListView.vue
│   │   │   ├── DepartmentCreateView.vue
│   │   │   └── DepartmentDetailView.vue
│   │   │
│   │   ├── positions/
│   │   │   ├── PositionListView.vue
│   │   │   ├── PositionCreateView.vue
│   │   │   └── PositionDetailView.vue
│   │   │
│   │   ├── employees/
│   │   │   ├── EmployeeListView.vue
│   │   │   └── EmployeeDetailView.vue
│   │   │
│   │   ├── invites/
│   │   │   ├── MyInvitesView.vue         # Pending invites
│   │   │   ├── OrganizationInvitesView.vue  # Org invites
│   │   │   └── CreateInviteView.vue
│   │   │
│   │   ├── permissions/
│   │   │   ├── PermissionsView.vue       # Manage permissions
│   │   │   └── GrantPermissionView.vue
│   │   │
│   │   └── misc/
│   │       ├── DashboardView.vue
│   │       ├── NotFoundView.vue
│   │       └── OfflineView.vue
│   │
│   ├── router/
│   │   ├── index.ts          # Router config
│   │   ├── guards.ts         # Navigation guards
│   │   └── routes.ts         # Route definitions
│   │
│   ├── utils/
│   │   ├── validation.ts     # Validation schemas (Zod)
│   │   ├── formatters.ts     # Date, number formatters
│   │   ├── storage.ts        # LocalStorage/IndexedDB wrapper
│   │   ├── errors.ts         # Error handling
│   │   └── constants.ts      # App constants
│   │
│   ├── types/                # Global TypeScript types
│   │   ├── common.ts
│   │   └── env.d.ts
│   │
│   ├── assets/               # Static assets
│   │   ├── styles/
│   │   │   ├── main.css      # Tailwind imports
│   │   │   └── variables.css
│   │   └── images/
│   │
│   └── workers/
│       └── service-worker.ts # PWA service worker
│
├── vite.config.ts
├── tsconfig.json
├── tailwind.config.js
├── .env.example
└── package.json
```

---

## 🔄 State Management Strategy (Pinia)

### Auth Store
```typescript
interface AuthState {
  token: string | null
  expiresAt: number | null
  isAuthenticated: boolean
  isLoading: boolean
}

actions:
- login(email: string)
- verifyCode(email: string, code: string)
- loginWithPassword(email: string, password: string)
- logout()
- refreshToken()
- checkAuth()
```

### Profile Store
```typescript
interface ProfileState {
  user: UserResponse | null
  isProfileComplete: boolean
  status: 'pending_profile' | 'active'
}

actions:
- fetchProfile()
- completeProfile(data)
- updateProfile(data)
- setPassword(password)
- changePassword(old, new)
- removePassword()
```

### Organization Store
```typescript
interface OrganizationState {
  activeOrganizationId: number | null
  organizations: OrganizationResponse[]
  myOrganizations: MyOrganizationInfo[]
  currentOrganization: OrganizationResponse | null
  isLoading: boolean
}

actions:
- fetchOrganizations()
- fetchMyOrganizations()
- setActiveOrganization(id)
- createOrganization(data)
- updateOrganization(id, data)
- deleteOrganization(id)
```

### Hierarchy Store (кэш структуры)
```typescript
interface HierarchyState {
  cache: Record<number, OrganizationHierarchyResponse>
  lastFetch: Record<number, number>
  locations: LocationResponse[]
  departments: DepartmentResponse[]
  positions: PositionResponse[]
  employees: EmployeeDetailResponse[]
}

actions:
- fetchHierarchy(orgId)
- invalidateCache(orgId)
- addLocation(orgId, location)
- updateDepartment(orgId, dept)
// etc...
```

### UI Store
```typescript
interface UIState {
  sidebarOpen: boolean
  currentModal: string | null
  toasts: Toast[]
  isOffline: boolean
}
```

---

## 🎨 Поток экранов приложения

### 1. Onboarding & Auth Flow
```
┌─────────────────┐
│   Landing Page  │ (если не залогинен)
│   (optional)    │
└────────┬────────┘
         │
         v
┌─────────────────┐
│   Login View    │ Enter email → Request code
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Verify Code     │ Enter 4-digit code
└────────┬────────┘
         │
         ├─→ [Existing user] ──→ Dashboard
         │
         └─→ [New user]
                  │
                  v
         ┌─────────────────┐
         │ Complete Profile│ First name, Last name, Middle name
         └────────┬────────┘
                  │
                  v
         ┌─────────────────┐
         │   Dashboard     │
         └─────────────────┘
```

### 2. Main Application Structure

```
┌───────────────────────────────────────────────────┐
│              App Header                           │
│  [Logo] [Org Selector]  ...  [Profile] [Logout]  │
├──────────────┬────────────────────────────────────┤
│              │                                    │
│   Sidebar    │         Main Content              │
│              │                                    │
│ • Dashboard  │    [Breadcrumbs]                  │
│ • Orgs       │                                    │
│ • Employees  │    <Router View>                  │
│ • Invites    │                                    │
│ • Settings   │                                    │
│              │                                    │
│              │                                    │
└──────────────┴────────────────────────────────────┘
```

### 3. Organization Management Flow

```
Dashboard
    │
    ├─→ My Organizations List
    │       │
    │       ├─→ Create Organization (Multi-step wizard)
    │       │       ├─ Step 1: Basic Info (name, legal_name)
    │       │       ├─ Step 2: Legal Data (INN, KPP, OGRN)
    │       │       ├─ Step 3: Address
    │       │       └─ Step 4: Share % & Review
    │       │
    │       └─→ Organization Detail
    │               ├─→ Overview (stats, status)
    │               ├─→ Hierarchy View (tree)
    │               ├─→ Locations
    │               │     ├─→ Create Location
    │               │     └─→ Location Detail
    │               │           └─→ Departments
    │               │                 ├─→ Create Department
    │               │                 └─→ Department Detail
    │               ├─→ Positions
    │               │     ├─→ Create Position
    │               │     └─→ Position Detail
    │               ├─→ Employees
    │               │     └─→ Employee Detail
    │               ├─→ Invites
    │               │     ├─→ Create Invite
    │               │     └─→ Manage Invites
    │               ├─→ Permissions
    │               └─→ Settings (Edit/Delete Org)
    │
    ├─→ My Invites (pending invitations)
    │       └─→ Accept/Decline
    │
    └─→ Profile
            ├─→ View/Edit Profile
            └─→ Security (Password management)
```

---

## 🔧 Технические решения

### 1. Валидация форм

**Библиотека**: Vee-Validate + Zod

**Подход**:
```typescript
// utils/validation.ts
import { z } from 'zod'

export const createOrganizationSchema = z.object({
  name: z.string().min(1, 'Название обязательно'),
  legal_name: z.string().optional(),
  inn: z.string().regex(/^\d{10}$|^\d{12}$/, 'ИНН должен быть 10 или 12 цифр').optional(),
  kpp: z.string().regex(/^\d{9}$/, 'КПП должен быть 9 цифр').optional(),
  ogrn: z.string().regex(/^\d{13}$/, 'ОГРН должен быть 13 цифр').optional(),
  legal_address: z.string().optional(),
  share_percent: z.number().min(0).max(100).optional()
})

export type CreateOrganizationFormData = z.infer<typeof createOrganizationSchema>
```

**Использование в компонентах**:
```vue
<script setup lang="ts">
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { createOrganizationSchema } from '@/utils/validation'

const { handleSubmit, errors, defineField } = useForm({
  validationSchema: toTypedSchema(createOrganizationSchema)
})

const [name, nameProps] = defineField('name')
const [inn, innProps] = defineField('inn')

const onSubmit = handleSubmit(async (values) => {
  await organizationStore.createOrganization(values)
})
</script>
```

**Преимущества**:
- Типобезопасность (TypeScript)
- Единая схема для клиента и документации
- Реактивная валидация
- Легко тестировать

---

### 2. Автосохранение данных

**Composable**: `useAutoSave`

```typescript
// composables/useAutoSave.ts
import { ref, watch, type Ref } from 'vue'
import { useDebounceFn } from '@vueuse/core'

interface AutoSaveOptions<T> {
  data: Ref<T>
  saveFn: (data: T) => Promise<void>
  delay?: number // ms
  storageKey?: string // для offline cache
}

export function useAutoSave<T>({ data, saveFn, delay = 2000, storageKey }: AutoSaveOptions<T>) {
  const isSaving = ref(false)
  const lastSaved = ref<Date | null>(null)
  const error = ref<string | null>(null)

  const save = async () => {
    try {
      isSaving.value = true
      error.value = null
      
      // Сохранение в localStorage для offline
      if (storageKey) {
        localStorage.setItem(storageKey, JSON.stringify(data.value))
      }
      
      await saveFn(data.value)
      lastSaved.value = new Date()
    } catch (e) {
      error.value = (e as Error).message
      console.error('Auto-save failed:', e)
    } finally {
      isSaving.value = false
    }
  }

  const debouncedSave = useDebounceFn(save, delay)

  // Watch data changes
  watch(data, () => {
    debouncedSave()
  }, { deep: true })

  return {
    isSaving,
    lastSaved,
    error,
    save // manual trigger
  }
}
```

**Использование**:
```vue
<script setup lang="ts">
const formData = ref({ name: '', inn: '' })

const { isSaving, lastSaved } = useAutoSave({
  data: formData,
  saveFn: async (data) => {
    await organizationService.updateOrganization(orgId, data)
  },
  delay: 3000,
  storageKey: `org-draft-${orgId}`
})
</script>

<template>
  <AutoSaveIndicator :is-saving="isSaving" :last-saved="lastSaved" />
</template>
```

---

### 3. Авторизация & JWT Management

**Стратегия**:

1. **Хранение токена**: Memory + HttpOnly Cookie (рекомендуется) или LocalStorage (для простоты)
2. **Interceptors** для автоматической подстановки токена
3. **Refresh strategy**: проверка expires_in перед каждым запросом
4. **Route guards**: защита маршрутов

**Реализация**:

```typescript
// api/client.ts
import axios, { type AxiosError } from 'axios'
import { useAuthStore } from '@/stores/auth.store'
import router from '@/router'

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  timeout: 30000
})

// Request interceptor: добавляем токен
apiClient.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor: обработка 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const authStore = useAuthStore()
    
    if (error.response?.status === 401) {
      // Token expired or invalid
      authStore.logout()
      router.push('/login')
    }
    
    return Promise.reject(error)
  }
)

export default apiClient
```

**Route Guards**:
```typescript
// router/guards.ts
import { useAuthStore } from '@/stores/auth.store'
import { useProfileStore } from '@/stores/profile.store'

export const authGuard = async (to: RouteLocationNormalized) => {
  const authStore = useAuthStore()
  const profileStore = useProfileStore()
  
  if (!authStore.isAuthenticated) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }
  
  // Check if profile is complete
  if (!profileStore.isProfileComplete && to.name !== 'CompleteProfile') {
    return { name: 'CompleteProfile' }
  }
  
  return true
}

export const guestGuard = (to: RouteLocationNormalized) => {
  const authStore = useAuthStore()
  
  if (authStore.isAuthenticated) {
    return { name: 'Dashboard' }
  }
  
  return true
}
```

**Проверка прав**:
```typescript
// composables/usePermissions.ts
import { computed } from 'vue'
import { useOrganizationStore } from '@/stores/organization.store'
import { useProfileStore } from '@/stores/profile.store'

export function usePermissions() {
  const orgStore = useOrganizationStore()
  const profileStore = useProfileStore()
  
  const isFounder = computed(() => {
    // Check if user is founder of active org
    const org = orgStore.myOrganizations.find(
      o => o.id === orgStore.activeOrganizationId
    )
    return org?.is_founder || false
  })
  
  const can = (permission: string, scopeType?: string, scopeId?: number): boolean => {
    // Founders have all permissions
    if (isFounder.value) return true
    
    // TODO: Check permission grants from API
    // This will require fetching user's permission grants
    return false
  }
  
  return {
    isFounder,
    can
  }
}
```

---

### 4. PWA Configuration

**manifest.json**:
```json
{
  "name": "OSI - Organization Management",
  "short_name": "OSI",
  "description": "Управление организациями, сотрудниками и структурой",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#ffffff",
  "theme_color": "#4F46E5",
  "orientation": "portrait",
  "icons": [
    {
      "src": "/icons/icon-72x72.png",
      "sizes": "72x72",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

**Service Worker (Workbox)**:
```typescript
// vite.config.ts
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.ico', 'robots.txt', 'icons/*'],
      manifest: {
        // ... manifest config
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/api\.example\.com\/api\/.*/i,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-cache',
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 60 * 60 * 24 // 24 hours
              },
              cacheableResponse: {
                statuses: [0, 200]
              }
            }
          }
        ]
      }
    })
  ]
})
```

**Offline detection**:
```typescript
// composables/useOffline.ts
import { ref, onMounted, onUnmounted } from 'vue'

export function useOffline() {
  const isOffline = ref(!navigator.onLine)
  
  const updateOnlineStatus = () => {
    isOffline.value = !navigator.onLine
  }
  
  onMounted(() => {
    window.addEventListener('online', updateOnlineStatus)
    window.addEventListener('offline', updateOnlineStatus)
  })
  
  onUnmounted(() => {
    window.removeEventListener('online', updateOnlineStatus)
    window.removeEventListener('offline', updateOnlineStatus)
  })
  
  return { isOffline }
}
```

---

### 5. Multi-step Form Wizard

**Component**: `FormWizard.vue`

```vue
<script setup lang="ts" generic="T extends Record<string, any>">
import { ref, computed, provide } from 'vue'

interface WizardStep {
  id: string
  title: string
  description?: string
  component: Component
  validate?: (data: Partial<T>) => boolean | Promise<boolean>
}

interface Props {
  steps: WizardStep[]
  initialData?: Partial<T>
  onComplete: (data: T) => Promise<void>
}

const props = defineProps<Props>()

const currentStepIndex = ref(0)
const formData = ref<Partial<T>>(props.initialData || {})

const currentStep = computed(() => props.steps[currentStepIndex.value])
const isFirstStep = computed(() => currentStepIndex.value === 0)
const isLastStep = computed(() => currentStepIndex.value === props.steps.length - 1)
const progress = computed(() => ((currentStepIndex.value + 1) / props.steps.length) * 100)

const goNext = async () => {
  if (currentStep.value.validate) {
    const isValid = await currentStep.value.validate(formData.value)
    if (!isValid) return
  }
  
  if (isLastStep.value) {
    await props.onComplete(formData.value as T)
  } else {
    currentStepIndex.value++
  }
}

const goBack = () => {
  if (!isFirstStep.value) {
    currentStepIndex.value--
  }
}

provide('wizardData', formData)
provide('updateWizardData', (key: keyof T, value: any) => {
  formData.value[key] = value
})
</script>

<template>
  <div class="wizard">
    <!-- Progress bar -->
    <div class="wizard-progress">
      <div class="progress-bar" :style="{ width: `${progress}%` }" />
    </div>
    
    <!-- Steps indicator -->
    <div class="wizard-steps">
      <div
        v-for="(step, index) in steps"
        :key="step.id"
        :class="{ active: index === currentStepIndex, completed: index < currentStepIndex }"
      >
        {{ step.title }}
      </div>
    </div>
    
    <!-- Current step content -->
    <div class="wizard-content">
      <component :is="currentStep.component" />
    </div>
    
    <!-- Navigation -->
    <div class="wizard-actions">
      <button v-if="!isFirstStep" @click="goBack">Назад</button>
      <button @click="goNext">
        {{ isLastStep ? 'Завершить' : 'Далее' }}
      </button>
    </div>
  </div>
</template>
```

**Использование**:
```vue
<script setup lang="ts">
import FormWizard from '@/components/forms/FormWizard.vue'
import Step1BasicInfo from './steps/Step1BasicInfo.vue'
import Step2LegalData from './steps/Step2LegalData.vue'
import Step3Address from './steps/Step3Address.vue'

const steps = [
  {
    id: 'basic',
    title: 'Основная информация',
    component: Step1BasicInfo,
    validate: (data) => !!data.name
  },
  {
    id: 'legal',
    title: 'Реквизиты',
    component: Step2LegalData
  },
  {
    id: 'address',
    title: 'Адрес',
    component: Step3Address
  }
]

const handleComplete = async (data) => {
  await organizationService.createOrganization(data)
  router.push('/organizations')
}
</script>

<template>
  <FormWizard :steps="steps" :on-complete="handleComplete" />
</template>
```

---

## 📦 Работа с API (Type-safe)

### Генерация типов из Swagger

Используем **openapi-typescript** для автогенерации типов:

```bash
npm install -D openapi-typescript
```

**package.json**:
```json
{
  "scripts": {
    "generate:types": "openapi-typescript ../backend/docs/swagger.yaml -o src/api/types/api.d.ts"
  }
}
```

### API Service Pattern

```typescript
// api/services/organization.service.ts
import apiClient from '../client'
import type {
  OrganizationResponse,
  CreateOrganizationRequest,
  UpdateOrganizationRequest,
  OrganizationsListResponse
} from '../types/api'

export const organizationService = {
  async getOrganizations(): Promise<OrganizationResponse[]> {
    const { data } = await apiClient.get<OrganizationsListResponse>('/organizations')
    return data.organizations || []
  },
  
  async getOrganization(id: number): Promise<OrganizationResponse> {
    const { data } = await apiClient.get<OrganizationResponse>(`/organizations/${id}`)
    return data
  },
  
  async createOrganization(payload: CreateOrganizationRequest): Promise<OrganizationResponse> {
    const { data } = await apiClient.post<OrganizationResponse>('/organizations', payload)
    return data
  },
  
  async updateOrganization(
    id: number,
    payload: UpdateOrganizationRequest
  ): Promise<OrganizationResponse> {
    const { data } = await apiClient.put<OrganizationResponse>(`/organizations/${id}`, payload)
    return data
  },
  
  async deleteOrganization(id: number): Promise<void> {
    await apiClient.delete(`/organizations/${id}`)
  }
}
```

---

## 🚀 План реализации (поэтапно)

### Phase 1: Foundation (Week 1)
- [ ] Setup Vite + Vue 3 + TypeScript
- [ ] Configure Tailwind CSS
- [ ] Install dependencies (Router, Pinia, VeeValidate, etc.)
- [ ] Generate API types from Swagger
- [ ] Setup API client with interceptors
- [ ] Create base UI components (Button, Input, Card, etc.)
- [ ] Configure PWA (manifest, icons)

### Phase 2: Authentication (Week 1-2)
- [ ] Auth Store (Pinia)
- [ ] Login flow (email → code → verify)
- [ ] Password login (alternative)
- [ ] Complete profile flow
- [ ] Route guards
- [ ] Logout functionality

### Phase 3: Profile & Navigation (Week 2)
- [ ] Profile Store
- [ ] App layout (Header, Sidebar, Footer)
- [ ] Profile view/edit
- [ ] Password management
- [ ] Breadcrumbs & navigation

### Phase 4: Organizations (Week 3)
- [ ] Organization Store
- [ ] Organization list view
- [ ] Create organization (multi-step wizard)
- [ ] Organization detail view
- [ ] Edit/delete organization
- [ ] Organization selector (switch context)

### Phase 5: Structure (Locations, Departments, Positions) (Week 4)
- [ ] Locations CRUD
- [ ] Departments CRUD (hierarchy support)
- [ ] Positions CRUD
- [ ] Hierarchy tree component
- [ ] Full hierarchy view

### Phase 6: Employees & Invites (Week 5)
- [ ] Invite Store
- [ ] My invites view
- [ ] Organization invites management
- [ ] Create invite flow
- [ ] Accept/decline invite
- [ ] Employee list & detail views

### Phase 7: Permissions (Week 6)
- [ ] Permission service
- [ ] usePermissions composable
- [ ] Grant/revoke permission UI
- [ ] Permission guards in components
- [ ] Position/Employee permission views

### Phase 8: Polish & PWA (Week 7)
- [ ] Auto-save implementation
- [ ] Offline detection & handling
- [ ] Service Worker optimization
- [ ] Install prompt
- [ ] Loading states & skeletons
- [ ] Error boundaries
- [ ] Toast notifications

### Phase 9: Testing & Optimization (Week 8)
- [ ] Unit tests (Vitest)
- [ ] E2E tests (Playwright)
- [ ] Performance optimization
- [ ] Bundle size optimization
- [ ] Accessibility audit
- [ ] Documentation

---

## 📱 Responsive Design Strategy

- **Mobile-first**: Design for mobile, enhance for desktop
- **Breakpoints**:
  - `sm`: 640px (mobile landscape)
  - `md`: 768px (tablet)
  - `lg`: 1024px (small desktop)
  - `xl`: 1280px (desktop)
  - `2xl`: 1536px (large desktop)

**Layout adaptations**:
- Mobile: Collapsed sidebar (hamburger menu), stacked forms
- Tablet: Side drawer, 2-column forms
- Desktop: Fixed sidebar, multi-column layouts

---

## 🎯 Key Features Summary

✅ **PWA**: Установка как приложение, offline support, push notifications  
✅ **Type-safe API**: Автогенерация типов из Swagger  
✅ **Form Validation**: Zod + VeeValidate  
✅ **Auto-save**: Debounced auto-save с offline cache  
✅ **Multi-step Wizards**: Reusable wizard component  
✅ **Permission System**: Hierarchical permission checks  
✅ **Responsive**: Mobile-first, adaptive layouts  
✅ **State Management**: Pinia stores для всех доменов  
✅ **Error Handling**: Graceful error handling & recovery  
✅ **Performance**: Code splitting, lazy loading  

---

## 🔜 Следующие шаги

1. **Согласовать архитектуру** с вашими ожиданиями
2. **Получить дизайн-макеты** для UI components
3. **Начать реализацию Phase 1** (Foundation)
4. **Настроить CI/CD** для автоматического деплоя

**Готов начать реализацию!** 🚀

Пришлите скрины дизайна, и я адаптирую компоненты под ваш фирменный стиль.
