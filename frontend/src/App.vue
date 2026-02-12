<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores'

const authStore = useAuthStore()

onMounted(() => {
  // Check auth on app load
  authStore.checkAuth()
})
</script>

<template>
  <RouterView />
</template>

<style>
/* Global styles in assets/styles/main.css */
</style>
