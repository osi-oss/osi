<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import ToggleGroup from '@/components/ui/ToggleGroup.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const activeTab = ref<'email' | 'phone'>(
  (sessionStorage.getItem('login_tab') as 'email' | 'phone') || 'email'
)
const email = ref(sessionStorage.getItem('login_email') || '')
const phone = ref(sessionStorage.getItem('login_phone') || '')
const isRegister = ref(false)
const validationError = ref('')

const toggleItems = [
  { label: 'Телефон', value: 'phone', disabled: true },
  { label: 'Почта', value: 'email' },
]

watch([activeTab, email, phone], () => {
  validationError.value = ''
  sessionStorage.setItem('login_tab', activeTab.value)
  sessionStorage.setItem('login_email', email.value)
  sessionStorage.setItem('login_phone', phone.value)
})

onMounted(() => {
  isRegister.value = route.query.register === 'true'
})

const title = computed(() => isRegister.value ? 'Регистрация' : 'Вход')

const isEmailValid = (v: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v)

const contactValue = computed(() => {
  if (activeTab.value === 'email') return email.value
  return `+7${phone.value.replace(/\D/g, '')}`
})

const validate = (): boolean => {
  validationError.value = ''
  if (activeTab.value === 'email') {
    if (!email.value) {
      validationError.value = 'Введите email'
      return false
    }
    if (!isEmailValid(email.value)) {
      validationError.value = 'Некорректный формат email'
      return false
    }
  } else {
    const digits = phone.value.replace(/\D/g, '')
    if (!digits) {
      validationError.value = 'Введите номер телефона'
      return false
    }
    if (digits.length !== 10) {
      validationError.value = 'Номер должен содержать 10 цифр'
      return false
    }
  }
  return true
}

const handleContinue = async () => {
  if (!validate()) return

  try {
    await authStore.requestCode({ email: contactValue.value })

    router.push({
      name: 'Verify',
      query: { email: contactValue.value }
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
        class="relative z-10 p-2 -ml-2 hover:bg-gray-100 rounded-lg transition"
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
        autocomplete="email"
        :error="activeTab === 'email' ? validationError : ''"
      />
    </div>

    <!-- Phone input -->
    <div v-else class="mb-6">
      <div
        class="flex items-center w-full border transition-colors rounded-[12px] overflow-hidden"
        :class="validationError && activeTab === 'phone' ? 'border-red-500 bg-red-50' : 'border-gray-200 bg-gray-50'"
      >
        <span class="px-4 py-3.5 text-base text-gray-500 bg-gray-100 border-r border-gray-200 select-none">+7</span>
        <input
          v-model="phone"
          type="tel"
          placeholder="999 123 45 67"
          autocomplete="tel"
          class="flex-1 px-3 py-3.5 text-base bg-transparent focus:outline-none placeholder:text-gray-400"
        />
      </div>
      <p v-if="validationError && activeTab === 'phone'" class="mt-1.5 text-sm text-red-600">
        {{ validationError }}
      </p>
    </div>

    <!-- Continue button -->
    <Button
      variant="primary"
      size="lg"
      full-width
      :loading="authStore.isLoading"
      :disabled="activeTab === 'email' ? !email : !phone"
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
