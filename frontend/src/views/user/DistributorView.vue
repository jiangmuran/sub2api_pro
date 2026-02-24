<template>
  <AppLayout>
    <div class="space-y-4">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('distributor.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('distributor.description') }}</p>
      </div>
      <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="card p-4">
          <div class="text-sm text-gray-500">{{ t('distributor.balance') }}</div>
          <div class="text-2xl font-semibold">{{ formatCNY(profile?.balance_cny_cents || 0) }}</div>
        </div>
      </div>

      <div class="card p-4">
        <h3 class="mb-3 text-lg font-semibold">{{ t('distributor.offers') }}</h3>
        <div class="space-y-3">
          <div v-for="offer in offers" :key="offer.id" class="rounded border border-gray-200 p-3 dark:border-dark-700">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <div>
                <div class="font-medium">{{ offer.name }}</div>
                <div class="text-xs text-gray-500">{{ t('distributor.validityDays') }}: {{ offer.validity_days }}</div>
              </div>
              <div class="flex items-center gap-2">
                <input v-model.number="sellPriceMap[offer.id]" type="number" min="1" class="input w-32" :placeholder="String(offer.cost_cny_cents)" />
                <button class="btn btn-primary" @click="buy(offer)">{{ t('distributor.buyCode') }}</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="mb-3 flex items-center gap-2">
          <input v-model="search" type="text" class="input max-w-72" :placeholder="t('distributor.searchOrders')" @input="loadOrders" />
          <Select v-model="status" :options="statusOptions" class="w-44" @change="loadOrders" />
        </div>
        <DataTable :columns="columns" :data="orders" :loading="loading">
          <template #cell-redeem_code="{ row }">
            <div class="flex items-center gap-2">
              <code>{{ row.redeem_code?.code }}</code>
              <button class="text-xs text-primary-600" @click="copyCode(row.redeem_code?.code || '')">{{ t('common.copy') }}</button>
            </div>
          </template>
          <template #cell-issued_at="{ value }">{{ formatDateTime(value) }}</template>
          <template #cell-redeemed_at="{ value }">{{ value ? formatDateTime(value) : '-' }}</template>
          <template #cell-actions="{ row }">
            <button v-if="row.status === 'issued'" class="btn btn-secondary btn-sm" @click="revoke(row.id)">{{ t('distributor.revoke') }}</button>
          </template>
        </DataTable>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { Column } from '@/components/common/types'
import Select from '@/components/common/Select.vue'
import { distributorAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import type { DistributorOffer, DistributorOrder, DistributorProfile } from '@/types'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const profile = ref<DistributorProfile | null>(null)
const offers = ref<DistributorOffer[]>([])
const orders = ref<DistributorOrder[]>([])
const loading = ref(false)
const search = ref('')
const status = ref('')
const sellPriceMap = ref<Record<number, number>>({})

const columns = computed<Column[]>(() => [
  { key: 'redeem_code', label: t('distributor.code') },
  { key: 'status', label: t('distributor.status') },
  { key: 'sell_price_cny_cents', label: t('distributor.sellPrice') },
  { key: 'issued_at', label: t('distributor.issuedAt') },
  { key: 'redeemed_at', label: t('distributor.redeemedAt') },
  { key: 'actions', label: t('common.actions') }
])

const statusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'issued', label: t('distributor.orderStatus.issued') },
  { value: 'redeemed', label: t('distributor.orderStatus.redeemed') },
  { value: 'revoked', label: t('distributor.orderStatus.revoked') }
])

const formatCNY = (cents: number) => `CNY ${(cents / 100).toFixed(2)}`

const loadProfile = async () => {
  profile.value = await distributorAPI.profile()
}

const loadOffers = async () => {
  offers.value = await distributorAPI.offers()
}

const loadOrders = async () => {
  loading.value = true
  try {
    const resp = await distributorAPI.orders(1, 50, { status: status.value, search: search.value })
    orders.value = resp.items
  } finally {
    loading.value = false
  }
}

const buy = async (offer: DistributorOffer) => {
  try {
    const sell = sellPriceMap.value[offer.id] || offer.cost_cny_cents
    const memo = window.prompt(t('distributor.memoPrompt')) || ''
    await distributorAPI.createOrder(offer.id, sell, memo)
    appStore.showSuccess(t('distributor.codeCreated'))
    await Promise.all([loadProfile(), loadOrders()])
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('common.operationFailed'))
  }
}

const revoke = async (orderId: number) => {
  try {
    await distributorAPI.revokeOrder(orderId)
    appStore.showSuccess(t('distributor.revoked'))
    await Promise.all([loadProfile(), loadOrders()])
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('common.operationFailed'))
  }
}

const copyCode = async (code: string) => {
  if (!code) return
  await navigator.clipboard.writeText(code)
}

onMounted(async () => {
  try {
    await Promise.all([loadProfile(), loadOffers(), loadOrders()])
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('common.operationFailed'))
  }
})
</script>
