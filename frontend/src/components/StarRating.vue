<template>
  <n-rate
    :value="modelValue"
    :count="5"
    size="large"
    @update:value="(v: number) => $emit('update:modelValue', v)"
  />
</template>

<script setup lang="ts">
defineProps<{
  modelValue: number
}>()

defineEmits<{
  'update:modelValue': [value: number]
}>()
</script>
