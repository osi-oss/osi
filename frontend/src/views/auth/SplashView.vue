<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'

const router = useRouter()

const goToLogin = () => {
  router.push({ name: 'Login' })
}

const goToRegister = () => {
  router.push({ name: 'Login', query: { register: 'true' } })
}
</script>

<template>
  <div class="min-h-screen bg-white flex flex-col items-center justify-center px-6 safe-top safe-bottom">
    <!-- Logo and branding -->
    <div class="flex-1 flex flex-col items-center justify-center max-w-sm w-full">
      <h1 class="text-7xl font-bold text-black mb-6">ОСИ</h1>
      <div class="text-center space-y-1">
        <p class="text-base text-black font-medium">оптимизация</p>
        <p class="text-base text-black font-medium">систематизация</p>
        <p class="text-base text-black font-medium">инновации</p>
      </div>
    </div>

    <!-- Actions -->
    <div class="w-full max-w-sm space-y-3 pb-8">
      <Button
        variant="primary"
        size="lg"
        full-width
        @click="goToLogin"
      >
        Вход
      </Button>
      
      <Button
        variant="secondary"
        size="lg"
        full-width
        @click="goToRegister"
      >
        Регистрация
      </Button>
    </div>
  </div>
</template>
