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
              <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                <span>{{ sessions.total }}</span>
                <button class="btn btn-ghost btn-xs" type="button" :disabled="!canBulkDelete" @click="bulkDelete">
                  {{ t('admin.security.bulkDelete') }} {{ bulkDeleteLabel }}
                </button>
              </div>
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
              <label class="flex items-center gap-2 text-xs text-gray-500">
                <input type="checkbox" :checked="selectAllMode" @change="toggleSelectAll" />
                {{ t('admin.security.selectAll') }}
              </label>
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
                  <div class="flex items-center gap-2">
                    <input
                      type="checkbox"
                      :checked="selectAllMode || selectedSessionIds.has(item.session_id)"
                      :disabled="selectAllMode"
                      @click.stop
                      @change="toggleSelect(item.session_id)"
                    />
                    <div class="text-sm font-semibold">
                      {{ item.session_id }}
                    </div>
                  </div>
                  <span class="text-xs text-gray-400">{{ item.request_count }}</span>
                </div>
                <div class="mt-1 text-xs text-gray-500">
                  {{ t('admin.security.user') }}: {{ item.user_email || '-' }} ·
                  {{ t('admin.security.apiKey') }}: {{ item.api_key_id || '-' }}
                </div>
                <div class="mt-1 text-xs text-gray-500">
                  {{ formatTime(item.last_at) }}
                </div>
                <div v-if="item.message_preview" class="mt-2 line-clamp-1 text-xs text-gray-600">
                  {{ item.message_preview }}
                </div>
              </button>
              <Pagination
                v-if="sessions.total > 0"
                :page="sessions.page"
                :total="sessions.total"
                :page-size="sessions.page_size"
                @update:page="handlePageChange"
                @update:pageSize="handlePageSizeChange"
              />
            </div>
          </div>
        </div>

        <div class="card lg:col-span-2">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="flex items-center justify-between">
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.security.chat') }}
              </h2>
              <div v-if="selectedSession" class="flex items-center gap-2 text-xs text-gray-500">
                <span>{{ selectedSession.session_id }}</span>
                <button class="btn btn-ghost btn-xs" type="button" @click="deleteSession">
                  {{ t('admin.security.deleteSession') }}
                </button>
              </div>
            </div>
          </div>
          <div class="space-y-4 p-6">
            <div class="flex items-center justify-between rounded-xl border border-gray-100 bg-gray-50 px-4 py-3 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-800">
              <div class="flex flex-wrap items-center gap-3">
                <span class="font-semibold text-gray-900 dark:text-gray-100">AI</span>
                <span class="text-xs text-gray-500">{{ t('admin.security.aiHint') }}</span>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <select v-model.number="selectedApiKeyId" class="input !h-8 !py-1 text-xs">
                  <option :value="0">{{ t('admin.security.selectApiKey') }}</option>
                  <option v-for="item in apiKeys" :key="item.id" :value="item.id">
                    {{ item.name }} (#{{ item.id }})
                  </option>
                </select>
                <button class="btn btn-secondary btn-sm" :disabled="aiLoading || !selectedApiKeyId" @click="summarize">
                {{ aiLoading ? t('common.loading') : t('admin.security.aiSummarize') }}
                </button>
              </div>
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
            <div v-if="aiSummary" class="rounded-xl border border-gray-100 bg-white px-4 py-3 dark:border-dark-700 dark:bg-dark-900">
              <div class="flex items-center justify-between">
                <div class="text-sm font-semibold text-gray-900 dark:text-gray-100">{{ t('admin.security.aiChat') }}</div>
                <button class="btn btn-ghost btn-xs" type="button" @click="clearAiChat">
                  {{ t('admin.security.aiClear') }}
                </button>
              </div>
              <div class="mt-3 space-y-3">
                <div v-if="aiChatMessages.length === 0" class="text-xs text-gray-500">
                  {{ t('admin.security.aiChatHint') }}
                </div>
                <div v-for="msg in aiChatMessages" :key="msg.key" class="flex" :class="msg.role === 'assistant' ? 'justify-start' : 'justify-end'">
                  <div class="max-w-[80%] rounded-2xl px-4 py-2 text-xs" :class="msg.role === 'assistant'
                    ? 'bg-gray-100 text-gray-800 dark:bg-dark-800 dark:text-gray-100'
                    : 'bg-primary-600 text-white'">
                    <div class="whitespace-pre-wrap break-words">{{ msg.content }}</div>
                  </div>
                </div>
              </div>
              <div class="mt-3 flex items-center gap-2">
                <input v-model="aiChatInput" class="input flex-1" :placeholder="t('admin.security.aiChatInput')" @keydown.enter.prevent="sendAiChat" />
                <button class="btn btn-primary btn-sm" :disabled="aiChatLoading || !aiChatInput || !selectedApiKeyId" @click="sendAiChat">
                  {{ aiChatLoading ? t('common.loading') : t('admin.security.aiChatSend') }}
                </button>
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
                    {{ formatMessageContent(msg) }}
                  </div>
                  <button v-if="isMessageLong(msg)" class="mt-2 text-[11px] text-blue-500" type="button" @click="toggleMessage(msg.key)">
                    {{ isMessageExpanded(msg.key) ? t('admin.security.collapse') : t('admin.security.expand') }}
                  </button>
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
import Pagination from '@/components/common/Pagination.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import { adminAPI } from '@/api'
import type { SecurityChatSession, SecurityChatLog, SecurityApiKey } from '@/api/admin/security'

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
const selectedSessionIds = ref<Set<string>>(new Set())
const selectAllMode = ref(false)

const canBulkDelete = computed(() => selectAllMode.value || selectedSessionIds.value.size > 0)
const bulkDeleteLabel = computed(() => {
  if (selectAllMode.value) return `(${t('admin.security.allResults')})`
  if (selectedSessionIds.value.size > 0) return `(${selectedSessionIds.value.size})`
  return ''
})


const apiKeys = ref<SecurityApiKey[]>([])
const selectedApiKeyId = ref<number>(0)

const aiLoading = ref(false)
const aiError = ref('')
const aiSummary = ref<null | {
  summary: string
  sensitive_findings?: string[]
  risk_level?: string
  recommended_actions?: string[]
}>(null)

const aiChatMessages = ref<Array<{ key: string; role: string; content: string }>>([])
const aiChatInput = ref('')
const aiChatLoading = ref(false)

const expandedMessages = ref<Set<string>>(new Set())

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

const handlePageChange = async (page: number) => {
  sessions.page = page
  await fetchSessions()
}

const handlePageSizeChange = async (size: number) => {
  sessions.page_size = size
  sessions.page = 1
  await fetchSessions()
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
  if (session.api_key_id) {
    selectedApiKeyId.value = session.api_key_id
  }
  messages.value = []
  await fetchMessages(session)
}

const refresh = async () => {
  selectedSession.value = null
  messages.value = []
  aiSummary.value = null
  aiError.value = ''
  aiChatMessages.value = []
  selectedSessionIds.value = new Set()
  selectAllMode.value = false
  await fetchSessions()
}

const onDateRangeChange = () => {
  refresh()
}

const fetchApiKeys = async () => {
  try {
    const list = await adminAPI.security.listApiKeys()
    apiKeys.value = list
    if (!selectedApiKeyId.value && list.length > 0) {
      selectedApiKeyId.value = list[0].id
    }
  } catch (err) {
    apiKeys.value = []
  }
}

const buildAiContext = () => {
  if (!selectedSession.value) return ''
  const slices = messageBlocks.value.slice(0, 6)
  const lines: string[] = []
  slices.forEach((block) => {
    lines.push(`[${block.created_at}] ${block.model || '-'} ${block.platform || '-'}`)
    block.messages.slice(0, 6).forEach((msg) => {
      const content = msg.content.length > 500 ? msg.content.slice(0, 500) + '...' : msg.content
      lines.push(`${msg.role}: ${content}`)
    })
  })
  return lines.join('\n')
}

const summarize = async () => {
  aiLoading.value = true
  aiError.value = ''
  try {
    const payload: Record<string, any> = {
      start_time: selectedSession.value ? undefined : (filters.startDate ? toISO(filters.startDate) : undefined),
      end_time: selectedSession.value ? undefined : (filters.endDate ? toISO(filters.endDate, true) : undefined),
      user_id: filters.userId ? Number(filters.userId) : undefined,
      session_id: selectedSession.value?.session_id,
      api_key_id: selectedSession.value?.api_key_id || selectedApiKeyId.value || undefined
    }
    const data = await adminAPI.security.summarize(payload)
    aiSummary.value = data
    aiChatMessages.value = data.summary
      ? [{ key: `ai-${Date.now()}`, role: 'assistant', content: data.summary }]
      : []
  } catch (err: any) {
    aiError.value = err?.message || t('common.unknownError')
  } finally {
    aiLoading.value = false
  }
}

const sendAiChat = async () => {
  if (!aiChatInput.value) return
  aiChatLoading.value = true
  aiError.value = ''
  const userMessage = aiChatInput.value
  aiChatInput.value = ''
  aiChatMessages.value.push({ key: `user-${Date.now()}`, role: 'user', content: userMessage })
  try {
    const payload = {
      api_key_id: selectedApiKeyId.value || undefined,
      context: buildAiContext(),
      messages: aiChatMessages.value.map((msg) => ({ role: msg.role, content: msg.content }))
    }
    const data = await adminAPI.security.chatWithAI(payload)
    aiChatMessages.value.push({ key: `ai-${Date.now()}`, role: 'assistant', content: data.summary })
  } catch (err: any) {
    aiError.value = err?.message || t('common.unknownError')
  } finally {
    aiChatLoading.value = false
  }
}

const clearAiChat = () => {
  aiChatMessages.value = []
}

const isMessageLong = (msg: { content: string }) => {
  return msg.content.split('\n').length > 10
}

const isMessageExpanded = (key: string) => expandedMessages.value.has(key)

const toggleMessage = (key: string) => {
  if (expandedMessages.value.has(key)) {
    expandedMessages.value.delete(key)
  } else {
    expandedMessages.value.add(key)
  }
}

const formatMessageContent = (msg: { key: string; content: string }) => {
  if (!isMessageLong(msg) || isMessageExpanded(msg.key)) return msg.content
  const lines = msg.content.split('\n')
  const head = lines.slice(0, 5)
  const tail = lines.slice(-5)
  return [...head, '... ...', ...tail].join('\n')
}

const deleteSession = async () => {
  if (!selectedSession.value) return
  if (!confirm(t('admin.security.deleteConfirm'))) return
  await adminAPI.security.deleteSession(selectedSession.value.session_id, {
    user_id: selectedSession.value.user_id,
    api_key_id: selectedSession.value.api_key_id
  })
  await refresh()
}

const toggleSelect = (sessionId: string) => {
  if (selectedSessionIds.value.has(sessionId)) {
    selectedSessionIds.value.delete(sessionId)
  } else {
    selectedSessionIds.value.add(sessionId)
  }
}

const toggleSelectAll = () => {
  if (selectAllMode.value) {
    selectAllMode.value = false
    return
  }
  selectAllMode.value = true
  selectedSessionIds.value = new Set()
}

const bulkDelete = async () => {
  if (!canBulkDelete.value) return
  if (!confirm(t('admin.security.bulkDeleteConfirm'))) return
  if (selectAllMode.value) {
    await adminAPI.security.bulkDeleteSessions({
      select_all: true,
      start_time: filters.startDate ? toISO(filters.startDate) : undefined,
      end_time: filters.endDate ? toISO(filters.endDate, true) : undefined,
      user_id: filters.userId ? Number(filters.userId) : undefined,
      api_key_id: filters.apiKeyId ? Number(filters.apiKeyId) : undefined,
      platform: filters.platform || undefined,
      model: filters.model || undefined,
      q: filters.query || undefined
    })
  } else {
    await adminAPI.security.bulkDeleteSessions({
      session_ids: Array.from(selectedSessionIds.value)
    })
  }
  await refresh()
}

onMounted(() => {
  refresh()
  fetchApiKeys()
})
</script>
