<script setup lang="ts">
import { computed } from 'vue'

export interface ToggleItem {
  label: string
  value: string
  disabled?: boolean
}

const props = defineProps<{
  items: ToggleItem[]
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const activeIndex = computed(() =>
  props.items.findIndex(i => i.value === props.modelValue)
)

const select = (item: ToggleItem) => {
  if (item.disabled) return
  emit('update:modelValue', item.value)
}
</script>

<template>
  <div class="relative flex p-1 bg-gray-200 rounded-2xl">

    <!-- SLIDER -->
    <div
      class="absolute top-1 bottom-1 w-[calc(50%-4px)] bg-white rounded-xl shadow-sm transition-transform duration-300 ease-out"
      :style="{
        transform: `translateX(${activeIndex * 100}%)`
      }"
    />

    <!-- BUTTONS -->
    <button
      v-for="(item, index) in items"
      :key="item.value"
      @click="select(item)"
      class="relative flex-1 py-3 text-sm font-medium transition-colors duration-200"
      :class="[
        modelValue === item.value
          ? 'text-black'
          : 'text-gray-500',
        item.disabled && 'opacity-50 cursor-not-allowed'
      ]"
    >
      {{ item.label }}
    </button>

  </div>
</template>
