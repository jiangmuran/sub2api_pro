<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <input v-model="search" type="text" class="input max-w-72" :placeholder="t('admin.distributor.searchProfile')" @input="loadProfiles" />
          <button class="btn btn-secondary" @click="loadAll">{{ t('common.refresh') }}</button>
        </div>
      </template>

      <template #table>
        <div class="mb-4 grid grid-cols-1 gap-4 md:grid-cols-3">
          <div class="card p-4">
            <div class="text-sm text-gray-500">{{ t('admin.distributor.unsettled') }}</div>
            <div class="text-2xl font-semibold">{{ formatCNY(summary?.unsettled_cny_cents || 0) }}</div>
            <div class="text-xs text-gray-500">{{ t('admin.distributor.delta') }}: {{ formatCNY(summary?.delta_since_last_settle_cny || 0) }}</div>
            <button class="btn btn-secondary mt-2" @click="markSettled">{{ t('admin.distributor.markSettled') }}</button>
          </div>
        </div>

        <div class="card p-4 mb-4">
          <h3 class="mb-2 text-lg font-semibold">{{ t('admin.distributor.statsByUser') }}</h3>
          <DataTable :columns="statsColumns" :data="summary?.by_user || []">
            <template #cell-distributor_user_id="{ value }">{{ userEmailMap[value] || value }}</template>
            <template #cell-sell_amount_cny="{ value }">{{ formatCNY(value) }}</template>
            <template #cell-cost_amount_cny="{ value }">{{ formatCNY(value) }}</template>
            <template #cell-refund_amount_cny="{ value }">{{ formatCNY(value) }}</template>
            <template #cell-gross_profit_cny="{ value }">{{ formatCNY(value) }}</template>
          </DataTable>
        </div>

        <div class="card p-4 mb-4">
          <div class="grid grid-cols-1 gap-3 md:grid-cols-4">
            <input v-model="createEmail" type="email" class="input" :placeholder="t('admin.distributor.email')" />
            <input v-model="createNotes" type="text" class="input" :placeholder="t('admin.distributor.notes')" />
            <Select v-model="createEnabled" :options="enabledOptions" />
            <button class="btn btn-primary" @click="upsertProfile">{{ t('common.save') }}</button>
          </div>
        </div>

        <DataTable :columns="profileColumns" :data="profiles" :loading="loadingProfiles">
          <template #cell-user="{ row }">{{ row.user?.email || row.user_id }}</template>
          <template #cell-balance_cny_cents="{ value }">{{ formatCNY(value) }}</template>
          <template #cell-actions="{ row }">
            <button class="btn btn-secondary btn-sm mr-2" @click="selectUser(row.user_id)">{{ t('admin.distributor.manage') }}</button>
            <button class="btn btn-secondary btn-sm" @click="adjust(row.user_id, 'topup')">+{{ t('admin.distributor.topup') }}</button>
            <button class="btn btn-secondary btn-sm ml-1" @click="adjust(row.user_id, 'refund')">-{{ t('admin.distributor.refund') }}</button>
          </template>
        </DataTable>

        <div v-if="selectedUserId" class="mt-4 space-y-4">
          <div class="card p-4">
            <h3 class="mb-2 text-lg font-semibold">{{ t('admin.distributor.offers') }}</h3>
            <div class="grid grid-cols-1 gap-2 md:grid-cols-6 mb-3">
              <input v-model="offerForm.name" type="text" class="input" :placeholder="t('admin.distributor.offerName')" />
              <Select v-model="offerForm.target_group_id" :options="groupOptions" />
              <input v-model.number="offerForm.validity_days" type="number" min="1" class="input" />
              <input v-model.number="offerForm.cost_cny" type="number" min="1" class="input" />
              <Select v-model="offerForm.enabled" :options="enabledOptions" />
              <button class="btn btn-primary" @click="createOffer">{{ t('common.create') }}</button>
            </div>
            <DataTable :columns="offerColumns" :data="offers">
              <template #cell-cost_cny_cents="{ value }">{{ formatCNY(value) }}</template>
              <template #cell-actions="{ row }">
                <button class="btn btn-danger btn-sm" @click="deleteOffer(row.id)">{{ t('common.delete') }}</button>
              </template>
            </DataTable>
          </div>

          <div class="card p-4">
            <h3 class="mb-2 text-lg font-semibold">{{ t('admin.distributor.orders') }}</h3>
            <div class="mb-2 flex gap-2">
              <input v-model="orderSearch" type="text" class="input max-w-72" :placeholder="t('admin.distributor.searchOrders')" @input="loadOrders" />
              <Select v-model="orderStatus" :options="orderStatusOptions" class="w-44" @change="loadOrders" />
            </div>
            <DataTable :columns="orderColumns" :data="orders" :loading="loadingOrders">
              <template #cell-redeem_code="{ row }">{{ row.redeem_code?.code }}</template>
              <template #cell-used_email="{ row }">{{ row.redeem_code?.user?.email || '-' }}</template>
              <template #cell-issued_at="{ value }">{{ formatDateTime(value) }}</template>
              <template #cell-redeemed_at="{ value }">{{ value ? formatDateTime(value) : '-' }}</template>
              <template #cell-sell_price_cny_cents="{ value }">{{ formatCNY(value) }}</template>
              <template #cell-actions="{ row }">
                <button v-if="row.status === 'issued'" class="btn btn-secondary btn-sm" @click="revokeOrder(row.id)">{{ t('admin.distributor.revoke') }}</button>
              </template>
            </DataTable>
          </div>
        </div>
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import type { Column } from '@/components/common/types'
import Select from '@/components/common/Select.vue'
import { adminAPI } from '@/api/admin'
import type { DistributorAdminSummary, DistributorOffer, DistributorOrder, DistributorProfile, Group } from '@/types'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const summary = ref<DistributorAdminSummary | null>(null)
const profiles = ref<DistributorProfile[]>([])
const offers = ref<DistributorOffer[]>([])
const orders = ref<DistributorOrder[]>([])
const groups = ref<Group[]>([])
const loadingProfiles = ref(false)
const loadingOrders = ref(false)
const selectedUserId = ref<number | null>(null)

