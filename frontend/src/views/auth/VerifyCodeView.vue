<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores'
import Button from '@/components/ui/Button.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const code = ref<string[]>(['', '', '', ''])
const inputRefs = ref<(HTMLInputElement | null)[]>([])
const countdown = ref(20)
const canResend = ref(false)
let verifying = false

onMounted(() => {
  email.value = (route.query.email as string) || ''
  startCountdown()
  nextTick(() => inputRefs.value[0]?.focus())
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

const isCodeComplete = computed(() => code.value.every(c => c !== ''))

watch(code, () => {
  if (isCodeComplete.value && !verifying) {
    handleVerify()
  }
}, { deep: true })

const setRef = (el: any, i: number) => {
  inputRefs.value[i] = el as HTMLInputElement
}

const handleInput = (index: number, event: Event) => {
  const target = event.target as HTMLInputElement
  const val = target.value.replace(/\D/g, '')

  if (val.length > 1) {
    // Handle paste of multiple digits
    const digits = val.slice(0, 4).split('')
    digits.forEach((d, i) => {
      if (index + i < 4) code.value[index + i] = d
    })
    const focusIdx = Math.min(index + digits.length, 3)
    nextTick(() => inputRefs.value[focusIdx]?.focus())
    return
  }

  code.value[index] = val
  if (val && index < 3) {
    nextTick(() => inputRefs.value[index + 1]?.focus())
  }
}

const handleKeydown = (index: number, event: KeyboardEvent) => {
  if (event.key === 'Backspace') {
    if (!code.value[index] && index > 0) {
      code.value[index - 1] = ''
      nextTick(() => inputRefs.value[index - 1]?.focus())
    } else {
      code.value[index] = ''
    }
    return
  }
  // Allow navigation & control keys, block non-digit characters
  const allowed = ['Tab', 'ArrowLeft', 'ArrowRight', 'Delete', 'Enter']
  if (!allowed.includes(event.key) && !/^\d$/.test(event.key)) {
    event.preventDefault()
  }
}

const handlePaste = (event: ClipboardEvent) => {
  event.preventDefault()
  const text = event.clipboardData?.getData('text') || ''
  const digits = text.replace(/\D/g, '').slice(0, 4).split('')
  digits.forEach((d, i) => {
    if (i < 4) code.value[i] = d
  })
  const focusIdx = Math.min(digits.length, 3)
  nextTick(() => inputRefs.value[focusIdx]?.focus())
}

const handleVerify = async () => {
  verifying = true
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
    code.value = ['', '', '', '']
    nextTick(() => inputRefs.value[0]?.focus())
  } finally {
    verifying = false
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
  router.back()
}

const goBack = () => {
  router.back()
}
</script>

<template>
  <div class="min-h-screen bg-gray-50 px-6 py-8 safe-top safe-bottom flex flex-col">
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

    <!-- Code inputs -->
    <div class="flex justify-center gap-3 mb-8" @paste="handlePaste">
      <input
        v-for="(digit, index) in code"
        :key="index"
        :ref="(el) => setRef(el, index)"
        type="text"
        inputmode="numeric"
        maxlength="1"
        :value="digit"
        @input="handleInput(index, $event)"
        @keydown="handleKeydown(index, $event)"
        :class="[
          'w-16 h-20 text-center text-3xl font-semibold rounded-[12px]',
          'border-2 transition bg-white focus:outline-none',
          digit ? 'border-gray-300' : index === code.findIndex(c => c === '') ? 'border-black' : 'border-gray-200'
        ]"
      />
    </div>

    <!-- Loading indicator -->
    <div v-if="authStore.isLoading" class="flex justify-center mb-6">
      <svg class="animate-spin h-6 w-6 text-black" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
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

    <!-- Error message -->
    <p v-if="authStore.error" class="mt-4 text-sm text-red-600 text-center">
      {{ authStore.error }}
    </p>
  </div>
</template>
