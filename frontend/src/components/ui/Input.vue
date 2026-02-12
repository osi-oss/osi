<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  modelValue?: string | number
  type?: 'text' | 'email' | 'password' | 'tel' | 'number'
  placeholder?: string
  disabled?: boolean
  error?: string
  icon?: string
  autocomplete?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'text',
  disabled: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string | number]
}>()

const inputClasses = computed(() => [
  'w-full px-4 py-3.5 text-base rounded-button border transition-colors',
  'focus:outline-none focus:ring-1 focus:ring-black/20',
  'placeholder:text-gray-400',
  {
    'border-gray-200 bg-gray-50': !props.error,
    'border-red-500 bg-red-50': props.error,
    'bg-gray-100 cursor-not-allowed': props.disabled,
    'pl-12': props.icon
  }
])

const handleInput = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}
</script>

<template>
  <div class="relative">
    <div v-if="icon" class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">
      <span class="text-xl">{{ icon }}</span>
    </div>
    
    <input
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :class="inputClasses"
      :autocomplete="autocomplete"
      @input="handleInput"
    />
    
    <p v-if="error" class="mt-1.5 text-sm text-red-600">
      {{ error }}
    </p>
  </div>
</template>
