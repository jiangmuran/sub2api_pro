<template>
  <AppLayout>
    <div class="space-y-6">
      <section
        class="rounded-2xl border border-gray-200 bg-gradient-to-br from-sky-50 via-white to-emerald-50 p-5 shadow-sm dark:border-dark-700 dark:from-dark-900 dark:via-dark-900 dark:to-dark-800"
      >
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('distributor.title') }}</h1>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">{{ t('distributor.description') }}</p>
          </div>

          <button class="btn btn-primary" :disabled="isRefreshing" @click="refreshAll">
            {{ t('common.refresh') }}
          </button>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <article class="rounded-xl border border-emerald-200 bg-white p-4 dark:border-emerald-800 dark:bg-dark-900">
            <p class="text-xs font-medium uppercase tracking-wide text-emerald-700 dark:text-emerald-300">
              {{ t('distributor.balance') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatCNY(profile?.balance_cny_cents || 0) }}
            </p>
          </article>

          <article class="rounded-xl border border-sky-200 bg-white p-4 dark:border-sky-800 dark:bg-dark-900">
            <p class="text-xs font-medium uppercase tracking-wide text-sky-700 dark:text-sky-300">
              {{ t('distributor.orderStatus.issued') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ issuedCount }}</p>
          </article>

          <article class="rounded-xl border border-cyan-200 bg-white p-4 dark:border-cyan-800 dark:bg-dark-900">
            <p class="text-xs font-medium uppercase tracking-wide text-cyan-700 dark:text-cyan-300">
              {{ t('distributor.orderStatus.redeemed') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ redeemedCount }}</p>
          </article>

          <article class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
            <p class="text-xs font-medium uppercase tracking-wide text-gray-600 dark:text-gray-300">
              {{ t('distributor.orderStatus.revoked') }}
            </p>
            <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ revokedCount }}</p>
          </article>
        </div>
      </section>

      <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('distributor.offers') }}</h2>
          <span class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('common.total') }} {{ offers.length }}
          </span>
        </div>

        <div class="grid grid-cols-1 gap-3 xl:grid-cols-2">
          <article
            v-for="offer in offers"
            :key="offer.id"
            class="rounded-xl border p-4"
            :class="offer.enabled
              ? 'border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900'
              : 'border-gray-200 bg-gray-50/80 opacity-80 dark:border-dark-700 dark:bg-dark-800/60'"
          >
            <div class="flex flex-wrap items-start justify-between gap-2">
              <div>
                <p class="text-base font-semibold text-gray-900 dark:text-white">{{ offer.name }}</p>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('distributor.validityDays') }}: {{ offer.validity_days }}
                </p>
              </div>
              <span
                class="inline-flex rounded-full px-2 py-1 text-xs font-medium"
                :class="offer.enabled
                  ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                  : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
              >
                {{ offer.enabled ? t('common.enabled') : t('common.disabled') }}
              </span>
            </div>

            <div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div class="rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-800">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.distributor.cost') }}</p>
                <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                  {{ formatCNY(offer.cost_cny_cents) }}
                </p>
              </div>
              <div class="rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-800">
                <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('distributor.sellPrice') }}</p>
                <input
                  v-model.number="sellPriceMap[offer.id]"
                  type="number"
                  min="0.01"
                  step="0.01"
                  class="input mt-1"
                  @blur="normalizeSellPrice(offer)"
                />
              </div>
            </div>

            <div class="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
              <input
                v-model="memoMap[offer.id]"
                type="text"
                maxlength="200"
                class="input"
                :placeholder="t('distributor.memoPrompt')"
              />
              <button
                class="btn btn-primary"
                :disabled="creatingOrderId === offer.id || !offer.enabled"
                @click="buy(offer)"
              >
                {{ creatingOrderId === offer.id ? t('common.processing') : t('distributor.buyCode') }}
              </button>
            </div>
          </article>
        </div>
      </section>

      <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <input
            v-model="search"
            type="text"
            class="input max-w-80"
            :placeholder="t('distributor.searchOrders')"
            @input="handleOrderSearch"
          />
          <Select v-model="status" :options="statusOptions" class="w-44" @change="loadOrders" />
          <button class="btn btn-secondary" :disabled="loadingOrders" @click="loadOrders">
            {{ t('common.refresh') }}
          </button>
        </div>

        <DataTable :columns="columns" :data="orders" :loading="loadingOrders" :sticky-actions-column="false">
          <template #cell-redeem_code="{ row }">
            <div class="flex items-center gap-2">
              <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-800 dark:bg-dark-700 dark:text-dark-100">
                {{ getRedeemCode(row) || '-' }}
              </code>
              <button class="text-xs text-primary-600 dark:text-primary-400" @click="copyCode(getRedeemCode(row))">
                {{ t('admin.redeem.copyCode') }}
              </button>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span class="inline-flex rounded-full px-2 py-1 text-xs font-medium" :class="orderStatusClass(value)">
              {{ t(`distributor.orderStatus.${value}`) }}
            </span>
          </template>

          <template #cell-sell_price_cny_cents="{ value }">{{ formatCNY(value) }}</template>
          <template #cell-issued_at="{ value }">{{ formatDateTime(value) }}</template>
          <template #cell-redeemed_at="{ value }">{{ value ? formatDateTime(value) : '-' }}</template>
          <template #cell-memo="{ value }">{{ value || '-' }}</template>

          <template #cell-actions="{ row }">
            <button
              v-if="row.status === 'issued'"
              class="btn btn-secondary btn-sm"
              :disabled="revokingOrderId === row.id"
              @click="revoke(row.id)"
            >
              {{ revokingOrderId === row.id ? t('common.processing') : t('distributor.revoke') }}
            </button>
            <span v-else class="text-xs text-gray-400">-</span>
          </template>
        </DataTable>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
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

