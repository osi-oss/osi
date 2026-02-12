<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores'
import Button from '@/components/ui/Button.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const code = ref<string[]>(['', '', '', ''])
const countdown = ref(20)
const canResend = ref(false)

onMounted(() => {
  email.value = (route.query.email as string) || ''
  startCountdown()
})

const startCountdown = () => {
  canResend.value = false
  countdown.value = 20
  
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
      canResend.value = true
    }
  }, 1000)
}

const isCodeComplete = computed(() => {
  return code.value.every(c => c !== '')
})

watch(code, () => {
  if (isCodeComplete.value) {
    handleVerify()
  }
}, { deep: true })

const handleDigitInput = (digit: string) => {
  const emptyIndex = code.value.findIndex(c => c === '')
  if (emptyIndex !== -1) {
    code.value[emptyIndex] = digit
  }
}

const handleBackspace = () => {
  const lastFilledIndex = code.value.map((c, i) => c !== '' ? i : -1).filter(i => i !== -1).pop()
  if (lastFilledIndex !== undefined) {
    code.value[lastFilledIndex] = ''
  }
}

const handleVerify = async () => {
  try {
    const codeString = code.value.join('')
    const response = await authStore.verifyCode({
      email: email.value,
      code: codeString
    })

    if (response.next_step === 'complete_profile') {
      router.push('/complete-profile')
    } else {
      router.push('/dashboard')
    }
  } catch (error) {
    console.error('Verify error:', error)
    // Очищаем код при ошибке
    code.value = ['', '', '', '']
  }
}

const handleResend = async () => {
  if (!canResend.value) return
  
  try {
    await authStore.requestCode({ email: email.value })
    startCountdown()
  } catch (error) {
    console.error('Resend error:', error)
  }
}

const handleChangeEmail = () => {
  router.push({ name: 'Login' })
}

const goBack = () => {
  router.push({ name: 'Login' })
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 px-6 py-8 safe-top safe-bottom flex flex-col">
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

    <!-- Title and instructions -->
    <div class="text-center mb-8">
      <h2 class="text-2xl font-bold mb-4">Введите код из письма</h2>
      <p class="text-base text-gray-600">
        Отправили на <span class="font-semibold text-black">{{ email }}</span>
      </p>
      <p class="text-sm text-gray-400 mt-1">
        Если во входящих нет, проверьте спам
      </p>
    </div>

    <!-- Code display -->
    <div class="flex justify-center gap-3 mb-8">
      <div
        v-for="(digit, index) in code"
        :key="index"
        :class="[
          'w-16 h-20 flex items-center justify-center text-3xl font-semibold rounded-button',
          'border-2 transition bg-white',
          digit ? 'border-gray-300' : index === code.findIndex(c => c === '') ? 'border-blue-500' : 'border-gray-300'
        ]"
      >
        {{ digit || '—' }}
      </div>
    </div>

    <!-- Resend timer -->
    <div class="text-center mb-6">
      <p v-if="!canResend" class="text-sm text-gray-500">
        отправить код снова через <span class="text-primary">0:{{ countdown.toString().padStart(2, '0') }}</span>
      </p>
      <button
        v-else
        @click="handleResend"
        class="text-sm text-primary font-medium hover:underline"
      >
        Отправить код снова
      </button>
    </div>

    <!-- Change email button -->
    <Button
      variant="secondary"
      size="lg"
      full-width
      class="mb-8"
      @click="handleChangeEmail"
    >
      Изменить почту
    </Button>

    <!-- Spacer -->
    <div class="flex-1"></div>

    <!-- Numeric keypad -->
    <div class="grid grid-cols-3 gap-3 max-w-sm mx-auto w-full">
      <button
        v-for="digit in ['1', '2', '3', '4', '5', '6', '7', '8', '9']"
        :key="digit"
        @click="handleDigitInput(digit)"
        class="aspect-square bg-white hover:bg-gray-50 active:bg-gray-100 rounded-button flex items-center justify-center text-2xl font-semibold transition shadow-sm border border-gray-200"
      >
        {{ digit }}
      </button>
      
      <!-- Empty cell -->
      <div></div>
      
      <!-- 0 -->
      <button
        @click="handleDigitInput('0')"
        class="aspect-square bg-white hover:bg-gray-50 active:bg-gray-100 rounded-button flex items-center justify-center text-2xl font-semibold transition shadow-sm border border-gray-200"
      >
        0
      </button>
      
      <!-- Backspace -->
      <button
        @click="handleBackspace"
        class="aspect-square bg-white hover:bg-gray-50 active:bg-gray-100 rounded-button flex items-center justify-center transition shadow-sm border border-gray-200"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2M3 12l6.414 6.414a2 2 0 001.414.586H19a2 2 0 002-2V7a2 2 0 00-2-2h-8.172a2 2 0 00-1.414.586L3 12z" />
        </svg>
      </button>
    </div>

    <!-- Error message -->
    <p v-if="authStore.error" class="mt-4 text-sm text-red-600 text-center">
      {{ authStore.error }}
    </p>
  </div>
</template>
