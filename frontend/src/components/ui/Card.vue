<script setup lang="ts">
interface Props {
  padding?: 'none' | 'sm' | 'md' | 'lg'
  shadow?: boolean
  border?: boolean
}

withDefaults(defineProps<Props>(), {
  padding: 'md',
  shadow: true,
  border: true
})
</script>

<template>
  <div 
    :class="[
      'bg-white rounded-card',
      {
        'p-4': padding === 'sm',
        'p-6': padding === 'md',
        'p-8': padding === 'lg',
        'shadow-sm': shadow,
        'border border-gray-border': border
      }
    ]"
  >
    <slot />
  </div>
</template>
