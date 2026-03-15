<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6 py-6">
      <!-- Hero Section -->
      <section class="text-center">
        <div class="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-violet-500 via-purple-500 to-pink-500 shadow-lg shadow-purple-500/30">
          <Icon name="microphone" size="lg" class="text-white" :stroke-width="2.5" />
        </div>
        <h1 class="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">{{ t('voiceChat.title') }}</h1>
        <p class="mt-3 text-lg text-gray-600 dark:text-gray-300">{{ t('voiceChat.description') }}</p>
      </section>

      <!-- Main Call Interface -->
      <section class="overflow-hidden rounded-3xl border border-gray-200 bg-gradient-to-br from-white via-purple-50/30 to-pink-50/30 shadow-xl dark:border-dark-600 dark:from-dark-800 dark:via-purple-950/20 dark:to-pink-950/20">
        <div class="p-8">
          <!-- Voice Visualizer -->
          <div class="mb-8">
            <div class="relative mx-auto aspect-square max-w-xs overflow-hidden rounded-full border-4 border-white bg-gradient-to-br from-violet-500 via-purple-500 to-pink-500 p-1 shadow-2xl shadow-purple-500/40 dark:border-dark-700">
              <div class="flex h-full w-full items-center justify-center rounded-full bg-gray-950">
                <div class="flex h-32 w-full items-end justify-center gap-1.5 px-8">
                  <div v-for="(bar, index) in visualizerBars" :key="index" class="w-full rounded-full bg-gradient-to-t from-violet-500 via-purple-400 to-pink-300 transition-all duration-150" :style="{ height: `${bar}px`, opacity: roomConnected ? '1' : '0.3' }" />
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

            <!-- Duration -->
            <div v-if="roomConnected" class="mt-3 text-center text-2xl font-bold tabular-nums text-gray-900 dark:text-white">
              {{ formattedDuration }}
            </div>
          </div>

          <!-- Voice Settings (Compact) -->
          <div v-if="!roomConnected" class="mb-6 grid gap-3 sm:grid-cols-3">
            <div class="rounded-xl border border-gray-200 bg-white/80 px-3 py-2 text-center dark:border-dark-600 dark:bg-dark-800/80">
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.setup.voice') }}</div>
              <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ selectedVoiceLabel }}</div>
            </div>
            <div class="rounded-xl border border-gray-200 bg-white/80 px-3 py-2 text-center dark:border-dark-600 dark:bg-dark-800/80">
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.setup.personality') }}</div>
              <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ selectedPersonalityLabel }}</div>
            </div>
            <div class="rounded-xl border border-gray-200 bg-white/80 px-3 py-2 text-center dark:border-dark-600 dark:bg-dark-800/80">
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.setup.speed') }}</div>
              <div class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">{{ speed.toFixed(1) }}x</div>
            </div>
          </div>

          <!-- Call Controls -->
          <div class="flex flex-col gap-3 sm:flex-row sm:justify-center">
            <button v-if="!roomConnected" type="button" class="btn-large bg-gradient-to-r from-violet-500 via-purple-500 to-pink-500 text-white shadow-lg shadow-purple-500/50 hover:shadow-xl hover:shadow-purple-500/60 disabled:opacity-50 disabled:shadow-none" :disabled="!canStart || startingCall" @click="startConversation">
              <Icon name="phone" size="sm" />
              {{ startingCall ? t('common.loading') + '...' : t('voiceChat.call.start') }}
            </button>
            <button v-else type="button" class="btn-large bg-gradient-to-r from-red-500 to-pink-500 text-white shadow-lg shadow-red-500/50 hover:shadow-xl hover:shadow-red-500/60" @click="stopConversation">
              <Icon name="phoneOff" size="sm" />
              {{ t('voiceChat.call.stop') }}
            </button>
            <button v-if="!roomConnected" type="button" class="btn-large border-2 border-gray-300 bg-white text-gray-700 hover:border-gray-400 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200 dark:hover:border-dark-500 dark:hover:bg-dark-700" @click="showSettings = !showSettings">
              <Icon name="settings" size="sm" />
              {{ showSettings ? t('common.hide') : t('common.settings') }}
            </button>
          </div>

          <!-- Quick Status Checks (Minimized) -->
          <div v-if="!roomConnected && preflight" class="mt-6 flex items-center justify-center gap-4 text-xs">
            <div class="flex items-center gap-1.5">
              <div class="h-2 w-2 rounded-full" :class="preflight?.function_ready ? 'bg-green-500' : 'bg-red-500'"></div>
              <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.checks.server') }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="h-2 w-2 rounded-full" :class="browserConnectivity.ok ? 'bg-green-500' : 'bg-red-500'"></div>
              <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.checks.browser') }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <div class="h-2 w-2 rounded-full" :class="microphoneReady ? 'bg-green-500' : 'bg-yellow-500'"></div>
              <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.checks.microphone') }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Expandable Settings Panel -->
      <section v-show="showSettings" class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-lg dark:border-dark-600 dark:bg-dark-800">
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
            <div class="rounded-xl border-2 border-dashed border-purple-200 bg-purple-50/50 p-4 dark:border-purple-900/40 dark:bg-purple-950/20">
              <div class="mb-3 text-sm font-medium text-purple-900 dark:text-purple-200">{{ t('voiceChat.keyPanel.generateTitle') }}</div>
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
                <span class="text-sm font-medium text-gray-600 dark:text-gray-400">{{ speed.toFixed(1) }}x</span>
              </div>
              <input v-model="speed" type="range" min="0.8" max="1.4" step="0.1" class="mt-2 w-full accent-purple-500" />
            </div>

            <!-- Pricing (Collapsed by default, small) -->
            <details class="rounded-lg border border-gray-200 bg-gray-50 dark:border-dark-600 dark:bg-dark-900/50">
              <summary class="cursor-pointer px-4 py-2.5 text-sm font-medium text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-800">
                {{ t('voiceChat.price.title') }}
              </summary>
              <div class="border-t border-gray-200 px-4 py-3 dark:border-dark-600">
                <div class="flex items-center justify-between text-sm">
                  <span class="text-gray-600 dark:text-gray-400">{{ t('voiceChat.price.singleActual') }}</span>
                  <span class="font-semibold text-gray-900 dark:text-white">{{ actualPriceLabel }}</span>
                </div>
                <div v-if="effectiveRateLabel" class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ effectiveRateLabel }}</div>
              </div>
            </details>
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
import { i18n } from '@/i18n'
import zhMessages from '@/i18n/locales/zh'
import enMessages from '@/i18n/locales/en'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import * as keysAPI from '@/api/keys'
import * as groupsAPI from '@/api/admin/groups'
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

