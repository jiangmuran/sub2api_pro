<template>
  <AppLayout>
    <div class="space-y-6">
      <section
        class="rounded-2xl border border-gray-200 bg-gradient-to-br from-emerald-50 via-white to-cyan-50 p-5 shadow-sm dark:border-dark-700 dark:from-dark-900 dark:via-dark-900 dark:to-dark-800"
      >
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
              {{ t('admin.distributor.title') }}
            </h1>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">
              {{ t('admin.distributor.description') }}
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <input
              v-model="settleNotes"
              type="text"
              maxlength="200"
              class="input w-52"
              :placeholder="t('admin.distributor.notes')"
            />
            <button class="btn btn-secondary" :disabled="settling" @click="markSettled">
              {{ settling ? t('common.processing') : t('admin.distributor.markSettled') }}
            </button>
            <button class="btn btn-primary" :disabled="isRefreshing" @click="loadAll">
              {{ t('common.refresh') }}
            </button>
          </div>
        </div>

        <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto_auto]">
          <input
            v-model="profileSearch"
            type="text"
            class="input"
            :placeholder="t('admin.distributor.searchProfile')"
            @input="handleProfileSearch"
          />
          <input
            v-model="createForm.email"
            type="email"
            class="input"
            :placeholder="t('admin.distributor.email')"
          />
          <Select v-model="createForm.copy_from_user_id" :options="copyProfileOptions" />
          <Select v-model="createForm.enabled" :options="enabledOptions" />
          <button class="btn btn-primary" :disabled="savingProfile" @click="upsertProfile">
            {{ savingProfile ? t('common.processing') : t('common.save') }}
          </button>
        </div>
        <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
          {{ t('admin.distributor.copyOffersHint') }}
        </p>
        <input
          v-model="createForm.notes"
          type="text"
          maxlength="200"
          class="input mt-3"
          :placeholder="t('admin.distributor.notes')"
        />
      </section>

      <section class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <article class="rounded-2xl border border-emerald-200 bg-white p-4 shadow-sm dark:border-emerald-800 dark:bg-dark-900">
          <p class="text-xs font-medium uppercase tracking-wide text-emerald-700 dark:text-emerald-300">
            {{ t('admin.distributor.unsettled') }}
          </p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ formatCNY(summary?.unsettled_cny_cents) }}
          </p>
        </article>

        <article class="rounded-2xl border border-cyan-200 bg-white p-4 shadow-sm dark:border-cyan-800 dark:bg-dark-900">
          <p class="text-xs font-medium uppercase tracking-wide text-cyan-700 dark:text-cyan-300">
            {{ t('admin.distributor.delta') }}
          </p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ formatCNY(summary?.delta_since_last_settle_cny) }}
          </p>
        </article>

        <article class="rounded-2xl border border-blue-200 bg-white p-4 shadow-sm dark:border-blue-800 dark:bg-dark-900">
          <p class="text-xs font-medium uppercase tracking-wide text-blue-700 dark:text-blue-300">
            {{ t('admin.distributor.grossProfit') }}
          </p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ formatCNY(totalGrossProfit) }}
          </p>
        </article>

        <article class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <p class="text-xs font-medium uppercase tracking-wide text-gray-600 dark:text-gray-300">
            {{ t('admin.distributor.orders') }}
          </p>
          <p class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
            {{ totalOrderCount }}
          </p>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t('common.enabled') }} {{ enabledProfileCount }} / {{ totalProfileCount }}
          </p>
        </article>
      </section>

      <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="mb-3 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.distributor.statsByUser') }}
          </h2>
          <span class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('common.total') }} {{ summaryRows.length }}
          </span>
        </div>
        <DataTable :columns="statsColumns" :data="summaryRows" :sticky-actions-column="false">
          <template #cell-distributor_user_id="{ value }">
            <span class="font-medium text-gray-800 dark:text-gray-100">{{ resolveUserEmail(value) }}</span>
          </template>
          <template #cell-sell_amount_cny="{ value }">{{ formatCNY(value) }}</template>
          <template #cell-cost_amount_cny="{ value }">{{ formatCNY(value) }}</template>
          <template #cell-refund_amount_cny="{ value }">{{ formatCNY(value) }}</template>
          <template #cell-gross_profit_cny="{ value }">{{ formatCNY(value) }}</template>
        </DataTable>
      </section>

      <div class="grid grid-cols-1 gap-6 xl:grid-cols-12">
        <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900 xl:col-span-5">
          <div class="mb-3 flex items-center justify-between">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.distributor.email') }}
            </h2>
            <span class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('common.total') }} {{ profiles.length }}
            </span>
          </div>

          <DataTable
            :columns="profileColumns"
            :data="profiles"
            :loading="loadingProfiles"
            :sticky-actions-column="false"
          >
            <template #cell-user="{ row }">
              <div class="min-w-40">
                <p class="font-medium text-gray-900 dark:text-gray-100">{{ getProfileUserEmail(row) }}</p>
                <p class="truncate text-xs text-gray-500 dark:text-gray-400">{{ row.notes || '-' }}</p>
              </div>
            </template>

            <template #cell-enabled="{ value }">
              <span
                class="inline-flex rounded-full px-2 py-1 text-xs font-medium"
                :class="value
                  ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                  : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
              >
                {{ value ? t('common.enabled') : t('common.disabled') }}
              </span>
            </template>

            <template #cell-balance_cny_cents="{ value }">{{ formatCNY(value) }}</template>

            <template #cell-actions="{ row }">
              <button
                class="btn btn-secondary btn-sm"
                :class="selectedUserId === row.user_id ? 'ring-2 ring-primary-300 dark:ring-primary-700' : ''"
                @click="selectUser(row.user_id)"
              >
                {{ t('admin.distributor.manage') }}
              </button>
            </template>
          </DataTable>
        </section>

        <section class="space-y-6 xl:col-span-7">
          <div
            v-if="!selectedProfile"
            class="rounded-2xl border border-dashed border-gray-300 bg-white p-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:bg-dark-900 dark:text-gray-400"
          >
            {{ t('admin.distributor.description') }}
          </div>

          <template v-else>
            <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ selectedUserEmail }}
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ selectedProfile.notes || '-' }}</p>
                </div>
                <span
                  class="inline-flex rounded-full px-2 py-1 text-xs font-medium"
                  :class="selectedProfile.enabled
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                    : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
                >
                  {{ selectedProfile.enabled ? t('common.enabled') : t('common.disabled') }}
                </span>
              </div>

              <div class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-3">
                <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.distributor.balance') }}</p>
                  <p class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
                    {{ formatCNY(selectedProfile.balance_cny_cents) }}
                  </p>
                </div>
                <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.distributor.orders') }}</p>
                  <p class="mt-1 text-base font-semibold text-gray-900 dark:text-white">{{ selectedOrderCount }}</p>
                </div>
                <div class="rounded-xl bg-gray-50 p-3 dark:bg-dark-800">
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.distributor.grossProfit') }}</p>
                  <p class="mt-1 text-base font-semibold text-gray-900 dark:text-white">
                    {{ formatCNY(selectedGrossProfit) }}
                  </p>
                </div>
              </div>

              <div class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-[minmax(0,120px)_minmax(0,1fr)_auto_auto]">
                <input
                  v-model.number="balanceForm.amount"
                  type="number"
                  step="0.01"
                  min="0.01"
                  class="input"
                  :placeholder="t('admin.distributor.amountPrompt')"
                />
                <input
                  v-model="balanceForm.notes"
                  type="text"
                  maxlength="200"
                  class="input"
                  :placeholder="t('admin.distributor.notes')"
                />
                <button
                  class="btn btn-secondary"
                  :disabled="adjustingBalance"
                  @click="submitBalanceAdjust('topup')"
                >
                  +{{ t('admin.distributor.topup') }}
                </button>
                <button
                  class="btn btn-secondary"
                  :disabled="adjustingBalance"
                  @click="submitBalanceAdjust('refund')"
                >
                  -{{ t('admin.distributor.refund') }}
                </button>
              </div>
            </section>

            <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
              <h3 class="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.distributor.offers') }}
              </h3>

              <div class="mb-4 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-6">
                <input
                  v-model="offerForm.name"
                  type="text"
                  class="input xl:col-span-2"
                  maxlength="128"
                  :placeholder="t('admin.distributor.offerName')"
                />
                <Select v-model="offerForm.target_group_id" :options="groupOptions" />
                <input v-model.number="offerForm.validity_days" type="number" min="1" max="36500" class="input" />
                <input v-model.number="offerForm.cost_cny" type="number" min="0.01" step="0.01" class="input" />
                <Select v-model="offerForm.enabled" :options="enabledOptions" />
              </div>

              <div class="mb-4 flex flex-wrap items-center gap-3">
                <input
                  v-model="offerForm.notes"
                  type="text"
                  maxlength="200"
                  class="input min-w-64 flex-1"
                  :placeholder="t('admin.distributor.notes')"
                />
                <button class="btn btn-primary" :disabled="creatingOffer" @click="createOffer">
                  {{ creatingOffer ? t('common.processing') : t('common.create') }}
                </button>
              </div>

              <DataTable :columns="offerColumns" :data="offers" :loading="loadingOffers" :sticky-actions-column="false">
                <template #cell-target_group_id="{ value }">
                  {{ groupNameMap[value] || value }}
                </template>
                <template #cell-cost_cny_cents="{ value }">{{ formatCNY(value) }}</template>
                <template #cell-enabled="{ value }">
                  <span
                    class="inline-flex rounded-full px-2 py-1 text-xs font-medium"
                    :class="value
                      ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                      : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'"
                  >
                    {{ value ? t('common.enabled') : t('common.disabled') }}
                  </span>
                </template>
                <template #cell-notes="{ value }">{{ value || '-' }}</template>
                <template #cell-actions="{ row }">
                  <button class="btn btn-danger btn-sm" :disabled="deletingOfferId === row.id" @click="deleteOffer(row.id)">
                    {{ deletingOfferId === row.id ? t('common.processing') : t('common.delete') }}
                  </button>
                </template>
              </DataTable>
            </section>

            <section class="rounded-2xl border border-gray-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
              <h3 class="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('admin.distributor.orders') }}
              </h3>

              <div class="mb-3 flex flex-wrap items-center gap-2">
                <input
                  v-model="orderSearch"
                  type="text"
                  class="input max-w-80"
                  :placeholder="t('admin.distributor.searchOrders')"
                  @input="handleOrderSearch"
                />
                <Select v-model="orderStatus" :options="orderStatusOptions" class="w-44" @change="loadOrders" />
                <button class="btn btn-secondary" :disabled="loadingOrders" @click="loadOrders">
                  {{ t('common.refresh') }}
                </button>
              </div>

              <DataTable :columns="orderColumns" :data="orders" :loading="loadingOrders" :sticky-actions-column="false">
                <template #cell-redeem_code="{ row }">
                  <code class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-800 dark:bg-dark-700 dark:text-dark-100">
                    {{ getRedeemCode(row) || '-' }}
                  </code>
                </template>
                <template #cell-status="{ value }">
                  <span
                    class="inline-flex rounded-full px-2 py-1 text-xs font-medium"
                    :class="orderStatusClass(value)"
                  >
                    {{ t(`admin.distributor.orderStatus.${value}`) }}
                  </span>
                </template>
                <template #cell-used_email="{ row }">{{ row.redeem_code?.user?.email || '-' }}</template>
                <template #cell-issued_at="{ value }">{{ formatDateTime(value) }}</template>
                <template #cell-redeemed_at="{ value }">{{ value ? formatDateTime(value) : '-' }}</template>
                <template #cell-sell_price_cny_cents="{ value }">{{ formatCNY(value) }}</template>
                <template #cell-memo="{ value }">{{ value || '-' }}</template>
                <template #cell-actions="{ row }">
                  <button
                    v-if="row.status === 'issued'"
                    class="btn btn-secondary btn-sm"
                    :disabled="revokeLoadingId === row.id"
                    @click="revokeOrder(row.id)"
                  >
                    {{ revokeLoadingId === row.id ? t('common.processing') : t('admin.distributor.revoke') }}
                  </button>
                  <span v-else class="text-xs text-gray-400">-</span>
                </template>
              </DataTable>
            </section>
          </template>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { Column } from '@/components/common/types'
