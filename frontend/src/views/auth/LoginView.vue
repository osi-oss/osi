<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import ToggleGroup from '@/components/ui/ToggleGroup.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const activeTab = ref<'email' | 'phone'>('email')
const email = ref('')
const isRegister = ref(false)

const toggleItems = [
  { label: 'Телефон', value: 'phone', disabled: false },
  { label: 'Почта', value: 'email' },
]

onMounted(() => {
  isRegister.value = route.query.register === 'true'
})

const title = computed(() => isRegister.value ? 'Регистрация' : 'Вход')

const handleContinue = async () => {
  try {
    if (!email.value) {
      return
    }

    // Отправляем код на email
    await authStore.requestCode({ email: email.value })
    
    // Переходим на экран ввода кода
    router.push({
      name: 'Verify',
      query: { email: email.value }
    })
  } catch (error) {
    console.error('Request code error:', error)
  }
}

const handlePasswordLogin = () => {
  // TODO: Implement password login modal
  console.log('Password login')
}

const goBack = () => {
  router.push({ name: 'Splash' })
}
</script>

<template>
  <div class="min-h-screen bg-white px-6 py-8 safe-top safe-bottom">
    <!-- Header -->
    <div class="flex items-center mb-12">
      <button
        @click="goBack"
        class="p-2 -ml-2 hover:bg-gray-100 rounded-lg transition"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
      </button>
      <h1 class="text-2xl font-bold text-center flex-1 -ml-10">ОСИ</h1>
    </div>

    <!-- Title -->
    <h2 class="text-2xl font-bold text-center mb-8">{{ title }}</h2>

    <!-- Tabs -->

    <ToggleGroup
      v-model="activeTab"
      :items="toggleItems"
      class="mb-6"
    />
    <!-- <div class="flex gap-2 mb-6">
      <button
        :class="[
          'flex-1 py-3.5 px-6 rounded-button text-base font-medium transition',
          activeTab === 'phone' 
            ? 'bg-gray-100 text-gray-500' 
            : 'bg-white text-black border border-gray-300'
        ]"
        @click="activeTab = 'phone'"
      >
        Телефон
      </button>
      <button
        :class="[
          'flex-1 py-3.5 px-6 rounded-button text-base font-medium transition',
          activeTab === 'email' 
            ? 'bg-white text-black border border-gray-900' 
            : 'bg-gray-100 text-gray-500'
        ]"
        @click="activeTab = 'email'"
      >
        Почта
      </button>
    </div> -->

    <!-- Email input -->
    <div v-if="activeTab === 'email'" class="mb-6">
      <Input
        v-model="email"
        type="email"
        placeholder="abc@mail.com"
        icon="📧"
        autocomplete="email"
      />
    </div>

    <!-- Phone input -->
    <div v-else class="mb-6">
      <Input
        v-model="email"
        type="tel"
        placeholder="+7 (999) 123-45-67"
        icon="📱"
        autocomplete="tel"
      />
    </div>

    <!-- Continue button -->
    <Button
      variant="primary"
      size="lg"
      full-width
      :loading="authStore.isLoading"
      :disabled="!email"
      class="mb-4"
      @click="handleContinue"
    >
      Дальше
    </Button>

    <!-- Password login -->
    <Button
      v-if="!isRegister"
      variant="secondary"
      size="lg"
      full-width
      @click="handlePasswordLogin"
    >
      Войти с паролем
    </Button>

    <!-- Error message -->
    <p v-if="authStore.error" class="mt-4 text-sm text-red-600 text-center">
      {{ authStore.error }}
    </p>
  </div>
</template>
