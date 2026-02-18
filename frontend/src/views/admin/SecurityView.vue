<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
          {{ t('admin.security.title') }}
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('admin.security.description') }}
        </p>
      </div>

      <div v-if="statsError" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
        {{ statsError }}
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div class="card">
          <div class="px-6 py-4">
            <div class="text-xs uppercase tracking-wide text-gray-500">
              {{ t('admin.security.stats.sessions') }}
            </div>
            <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatNumber(statsValue.session_count) }}
            </div>
          </div>
        </div>
        <div class="card">
          <div class="px-6 py-4">
            <div class="text-xs uppercase tracking-wide text-gray-500">
              {{ t('admin.security.stats.avgSessionsPerDay') }}
            </div>
            <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatNumber(statsValue.avg_sessions_per_day, 1) }}
            </div>
          </div>
        </div>
        <div class="card">
          <div class="px-6 py-4">
            <div class="text-xs uppercase tracking-wide text-gray-500">
              {{ t('admin.security.stats.estimatedSize') }}
            </div>
            <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatBytes(statsValue.estimated_bytes || 0) }}
            </div>
          </div>
        </div>
        <div class="card">
          <div class="px-6 py-4">
            <div class="text-xs uppercase tracking-wide text-gray-500">
              {{ t('admin.security.stats.tableSize') }}
            </div>
            <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">
              {{ formatBytes(statsValue.table_bytes || 0) }}
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.security.platformShare') }}
            </h2>
            <span class="text-xs text-gray-500">
              {{ t('admin.security.platformShareHint') }}
            </span>
          </div>
        </div>
        <div class="grid grid-cols-1 gap-4 p-6 md:grid-cols-3">
          <div
            v-for="item in platformItems"
            :key="item.key"
            class="rounded-xl border border-gray-100 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-900"
          >
            <div class="flex items-center justify-between text-sm">
              <span class="font-semibold text-gray-900 dark:text-gray-100">
                {{ item.label }}
              </span>
              <span class="text-gray-500">
                {{ formatPercent(item.ratio) }}
              </span>
            </div>
            <div class="mt-2 h-2 w-full rounded-full bg-gray-100 dark:bg-dark-800">
              <div
                class="h-2 rounded-full bg-primary-500"
                :style="{ width: item.ratio * 100 + '%' }"
              ></div>
            </div>
            <div class="mt-2 text-xs text-gray-500">
              {{ formatNumber(item.count) }} {{ t('admin.security.requests') }}
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div class="card lg:col-span-2">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.security.filters') }}
            </h2>
          </div>
          <div class="space-y-4 p-6">
            <DateRangePicker
              v-model:startDate="filters.startDate"
              v-model:endDate="filters.endDate"
              @change="onDateRangeChange"
            />
            <div class="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
              <input
                v-model="filters.userId"
                type="text"
                class="input"
                :placeholder="t('admin.security.userId')"
              />
              <input
                v-model="filters.apiKeyId"
                type="text"
                class="input"
                :placeholder="t('admin.security.apiKeyId')"
              />
              <input
                v-model="filters.accountId"
                type="text"
                class="input"
                :placeholder="t('admin.security.accountId')"
              />
              <input
                v-model="filters.groupId"
                type="text"
                class="input"
                :placeholder="t('admin.security.groupId')"
              />
              <input
                v-model="filters.platform"
                type="text"
                class="input"
                :placeholder="t('admin.security.platform')"
              />
              <input
                v-model="filters.model"
                type="text"
                class="input"
                :placeholder="t('admin.security.model')"
              />
              <input
                v-model="filters.sessionId"
                type="text"
                class="input"
                :placeholder="t('admin.security.sessionId')"
              />
            </div>
            <div class="flex flex-wrap gap-2">
              <button class="btn btn-primary btn-sm" type="button" :disabled="loadingStats" @click="refresh">
                {{ loadingStats ? t('common.loading') : t('admin.security.applyFilters') }}
              </button>
              <button class="btn btn-secondary btn-sm" type="button" @click="resetFilters">
                {{ t('common.reset') }}
              </button>
            </div>
          </div>
        </div>
        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">
              {{ t('admin.security.actions') }}
            </h2>
          </div>
          <div class="space-y-3 p-6">
            <button
              class="btn btn-secondary btn-sm w-full"
              type="button"
              :disabled="exportLoading"
              @click="exportLogs"
            >
              {{ exportLoading ? t('common.loading') : t('admin.security.exportTxt') }}
            </button>
            <button
              class="btn btn-danger btn-sm w-full"
              type="button"
              :disabled="deleteLoading"
              @click="deleteLogs"
            >
              {{ deleteLoading ? t('common.loading') : t('admin.security.deleteLogs') }}
            </button>
            <p class="text-xs text-gray-500">
              {{ t('admin.security.deleteHint') }}
            </p>
            <div
              v-if="deleteResult"
              class="rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-xs text-emerald-700"
            >
              {{ t('admin.security.deleteResult', {
                logs: formatNumber(deleteResult.logs_deleted),
                sessions: formatNumber(deleteResult.sessions_deleted)
              }) }}
            </div>
            <div v-if="actionError" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700">
              {{ actionError }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { saveAs } from 'file-saver'
import AppLayout from '@/components/layout/AppLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import { adminAPI } from '@/api'
import { formatBytes } from '@/utils/format'
import type { SecurityChatDeleteResult, SecurityChatStats } from '@/api/admin/security'

const { t } = useI18n()

const loadingStats = ref(false)
const exportLoading = ref(false)
const deleteLoading = ref(false)

const statsError = ref('')
const actionError = ref('')

const stats = ref<SecurityChatStats | null>(null)
const deleteResult = ref<SecurityChatDeleteResult | null>(null)

const filters = reactive({
  startDate: '',
  endDate: '',
  userId: '',
  apiKeyId: '',
  accountId: '',
  groupId: '',
  sessionId: '',
  platform: '',
  model: ''
})

const statsValue = computed(() => {
  return (
    stats.value ?? {
      start_time: '',
      end_time: '',
      request_count: 0,
      session_count: 0,
      avg_requests_per_day: 0,
      avg_sessions_per_day: 0,
      estimated_bytes: 0,
      table_bytes: 0,
      platform_share: {
        opencode: { count: 0, ratio: 0 },
        codex: { count: 0, ratio: 0 },
        other: { count: 0, ratio: 0 }
      },
      platform_share_basis: 'request'
    }
  )
})

const platformItems = computed(() => {
  const share = statsValue.value.platform_share
  return [
    { key: 'opencode', label: t('admin.security.platforms.opencode'), count: share.opencode.count, ratio: share.opencode.ratio },
    { key: 'codex', label: t('admin.security.platforms.codex'), count: share.codex.count, ratio: share.codex.ratio },
    { key: 'other', label: t('admin.security.platforms.other'), count: share.other.count, ratio: share.other.ratio }
  ]
})

const toISO = (dateString: string, endOfDay = false) => {
  if (!dateString) return ''
  const base = new Date(dateString + (endOfDay ? 'T23:59:59' : 'T00:00:00'))
  return base.toISOString()
}

const toNumber = (value: string) => {
  if (!value) return undefined
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) return undefined
  return parsed
}