import Select from '@/components/common/Select.vue'
import { adminAPI } from '@/api/admin'
import type {
  DistributorAdminSummary,
  DistributorOffer,
  DistributorOrder,
  DistributorProfile,
  Group,
  DistributorUserStats
} from '@/types'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const summary = ref<DistributorAdminSummary | null>(null)
const profiles = ref<DistributorProfile[]>([])
const offers = ref<DistributorOffer[]>([])
const orders = ref<DistributorOrder[]>([])
const groups = ref<Group[]>([])

const isRefreshing = ref(false)
const loadingProfiles = ref(false)
const loadingOffers = ref(false)
const loadingOrders = ref(false)
const savingProfile = ref(false)
const creatingOffer = ref(false)
const deletingOfferId = ref<number | null>(null)
const revokeLoadingId = ref<number | null>(null)
const adjustingBalance = ref(false)
const settling = ref(false)

const selectedUserId = ref<number | null>(null)

const profileSearch = ref('')
const orderSearch = ref('')
const orderStatus = ref('')
const settleNotes = ref('')

const createForm = reactive({
  email: '',
  enabled: true,
  notes: '',
  copy_from_user_id: 0
})

const offerForm = reactive({
  name: '',
  target_group_id: 0,
  validity_days: 30,
  cost_cny: 1,
  enabled: true,
  notes: ''
})