const search = ref('')
const createEmail = ref('')
const createEnabled = ref(true)
const createNotes = ref('')
const orderSearch = ref('')
const orderStatus = ref('')

const offerForm = reactive({
  name: '',
  target_group_id: 0,
  validity_days: 30,
  cost_cny: 100,
  enabled: true
})

const enabledOptions = computed(() => [
  { value: true, label: t('common.enabled') },
  { value: false, label: t('common.disabled') }
])

const groupOptions = computed(() => groups.value.filter(g => g.subscription_type === 'subscription').map(g => ({ value: g.id, label: g.name })))

const profileColumns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.distributor.email') },
  { key: 'enabled', label: t('common.status') },
  { key: 'balance_cny_cents', label: t('admin.distributor.balance') },
  { key: 'actions', label: t('common.actions') }
])

const offerColumns = computed<Column[]>(() => [
  { key: 'name', label: t('admin.distributor.offerName') },
  { key: 'validity_days', label: t('admin.distributor.validityDays') },
  { key: 'cost_cny_cents', label: t('admin.distributor.cost') },
  { key: 'enabled', label: t('common.status') },
  { key: 'actions', label: t('common.actions') }
])

const orderColumns = computed<Column[]>(() => [
  { key: 'redeem_code', label: t('admin.distributor.code') },
  { key: 'status', label: t('admin.distributor.status') },
  { key: 'sell_price_cny_cents', label: t('admin.distributor.sellPrice') },
  { key: 'used_email', label: t('admin.distributor.usedByEmail') },
  { key: 'issued_at', label: t('admin.distributor.issuedAt') },
  { key: 'redeemed_at', label: t('admin.distributor.redeemedAt') },
  { key: 'actions', label: t('common.actions') }
])

const orderStatusOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'issued', label: t('admin.distributor.orderStatus.issued') },
  { value: 'redeemed', label: t('admin.distributor.orderStatus.redeemed') },
  { value: 'revoked', label: t('admin.distributor.orderStatus.revoked') }
])

const statsColumns = computed<Column[]>(() => [
  { key: 'distributor_user_id', label: t('admin.distributor.email') },
  { key: 'orders_total', label: t('admin.distributor.orders') },
  { key: 'sell_amount_cny', label: t('admin.distributor.sellPrice') },
  { key: 'cost_amount_cny', label: t('admin.distributor.cost') },
  { key: 'refund_amount_cny', label: t('admin.distributor.refund') },
  { key: 'gross_profit_cny', label: t('admin.distributor.grossProfit') }
])

const userEmailMap = computed<Record<number, string>>(() => {
  const m: Record<number, string> = {}
  for (const item of profiles.value) {
    m[item.user_id] = item.user?.email || String(item.user_id)
  }
  return m
})

const formatCNY = (cents: number) => `CNY ${(cents / 100).toFixed(2)}`

const loadSummary = async () => { summary.value = await adminAPI.distributors.summary() }
const loadProfiles = async () => {
  loadingProfiles.value = true
  try {
    const resp = await adminAPI.distributors.listProfiles(1, 50, search.value)
    profiles.value = resp.items
  } finally {
    loadingProfiles.value = false
  }
}
const loadOrders = async () => {
  if (!selectedUserId.value) return
  loadingOrders.value = true
  try {
    const resp = await adminAPI.distributors.listOrders(1, 100, {
      distributor_user_id: selectedUserId.value,
      status: orderStatus.value,
      search: orderSearch.value
    })
    orders.value = resp.items
  } finally {
    loadingOrders.value = false
  }
}
const loadOffers = async () => {
  if (!selectedUserId.value) return
  offers.value = await adminAPI.distributors.listOffers(selectedUserId.value)
}
const loadGroups = async () => {
  const resp = await adminAPI.groups.getAll()
  groups.value = resp
}

const loadAll = async () => {
  await Promise.all([loadSummary(), loadProfiles()])
  if (selectedUserId.value) {
    await Promise.all([loadOffers(), loadOrders()])
  }
}

const upsertProfile = async () => {
  await adminAPI.distributors.upsertProfile(createEmail.value, createEnabled.value, createNotes.value)
  appStore.showSuccess(t('common.saved'))
  await loadProfiles()
}

const adjust = async (userId: number, operation: 'topup' | 'refund') => {
  const raw = window.prompt(t('admin.distributor.amountPrompt'))
  const amount = Number(raw)
  if (!Number.isFinite(amount) || amount <= 0) return
  await adminAPI.distributors.adjustBalance(userId, operation, Math.round(amount * 100))
  await Promise.all([loadProfiles(), loadSummary()])
}

const selectUser = async (userId: number) => {
  selectedUserId.value = userId
  await Promise.all([loadOffers(), loadOrders()])
}

const createOffer = async () => {
  if (!selectedUserId.value) return
  await adminAPI.distributors.createOffer({
    distributor_user_id: selectedUserId.value,
    name: offerForm.name,
    target_group_id: offerForm.target_group_id,
    validity_days: offerForm.validity_days,
    cost_cny: offerForm.cost_cny,
    enabled: offerForm.enabled
  })
  await loadOffers()
}

const deleteOffer = async (id: number) => {
  await adminAPI.distributors.deleteOffer(id)
  await loadOffers()
}

const revokeOrder = async (id: number) => {
  await adminAPI.distributors.revokeOrder(id)
  await Promise.all([loadOrders(), loadSummary(), loadProfiles()])
}

const markSettled = async () => {
  await adminAPI.distributors.settle()
  await loadSummary()
}

onMounted(async () => {
  try {
    await Promise.all([loadGroups(), loadAll()])
  } catch (error: any) {
    appStore.showError(error?.response?.data?.detail || t('common.operationFailed'))
  }
})
</script>
