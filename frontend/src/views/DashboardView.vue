<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import Card from '@/components/ui/Card.vue'

const router = useRouter()
const authStore = useAuthStore()

const user = computed(() => authStore.user)

const initials = computed(() => {
  if (!user.value) return '?'
  const first = user.value.first_name?.[0] || ''
  const last = user.value.last_name?.[0] || ''
  return (first + last).toUpperCase()
})

const fullName = computed(() => {
  if (!user.value) return 'Пользователь'
  return `${user.value.first_name || ''} ${user.value.last_name || ''}`.trim()
})

const handleLogout = async () => {
  await authStore.logout()
  router.push({ name: 'Splash' })
}

const handleCreateOrganization = () => {
  // TODO: Navigate to create organization
  console.log('Create organization')
}
</script>

<template>
  <div class="min-h-screen bg-white safe-top safe-bottom">
    <!-- Header -->
    <div class="bg-black text-white px-6 py-4 flex items-center justify-between">
      <h1 class="text-xl font-bold">ОСИ</h1>
      <button
        @click="handleLogout"
        class="text-sm hover:underline"
      >
        Выйти
      </button>
    </div>

    <div class="px-6 py-6 space-y-6">
      <!-- Profile Card -->
      <Card>
        <div class="flex items-center gap-4">
          <!-- Avatar -->
          <div class="w-16 h-16 rounded-full bg-gray-light flex items-center justify-center text-2xl font-bold text-gray-600">
            {{ initials }}
          </div>

          <!-- User info -->
          <div class="flex-1">
            <h2 class="text-lg font-semibold">{{ fullName }}</h2>
            <p class="text-sm text-gray-500">{{ user?.email }}</p>
          </div>

          <!-- Arrow -->
          <button class="p-2">
            <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>
      </Card>

      <!-- Organizations Section -->
      <div>
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-xl font-bold">Все организации</h3>
          <button
            @click="handleCreateOrganization"
            class="text-sm px-4 py-2 border border-black rounded-button hover:bg-gray-50 transition font-medium"
          >
            Создать организацию
          </button>
        </div>

        <!-- Organization Card (example) -->
        <Card class="mb-4">
          <div class="flex items-center gap-4">
            <!-- Avatar -->
            <div class="w-16 h-16 rounded-full bg-gray-light flex items-center justify-center text-2xl font-bold text-gray-600">
              ОВ
            </div>

            <!-- Org info -->
            <div class="flex-1">
              <h4 class="text-base font-semibold">ООО "Оверсииз"</h4>
              <p class="text-sm text-gray-500">Админ</p>
              <p class="text-sm text-gray-500">Официант</p>
            </div>

            <!-- Arrow -->
            <button class="p-2">
              <svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        </Card>
      </div>

      <!-- Invites Section -->
      <div>
        <h3 class="text-xl font-bold mb-4">Мои приглашения</h3>
        
        <Card>
          <p class="text-gray-400 text-center py-8">
            У вас пока нет приглашений
          </p>
        </Card>
      </div>
    </div>
  </div>
</template>