const balanceForm = reactive({
  amount: 10,
  notes: ''
})

const summaryRows = computed<DistributorUserStats[]>(() => summary.value?.by_user || [])

const selectedProfile = computed(() =>
  profiles.value.find((item) => item.user_id === selectedUserId.value) || null
)

const selectedUserEmail = computed(() => {
  if (!selectedProfile.value) {
    return '-'
  }
  return getProfileUserEmail(selectedProfile.value)
})

const selectedStats = computed(
  () => summaryRows.value.find((item) => item.distributor_user_id === selectedUserId.value) || null
)

const selectedOrderCount = computed(() => selectedStats.value?.orders_total || 0)
const selectedGrossProfit = computed(() => selectedStats.value?.gross_profit_cny || 0)
const enabledProfileCount = computed(() => profiles.value.filter((item) => item.enabled).length)
const totalProfileCount = computed(() => profiles.value.length)
const totalOrderCount = computed(() =>
  summaryRows.value.reduce((acc, item) => acc + (item.orders_total || 0), 0)
)
const totalGrossProfit = computed(() =>
  summaryRows.value.reduce((acc, item) => acc + (item.gross_profit_cny || 0), 0)
)

const enabledOptions = computed(() => [
  { value: true, label: t('common.enabled') },
  { value: false, label: t('common.disabled') }
])

