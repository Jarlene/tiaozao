<template>
  <n-card title="编辑商品" style="max-width: 700px; margin: 0 auto">
    <n-spin :show="loading">
      <n-form ref="formRef" :rules="rules" :model="form" @submit.prevent="handleSubmit">
        <n-form-item label="标题" path="title">
          <n-input v-model:value="form.title" placeholder="商品标题" :maxlength="200" show-count />
        </n-form-item>

        <n-form-item label="描述" path="description">
          <n-input
            v-model:value="form.description"
            type="textarea"
            placeholder="商品描述"
            :rows="4"
          />
        </n-form-item>

        <n-form-item label="价格" path="price">
          <n-input-number
            v-model:value="form.price"
            placeholder="价格（元）"
            :min="0"
            :precision="2"
            style="width: 200px"
          >
            <template #prefix>¥</template>
          </n-input-number>
        </n-form-item>

        <n-form-item label="分类" path="category_id">
          <n-select
            v-model:value="form.category_id"
            :options="categoryOptions"
            placeholder="选择分类"
            clearable
          />
        </n-form-item>

        <n-space justify="center" style="margin-top: 24px">
          <n-button type="primary" attr-type="submit" :loading="submitting">
            保存修改
          </n-button>
          <n-button @click="$router.back()">取消</n-button>
        </n-space>
      </n-form>
    </n-spin>
  </n-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import type { FormRules } from 'naive-ui'
import { productAPI, categoryAPI } from '@/api/products'
import type { CategoryTreeItem } from '@/api/products'
import { flattenCategories } from '@/utils/categories'

const route = useRoute()
const router = useRouter()
const message = useMessage()

const loading = ref(true)
const submitting = ref(false)

interface ProductEditForm {
  title: string
  description: string
  price: number | null
  category_id: number | null
}

const form = reactive<ProductEditForm>({
  title: '',
  description: '',
  price: null,
  category_id: null,
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入商品标题', trigger: 'blur' },
    { max: 200, message: '标题最多200字', trigger: 'blur' },
  ],
  price: [
    { required: true, message: '请输入价格', trigger: 'blur' },
    { type: 'number', min: 0, message: '价格不能为负', trigger: 'blur' },
  ],
}

const categories = ref<CategoryTreeItem[]>([])
const categoryOptions = ref<Array<{ label: string; value: number }>>([])

async function loadProduct() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const data = await productAPI.getById(id)
    form.title = data.title
    form.description = data.description || ''
    form.price = data.price / 100 // 分转元
    form.category_id = data.category_id
  } catch {
    message.error('加载商品信息失败')
    router.push('/products')
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (!form.price && form.price !== 0) {
    message.error('请输入价格')
    return
  }

  submitting.value = true
  try {
    const id = Number(route.params.id)
    await productAPI.update(id, {
      title: form.title,
      description: form.description,
      price: Math.round(form.price * 100),
      category_id: form.category_id || null,
    })
    message.success('保存成功')
    router.push(`/products/${id}`)
  } catch {
    message.error('保存失败')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await Promise.all([
    loadProduct(),
    categoryAPI.getTree().then((data) => {
      categories.value = data
      categoryOptions.value = flattenCategories(data)
    }).catch(() => {}),
  ])
})
</script>
