<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary'
  size?: 'sm' | 'md' | 'lg'
  fullWidth?: boolean
  disabled?: boolean
  loading?: boolean
  type?: 'button' | 'submit' | 'reset'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  fullWidth: false,
  disabled: false,
  loading: false,
  type: 'button'
})

const buttonClasses = computed(() => [
  'inline-flex items-center justify-center font-medium transition-colors',
  'focus:outline-none focus:ring-2 focus:ring-offset-2',
  'disabled:opacity-50 disabled:cursor-not-allowed',
  {
    // Variants
    'bg-black text-white hover:bg-gray-800 focus:ring-black': props.variant === 'primary',
    'bg-gray-400 text-white hover:bg-gray-500 focus:ring-gray-400': props.variant === 'secondary',
    
    // Sizes
    'px-4 py-2.5 text-sm rounded-xl': props.size === 'sm',
    'px-6 py-3.5 text-base rounded-xl': props.size === 'md',
    'px-6 py-4 text-base font-semibold rounded-xl': props.size === 'lg',
    
    // Full width
    'w-full': props.fullWidth
  }
])
</script>

<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    :class="buttonClasses"
  >
    <span v-if="loading" class="mr-2">
      <svg class="animate-spin h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
      </svg>
    </span>
    <slot />
  </button>
</template>