const copyProfileOptions = computed(() => [
  { value: 0, label: t('admin.distributor.copyFromNone') },
  ...profiles.value.map((item) => ({
    value: item.user_id,
    label: `${getProfileUserEmail(item)} (#${item.user_id})`
  }))
])

const groupOptions = computed(() =>
  groups.value
    .filter((item) => item.subscription_type === 'subscription')
    .map((item) => ({ value: item.id, label: item.name }))
)

const groupNameMap = computed<Record<number, string>>(() => {
  const map: Record<number, string> = {}
  for (const group of groups.value) {
    map[group.id] = group.name
  }
  return map
})

const profileColumns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.distributor.email') },
  { key: 'enabled', label: t('common.status') },
  { key: 'balance_cny_cents', label: t('admin.distributor.balance') },
  { key: 'actions', label: t('common.actions') }
])

const offerColumns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.distributor.offerName') },
  { key: 'target_group_id', label: t('keys.group') },
  { key: 'validity_days', label: t('admin.distributor.validityDays') },
  { key: 'cost_cny_cents', label: t('admin.distributor.cost') },
  { key: 'enabled', label: t('common.status') },
  { key: 'notes', label: t('admin.distributor.notes') },
  { key: 'actions', label: t('common.actions') }
])

const orderColumns = computed<Column[]>(() => [
  { key: 'redeem_code', label: t('admin.distributor.code') },
  { key: 'status', label: t('admin.distributor.status') },
  { key: 'sell_price_cny_cents', label: t('admin.distributor.sellPrice') },
  { key: 'used_email', label: t('admin.distributor.usedByEmail') },
  { key: 'memo', label: t('admin.distributor.notes') },
  { key: 'issued_at', label: t('admin.distributor.issuedAt') },
  { key: 'redeemed_at', label: t('admin.distributor.redeemedAt') },
  { key: 'actions', label: t('common.actions') }
])

