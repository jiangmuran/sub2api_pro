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
            <p v-if="exportLoading && exportProgress.total" class="text-xs text-gray-500">
              {{ t('admin.security.exportProgress', {
                current: formatNumber(exportProgress.current),
                total: formatNumber(exportProgress.total)
              }) }}
            </p>
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
import JSZip from 'jszip'
import AppLayout from '@/components/layout/AppLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import { adminAPI } from '@/api'
import { formatBytes } from '@/utils/format'
import type { SecurityChatDeleteResult, SecurityChatLog, SecurityChatStats } from '@/api/admin/security'

const { t } = useI18n()

const loadingStats = ref(false)
const exportLoading = ref(false)
const deleteLoading = ref(false)

const statsError = ref('')
const actionError = ref('')

const stats = ref<SecurityChatStats | null>(null)
const deleteResult = ref<SecurityChatDeleteResult | null>(null)
const exportProgress = reactive({ current: 0, total: 0 })

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

const sanitizeZipSegment = (value: string, fallback: string) => {
  const trimmed = value.trim()
  if (!trimmed) return fallback
  return trimmed
    .replace(/[\\/]/g, '_')
    .replace(/\.\./g, '_')
    .replace(/[:*?"<>|]/g, '_')
    .replace(/[\u0000-\u001F]/g, '_')
}

const resolveExportFolder = (log: SecurityChatLog) => {
  if (log.user_email?.trim()) {
    return sanitizeZipSegment(log.user_email, 'unknown')
  }
  if (log.user_id) {
    return `user_${log.user_id}`
  }
  return 'unknown'
}

const formatFileTimestamp = (value: string) => {
  const date = new Date(value)
  const safeDate = Number.isNaN(date.getTime()) ? new Date() : date
  const yyyy = safeDate.getUTCFullYear()
  const mm = String(safeDate.getUTCMonth() + 1).padStart(2, '0')
  const dd = String(safeDate.getUTCDate()).padStart(2, '0')
  const hh = String(safeDate.getUTCHours()).padStart(2, '0')
  const min = String(safeDate.getUTCMinutes()).padStart(2, '0')
  const ss = String(safeDate.getUTCSeconds()).padStart(2, '0')
  return `${yyyy}${mm}${dd}_${hh}${min}${ss}`
}

const buildExportFileName = (log: SecurityChatLog) => {
  const ts = formatFileTimestamp(log.created_at)
  return `${ts}_${log.id}.txt`
}

const formatLogTimestamp = (value: string) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toISOString()
}

const buildLogText = (log: SecurityChatLog) => {
  const lines: string[] = []
  lines.push('----')
  lines.push(`id: ${log.id}`)
  lines.push(`created_at: ${formatLogTimestamp(log.created_at)}`)
  lines.push(`session_id: ${log.session_id || ''}`)
  lines.push(`request_id: ${log.request_id || ''}`)
  lines.push(`client_request_id: ${log.client_request_id || ''}`)
  lines.push(`user_id: ${log.user_id ?? ''}`)
  lines.push(`user_email: ${log.user_email || ''}`)
  lines.push(`api_key_id: ${log.api_key_id ?? ''}`)
  lines.push(`account_id: ${log.account_id ?? ''}`)
  lines.push(`group_id: ${log.group_id ?? ''}`)
  lines.push(`platform: ${log.platform || ''}`)
  lines.push(`model: ${log.model || ''}`)
  lines.push(`request_path: ${log.request_path || ''}`)
  lines.push(`status_code: ${log.status_code ?? ''}`)
  lines.push(`stream: ${log.stream ? 'true' : 'false'}`)
  lines.push('messages:')
  const messages = Array.isArray(log.messages) ? log.messages : []
  messages.forEach((msg, index) => {
    const labelIndex = Number.isFinite(msg.index) ? msg.index : index
    lines.push(`[${labelIndex}][${msg.source || ''}][${msg.role || ''}]`)
    if (msg.content) {
      lines.push(msg.content)
    }
    lines.push('')
  })
  lines.push('')
  return lines.join('\n')
}

const exportLogs = async () => {
  actionError.value = ''
  exportLoading.value = true
  exportProgress.current = 0
  exportProgress.total = 0
  try {
    const zip = new JSZip()
    const baseParams = buildParams()
    const pageSize = 200
    const maxRows = 200000
    let page = 1
    let total = 0
    while (true) {
      const res = await adminAPI.security.listLogs({ ...baseParams, page, page_size: pageSize })
      if (page === 1) {
        total = res.total || 0
        exportProgress.total = total
      }
      if (!res.items?.length) break
      for (const log of res.items) {
        if (exportProgress.current >= maxRows) break
        const folder = resolveExportFolder(log)
        const fileName = buildExportFileName(log)
        zip.file(`${folder}/${fileName}`, buildLogText(log))
        exportProgress.current += 1
      }
      if (exportProgress.current >= maxRows || res.items.length < pageSize) break
      page += 1
    }

    if (exportProgress.current === 0) {
      actionError.value = t('admin.security.exportEmpty')
      return
    }

    const start = filters.startDate ? filters.startDate.replace(/-/g, '') : new Date().toISOString().slice(0, 10).replace(/-/g, '')
    const end = filters.endDate ? filters.endDate.replace(/-/g, '') : start
    const name = `security_logs_${start}_${end}.zip`
    const blob = await zip.generateAsync({ type: 'blob', compression: 'DEFLATE', compressionOptions: { level: 6 } })
    saveAs(blob, name)
  } catch (err: any) {
    actionError.value = err?.message || t('common.unknownError')
  } finally {
    exportLoading.value = false
    exportProgress.current = 0
    exportProgress.total = 0
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
