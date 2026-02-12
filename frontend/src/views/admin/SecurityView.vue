<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.security.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.security.description') }}
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <SearchInput v-model="filters.query" :placeholder="t('common.search')" @search="refresh" />
          <DateRangePicker
            v-model:startDate="filters.startDate"
            v-model:endDate="filters.endDate"
            @change="onDateRangeChange"
          />
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div class="card lg:col-span-1">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="flex items-center justify-between">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.security.sessions') }}
              </h2>
              <span class="text-xs text-gray-500 dark:text-gray-400">
                {{ sessions.total }}
              </span>
            </div>
          </div>
          <div class="space-y-3 p-4">
            <div class="grid grid-cols-2 gap-2">
              <input
                v-model="filters.userId"
                type="text"
                class="input"
                :placeholder="t('admin.security.userId')"
                @keydown.enter="refresh"
              />
              <input
                v-model="filters.apiKeyId"
                type="text"
                class="input"
                :placeholder="t('admin.security.apiKeyId')"
                @keydown.enter="refresh"
              />
            </div>
            <div class="grid grid-cols-2 gap-2">
              <input
                v-model="filters.platform"
                type="text"
                class="input"
                :placeholder="t('admin.security.platform')"
                @keydown.enter="refresh"
              />
              <input
                v-model="filters.model"
                type="text"
                class="input"
                :placeholder="t('admin.security.model')"
                @keydown.enter="refresh"
              />
            </div>
            <input
              v-model="filters.sessionId"
              type="text"
              class="input"
              :placeholder="t('admin.security.sessionId')"
              @keydown.enter="refresh"
            />

            <div class="space-y-2">
              <button class="btn btn-secondary btn-sm w-full" @click="refresh">
                {{ t('common.search') }}
              </button>
            </div>

            <div v-if="loadingSessions" class="py-6 text-center text-sm text-gray-500">
              {{ t('common.loading') }}
            </div>

            <div v-else-if="sessions.items.length === 0" class="py-6 text-center text-sm text-gray-500">
              {{ t('admin.security.empty') }}
            </div>

            <div v-else class="space-y-2">
              <button
                v-for="item in sessions.items"
                :key="sessionKey(item)"
                type="button"
                class="w-full rounded-xl border px-4 py-3 text-left transition"
                :class="selectedSessionKey === sessionKey(item)
                  ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-500/10 dark:text-primary-100'
                  : 'border-gray-100 bg-white text-gray-700 hover:border-primary-200 hover:bg-primary-50/40 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-200'
                "
                @click="selectSession(item)"
              >
                <div class="flex items-center justify-between">
                  <div class="text-sm font-semibold">
                    {{ item.session_id }}
                  </div>
                  <span class="text-xs text-gray-400">{{ item.request_count }}</span>
                </div>
                <div class="mt-1 text-xs text-gray-500">
                  {{ t('admin.security.user') }}: {{ item.user_id || '-' }} ·
                  {{ t('admin.security.apiKey') }}: {{ item.api_key_id || '-' }}
                </div>
                <div class="mt-1 text-xs text-gray-500">
                  {{ formatTime(item.last_at) }}
                </div>
                <div v-if="item.message_preview" class="mt-2 line-clamp-1 text-xs text-gray-600">
                  {{ item.message_preview }}
                </div>
              </button>
            </div>
          </div>
        </div>

        <div class="card lg:col-span-2">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="flex items-center justify-between">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.security.chat') }}
              </h2>
              <span v-if="selectedSession" class="text-xs text-gray-500">
                {{ selectedSession.session_id }}
              </span>
            </div>
          </div>
          <div class="space-y-4 p-6">
            <div class="flex items-center justify-between rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-800">
              <div class="flex flex-wrap items-center gap-3">
                <span class="font-semibold text-gray-900 dark:text-gray-100">AI</span>
                <span class="text-xs text-gray-500">{{ t('admin.security.aiHint') }}</span>
              </div>
              <button class="btn btn-secondary btn-sm" :disabled="aiLoading" @click="summarize">
                {{ aiLoading ? t('common.loading') : t('admin.security.aiSummarize') }}
              </button>
            </div>
            <div v-if="aiError" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
              {{ aiError }}
            </div>
            <div v-if="aiSummary" class="space-y-2 rounded-xl border border-gray-100 bg-white px-4 py-3 text-sm text-gray-700 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-100">
              <div class="text-sm font-semibold">{{ t('admin.security.aiSummary') }}</div>
              <div class="whitespace-pre-wrap">{{ aiSummary.summary }}</div>
              <div v-if="aiSummary.risk_level" class="text-xs text-gray-500">
                {{ t('admin.security.aiRisk') }}: {{ aiSummary.risk_level }}
              </div>
              <div v-if="aiSummary.sensitive_findings?.length" class="mt-2">
                <div class="text-xs font-semibold text-gray-500">{{ t('admin.security.aiFindings') }}</div>
                <ul class="mt-1 list-disc space-y-1 pl-5 text-xs text-gray-600">
                  <li v-for="item in aiSummary.sensitive_findings" :key="item">{{ item }}</li>
                </ul>
              </div>
              <div v-if="aiSummary.recommended_actions?.length" class="mt-2">
                <div class="text-xs font-semibold text-gray-500">{{ t('admin.security.aiActions') }}</div>
                <ul class="mt-1 list-disc space-y-1 pl-5 text-xs text-gray-600">
                  <li v-for="item in aiSummary.recommended_actions" :key="item">{{ item }}</li>
                </ul>
              </div>
            </div>
            <div v-if="!selectedSession" class="py-12 text-center text-sm text-gray-500">
              {{ t('admin.security.selectSession') }}
            </div>
            <div v-else-if="loadingMessages" class="py-12 text-center text-sm text-gray-500">
              {{ t('common.loading') }}
            </div>
            <div v-else-if="messages.length === 0" class="py-12 text-center text-sm text-gray-500">
              {{ t('admin.security.emptyMessages') }}
            </div>
            <div v-else class="space-y-6">
              <div v-for="block in messageBlocks" :key="block.key" class="space-y-3">
                <div class="rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 text-xs text-gray-500 dark:border-dark-700 dark:bg-dark-800">
                  <div class="flex flex-wrap items-center justify-between gap-2">
                    <span>{{ formatTime(block.created_at) }}</span>
                    <span>{{ t('admin.security.requestId') }}: {{ block.request_id || '-' }}</span>
                    <span>{{ t('admin.security.status') }}: {{ block.status_code || '-' }}</span>
                  </div>
                  <div class="mt-1 text-[11px] text-gray-400">
                    {{ block.model || '-' }} · {{ block.platform || '-' }}
                  </div>
                </div>
                <div
                  v-for="msg in block.messages"
                  :key="msg.key"
                  class="flex"
                  :class="msg.role === 'assistant' ? 'justify-start' : 'justify-end'"
                >
                  <div
                    class="max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed"
                    :class="msg.role === 'assistant'
                      ? 'bg-gray-100 text-gray-800 dark:bg-dark-800 dark:text-gray-100'
                      : 'bg-primary-600 text-white'
                    "
                  >
                    <div class="text-xs uppercase tracking-wide opacity-60">
                      {{ msg.role }} · {{ msg.source }}
                    </div>
                    <div class="mt-2 whitespace-pre-wrap break-words">
                      {{ msg.content }}
                    </div>
                  </div>
                </div>
              </div>
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
import AppLayout from '@/components/layout/AppLayout.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import { adminAPI } from '@/api'
import type { SecurityChatSession, SecurityChatLog } from '@/api/admin/security'