const statsColumns = computed<Column[]>(() => [
  { key: 'distributor_user_id', label: t('admin.distributor.email') },
  { key: 'orders_total', label: t('admin.distributor.orders') },
  { key: 'sell_amount_cny', label: t('admin.distributor.sellPrice') },
  { key: 'cost_amount_cny', label: t('admin.distributor.cost') },
  { key: 'refund_amount_cny', label: t('admin.distributor.refund') },
  { key: 'gross_profit_cny', label: t('admin.distributor.grossProfit') }
])

const orderStatusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'issued', label: t('admin.distributor.orderStatus.issued') },
  { value: 'redeemed', label: t('admin.distributor.orderStatus.redeemed') },
  { value: 'revoked', label: t('admin.distributor.orderStatus.revoked') }
])

const userEmailMap = computed<Record<number, string>>(() => {
  const map: Record<number, string> = {}
  for (const item of profiles.value) {
    map[item.user_id] = getProfileUserEmail(item)
  }
  return map
})

const resolvedUserEmailMap = ref<Record<number, string>>({})

const getUserEmailFromObject = (user: any): string => {
  if (!user || typeof user !== 'object') {
    return ''
  }
  return String(user.email ?? user.Email ?? '').trim()
}

const getProfileUserEmail = (profile: DistributorProfile): string => {
  const fromEmbedded = getUserEmailFromObject((profile as any).user)
  if (fromEmbedded) {
    return fromEmbedded
  }
  const fromResolved = resolvedUserEmailMap.value[profile.user_id]
  if (fromResolved) {
    return fromResolved
  }
  return String(profile.user_id)
}

const resolveUserEmail = (userId: number | string): string => {
  const numericID = Number(userId)
  if (!Number.isFinite(numericID) || numericID <= 0) {
    return String(userId)
  }
  return userEmailMap.value[numericID] || resolvedUserEmailMap.value[numericID] || String(userId)
}

const formatCNY = (cents?: number | null) => {
  const value = Number.isFinite(cents as number) ? Number(cents) : 0
  return `CNY ${(value / 100).toFixed(2)}`
}

const getRedeemCode = (order: DistributorOrder): string => {
  const redeem = (order as any)?.redeem_code
  if (!redeem) {
    return ''
  }
  return String(redeem.code ?? redeem.Code ?? '').trim()
}

const orderStatusClass = (status: string) => {
  if (status === 'issued') {
    return 'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-300'
  }
  if (status === 'redeemed') {
    return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  }
  return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-300'
}

const showRequestError = (error: any) => {
  appStore.showError(error?.response?.data?.detail || t('common.unknownError'))
}

const loadSummary = async () => {
  summary.value = await adminAPI.distributors.summary()
  await hydrateUserEmails()
}

const loadProfiles = async () => {
  loadingProfiles.value = true
  try {
    const resp = await adminAPI.distributors.listProfiles(1, 50, profileSearch.value.trim())
    profiles.value = resp.items
    await hydrateUserEmails()
    if (selectedUserId.value && !profiles.value.some((item) => item.user_id === selectedUserId.value)) {
      selectedUserId.value = null
      offers.value = []
      orders.value = []
    }
  } finally {
    loadingProfiles.value = false
  }
}

const loadOffers = async () => {
  if (!selectedUserId.value) {
    offers.value = []
    return
  }
  loadingOffers.value = true
  try {
    offers.value = await adminAPI.distributors.listOffers(selectedUserId.value)
  } finally {
    loadingOffers.value = false
  }
}

