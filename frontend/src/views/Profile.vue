<template>
  <n-card title="个人资料" style="max-width: 600px; margin: 0 auto;">
    <n-descriptions label-placement="left" bordered>
      <n-descriptions-item label="用户 ID">
        {{ profile.id }}
      </n-descriptions-item>
      <n-descriptions-item label="邮箱">
        {{ profile.email }}
      </n-descriptions-item>
      <n-descriptions-item label="昵称">
        <n-input v-if="editing" v-model:value="editForm.nickname" />
        <template v-else>{{ profile.nickname || '未设置' }}</template>
      </n-descriptions-item>
      <n-descriptions-item label="注册时间">
        {{ profile.created_at }}
      </n-descriptions-item>
    </n-descriptions>

    <n-space style="margin-top: 16px" justify="center">
      <n-button v-if="!editing" type="primary" @click="startEdit">编辑</n-button>
      <template v-if="editing">
        <n-button type="primary" :loading="saving" @click="save">保存</n-button>
        <n-button @click="cancelEdit">取消</n-button>
      </template>
    </n-space>
  </n-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const message = useMessage()
const editing = ref(false)
const saving = ref(false)

const profile = reactive({
  id: 0,
  email: '',
  nickname: '',
  avatar_url: '',
  created_at: '',
})

const editForm = reactive({
  nickname: '',
})

onMounted(async () => {
  try {
    const data = await authStore.getProfile()
    Object.assign(profile, data)
  } catch {
    message.error('获取个人资料失败')
  }
})

function startEdit() {
  editForm.nickname = profile.nickname
  editing.value = true
}

async function save() {
  saving.value = true
  try {
    const data = await authStore.updateProfile({ nickname: editForm.nickname })
    Object.assign(profile, data)
    editing.value = false
    message.success('保存成功')
  } catch {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

function cancelEdit() {
  editing.value = false
}
</script>
