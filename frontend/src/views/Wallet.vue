<template>
  <n-space vertical :size="20">
    <n-h2>我的钱包</n-h2>

    <!-- 余额概览 -->
    <n-grid :cols="3" :x-gap="16">
      <n-gi>
        <n-card title="可用余额" :bordered="true" hoverable>
          <n-statistic>
            <n-number-animation :from="0" :to="walletInfo?.balance || 0" :format="formatBalance" />
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="冻结余额" :bordered="true" hoverable>
          <n-statistic>
            <n-number-animation :from="0" :to="walletInfo?.frozen_balance || 0" :format="formatBalance" />
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="总资产" :bordered="true" hoverable>
          <n-statistic>
            <n-number-animation :from="0" :to="walletInfo?.total_balance || 0" :format="formatBalance" />
          </n-statistic>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 充值操作 -->
    <n-card title="充值" :bordered="true" size="small">
      <n-space align="center">
        <n-input-number
          v-model:value="topUpAmount"
          :min="1"
          :max="10000000"
          placeholder="输入充值金额（分）"
          style="width: 200px"
        />
        <n-text depth="3" style="font-size: 12px">（1元 = 100分）</n-text>
        <n-button type="primary" :loading="topUpLoading" @click="handleTopUp">
          充值
        </n-button>
      </n-space>
    </n-card>

    <!-- 交易流水 -->
    <n-card title="交易流水" :bordered="true">
      <n-spin :show="loading">
        <n-empty v-if="transactions.length === 0 && !loading" description="暂无交易记录" style="padding: 40px 0" />

        <n-list v-else>
          <n-list-item v-for="tx in transactions" :key="tx.id">
            <n-space align="center" justify="space-between" style="width: 100%">
              <n-space vertical :size="2">
                <n-space align="center" size="small">
                  <n-tag :type="getTxTypeTag(tx.type)" size="tiny">
                    {{ tx.type_name }}
                  </n-tag>
                  <n-text depth="3" style="font-size: 12px">{{ tx.created_at }}</n-text>
                </n-space>
                <n-text v-if="tx.description" depth="3" style="font-size: 12px">
                  {{ tx.description }}
                </n-text>
              </n-space>
              <n-text :style="{ color: tx.amount >= 0 ? '#18a058' : '#f5222d', fontWeight: 'bold' }">
                {{ tx.amount >= 0 ? '+' : '' }}{{ (tx.amount / 100).toFixed(2) }} 元
              </n-text>
            </n-space>
          </n-list-item>
        </n-list>

        <!-- 分页 -->
        <n-space justify="center" style="margin-top: 16px" v-if="totalPages > 1">
          <n-pagination
            :page="currentPage"
            :page-size="pageSize"
            :item-count="total"
            @update:page="changePage"
          />
        </n-space>
      </n-spin>
    </n-card>
  </n-space>
</template>

<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { walletAPI } from '@/api/wallet'
import type { WalletInfo, TransactionItem } from '@/api/wallet'

const message = useMessage()

const walletInfo = ref<WalletInfo | null>(null)
const loading = ref(false)
const transactions = ref<TransactionItem[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const topUpAmount = ref(10000) // 默认100元
const topUpLoading = ref(false)

function formatBalance(value: number): string {
  return (value / 100).toFixed(2) + ' 元'
}

function getTxTypeTag(type: string): 'success' | 'warning' | 'info' | 'error' | 'default' {
  switch (type) {
    case 'top_up':
    case 'complete':
      return 'success'
    case 'pay':
      return 'warning'
    case 'refund':
      return 'info'
    case 'withdraw':
      return 'error'
    default:
      return 'default'
  }
}

async function loadWalletInfo() {
  try {
    walletInfo.value = await walletAPI.getWallet()
  } catch {
    message.error('加载钱包信息失败')
  }
}

async function loadTransactions() {
  loading.value = true
  try {
    const result = await walletAPI.listTransactions({
      page: currentPage.value,
      size: pageSize.value,
    })
    transactions.value = result.items
    total.value = result.total
  } catch {
    transactions.value = []
    message.error('加载交易记录失败')
  } finally {
    loading.value = false
  }
}

async function handleTopUp() {
  if (!topUpAmount.value || topUpAmount.value <= 0) {
    message.warning('请输入充值金额')
    return
  }

  topUpLoading.value = true
  try {
    const info = await walletAPI.topUp(topUpAmount.value)
    walletInfo.value = info
    message.success(`充值成功！充值 ${(topUpAmount.value / 100).toFixed(2)} 元`)
    topUpAmount.value = 10000
    // 刷新交易流水
    loadTransactions()
  } catch (err: any) {
    const msg = err?.response?.data?.message || '充值失败'
    message.error(msg)
  } finally {
    topUpLoading.value = false
  }
}

function changePage(page: number) {
  currentPage.value = page
  loadTransactions()
}

onMounted(() => {
  loadWalletInfo()
  loadTransactions()
})
</script>