const loadingOrders = ref(false)
const loadingProfile = ref(false)
const loadingOffers = ref(false)
const isRefreshing = ref(false)
const creatingOrderId = ref<number | null>(null)
const revokingOrderId = ref<number | null>(null)

const search = ref('')
const status = ref('')
const sellPriceMap = ref<Record<number, number>>({})
const memoMap = ref<Record<number, string>>({})

const columns = computed<Column[]>(() => [
  { key: 'redeem_code', label: t('distributor.code') },
  { key: 'status', label: t('distributor.status') },
  { key: 'sell_price_cny_cents', label: t('distributor.sellPrice') },
  { key: 'memo', label: t('admin.distributor.notes') },
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

const issuedCount = computed(() => orders.value.filter((item) => item.status === 'issued').length)
const redeemedCount = computed(() => orders.value.filter((item) => item.status === 'redeemed').length)
const revokedCount = computed(() => orders.value.filter((item) => item.status === 'revoked').length)

const formatCNY = (cents: number) => `CNY ${(cents / 100).toFixed(2)}`

const getRedeemCode = (order: DistributorOrder): string => {
  const redeem = (order as any)?.redeem_code
  if (!redeem) {
    return ''
  }
  return String(redeem.code ?? redeem.Code ?? '').trim()
}

const orderStatusClass = (statusValue: string) => {
  if (statusValue === 'issued') {
    return 'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-300'
  }
  if (statusValue === 'redeemed') {
    return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  }
  return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
}

const showRequestError = (error: any) => {
  appStore.showError(error?.response?.data?.detail || t('common.unknownError'))
}

const loadProfile = async () => {
  loadingProfile.value = true
  try {
    profile.value = await distributorAPI.profile()
  } finally {
    loadingProfile.value = false
  }
}

const loadOffers = async () => {
  loadingOffers.value = true
  try {
    offers.value = await distributorAPI.offers()
    for (const offer of offers.value) {
      if (!Number.isFinite(sellPriceMap.value[offer.id]) || sellPriceMap.value[offer.id] <= 0) {
        sellPriceMap.value[offer.id] = Number((offer.cost_cny_cents / 100).toFixed(2))
      }
      if (memoMap.value[offer.id] === undefined) {
        memoMap.value[offer.id] = ''
      }
    }
  } finally {
    loadingOffers.value = false
  }
}

const loadOrders = async () => {
  loadingOrders.value = true
  try {
    const resp = await distributorAPI.orders(1, 50, {
      status: status.value,
      search: search.value.trim()
    })
    orders.value = resp.items
  } finally {
    loadingOrders.value = false
  }
}

const normalizeSellPrice = (offer: DistributorOffer) => {
  const value = Number(sellPriceMap.value[offer.id])
  if (!Number.isFinite(value) || value <= 0) {
    sellPriceMap.value[offer.id] = Number((offer.cost_cny_cents / 100).toFixed(2))
    return
  }
  sellPriceMap.value[offer.id] = Number(value.toFixed(2))
}

const buy = async (offer: DistributorOffer) => {
  if (!offer.enabled) {
    return
  }

  normalizeSellPrice(offer)
  const sellPriceYuan = Number(sellPriceMap.value[offer.id])
  const sellPriceCents = Math.round(sellPriceYuan * 100)
  if (!Number.isFinite(sellPriceCents) || sellPriceCents <= 0) {
    appStore.showError(t('common.unknownError'))
    return
  }

  creatingOrderId.value = offer.id
  try {
    await distributorAPI.createOrder(offer.id, sellPriceCents, memoMap.value[offer.id]?.trim() || '')
    memoMap.value[offer.id] = ''
    appStore.showSuccess(t('distributor.codeCreated'))
    await Promise.all([loadProfile(), loadOrders()])
  } catch (error: any) {
    showRequestError(error)
  } finally {
    creatingOrderId.value = null
  }
}

const revoke = async (orderId: number) => {
  revokingOrderId.value = orderId
  try {
    await distributorAPI.revokeOrder(orderId)
    appStore.showSuccess(t('distributor.revoked'))
    await Promise.all([loadProfile(), loadOrders()])
  } catch (error: any) {
    showRequestError(error)
  } finally {
    revokingOrderId.value = null
  }
}

const copyCode = async (code: string) => {
  if (!code) {
    return
  }

  try {
    await navigator.clipboard.writeText(code)
    appStore.showSuccess(t('common.copiedToClipboard'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

let searchTimer: ReturnType<typeof setTimeout> | null = null

const handleOrderSearch = () => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
  searchTimer = setTimeout(() => {
    loadOrders().catch((error: any) => showRequestError(error))
  }, 250)
}

const refreshAll = async () => {
  isRefreshing.value = true
  try {
    await Promise.all([loadProfile(), loadOffers(), loadOrders()])
  } catch (error: any) {
    showRequestError(error)
  } finally {
    isRefreshing.value = false
  }
}

onMounted(async () => {
  await refreshAll()
})

onUnmounted(() => {
  if (searchTimer) {
    clearTimeout(searchTimer)
  }
})
</script>