const t = (key: string, params?: Record<string, unknown> | undefined) => {
  const value = params ? useI18n().t(key, params) : useI18n().t(key)
  if (typeof value === 'string') {
    return value
  }
  return key
}

const ensureVoiceChatMessages = () => {
  const locale = i18n.global.locale.value
  const base = i18n.global.getLocaleMessage(locale)
  if ((base as any)?.voiceChat) {
    return
  }
  const source = locale === 'zh' ? (zhMessages as any) : (enMessages as any)
  const fallback = source.voiceChat || source.modelTest?.voiceChat
  if (fallback) {
    i18n.global.setLocaleMessage(locale, { ...base, voiceChat: fallback })
  }
}
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const showSettings = ref(false)
const loadingBootstrap = ref(false)
const apiKeys = ref<ApiKey[]>([])
const groupOptions = ref<Array<{ value: number; label: string }>>([])
const selectedGroupId = ref<number | null>(null)
const selectedApiKeyId = ref<number | null>(null)
const apiKeyInput = ref('')
const generatingKey = ref(false)
const preflighting = ref(false)
const preflight = ref<VoicePreflightResponse | null>(null)
const browserConnectivity = ref<{ ok: boolean; checked: boolean; label: string; shortLabel: string }>({ ok: false, checked: false, label: t('voiceChat.checks.pending'), shortLabel: t('common.notAvailable') })
const microphoneReady = ref(false)

const selectedVoice = ref('shimmer')
const selectedPersonality = ref('friendly')
const speed = ref(1.0)

const voiceOptions = [
  { value: 'alloy', label: 'Alloy' },
  { value: 'echo', label: 'Echo' },
  { value: 'shimmer', label: 'Shimmer' },
  { value: 'verse', label: 'Verse' }
]

const personalityOptions = [
  { value: 'friendly', label: t('voiceChat.personality.friendly') },
  { value: 'professional', label: t('voiceChat.personality.professional') },
  { value: 'casual', label: t('voiceChat.personality.casual') }
]

const startingCall = ref(false)
const roomConnected = ref(false)
let livekitRoom: LivekitRoom | null = null
const audioRoot = ref<HTMLDivElement | null>(null)
const visualizerBars = ref(Array.from({ length: 24 }, () => 8))
let visualizerId: number | null = null
let timerId: number | null = null
let callStartedAt = 0
const callDurationSeconds = ref(0)

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
const effectiveRateLabel = computed(() => (preflight.value?.subscription_mode ? t('voiceChat.price.subscriptionMode') : ''))
const formatPrice = (p: number) => (p > 0 ? `¥${p.toFixed(4)}` : t('common.free'))
const actualPriceLabel = computed(() => (preflight.value ? formatPrice(actualSinglePrice.value) : '--'))
const selectedVoiceLabel = computed(() => voiceOptions.find((item) => item.value === selectedVoice.value)?.label || selectedVoice.value)
const selectedPersonalityLabel = computed(() => personalityOptions.find((item) => item.value === selectedPersonality.value)?.label || selectedPersonality.value)
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
    const [fetchedKeysData, fetchedGroups] = await Promise.all([keysAPI.list(), groupsAPI.list()])
    apiKeys.value = fetchedKeysData.items
    groupOptions.value = fetchedGroups.items.map((g: Group) => ({ value: g.id, label: g.name }))
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
}

watch(() => apiKeyInput.value.trim(), (value, oldValue) => {
  if (value !== oldValue) {
    preflight.value = null
    browserConnectivity.value = { ok: false, checked: false, label: t('voiceChat.checks.pending'), shortLabel: t('common.notAvailable') }
  }
})

onMounted(() => {
  ensureVoiceChatMessages()
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