const loadOrders = async () => {
  if (!selectedUserId.value) {
    orders.value = []
    return
  }
  loadingOrders.value = true
  try {
    const resp = await adminAPI.distributors.listOrders(1, 100, {
      distributor_user_id: selectedUserId.value,
      status: orderStatus.value,
      search: orderSearch.value.trim()
    })
    orders.value = resp.items
  } finally {
    loadingOrders.value = false
  }
}

const loadGroups = async () => {
  groups.value = await adminAPI.groups.getAll()
  if (!offerForm.target_group_id && groupOptions.value.length > 0) {
    offerForm.target_group_id = Number(groupOptions.value[0].value)
  }
}

const loadAll = async () => {
  isRefreshing.value = true
  try {
    await Promise.all([loadSummary(), loadProfiles()])
    if (selectedUserId.value) {
      await Promise.all([loadOffers(), loadOrders()])
    }
  } catch (error: any) {
    showRequestError(error)
  } finally {
    isRefreshing.value = false
  }
}

const hydrateUserEmails = async () => {
  const ids = new Set<number>()
  for (const item of profiles.value) {
    ids.add(item.user_id)
  }
  for (const item of summaryRows.value) {
    ids.add(Number(item.distributor_user_id))
  }

  const toResolve: number[] = []
  for (const id of ids) {
    if (!id || userEmailMap.value[id] || resolvedUserEmailMap.value[id]) {
      continue
    }
    toResolve.push(id)
  }
  if (toResolve.length === 0) {
    return
  }

  const updates: Record<number, string> = {}
  await Promise.all(
    toResolve.map(async (id) => {
      try {
        const user = await adminAPI.users.getById(id)
        if (user?.email) {
          updates[id] = user.email
        }
      } catch {
        // ignore lookup failures, fallback to id display
      }
    })
  )

  if (Object.keys(updates).length > 0) {
    resolvedUserEmailMap.value = {
      ...resolvedUserEmailMap.value,
      ...updates
    }
  }
}

const copyOffersFromTemplateUser = async (sourceUserID: number, targetUserID: number) => {
  const sourceOffers = await adminAPI.distributors.listOffers(sourceUserID)
  if (sourceOffers.length === 0) {
    return 0
  }
  let copiedCount = 0
  for (const offer of sourceOffers) {
    await adminAPI.distributors.createOffer({
      distributor_user_id: targetUserID,
      name: offer.name,
      target_group_id: offer.target_group_id,
      validity_days: offer.validity_days,
      cost_cny: offer.cost_cny_cents,
      enabled: offer.enabled,
      notes: offer.notes || ''
    })
    copiedCount++
  }
  return copiedCount
}

const upsertProfile = async () => {
  const email = createForm.email.trim()
  if (!email) {
    appStore.showError(t('auth.emailRequired'))
    return
  }

  savingProfile.value = true
  try {
    const templateUserID = Number(createForm.copy_from_user_id || 0)
    const profile = await adminAPI.distributors.upsertProfile(email, createForm.enabled, createForm.notes.trim())
    let copiedCount = 0
    if (templateUserID > 0) {
      if (templateUserID === profile.user_id) {
        appStore.showInfo(t('admin.distributor.copyFromSelfNotAllowed'))
      } else {
        copiedCount = await copyOffersFromTemplateUser(templateUserID, profile.user_id)
      }
    }

    appStore.showSuccess(t('common.success'))
    createForm.email = ''
    createForm.notes = ''
    createForm.copy_from_user_id = 0

    await loadProfiles()
    await loadSummary()
    await selectUser(profile.user_id)
    if (copiedCount > 0) {
      appStore.showSuccess(t('admin.distributor.copiedOffers', { count: copiedCount }))
      if (selectedUserId.value === profile.user_id) {
        await loadOffers()
      }
    }
  } catch (error: any) {
    showRequestError(error)
  } finally {
    savingProfile.value = false
  }
}