const buildParams = () => {
  const params: Record<string, any> = {}
  if (filters.startDate) params.start_time = toISO(filters.startDate)
  if (filters.endDate) params.end_time = toISO(filters.endDate, true)
  const userId = toNumber(filters.userId)
  if (userId) params.user_id = userId
  const apiKeyId = toNumber(filters.apiKeyId)
  if (apiKeyId) params.api_key_id = apiKeyId
  const accountId = toNumber(filters.accountId)
  if (accountId) params.account_id = accountId
  const groupId = toNumber(filters.groupId)
  if (groupId) params.group_id = groupId
  if (filters.sessionId.trim()) params.session_id = filters.sessionId.trim()
  if (filters.platform.trim()) params.platform = filters.platform.trim()
  if (filters.model.trim()) params.model = filters.model.trim()
  return params
}

const refresh = async () => {
  loadingStats.value = true
  statsError.value = ''
  try {
    stats.value = await adminAPI.security.getStats(buildParams())
  } catch (err: any) {
    statsError.value = err?.message || t('common.unknownError')
  } finally {
    loadingStats.value = false
  }
}

const onDateRangeChange = () => {
  refresh()
}

const resetFilters = async () => {
  filters.startDate = ''
  filters.endDate = ''
  filters.userId = ''
  filters.apiKeyId = ''
  filters.accountId = ''
  filters.groupId = ''
  filters.sessionId = ''
  filters.platform = ''
  filters.model = ''
  await refresh()
}

const exportLogs = async () => {
  actionError.value = ''
  exportLoading.value = true
  try {
    const blob = await adminAPI.security.exportLogs(buildParams())
    const start = filters.startDate ? filters.startDate.replace(/-/g, '') : new Date().toISOString().slice(0, 10).replace(/-/g, '')
    const end = filters.endDate ? filters.endDate.replace(/-/g, '') : start
    const name = `security_logs_${start}_${end}.txt.zip`
    saveAs(blob, name)
  } catch (err: any) {
    actionError.value = err?.message || t('common.unknownError')
  } finally {
    exportLoading.value = false
  }
}

const deleteLogs = async () => {
  if (!confirm(t('admin.security.deleteConfirm'))) return
  actionError.value = ''
  deleteLoading.value = true
  try {
    const result = await adminAPI.security.deleteLogs(buildParams())
    deleteResult.value = result
    await refresh()
  } catch (err: any) {
    actionError.value = err?.message || t('common.unknownError')
  } finally {
    deleteLoading.value = false
  }
}

const formatNumber = (value: number | null | undefined, digits = 0) => {
  if (value === null || value === undefined || Number.isNaN(value)) return '-'
  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits
  }).format(value)
}

const formatPercent = (ratio: number) => {
  if (!Number.isFinite(ratio)) return '0%'
  return `${(ratio * 100).toFixed(1)}%`
}

onMounted(() => {
  refresh()
})
</script>
