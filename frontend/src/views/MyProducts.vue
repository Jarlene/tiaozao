<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import type { Product, ProductQuery } from '@/api'
import { productApi } from '@/api'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()

const products = ref<Product[]>([])
const loading = ref(false)
const error = ref('')
const total = ref(0)
const currentPage = ref(1)

/** 是否显示创建/编辑对话框 */
const dialogVisible = ref(false)
const dialogTitle = ref('发布商品')
const isEditing = ref(false) // 编辑模式还是新建模式

// 当前正在编辑的商品 ID
const editingProductId = ref('')

// 商品表单
const productForm = ref({
  title: '',
  description: '',
  price: 0,
  originalPrice: undefined as number | undefined,
  category: '',
  condition: 'good' as string,
})

const productFormRef = ref()
const submitting = ref(false)

/** 商品成色选项 */
const conditionOptions = [
  { value: 'new', label: '全新' },
  { value: 'like_new', label: '几乎全新' },
  { value: 'good', label: '良好' },
  { value: 'fair', label: '一般' },
  { value: 'poor', label: '较差' },
]

/** 表单校验规则 */
const rules = {
  title: [
    { required: true, message: '请输入商品标题', trigger: 'blur' },
    { min: 2, max: 100, message: '标题长度为 2-100 个字符', trigger: 'blur' },
  ],
  description: [
    { required: true, message: '请输入商品描述', trigger: 'blur' },
    { min: 10, message: '描述至少 10 个字符', trigger: 'blur' },
  ],
  price: [
    { required: true, message: '请输入价格', trigger: 'blur' },
    { type: 'number', min: 0.01, message: '价格必须大于 0', trigger: 'blur' },
  ],
  category: [{ required: true, message: '请输入商品分类', trigger: 'blur' }],
  condition: [{ required: true, message: '请选择商品成色', trigger: 'change' }],
}

onMounted(async () => {
  await loadProducts()
})

/** 加载我的商品列表 */
async function loadProducts() {
  loading.value = true
  error.value = ''
  try {
    const query: ProductQuery = { page: currentPage.value, pageSize: 10 }
    const res = await productApi.getMyProducts(query)
    products.value = res.items
    total.value = res.total
  } catch (err: any) {
    error.value = err?.response?.data?.message || '加载商品列表失败'
  } finally {
    loading.value = false
  }
}

/** 打开创建商品对话框 */
function openCreateDialog() {
  isEditing.value = false
  dialogTitle.value = '发布商品'
  editingProductId.value = ''
  productForm.value = {
    title: '',
    description: '',
    price: 0,
    originalPrice: undefined,
    category: '',
    condition: 'good',
  }
  dialogVisible.value = true
}

/** 打开编辑商品对话框 */
function openEditDialog(product: Product) {
  isEditing.value = true
  dialogTitle.value = '编辑商品'
  editingProductId.value = product.id
  productForm.value = {
    title: product.title,
    description: product.description,
    price: product.price,
    originalPrice: product.originalPrice,
    category: product.category,
    condition: product.condition,
  }
  dialogVisible.value = true
}

