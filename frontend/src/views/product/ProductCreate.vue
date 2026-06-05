<template>
  <n-card title="发布商品" style="max-width: 700px; margin: 0 auto">
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

      <n-form-item label="商品图片" path="image_ids">
        <n-space vertical>
          <!-- 上传按钮 -->
          <n-upload
            :disabled="uploadedImages.length >= 9"
            :show-file-list="false"
            :accept="'image/jpeg,image/png,image/webp'"
            @before-upload="handleBeforeUpload"
          >
            <n-button :disabled="uploadedImages.length >= 9" :loading="uploading">
              上传图片（{{ uploadedImages.length }}/9）
            </n-button>
          </n-upload>

          <!-- 已上传预览 -->
          <n-grid :cols="4" :x-gap="8" :y-gap="8" v-if="uploadedImages.length > 0">
            <n-grid-item v-for="(img, index) in uploadedImages" :key="img.id">
              <n-card size="small" :bordered="true" style="text-align: center">
                <n-image
                  :src="img.url"
                  style="width: 100%; height: 100px; object-fit: cover"
                  fallback-src="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><rect fill='%23e0e0e0' width='100' height='100'/></svg>"
                />
                <template #footer>
                  <n-button text type="error" size="tiny" @click="removeImage(index)">
                    删除
                  </n-button>
                </template>
              </n-card>
            </n-grid-item>
          </n-grid>
        </n-space>
      </n-form-item>

      <n-space justify="center" style="margin-top: 24px">
        <n-button type="primary" attr-type="submit" :loading="submitting">
          发布商品
        </n-button>
        <n-button @click="$router.push('/products')">取消</n-button>
      </n-space>
    </n-form>
  </n-card>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from 'naive-ui'
import type { FormRules, UploadFileInfo } from 'naive-ui'
import { productAPI, categoryAPI } from '@/api/products'
import type { ProductImageItem, CategoryTreeItem } from '@/api/products'
import { flattenCategories } from '@/utils/categories'
import { yuanToCents } from '@/utils/format'

const router = useRouter()
const message = useMessage()

const submitting = ref(false)
const uploading = ref(false)
const uploadedImages = ref<ProductImageItem[]>([])

interface ProductCreateForm {
  title: string
  description: string
  price: number | null
  category_id: number | null
}

const form = reactive<ProductCreateForm>({
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

async function handleBeforeUpload(data: { file: UploadFileInfo; fileList: UploadFileInfo[] }): Promise<boolean> {
  const file = data.file.file
  if (!file) return false

  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    message.error('仅支持 jpg/png/webp 格式')
    return false
  }

  // 验证文件大小（5MB）
  if (file.size > 5 * 1024 * 1024) {
    message.error('图片大小不能超过 5MB')
    return false
  }

  try {
    uploading.value = true
    const result = await productAPI.uploadImage(file)
    uploadedImages.value.push(result)
    message.success('上传成功')
  } catch {
    message.error('上传失败')
  } finally {
    uploading.value = false
  }

  return false // 阻止默认上传行为
}

async function removeImage(index: number) {
  const img = uploadedImages.value[index]
  try {
    await productAPI.deleteImage(img.id)
    uploadedImages.value.splice(index, 1)
  } catch {
    message.error('删除失败')
  }
}

async function handleSubmit() {
  if (!form.price && form.price !== 0) {
    message.error('请输入价格')
    return
  }

  submitting.value = true
  try {
    const data = await productAPI.create({
      title: form.title,
      description: form.description,
      price: yuanToCents(form.price),
      category_id: form.category_id || null,
      image_ids: uploadedImages.value.map((img) => img.id),
    })
    message.success('发布成功')
    router.push(`/products/${data.id}`)
  } catch {
    message.error('发布失败')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  try {
    const data = await categoryAPI.getTree()
    categories.value = data
    categoryOptions.value = flattenCategories(data)
  } catch {
    // 分类加载失败不影响
  }
})
</script>