const { t } = useI18n()

const loadingSessions = ref(false)
const loadingMessages = ref(false)

const sessions = reactive({ items: [] as SecurityChatSession[], total: 0, page: 1, page_size: 20 })
const messages = ref(
  [] as Array<{
    key: string
    role: string
    content: string
    source: string
    created_at: string
    request_id?: string
    status_code?: number
    model?: string
    platform?: string
  }>
)

const selectedSession = ref<SecurityChatSession | null>(null)
const selectedSessionKey = computed(() => (selectedSession.value ? sessionKey(selectedSession.value) : ''))

const aiLoading = ref(false)
const aiError = ref('')
const aiSummary = ref<null | {
  summary: string
  sensitive_findings?: string[]
  risk_level?: string
  recommended_actions?: string[]
}>(null)

const filters = reactive({
  query: '',
  userId: '',
  apiKeyId: '',
  sessionId: '',
  platform: '',
  model: '',
  startDate: '',
  endDate: ''
})

const formatTime = (value: string) => {
  if (!value) return ''
  const date = new Date(value)
  return isNaN(date.getTime()) ? value : date.toLocaleString()
}

const toISO = (dateString: string, endOfDay = false) => {
  if (!dateString) return ''
  const base = new Date(dateString + (endOfDay ? 'T23:59:59' : 'T00:00:00'))
  return base.toISOString()
}

