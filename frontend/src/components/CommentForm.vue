<template>
  <div class="comment-form">
    <textarea
      v-model="content"
      :placeholder="placeholder"
      :disabled="disabled"
      rows="3"
      class="form-textarea"
    />
    <div class="form-actions">
      <button
        :disabled="!content.trim() || disabled"
        class="btn-submit"
        @click="handleSubmit"
      >
        {{ submitting ? '提交中...' : '发表评论' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(
  defineProps<{
    placeholder?: string
    submitting?: boolean
  }>(),
  {
    placeholder: '写下你的评论...',
    submitting: false,
  },
)

const emit = defineEmits<{
  submit: [content: string]
}>()

const content = ref('')
const disabled = ref(false)

async function handleSubmit() {
  if (!content.value.trim()) return
  disabled.value = true
  emit('submit', content.value.trim())
  content.value = ''
  disabled.value = false
}
</script>

<style scoped>
.comment-form {
  margin-bottom: 16px;
}
.form-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  font-size: 14px;
  resize: vertical;
  box-sizing: border-box;
  transition: border-color 0.2s;
}
.form-textarea:focus {
  outline: none;
  border-color: #1677ff;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}
.btn-submit {
  padding: 8px 20px;
  background: #1677ff;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s;
}
.btn-submit:hover {
  background: #4096ff;
}
.btn-submit:disabled {
  background: #d9d9d9;
  cursor: not-allowed;
}
</style>