/** 提交商品表单（创建或更新） */
async function handleSubmit() {
  if (!productFormRef.value) return

  const valid = await productFormRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    // 构建 FormData 以支持图片上传
    const formData = new FormData()
    formData.append('title', productForm.value.title)
    formData.append('description', productForm.value.description)
    formData.append('price', String(productForm.value.price))
    if (productForm.value.originalPrice) {
      formData.append('originalPrice', String(productForm.value.originalPrice))
    }
    formData.append('category', productForm.value.category)
    formData.append('condition', productForm.value.condition)

    if (isEditing.value && editingProductId.value) {
      await productApi.update(editingProductId.value, formData)
      ElMessage.success('商品已更新')
    } else {
      await productApi.create(formData)
      ElMessage.success('商品已发布')
    }

    dialogVisible.value = false
    await loadProducts()
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

/** 删除商品 */
async function handleDelete(product: Product) {
  await ElMessageBox.confirm(
    `确定要删除商品"${product.title}"吗？此操作不可恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )

  try {
    await productApi.delete(product.id)
    ElMessage.success('商品已删除')
    await loadProducts()
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '删除失败')
  }
}

/** 格式化价格 */
function formatPrice(price: number) {
  return `¥${price.toFixed(2)}`
}

/** 获取成色标签 */
function getConditionLabel(condition: string) {
  const map: Record<string, string> = {
    new: '全新',
    like_new: '几乎全新',
    good: '良好',
    fair: '一般',
    poor: '较差',
  }
  return map[condition] || condition
}

/** 获取状态标签类型 */
function getStatusType(status: string) {
  const map: Record<string, string> = {
    active: 'success',
    sold: 'danger',
    inactive: 'info',
  }
  return map[status] || 'info'
}

/** 获取状态文本 */
function getStatusLabel(status: string) {
  const map: Record<string, string> = {
    active: '在售',
    sold: '已售出',
    inactive: '下架',
  }
  return map[status] || status
}
</script>

<template>
  <div class="my-products-page">
    <div class="page-header">
      <h2 class="page-title">我的商品</h2>
      <el-button type="primary" @click="openCreateDialog">
        <el-icon><Plus /></el-icon>
        发布商品
      </el-button>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <el-skeleton :rows="4" animated />
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="error-state">
      <el-result icon="error" title="加载失败" :sub-title="error">
        <template #extra>
          <el-button type="primary" @click="loadProducts">重新加载</el-button>
        </template>
      </el-result>
    </div>

    <!-- 空状态 -->
    <div v-else-if="products.length === 0" class="empty-state">
      <el-empty description="还没有发布商品">
        <el-button type="primary" @click="openCreateDialog">发布第一个商品</el-button>
      </el-empty>
    </div>

    <!-- 商品表格 -->
    <template v-else>
      <el-card class="table-card">
        <el-table :data="products" stripe style="width: 100%">
          <el-table-column label="商品信息" min-width="300">
            <template #default="{ row }: { row: Product }">
              <div class="product-cell">
                <img
                  :src="row.images[0] || '/placeholder.svg'"
                  :alt="row.title"
                  class="cell-image"
                />
                <div class="cell-info">
                  <span class="cell-title">{{ row.title }}</span>
                  <span class="cell-category">{{ row.category }}</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="价格" width="120" align="center">
            <template #default="{ row }: { row: Product }">
              <span class="cell-price">{{ formatPrice(row.price) }}</span>
            </template>
          </el-table-column>

          <el-table-column label="成色" width="100" align="center">
            <template #default="{ row }: { row: Product }">
              <el-tag size="small">{{ getConditionLabel(row.condition) }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="100" align="center">
            <template #default="{ row }: { row: Product }">
              <el-tag :type="getStatusType(row.status)" size="small">
                {{ getStatusLabel(row.status) }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column label="发布时间" width="180" align="center">
            <template #default="{ row }: { row: Product }">
              {{ new Date(row.createdAt).toLocaleDateString('zh-CN') }}
            </template>
          </el-table-column>

          <el-table-column label="操作" width="200" align="center" fixed="right">
            <template #default="{ row }: { row: Product }">
              <el-button
                text
                type="primary"
                size="small"
                @click="router.push(`/products/${row.id}`)"
              >
                查看
              </el-button>
              <el-button
                text
                type="warning"
                size="small"
                @click="openEditDialog(row)"
              >
                编辑
              </el-button>
              <el-button
                text
                type="danger"
                size="small"
                @click="handleDelete(row)"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 分页 -->
        <div v-if="total > 10" class="pagination-wrapper">
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="10"
            :total="total"
            layout="prev, pager, next"
            background
            @current-change="loadProducts"
          />
        </div>
      </el-card>
    </template>

    <!-- 创建/编辑商品对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="productFormRef"
        :model="productForm"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="商品标题" prop="title">
          <el-input v-model="productForm.title" placeholder="请输入商品标题" />
        </el-form-item>

        <el-form-item label="商品描述" prop="description">
          <el-input
            v-model="productForm.description"
            type="textarea"
            :rows="4"
            placeholder="请描述商品的品牌、型号、使用情况等"
          />
        </el-form-item>

        <el-form-item label="价格" prop="price">
          <el-input-number
            v-model="productForm.price"
            :min="0.01"
            :precision="2"
            :step="10"
            style="width: 200px"
          />
        </el-form-item>

        <el-form-item label="原价">
          <el-input-number
            v-model="productForm.originalPrice"
            :min="0"
            :precision="2"
            :step="10"
            style="width: 200px"
            placeholder="选填"
          />
        </el-form-item>

        <el-form-item label="分类" prop="category">
          <el-input v-model="productForm.category" placeholder="例如：电子产品、家具等" />
        </el-form-item>

        <el-form-item label="商品成色" prop="condition">
          <el-select v-model="productForm.condition" style="width: 200px">
            <el-option
              v-for="opt in conditionOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEditing ? '保存修改' : '发布商品' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.my-products-page {
  padding-bottom: 40px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-title {
  font-size: 22px;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.loading-state {
  background: #fff;
  border-radius: 8px;
  padding: 40px;
}

.error-state {
  padding: 40px;
}

.empty-state {
  background: #fff;
  border-radius: 12px;
  padding: 60px;
}

.table-card {
  border-radius: 8px;
}

.product-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cell-image {
  width: 60px;
  height: 60px;
  border-radius: 6px;
  object-fit: cover;
  background: #f5f5f5;
}

.cell-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cell-title {
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.cell-category {
  font-size: 12px;
  color: #999;
}

.cell-price {
  font-size: 16px;
  font-weight: 500;
  color: #f56c6c;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}
</style>
