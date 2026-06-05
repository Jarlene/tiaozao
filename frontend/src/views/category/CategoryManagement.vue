<template>
  <n-space vertical>
    <n-h2>分类管理</n-h2>

    <n-space>
      <n-button type="primary" @click="openCreateModal(null)">
        创建根分类
      </n-button>
    </n-space>

    <n-spin :show="loading">
      <n-empty v-if="!loading && categories.length === 0" description="暂无分类" style="padding: 60px 0" />

      <n-space vertical v-else>
        <n-card
          v-for="item in categories"
          :key="item.id"
          size="small"
          :title="item.name"
          style="margin-bottom: 8px"
        >
          <template #header-extra>
            <n-space>
              <n-button size="tiny" @click="openCreateModal(item.id)">添加子分类</n-button>
              <n-button size="tiny" @click="openEditModal(item)">编辑</n-button>
              <n-popconfirm @positive-click="handleDelete(item.id)">
                <template #trigger>
                  <n-button size="tiny" type="error">删除</n-button>
                </template>
                确定要删除「{{ item.name }}」吗？如果分类下有子分类或商品，将无法删除。
              </n-popconfirm>
            </n-space>
          </template>

          <!-- 子分类 -->
          <n-space vertical v-if="item.children && item.children.length > 0">
            <n-card
              v-for="child in item.children"
              :key="child.id"
              size="small"
              :title="child.name"
              embedded
              style="margin-bottom: 4px"
            >
              <template #header-extra>
                <n-space>
                  <n-button size="tiny" @click="openEditModal(child)">编辑</n-button>
                  <n-popconfirm @positive-click="handleDelete(child.id)">
                    <template #trigger>
                      <n-button size="tiny" type="error">删除</n-button>
                    </template>
                    确定要删除「{{ child.name }}」吗？如果分类下有子分类或商品，将无法删除。
                  </n-popconfirm>
                </n-space>
              </template>
            </n-card>
          </n-space>
        </n-card>
      </n-space>
    </n-spin>

    <!-- 创建/编辑分类弹窗 -->
    <n-modal v-model:show="showModal" :title="isEditing ? '编辑分类' : '创建分类'" preset="card" style="width: 450px">
      <n-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleSubmit">
        <n-form-item label="分类名称" path="name">
          <n-input v-model:value="form.name" placeholder="请输入分类名称" :maxlength="100" show-count />
        </n-form-item>

        <n-space justify="end" style="margin-top: 16px">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" attr-type="submit" :loading="submitting">确定</n-button>
        </n-space>
      </n-form>
    </n-modal>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useMessage } from 'naive-ui'
import { categoryAPI } from '@/api/products'
import type { CategoryTreeItem } from '@/api/products'

const message = useMessage()

const loading = ref(false)
const submitting = ref(false)
const categories = ref<CategoryTreeItem[]>([])

const showModal = ref(false)
const isEditing = ref(false)
const editingId = ref<number | null>(null)
const editingParentId = ref<number | null>(null)

interface CategoryForm {
  name: string
}

const form = reactive<CategoryForm>({
  name: '',
})

const rules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' },
    { max: 100, message: '名称最多100字', trigger: 'blur' },
  ],
}

async function fetchCategories() {
  loading.value = true
  try {
    categories.value = await categoryAPI.getTree()
  } catch {
    message.error('加载分类失败')
  } finally {
    loading.value = false
  }
}

function openCreateModal(parentId: number | null) {
  isEditing.value = false
  editingId.value = null
  editingParentId.value = parentId
  form.name = ''
  showModal.value = true
}

function openEditModal(item: CategoryTreeItem) {
  isEditing.value = true
  editingId.value = item.id
  editingParentId.value = null
  form.name = item.name
  showModal.value = true
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (isEditing.value && editingId.value) {
      await categoryAPI.update(editingId.value, { name: form.name })
      message.success('更新成功')
    } else {
      await categoryAPI.create({ name: form.name, parent_id: editingParentId.value })
      message.success('创建成功')
    }
    showModal.value = false
    await fetchCategories()
  } catch {
    message.error(isEditing.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

async function handleDelete(id: number) {
  try {
    await categoryAPI.delete(id)
    message.success('删除成功')
    await fetchCategories()
  } catch {
    message.error('删除失败，该分类下可能有子分类或商品')
  }
}

onMounted(() => {
  fetchCategories()
})
</script>
