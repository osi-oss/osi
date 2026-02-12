<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'

const router = useRouter()
const authStore = useAuthStore()

const firstName = ref('')
const lastName = ref('')
const middleName = ref('')

const handleSubmit = async () => {
  try {
    if (!firstName.value || !lastName.value) {
      return
    }

    await authStore.completeProfile({
      first_name: firstName.value,
      last_name: lastName.value,
      middle_name: middleName.value || undefined
    })

    router.push('/dashboard')
  } catch (error) {
    console.error('Complete profile error:', error)
  }
}

const goBack = () => {
  authStore.logout()
  router.push({ name: 'Login' })
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
    <h2 class="text-2xl font-bold text-center mb-8">Давайте познакомимся</h2>

    <!-- Form -->
    <div class="space-y-4 mb-8">
      <Input
        v-model="firstName"
        type="text"
        placeholder="Имя"
        autocomplete="given-name"
      />

      <Input
        v-model="lastName"
        type="text"
        placeholder="Фамилия"
        autocomplete="family-name"
      />

      <Input
        v-model="middleName"
        type="text"
        placeholder="Отчество"
        autocomplete="additional-name"
      />
    </div>

    <!-- Submit button -->
    <Button
      variant="primary"
      size="lg"
      full-width
      :loading="authStore.isLoading"
      :disabled="!firstName || !lastName"
      @click="handleSubmit"
    >
      Дальше
    </Button>

    <!-- Error message -->
    <p v-if="authStore.error" class="mt-4 text-sm text-red-600 text-center">
      {{ authStore.error }}
    </p>
  </div>
</template>