const sessionKey = (item: SecurityChatSession) => {
  return `${item.session_id}:${item.user_id || 0}:${item.api_key_id || 0}`
}

const fetchSessions = async () => {
  loadingSessions.value = true
  try {
    const params: Record<string, any> = {
      page: sessions.page,
      page_size: sessions.page_size
    }
    if (filters.query) params.q = filters.query
    if (filters.userId) params.user_id = filters.userId
    if (filters.apiKeyId) params.api_key_id = filters.apiKeyId
    if (filters.sessionId) params.session_id = filters.sessionId
    if (filters.platform) params.platform = filters.platform
    if (filters.model) params.model = filters.model
    if (filters.startDate) params.start_time = toISO(filters.startDate)
    if (filters.endDate) params.end_time = toISO(filters.endDate, true)

    const data = await adminAPI.security.listSessions(params)
    sessions.items = data.items
    sessions.total = data.total
    sessions.page = data.page
    sessions.page_size = data.page_size
  } finally {
    loadingSessions.value = false
  }
}

const fetchMessages = async (session: SecurityChatSession) => {
  loadingMessages.value = true
  try {
    const params: Record<string, any> = {
      session_id: session.session_id,
      page: 1,
      page_size: 500
    }
    if (session.user_id) params.user_id = session.user_id
    if (session.api_key_id) params.api_key_id = session.api_key_id
    if (filters.startDate) params.start_time = toISO(filters.startDate)
    if (filters.endDate) params.end_time = toISO(filters.endDate, true)

    const data = await adminAPI.security.listMessages(params)
    const flattened: Array<{
      key: string
      role: string
      content: string
      source: string
      created_at: string
      request_id?: string
      status_code?: number
      model?: string
      platform?: string
    }> = []
    data.items.forEach((log: SecurityChatLog) => {
      log.messages.forEach((msg) => {
        flattened.push({
          key: `${log.id}-${msg.index}-${msg.role}`,
          role: msg.role || 'assistant',
          content: msg.content,
          source: msg.source || 'request',
          created_at: log.created_at,
          request_id: log.request_id,
          status_code: log.status_code,
          model: log.model,
          platform: log.platform
        })
      })
    })
    messages.value = flattened
  } finally {
    loadingMessages.value = false
  }
}

const messageBlocks = computed(() => {
  const blocks: Array<{
    key: string
    request_id?: string
    status_code?: number
    model?: string
    platform?: string
    created_at: string
    messages: typeof messages.value
  }> = []
  let currentKey = ''
  messages.value.forEach((msg) => {
    const key = `${msg.request_id || 'req'}-${msg.created_at}`
    if (key !== currentKey) {
      blocks.push({
        key,
        request_id: msg.request_id,
        status_code: msg.status_code,
        model: msg.model,
        platform: msg.platform,
        created_at: msg.created_at,
        messages: [] as any
      })
      currentKey = key
    }
    const target = blocks[blocks.length - 1]
    const last = target.messages[target.messages.length - 1]
    if (last && last.role === msg.role && last.source === msg.source) {
      last.content = `${last.content}\n${msg.content}`
    } else {
      target.messages.push({ ...msg })
    }
  })
  return blocks
})

const selectSession = async (session: SecurityChatSession) => {
  selectedSession.value = session
  messages.value = []
  await fetchMessages(session)
}

const refresh = async () => {
  selectedSession.value = null
  messages.value = []
  aiSummary.value = null
  aiError.value = ''
  await fetchSessions()
}

const onDateRangeChange = () => {
  refresh()
}

const summarize = async () => {
  aiLoading.value = true
  aiError.value = ''
  try {
    const payload: Record<string, any> = {
      start_time: filters.startDate ? toISO(filters.startDate) : undefined,
      end_time: filters.endDate ? toISO(filters.endDate, true) : undefined,
      user_id: filters.userId ? Number(filters.userId) : undefined,
      api_key_id: filters.apiKeyId ? Number(filters.apiKeyId) : undefined,
      session_id: selectedSession.value?.session_id
    }
    const data = await adminAPI.security.summarize(payload)
    aiSummary.value = data
  } catch (err: any) {
    aiError.value = err?.message || t('common.unknownError')
  } finally {
    aiLoading.value = false
  }
}

onMounted(() => {
  refresh()
})
</script>
