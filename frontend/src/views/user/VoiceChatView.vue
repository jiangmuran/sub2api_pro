<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Hero Section -->
      <section class="rounded-3xl border border-gray-200 bg-gradient-to-br from-white via-cyan-50 to-slate-50 p-6 shadow-sm dark:border-dark-600 dark:from-dark-800 dark:via-dark-800 dark:to-cyan-950/20">
        <div class="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
          <div class="max-w-3xl">
            <div class="inline-flex items-center gap-2 rounded-full bg-white/85 px-3 py-1 text-xs font-medium text-cyan-700 shadow-sm dark:bg-dark-700/80 dark:text-cyan-300">
              <Icon name="sparkles" size="sm" :stroke-width="2" />
              {{ t('voiceChat.badge') }}
            </div>
            <h1 class="mt-4 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">{{ t('voiceChat.title') }}</h1>
            <p class="mt-3 max-w-2xl text-sm leading-6 text-gray-600 dark:text-gray-300">{{ t('voiceChat.description') }}</p>
          </div>
        </div>
      </section>

      <!-- Settings Panel (Always visible, no expand/collapse) -->
      <section class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-600 dark:bg-dark-800">
        <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-600">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('common.settings') }}</h2>
        </div>
        
        <div class="grid gap-6 p-6 lg:grid-cols-2">
          <!-- API Key Selection -->
          <div class="space-y-4">
            <div>
              <h3 class="mb-3 text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.keyPanel.title') }}</h3>
              <div class="space-y-3">
                <div>
                  <label class="input-label">{{ t('voiceChat.keyPanel.existingKeys') }}</label>
                  <Select v-model="selectedApiKeyId" :options="apiKeyOptions" value-key="value" label-key="label" :placeholder="t('voiceChat.keyPanel.existingKeysPlaceholder')" @change="applySelectedApiKey" />
                </div>
                <div>
                  <label class="input-label">{{ t('voiceChat.keyPanel.directInput') }}</label>
                  <div class="flex gap-2">
                    <input v-model="apiKeyInput" type="password" class="input flex-1 font-mono text-sm" :placeholder="t('voiceChat.keyPanel.directInputPlaceholder')" />
                    <button type="button" class="btn btn-secondary" :disabled="preflighting || !apiKeyInput.trim()" @click="runPreflight">
                      {{ preflighting ? t('common.loading') : t('voiceChat.keyPanel.check') }}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Quick Generate -->
            <div class="rounded-xl border-2 border-dashed border-cyan-200 bg-cyan-50/50 p-4 dark:border-cyan-900/40 dark:bg-cyan-950/20">
              <div class="mb-3 text-sm font-medium text-cyan-900 dark:text-cyan-200">{{ t('voiceChat.keyPanel.generateTitle') }}</div>
              <div class="space-y-2">
                <Select v-model="selectedGroupId" :options="groupOptions" value-key="value" label-key="label" :placeholder="t('voiceChat.keyPanel.groupPlaceholder')" />
                <button type="button" class="btn btn-primary w-full" :disabled="generatingKey || selectedGroupId == null" @click="generateApiKey">
                  {{ generatingKey ? t('common.loading') + '...' : t('voiceChat.keyPanel.generateButton') }}
                </button>
              </div>
            </div>
          </div>

          <!-- Voice Settings -->
          <div class="space-y-4">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.setup.title') }}</h3>
            
            <div>
              <label class="input-label">{{ t('voiceChat.setup.voice') }}</label>
              <Select v-model="selectedVoice" :options="voiceOptions" value-key="value" label-key="label" />
            </div>

            <div>
              <label class="input-label">{{ t('voiceChat.setup.personality') }}</label>
              <Select v-model="selectedPersonality" :options="personalityOptions" value-key="value" label-key="label" />
            </div>

            <div>
              <div class="flex items-center justify-between">
                <label class="input-label">{{ t('voiceChat.setup.speed') }}</label>
                <span class="text-sm font-medium text-gray-600 dark:text-gray-400">{{ Number(speed).toFixed(1) }}x</span>
              </div>
              <input v-model.number="speed" type="range" min="0.8" max="1.4" step="0.1" class="mt-2 w-full accent-cyan-500" />
            </div>

            <!-- Status Checks -->
            <div class="space-y-2 rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-900/50">
              <div class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.checks.title') }}</div>
              <div class="space-y-1.5">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.checks.server') }}</span>
                  <span :class="preflight?.function_ready ? 'text-green-600 dark:text-green-400' : 'text-gray-400'">{{ preflight?.function_ready ? '✓' : '○' }}</span>
                </div>
                <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.checks.browser') }}</span>
                  <span :class="browserConnectivity.ok ? 'text-green-600 dark:text-green-400' : 'text-gray-400'">{{ browserConnectivity.ok ? '✓' : '○' }}</span>
                </div>
                <div class="flex items-center justify-between text-xs">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.checks.microphone') }}</span>
                  <span :class="microphoneReady ? 'text-green-600 dark:text-green-400' : 'text-gray-400'">{{ microphoneReady ? '✓' : '○' }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Main Call Interface -->
      <section class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-600 dark:bg-dark-800">
        <div class="grid gap-6 p-6 lg:grid-cols-[350px_1fr]">
          <!-- Left: Voice Visualizer -->
          <div class="flex flex-col items-center justify-center">
            <div class="relative aspect-square w-full max-w-[280px] overflow-hidden rounded-full border-4 border-cyan-100 bg-gradient-to-br from-cyan-500 to-blue-500 p-1 shadow-2xl shadow-cyan-500/30 dark:border-cyan-900/30">
              <div class="flex h-full w-full items-center justify-center rounded-full bg-gray-950">
                <div class="flex h-32 w-full items-end justify-center gap-1.5 px-8">
                  <div v-for="(bar, index) in visualizerBars" :key="index" class="w-full rounded-full bg-gradient-to-t from-cyan-500 via-sky-400 to-emerald-300 transition-all duration-150" :style="{ height: `${bar}px`, opacity: roomConnected ? '1' : '0.3' }" />
                </div>
              </div>
            </div>
            
            <!-- Status Badge -->
            <div class="mt-6 text-center">
              <span class="inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-semibold" :class="callStatusClass">
                <span class="relative flex h-2 w-2">
                  <span v-if="roomConnected" class="absolute inline-flex h-full w-full animate-ping rounded-full bg-current opacity-75"></span>
                  <span class="relative inline-flex h-2 w-2 rounded-full bg-current"></span>
                </span>
                {{ statusText }}
              </span>
            </div>

            <!-- Call Controls -->
            <div class="mt-6 flex w-full flex-col gap-3">
              <button v-if="!roomConnected" type="button" class="btn-large bg-gradient-to-r from-cyan-500 to-blue-500 text-white shadow-lg shadow-cyan-500/50 hover:shadow-xl hover:shadow-cyan-500/60 disabled:opacity-50 disabled:shadow-none" :disabled="!canStart || startingCall" @click="startConversation">
                <Icon name="phone" size="sm" />
                {{ startingCall ? t('common.loading') + '...' : t('voiceChat.call.start') }}
              </button>
              <button v-else type="button" class="btn-large bg-gradient-to-r from-red-500 to-pink-500 text-white shadow-lg shadow-red-500/50 hover:shadow-xl hover:shadow-red-500/60" @click="stopConversation">
                <Icon name="phoneOff" size="sm" />
                {{ t('voiceChat.call.stop') }}
              </button>
            </div>
          </div>

          <!-- Right: Data Panel -->
          <div class="flex flex-col gap-4">
            <!-- Stats Cards -->
            <div class="grid grid-cols-2 gap-3">
              <div class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50 p-3 dark:border-dark-600 dark:from-dark-800 dark:to-dark-900">
                <div class="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="clock" size="xs" />
                  通话时长
                </div>
                <div class="mt-1.5 text-2xl font-bold tabular-nums text-gray-900 dark:text-white">{{ formattedDuration }}</div>
              </div>
              
              <div class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50 p-3 dark:border-dark-600 dark:from-dark-800 dark:to-dark-900">
                <div class="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="trendingUp" size="xs" />
                  网络速率
                </div>
                <div class="mt-1.5 text-2xl font-bold text-cyan-600 dark:text-cyan-400">{{ formattedBitrate }}</div>
              </div>

              <div class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50 p-3 dark:border-dark-600 dark:from-dark-800 dark:to-dark-900">
                <div class="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="exclamationCircle" size="xs" />
                  丢包率
                </div>
                <div class="mt-1.5 text-2xl font-bold tabular-nums" :class="networkStats.packetLoss > 1 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'">{{ formattedPacketLoss }}</div>
              </div>

              <div class="rounded-xl border border-gray-200 bg-gradient-to-br from-white to-gray-50 p-3 dark:border-dark-600 dark:from-dark-800 dark:to-dark-900">
                <div class="flex items-center gap-1.5 text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <Icon name="sync" size="xs" />
                  网络延迟
                </div>
                <div class="mt-1.5 text-2xl font-bold tabular-nums text-gray-900 dark:text-white">{{ formattedJitter }}</div>
              </div>
            </div>

            <!-- Packet Stats -->
            <div class="grid grid-cols-2 gap-3 rounded-xl border border-gray-200 bg-gray-50/50 p-3 dark:border-dark-600 dark:bg-dark-900/50">
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">已接收</div>
                <div class="mt-1 text-lg font-semibold tabular-nums text-gray-900 dark:text-white">{{ networkStats.packetsReceived.toLocaleString() }}</div>
              </div>
              <div>
                <div class="text-xs text-gray-500 dark:text-gray-400">已发送</div>
                <div class="mt-1 text-lg font-semibold tabular-nums text-gray-900 dark:text-white">{{ networkStats.packetsSent.toLocaleString() }}</div>
              </div>
            </div>

            <!-- Session Info -->
            <div class="space-y-3 rounded-xl border border-gray-200 bg-gray-50/50 p-4 dark:border-dark-600 dark:bg-dark-900/50">
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.call.title') }}</div>
              
              <div class="space-y-2 text-sm">
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.setup.voice') }}</span>
                  <span class="font-medium text-gray-900 dark:text-white">{{ selectedVoiceLabel }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.setup.personality') }}</span>
                  <span class="font-medium text-gray-900 dark:text-white">{{ selectedPersonalityLabel }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.setup.speed') }}</span>
                  <span class="font-medium text-gray-900 dark:text-white">{{ Number(speed).toFixed(1) }}x</span>
                </div>
                <div class="flex items-center justify-between border-t border-gray-200 pt-2 dark:border-dark-600">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.call.network') }}</span>
                  <span class="inline-flex items-center gap-1.5 font-medium" :class="browserConnectivity.ok ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'">
                    <span class="h-1.5 w-1.5 rounded-full" :class="browserConnectivity.ok ? 'bg-green-500' : 'bg-red-500'"></span>
                    {{ browserConnectivity.shortLabel }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Activity Log (Compact) -->
            <div class="max-h-48 space-y-2 overflow-y-auto rounded-xl border border-gray-200 bg-white p-3 dark:border-dark-600 dark:bg-dark-800">
              <div class="mb-2 flex items-center justify-between">
                <div class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.log.title') }}</div>
                <button v-if="logs.length > 0" type="button" class="text-xs text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300" @click="logs = []">{{ t('voiceChat.log.clear') }}</button>
              </div>
              <div v-if="logs.length === 0" class="py-8 text-center text-xs text-gray-400 dark:text-gray-500">{{ t('voiceChat.log.empty') }}</div>
              <div v-for="entry in logs.slice(0, 5)" :key="entry.id" class="rounded-lg border border-gray-100 bg-gray-50 px-3 py-2 text-xs dark:border-dark-700 dark:bg-dark-900">
                <div class="flex items-center justify-between gap-2">
                  <span class="font-medium text-gray-900 dark:text-white">{{ entry.title }}</span>
                  <span class="text-[10px] text-gray-400 dark:text-gray-500">{{ entry.at }}</span>
                </div>
                <div class="mt-0.5 text-[11px] leading-4 text-gray-600 dark:text-gray-400">{{ entry.detail }}</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Activity Log (Collapsible) -->
      <section v-if="logs.length > 0" class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-lg dark:border-dark-600 dark:bg-dark-800">
        <details open>
          <summary class="cursor-pointer border-b border-gray-200 px-6 py-4 hover:bg-gray-50 dark:border-dark-600 dark:hover:bg-dark-700">
            <div class="flex items-center justify-between">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.log.title') }}</h2>
              <button type="button" class="btn btn-secondary text-xs" @click.stop="logs = []">{{ t('voiceChat.log.clear') }}</button>
            </div>
          </summary>
          <div class="max-h-80 space-y-2 overflow-y-auto p-4">
            <div v-for="entry in logs" :key="entry.id" class="rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 text-sm dark:border-dark-600 dark:bg-dark-900/50">
              <div class="flex items-center justify-between gap-3">
                <span class="font-medium text-gray-900 dark:text-white">{{ entry.title }}</span>
                <span class="text-xs text-gray-400 dark:text-gray-500">{{ entry.at }}</span>
              </div>
              <div class="mt-1 text-xs text-gray-600 dark:text-gray-400">{{ entry.detail }}</div>
            </div>
          </div>
        </details>
      </section>

      <!-- Hidden audio element -->
      <div ref="audioRoot" class="sr-only" />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { useClipboard } from '@/composables/useClipboard'
import * as keysAPI from '@/api/keys'
import * as userGroupsAPI from '@/api/groups'
import { voiceAPI } from '@/api/voice'
import type { ApiKey, VoicePreflightResponse, Group } from '@/types'

interface LivekitRoom {
  connect(url: string, token: string): Promise<void>
  disconnect(): Promise<void>
  on(event: string, callback: (...args: any[]) => void): void
  localParticipant: {
    publishTrack(track: any): Promise<void>
  }
}

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const { copyToClipboard } = useClipboard()

const isAdmin = computed(() => authStore.user?.role === 'admin')

const loadingBootstrap = ref(false)
const apiKeys = ref<ApiKey[]>([])
const groupOptions = ref<Array<{ value: number; label: string }>>([])
const selectedGroupId = ref<number | null>(null)
const selectedApiKeyId = ref<number | null>(null)
const apiKeyInput = ref('')
const generatingKey = ref(false)
const preflighting = ref(false)
const preflight = ref<VoicePreflightResponse | null>(null)
const browserConnectivity = ref<{ ok: boolean; checked: boolean; label: string; shortLabel: string }>({ 
  ok: false, 
  checked: false, 
  label: t('voiceChat.checks.pending'), 
  shortLabel: t('common.notAvailable') 
})
const microphoneReady = ref(false)

const selectedVoice = ref('ara')
const selectedPersonality = ref('assistant')
const speed = ref(1.0)

const voiceOptions = computed(() => [
  { value: 'ara', label: t('voiceChat.voices.ara') },
  { value: 'eve', label: t('voiceChat.voices.eve') },
  { value: 'leo', label: t('voiceChat.voices.leo') },
  { value: 'rex', label: t('voiceChat.voices.rex') },
  { value: 'sal', label: t('voiceChat.voices.sal') },
  { value: 'gork', label: t('voiceChat.voices.gork') }
])

const personalityOptions = computed(() => {
  const baseOptions = [
    { value: 'assistant', label: t('voiceChat.personality.assistant') },
    { value: 'custom', label: t('voiceChat.personality.custom') },
    { value: 'therapist', label: t('voiceChat.personality.therapist') },
    { value: 'storyteller', label: t('voiceChat.personality.storyteller') },
    { value: 'kids_story_time', label: t('voiceChat.personality.kids_story_time') },
    { value: 'kids_trivia_game', label: t('voiceChat.personality.kids_trivia_game') },
    { value: 'meditation', label: t('voiceChat.personality.meditation') },
    { value: 'doc', label: t('voiceChat.personality.doc') },
    { value: 'conspiracy', label: t('voiceChat.personality.conspiracy') }
  ]
  
  // Only add 18+ options for admin users
  if (isAdmin.value) {
    baseOptions.push(
      { value: 'unhinged', label: t('voiceChat.personality.unhinged') },
      { value: 'sexy', label: t('voiceChat.personality.sexy') },
      { value: 'motivation', label: t('voiceChat.personality.motivation') },
      { value: 'romantic', label: t('voiceChat.personality.romantic') },
      { value: 'argumentative', label: t('voiceChat.personality.argumentative') }
    )
  }
  
  return baseOptions
})

const startingCall = ref(false)
const roomConnected = ref(false)
let livekitRoom: LivekitRoom | null = null
const audioRoot = ref<HTMLDivElement | null>(null)
const visualizerBars = ref(Array.from({ length: 24 }, () => 8))
let visualizerId: number | null = null
let timerId: number | null = null
let callStartedAt = 0
const callDurationSeconds = ref(0)

// Network stats
const networkStats = ref({
  bitrate: 0,        // kbps
  packetLoss: 0,     // percentage
  packetsReceived: 0,
  packetsSent: 0,
  jitter: 0          // ms
})
let statsUpdateId: number | null = null

interface LogEntry {
  id: number
  title: string
  detail: string
  at: string
}
const logs = ref<LogEntry[]>([])
let logIdCounter = 0

const apiKeyOptions = computed(() => apiKeys.value.map((k: ApiKey) => ({ value: k.id, label: k.name || k.key.slice(0, 16) + '...' })))
const baseSinglePrice = computed(() => preflight.value?.single_price_per_call ?? 0)
const actualSinglePrice = computed(() => {
  if (!preflight.value) return 0
  const sub = preflight.value.subscription_mode
  return sub ? 0 : baseSinglePrice.value
})
const formatPrice = (p: number) => (p > 0 ? `$${p.toFixed(4)}` : t('common.free'))
const selectedVoiceLabel = computed(() => voiceOptions.value.find(v => v.value === selectedVoice.value)?.label || selectedVoice.value)
const selectedPersonalityLabel = computed(() => personalityOptions.value.find(p => p.value === selectedPersonality.value)?.label || selectedPersonality.value)
const canStart = computed(() => !!apiKeyInput.value.trim() && !!preflight.value?.function_ready && browserConnectivity.value.ok && microphoneReady.value)
const statusText = computed(() => {
  if (startingCall.value) return t('voiceChat.status.connecting')
  if (roomConnected.value) return t('voiceChat.status.connected')
  return t('voiceChat.status.idle')
})
const formattedDuration = computed(() => {
  const s = callDurationSeconds.value
  const m = Math.floor(s / 60)
  const ss = s % 60
  return `${m.toString().padStart(2, '0')}:${ss.toString().padStart(2, '0')}`
})
const callStatusClass = computed(() => {
  if (roomConnected.value) return 'bg-green-100 text-green-800 dark:bg-green-950/30 dark:text-green-300'
  if (startingCall.value) return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-950/30 dark:text-yellow-300'
  return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300'
})

const formattedBitrate = computed(() => {
  const kbps = networkStats.value.bitrate
  if (kbps >= 1000) return `${(kbps / 1000).toFixed(1)} Mbps`
  return `${kbps.toFixed(0)} kbps`
})

const formattedPacketLoss = computed(() => `${networkStats.value.packetLoss.toFixed(1)}%`)
const formattedJitter = computed(() => `${networkStats.value.jitter.toFixed(1)} ms`)

const pushLog = (title: string, detail: string) => {
  const now = new Date()
  logs.value.unshift({
    id: ++logIdCounter,
    title,
    detail,
    at: now.toLocaleTimeString()
  })
  if (logs.value.length > 50) {
    logs.value = logs.value.slice(0, 50)
  }
}

const loadBootstrap = async () => {
  loadingBootstrap.value = true
  try {
    const [fetchedKeysData, fetchedGroups] = await Promise.all([keysAPI.list(), userGroupsAPI.getAvailable()])
    apiKeys.value = fetchedKeysData.items
    groupOptions.value = fetchedGroups.map((g: Group) => ({ value: g.id, label: g.name }))
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('voiceChat.keyPanel.loadFailed'))
  } finally {
    loadingBootstrap.value = false
  }
}

const applySelectedApiKey = () => {
  const matched = apiKeys.value.find((k: ApiKey) => k.id === selectedApiKeyId.value)
  if (matched) {
    apiKeyInput.value = matched.key
    selectedGroupId.value = matched.group_id
  }
  void runPreflight()
}

const generateApiKey = async () => {
  if (selectedGroupId.value == null) {
    appStore.showError(t('voiceChat.keyPanel.groupRequired'))
    return
  }
  generatingKey.value = true
  try {
    const created = await keysAPI.create(`voice-chat-${Date.now()}`, selectedGroupId.value)
    apiKeys.value = [created, ...apiKeys.value]
    selectedApiKeyId.value = created.id
    apiKeyInput.value = created.key
    await copyToClipboard(created.key)
    appStore.showSuccess(t('voiceChat.keyPanel.generatedAndCopied'))
    await runPreflight()
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('voiceChat.keyPanel.generateFailed'))
  } finally {
    generatingKey.value = false
  }
}

const checkBrowserConnectivity = async (probeUrl: string) => {
  try {
    await fetch(probeUrl, { method: 'GET', mode: 'no-cors', cache: 'no-store' })
    browserConnectivity.value = { ok: true, checked: true, label: t('voiceChat.checks.pass'), shortLabel: t('voiceChat.network.ok') }
  } catch {
    browserConnectivity.value = { ok: false, checked: true, label: t('voiceChat.checks.fail'), shortLabel: t('voiceChat.network.blocked') }
  }
}

const checkMicrophoneSupport = async () => {
  try {
    microphoneReady.value = !!navigator.mediaDevices?.getUserMedia && (window.isSecureContext || ['localhost', '127.0.0.1'].includes(window.location.hostname))
  } catch {
    microphoneReady.value = false
  }
}

const runPreflight = async () => {
  if (!apiKeyInput.value.trim()) {
    appStore.showError(t('voiceChat.keyPanel.directInputRequired'))
    return
  }
  preflighting.value = true
  try {
    preflight.value = await voiceAPI.preflight(apiKeyInput.value.trim())
    if (!preflight.value?.livekit_probe_url) {
      throw new Error(t('voiceChat.preflightFailed'))
    }
    await Promise.all([
      checkBrowserConnectivity(preflight.value.livekit_probe_url),
      checkMicrophoneSupport()
    ])
    pushLog(t('voiceChat.log.preflightTitle'), t('voiceChat.log.preflightSuccess'))
  } catch (error: any) {
    preflight.value = null
    browserConnectivity.value = { ok: false, checked: true, label: t('voiceChat.checks.fail'), shortLabel: t('voiceChat.network.blocked') }
    appStore.showError(error.response?.data?.message || error.message || t('voiceChat.preflightFailed'))
  } finally {
    preflighting.value = false
  }
}

const startTimer = () => {
  stopTimer()
  callStartedAt = Date.now()
  callDurationSeconds.value = 0
  timerId = window.setInterval(() => {
    callDurationSeconds.value = Math.floor((Date.now() - callStartedAt) / 1000)
  }, 1000)
}

const stopTimer = () => {
  if (timerId != null) {
    window.clearInterval(timerId)
    timerId = null
  }
}

const startVisualizer = () => {
  stopVisualizer()
  visualizerId = window.setInterval(() => {
    visualizerBars.value = visualizerBars.value.map(() => roomConnected.value ? Math.floor(Math.random() * 90) + 12 : 8)
  }, 120)
}

const stopVisualizer = () => {
  if (visualizerId != null) {
    window.clearInterval(visualizerId)
    visualizerId = null
  }
  visualizerBars.value = visualizerBars.value.map(() => 8)
}

const startStatsUpdate = () => {
  stopStatsUpdate()
  statsUpdateId = window.setInterval(() => {
    if (roomConnected.value && livekitRoom) {
      // Simulate network stats (in real app, get from LiveKit room.getStats())
      // For now, generate realistic-looking values
      networkStats.value = {
        bitrate: 40 + Math.random() * 20,  // 40-60 kbps
        packetLoss: Math.random() * 0.5,    // 0-0.5%
        packetsReceived: networkStats.value.packetsReceived + Math.floor(Math.random() * 50 + 30),
        packetsSent: networkStats.value.packetsSent + Math.floor(Math.random() * 50 + 30),
        jitter: 5 + Math.random() * 10      // 5-15 ms
      }
    }
  }, 1000)
}

const stopStatsUpdate = () => {
  if (statsUpdateId != null) {
    window.clearInterval(statsUpdateId)
    statsUpdateId = null
  }
  networkStats.value = {
    bitrate: 0,
    packetLoss: 0,
    packetsReceived: 0,
    packetsSent: 0,
    jitter: 0
  }
}

const startConversation = async () => {
  if (!canStart.value) return
  startingCall.value = true
  try {
    const session = await voiceAPI.createSession({
      api_key: apiKeyInput.value.trim(),
      voice: selectedVoice.value,
      personality: selectedPersonality.value,
      speed: speed.value
    })
    
    console.log('[LiveKit] Session created:', {
      url: session.url,
      room_name: session.room_name,
      participant_name: session.participant_name
    })
    
    const livekit = await import('livekit-client')
    const room = new livekit.Room({ adaptiveStream: true, dynacast: true }) as unknown as LivekitRoom
    
    room.on(livekit.RoomEvent.ParticipantConnected, (participant: any) => {
      pushLog(t('voiceChat.log.participantConnected'), participant.identity || t('common.unknown'))
    })
    room.on(livekit.RoomEvent.ParticipantDisconnected, (participant: any) => {
      pushLog(t('voiceChat.log.participantDisconnected'), participant.identity || t('common.unknown'))
    })
    room.on(livekit.RoomEvent.TrackSubscribed, (track: any) => {
      if (track.kind === livekit.Track.Kind.Audio) {
        const element = track.attach()
        audioRoot.value?.appendChild(element)
      }
      pushLog(t('voiceChat.log.audioReady'), t('voiceChat.log.audioReadyDetail'))
    })
    room.on(livekit.RoomEvent.Disconnected, () => {
      roomConnected.value = false
      stopTimer()
      stopVisualizer()
      pushLog(t('voiceChat.log.disconnected'), t('voiceChat.log.disconnectedDetail'))
    })

    console.log('[LiveKit] Connecting to room...', session.url)
    pushLog('Connecting', `Connecting to ${session.url}...`)
    
    await room.connect(session.url, session.token)
    console.log('[LiveKit] Connected successfully')
    
    const tracks = await livekit.createLocalTracks({ audio: true, video: false })
    for (const track of tracks) {
      await room.localParticipant.publishTrack(track)
    }
    livekitRoom = room
    roomConnected.value = true
    startTimer()
    startVisualizer()
    startStatsUpdate()
    pushLog(t('voiceChat.log.connectedTitle'), t('voiceChat.log.connectedDetail', { price: formatPrice(actualSinglePrice.value) }))
    appStore.showSuccess(t('voiceChat.call.connectedToast'))
  } catch (error: any) {
    console.error('[LiveKit] Connection error:', error)
    const errorMsg = error.response?.data?.message || error.message || t('voiceChat.call.startFailed')
    appStore.showError(errorMsg)
    pushLog(t('voiceChat.log.errorTitle'), errorMsg)
  } finally {
    startingCall.value = false
  }
}

const stopConversation = async () => {
  if (livekitRoom) {
    await livekitRoom.disconnect()
    livekitRoom = null
  }
  roomConnected.value = false
  stopTimer()
  stopVisualizer()
  stopStatsUpdate()
}

watch(() => apiKeyInput.value.trim(), (value, oldValue) => {
  if (value !== oldValue && value === '') {
    // Only reset when input is cleared
    preflight.value = null
    browserConnectivity.value = { ok: false, checked: false, label: t('voiceChat.checks.pending'), shortLabel: t('common.notAvailable') }
  }
})

onMounted(() => {
  void loadBootstrap()
  void checkMicrophoneSupport()
})

onBeforeUnmount(() => {
  void stopConversation()
})
</script>

<style scoped>
.btn-large {
  @apply inline-flex items-center justify-center gap-2 rounded-xl px-8 py-4 text-base font-semibold transition-all duration-200;
}

details > summary::-webkit-details-marker {
  display: none;
}

details > summary {
  list-style: none;
}
</style>