const submitBalanceAdjust = async (operation: 'topup' | 'refund') => {
  if (!selectedUserId.value) {
    return
  }

  const amount = Number(balanceForm.amount)
  if (!Number.isFinite(amount) || amount <= 0) {
    appStore.showError(t('admin.distributor.amountPrompt'))
    return
  }

  adjustingBalance.value = true
  try {
    await adminAPI.distributors.adjustBalance(
      selectedUserId.value,
      operation,
      Math.round(amount * 100),
      balanceForm.notes.trim()
    )
    appStore.showSuccess(t('common.success'))
    await Promise.all([loadProfiles(), loadSummary(), loadOrders()])
  } catch (error: any) {
    showRequestError(error)
  } finally {
    adjustingBalance.value = false
  }
}

const selectUser = async (userId: number) => {
  selectedUserId.value = userId
  const selected = profiles.value.find((item) => item.user_id === userId)
  if (selected) {
    createForm.email = getProfileUserEmail(selected)
    createForm.enabled = selected.enabled
    createForm.notes = selected.notes || ''
  }
  try {
    await Promise.all([loadOffers(), loadOrders()])
  } catch (error: any) {
    showRequestError(error)
  }
}

const createOffer = async () => {
  if (!selectedUserId.value) {
    return
  }

  const name = offerForm.name.trim()
  const costCNY = Number(offerForm.cost_cny)
  if (!name) {
    appStore.showError(t('admin.distributor.offerName'))
    return
  }
  if (!offerForm.target_group_id) {
    appStore.showError(t('admin.redeem.groupRequired'))
    return
  }
  if (!Number.isFinite(costCNY) || costCNY <= 0) {
    appStore.showError(t('admin.distributor.cost'))
    return
  }

  creatingOffer.value = true
  try {
    await adminAPI.distributors.createOffer({
      distributor_user_id: selectedUserId.value,
      name,
      target_group_id: offerForm.target_group_id,
      validity_days: offerForm.validity_days,
      cost_cny: Math.round(costCNY * 100),
      enabled: offerForm.enabled,
      notes: offerForm.notes.trim()
    })
    appStore.showSuccess(t('common.success'))
    offerForm.name = ''
    offerForm.notes = ''
    await loadOffers()
  } catch (error: any) {
    showRequestError(error)
  } finally {
    creatingOffer.value = false
  }
}

const deleteOffer = async (id: number) => {
  deletingOfferId.value = id
  try {
    await adminAPI.distributors.deleteOffer(id)
    appStore.showSuccess(t('common.success'))
    await loadOffers()
  } catch (error: any) {
    showRequestError(error)
  } finally {
    deletingOfferId.value = null
  }
}

const revokeOrder = async (id: number) => {
  revokeLoadingId.value = id
  try {
    await adminAPI.distributors.revokeOrder(id)
    appStore.showSuccess(t('common.success'))
    await Promise.all([loadOrders(), loadSummary(), loadProfiles()])
  } catch (error: any) {
    showRequestError(error)
  } finally {
    revokeLoadingId.value = null
  }
}

const markSettled = async () => {
  settling.value = true
  try {
    await adminAPI.distributors.settle(settleNotes.value.trim())
    settleNotes.value = ''
    appStore.showSuccess(t('common.success'))
    await loadSummary()
  } catch (error: any) {
    showRequestError(error)
  } finally {
    settling.value = false
  }
}

let profileSearchTimer: ReturnType<typeof setTimeout> | null = null
let orderSearchTimer: ReturnType<typeof setTimeout> | null = null

const handleProfileSearch = () => {
  if (profileSearchTimer) {
    clearTimeout(profileSearchTimer)
  }
  profileSearchTimer = setTimeout(() => {
    loadProfiles().catch((error: any) => showRequestError(error))
  }, 250)
}

const handleOrderSearch = () => {
  if (orderSearchTimer) {
    clearTimeout(orderSearchTimer)
  }
  orderSearchTimer = setTimeout(() => {
    loadOrders().catch((error: any) => showRequestError(error))
  }, 250)
}

onMounted(async () => {
  try {
    await Promise.all([loadGroups(), loadAll()])
  } catch (error: any) {
    showRequestError(error)
  }
})

onUnmounted(() => {
  if (profileSearchTimer) {
    clearTimeout(profileSearchTimer)
  }
  if (orderSearchTimer) {
    clearTimeout(orderSearchTimer)
  }
})
</script>
