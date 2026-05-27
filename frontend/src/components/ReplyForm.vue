<template>
  <div class="reply-form">
    <textarea
      v-model="content"
      :placeholder="`回复 @${username}`"
      rows="2"
      class="form-textarea"
    />
    <div class="form-actions">
      <button class="btn-cancel" @click="$emit('cancel')">取消</button>
      <button
        :disabled="!content.trim()"
        class="btn-submit"
        @click="handleSubmit"
      >
        回复
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  username: string
}>()

const emit = defineEmits<{
  submit: [content: string]
  cancel: []
}>()

const content = ref('')

function handleSubmit() {
  if (!content.value.trim()) return
  emit('submit', content.value.trim())
  content.value = ''
}
</script>

<style scoped>
.reply-form {
  margin-top: 12px;
  padding-left: 48px;
}
.form-textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  font-size: 14px;
  resize: vertical;
  box-sizing: border-box;
}
.form-textarea:focus {
  outline: none;
  border-color: #1677ff;
}
.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}
.btn-submit {
  padding: 6px 16px;
  background: #1677ff;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}
.btn-submit:disabled {
  background: #d9d9d9;
  cursor: not-allowed;
}
.btn-cancel {
  padding: 6px 16px;
  background: #fff;
  color: #666;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
}
</style>
