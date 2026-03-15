<template>
  <AppLayout>
    <div class="space-y-6">
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

          <div class="grid gap-3 sm:grid-cols-3 xl:w-[360px] xl:grid-cols-1">
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.summary.keyCount') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ apiKeys.length }}</div>
            </div>
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.summary.price') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ actualPriceLabel }}</div>
            </div>
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.summary.status') }}</div>
              <div class="mt-2 text-sm font-semibold text-gray-900 dark:text-white">{{ statusText }}</div>
            </div>
          </div>
        </div>
      </section>

      <div class="grid gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
        <aside class="space-y-6">
          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.keyPanel.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.keyPanel.description') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" :disabled="loadingBootstrap" @click="loadBootstrap">{{ t('common.refresh') }}</button>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="input-label">{{ t('voiceChat.keyPanel.existingKeys') }}</label>
                <Select v-model="selectedApiKeyId" :options="apiKeyOptions" value-key="value" label-key="label" :placeholder="t('voiceChat.keyPanel.existingKeysPlaceholder')" @change="applySelectedApiKey" />
              </div>

              <div>
                <label class="input-label">{{ t('voiceChat.keyPanel.directInput') }}</label>
                <div class="flex gap-2">
                  <input v-model="apiKeyInput" type="password" class="input flex-1 font-mono" :placeholder="t('voiceChat.keyPanel.directInputPlaceholder')" />
                  <button type="button" class="btn btn-secondary" :disabled="preflighting || !apiKeyInput.trim()" @click="runPreflight">{{ t('voiceChat.keyPanel.check') }}</button>
                </div>
              </div>

              <div class="rounded-xl border border-cyan-200 bg-cyan-50/70 p-4 dark:border-cyan-900/40 dark:bg-cyan-950/20">
                <div class="text-sm font-medium text-cyan-900 dark:text-cyan-200">{{ t('voiceChat.keyPanel.generateTitle') }}</div>
                <div class="mt-1 text-xs text-cyan-700 dark:text-cyan-300">{{ t('voiceChat.keyPanel.generateHint') }}</div>

                <div class="mt-4 space-y-3">
                  <div>
                    <label class="input-label">{{ t('voiceChat.keyPanel.groupLabel') }}</label>
                    <Select v-model="selectedGroupId" :options="groupOptions" value-key="value" label-key="label" :placeholder="t('voiceChat.keyPanel.groupPlaceholder')" />
                  </div>
                  <button type="button" class="btn btn-primary w-full" :disabled="generatingKey || selectedGroupId == null" @click="generateApiKey">
                    {{ generatingKey ? t('common.loading') + '...' : t('voiceChat.keyPanel.generateButton') }}
                  </button>
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div>
              <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.setup.title') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.setup.description') }}</p>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="input-label">{{ t('voiceChat.setup.model') }}</label>
                <div class="rounded-xl border border-gray-200 bg-gray-50 px-4 py-3 text-sm font-medium text-gray-900 dark:border-dark-600 dark:bg-dark-900/40 dark:text-white">grok-livechat</div>
              </div>

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
                  <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ speed.toFixed(1) }}x</span>
                </div>
                <input v-model="speed" type="range" min="0.8" max="1.4" step="0.1" class="mt-2 w-full accent-cyan-500" />
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.price.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.price.description') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" :disabled="preflighting || !apiKeyInput.trim()" @click="runPreflight">{{ preflighting ? t('common.loading') + '...' : t('voiceChat.price.refresh') }}</button>
            </div>

            <div class="mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
              <div class="rounded-xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-900/30">
                <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.price.singleBase') }}</div>
                <div class="mt-2 text-xl font-semibold text-gray-900 dark:text-white">{{ basePriceLabel }}</div>
              </div>
              <div class="rounded-xl border border-cyan-200 bg-cyan-50 p-4 dark:border-cyan-900/40 dark:bg-cyan-950/20">
                <div class="text-xs uppercase tracking-wide text-cyan-700 dark:text-cyan-300">{{ t('voiceChat.price.singleActual') }}</div>
                <div class="mt-2 text-xl font-semibold text-cyan-900 dark:text-cyan-100">{{ actualPriceLabel }}</div>
                <div class="mt-1 text-xs text-cyan-700/80 dark:text-cyan-200/80">{{ effectiveRateLabel }}</div>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div>
              <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.checks.title') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.checks.description') }}</p>
            </div>

            <div class="mt-4 space-y-3">
              <div class="flex items-center justify-between rounded-xl border border-gray-200 px-4 py-3 dark:border-dark-600">
                <div>
                  <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('voiceChat.checks.server') }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.checks.serverHint') }}</div>
                </div>
                <span class="text-xs font-semibold" :class="statusClass(preflight?.function_ready && preflight?.server_livekit_ready)">{{ statusLabel(preflight?.function_ready && preflight?.server_livekit_ready) }}</span>
              </div>
              <div class="flex items-center justify-between rounded-xl border border-gray-200 px-4 py-3 dark:border-dark-600">
                <div>
                  <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('voiceChat.checks.browser') }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.checks.browserHint') }}</div>
                </div>
                <span class="text-xs font-semibold" :class="statusClass(browserConnectivity.ok)">{{ browserConnectivity.label }}</span>
              </div>
              <div class="flex items-center justify-between rounded-xl border border-gray-200 px-4 py-3 dark:border-dark-600">
                <div>
                  <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('voiceChat.checks.microphone') }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.checks.microphoneHint') }}</div>
                </div>
                <span class="text-xs font-semibold" :class="statusClass(microphoneReady)">{{ statusLabel(microphoneReady) }}</span>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-amber-200 bg-amber-50 p-5 shadow-sm dark:border-amber-900/40 dark:bg-amber-950/20">
            <div class="flex items-start gap-3">
              <Icon name="infoCircle" size="sm" class="mt-0.5 text-amber-600 dark:text-amber-300" />
              <div>
                <h2 class="text-sm font-semibold text-amber-900 dark:text-amber-200">{{ t('voiceChat.disclaimer.title') }}</h2>
                <p class="mt-2 text-xs leading-6 text-amber-800 dark:text-amber-200/90">{{ t('voiceChat.disclaimer.body') }}</p>
              </div>
            </div>
          </section>
        </aside>

        <section class="space-y-6">
          <div class="overflow-hidden rounded-3xl border border-gray-200 bg-white shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="border-b border-gray-200 px-5 py-4 dark:border-dark-600">
              <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
                <div>
                  <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.call.title') }}</h2>
                  <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('voiceChat.call.description') }}</p>
                </div>
                <div class="flex flex-wrap items-center gap-3">
                  <span class="rounded-full px-3 py-1 text-xs font-semibold" :class="callStatusClass">{{ statusText }}</span>
                  <span class="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">{{ formattedDuration }}</span>
                </div>
              </div>
            </div>

            <div class="p-5">
              <div class="rounded-3xl border border-cyan-100 bg-gradient-to-br from-cyan-50 via-white to-slate-50 p-6 dark:border-cyan-900/30 dark:from-cyan-950/20 dark:via-dark-800 dark:to-dark-800">
                <div class="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
                  <div>
                    <div class="text-sm font-medium text-gray-900 dark:text-white">{{ selectedVoiceLabel }}</div>
                    <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ selectedPersonalityLabel }} · {{ speed.toFixed(1) }}x</div>
                  </div>
                  <div class="flex items-center gap-3">
                    <button type="button" class="btn btn-primary min-w-[140px]" :disabled="!canStart || startingCall" @click="startConversation">
                      {{ startingCall ? t('common.loading') + '...' : t('voiceChat.call.start') }}
                    </button>
                    <button type="button" class="btn btn-secondary min-w-[140px]" :disabled="!roomConnected && !startingCall" @click="stopConversation">
                      {{ t('voiceChat.call.stop') }}
                    </button>
                  </div>
                </div>

                <div class="mt-6">
                  <div class="flex h-40 items-end gap-1 rounded-2xl bg-gray-950/95 px-4 py-4">
                    <div v-for="(bar, index) in visualizerBars" :key="index" class="w-full rounded-full bg-gradient-to-t from-cyan-500 via-sky-400 to-emerald-300 transition-all duration-150" :style="{ height: `${bar}px`, opacity: roomConnected ? '1' : '0.35' }" />
                  </div>
                </div>

                <div class="mt-4 grid gap-3 md:grid-cols-3">
                  <div class="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-900/30">
                    <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.call.charge') }}</div>
                    <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ actualPriceLabel }}</div>
                  </div>
                  <div class="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-900/30">
                    <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.call.timer') }}</div>
                    <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ formattedDuration }}</div>
                  </div>
                  <div class="rounded-2xl border border-white/70 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-900/30">
                    <div class="text-xs uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('voiceChat.call.network') }}</div>
                    <div class="mt-2 text-lg font-semibold text-gray-900 dark:text-white">{{ browserConnectivity.shortLabel }}</div>
                  </div>
                </div>
              </div>

              <div class="mt-5 rounded-2xl border border-gray-200 bg-gray-50/80 p-4 dark:border-dark-600 dark:bg-dark-900/30">
                <div class="flex items-center justify-between gap-3">
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('voiceChat.log.title') }}</h3>
                  <button type="button" class="btn btn-secondary" @click="logs = []">{{ t('voiceChat.log.clear') }}</button>
                </div>
                <div class="mt-4 min-h-[220px] max-h-[360px] space-y-2 overflow-y-auto">
                  <div v-if="logs.length === 0" class="flex h-[180px] items-center justify-center text-sm text-gray-500 dark:text-gray-400">{{ t('voiceChat.log.empty') }}</div>
                  <div v-for="entry in logs" :key="entry.id" class="rounded-xl border border-gray-200 bg-white px-4 py-3 text-sm shadow-sm dark:border-dark-600 dark:bg-dark-800">
                    <div class="flex items-center justify-between gap-3">
                      <span class="font-medium text-gray-900 dark:text-white">{{ entry.title }}</span>
                      <span class="text-[11px] text-gray-400 dark:text-gray-500">{{ entry.at }}</span>
                    </div>
                    <div class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">{{ entry.detail }}</div>
                  </div>
                </div>
              </div>

              <div ref="audioRoot" class="sr-only" />
            </div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI, userGroupsAPI, voiceAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey, Group, VoicePreflightResponse } from '@/types'

type VoiceLogEntry = { id: number; title: string; detail: string; at: string }

type LivekitRoom = {
  connect: (url: string, token: string) => Promise<void>
  disconnect: () => Promise<void>
  on: (event: unknown, cb: (...args: any[]) => void) => void
  localParticipant: { publishTrack: (track: any) => Promise<void> }
}

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loadingBootstrap = ref(false)
const generatingKey = ref(false)
const preflighting = ref(false)
const startingCall = ref(false)
const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const userGroupRates = ref<Record<number, number>>({})
const selectedApiKeyId = ref<number | null>(null)
const selectedGroupId = ref<number | null>(null)
const apiKeyInput = ref('')
const selectedVoice = ref('ara')
const selectedPersonality = ref('assistant')
const speed = ref(1.0)
const preflight = ref<VoicePreflightResponse | null>(null)
const roomConnected = ref(false)
const callDurationSeconds = ref(0)
const logs = ref<VoiceLogEntry[]>([])
const visualizerBars = ref<number[]>(Array.from({ length: 40 }, () => 8))
const browserConnectivity = ref({ ok: false, checked: false, label: t('voiceChat.checks.pending'), shortLabel: t('common.notAvailable') })
const audioRoot = ref<HTMLElement | null>(null)
const microphoneReady = ref(false)

let livekitRoom: LivekitRoom | null = null
let timerId: number | null = null
let visualizerId: number | null = null
let callStartedAt = 0

const voiceOptions = [
  { value: 'ara', label: 'Ara' },
  { value: 'eve', label: 'Eve' },
  { value: 'leo', label: 'Leo' },
  { value: 'rex', label: 'Rex' },
  { value: 'sal', label: 'Sal' },
  { value: 'gork', label: 'Gork' }
]

const personalityOptions = [
  { value: 'assistant', label: 'Assistant' },
  { value: 'therapist', label: 'Therapist' },
  { value: 'storyteller', label: 'Storyteller' },
  { value: 'kids_story_time', label: 'Kids Story Time' },
  { value: 'kids_trivia_game', label: 'Kids Trivia Game' },
  { value: 'meditation', label: 'Meditation' },
  { value: 'doc', label: 'Doc' },
  { value: 'conspiracy', label: 'Conspiracy' }
]

const apiKeyOptions = computed(() => apiKeys.value.map((key) => ({ value: key.id, label: `${key.name} · ${maskKey(key.key)}` })))
const groupOptions = computed(() => groups.value.map((group) => ({ value: group.id, label: `${group.name} · ${effectiveRateForGroup(group.id).toFixed(2)}x` })))

const activeApiKey = computed(() => {
  if (selectedApiKeyId.value != null) {
    return apiKeys.value.find((item) => item.id === selectedApiKeyId.value) || null
  }
  return apiKeys.value.find((item) => item.key === apiKeyInput.value.trim()) || null
})

const effectiveRateForGroup = (groupId?: number | null) => {
  if (!groupId) return 1
  const group = groups.value.find((item) => item.id === groupId) || apiKeys.value.find((item) => item.group_id === groupId)?.group
  const userRate = userGroupRates.value[groupId]
  if (userRate != null && userRate > 0) return userRate
  return group?.rate_multiplier || 1
}

const effectiveRate = computed(() => effectiveRateForGroup(activeApiKey.value?.group_id ?? selectedGroupId.value))
const baseSinglePrice = computed(() => preflight.value?.single_price_per_call || 0)
const actualSinglePrice = computed(() => baseSinglePrice.value * effectiveRate.value)
const effectiveRateLabel = computed(() => t('voiceChat.price.rateHint', { rate: effectiveRate.value.toFixed(2) }))
const basePriceLabel = computed(() => (preflight.value ? formatPrice(baseSinglePrice.value) : '--'))
const actualPriceLabel = computed(() => (preflight.value ? formatPrice(actualSinglePrice.value) : '--'))
const selectedVoiceLabel = computed(() => voiceOptions.find((item) => item.value === selectedVoice.value)?.label || selectedVoice.value)
const selectedPersonalityLabel = computed(() => personalityOptions.find((item) => item.value === selectedPersonality.value)?.label || selectedPersonality.value)
const canStart = computed(() => !!apiKeyInput.value.trim() && !!preflight.value?.function_ready && !!preflight.value?.server_livekit_ready && browserConnectivity.value.ok && microphoneReady.value)
const statusText = computed(() => {
  if (startingCall.value) return t('voiceChat.status.connecting')
  if (roomConnected.value) return t('voiceChat.status.connected')
  if (preflighting.value) return t('voiceChat.status.checking')
  return t('voiceChat.status.ready')
})
const callStatusClass = computed(() => roomConnected.value ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-300' : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-300')
const formattedDuration = computed(() => {
  const minutes = Math.floor(callDurationSeconds.value / 60).toString().padStart(2, '0')
  const seconds = (callDurationSeconds.value % 60).toString().padStart(2, '0')
  return `${minutes}:${seconds}`
})

const maskKey = (value: string) => (value.length <= 12 ? value : `${value.slice(0, 8)}...${value.slice(-4)}`)
const formatPrice = (value: number) => `$${value.toFixed(2)}`

const statusClass = (ok?: boolean) => ok ? 'text-emerald-600 dark:text-emerald-300' : 'text-amber-600 dark:text-amber-300'
const statusLabel = (ok?: boolean) => ok ? t('voiceChat.checks.pass') : t('voiceChat.checks.fail')

const pushLog = (title: string, detail: string) => {
  logs.value.unshift({ id: Date.now() + Math.random(), title, detail, at: new Date().toLocaleTimeString() })
}

const loadBootstrap = async () => {
  loadingBootstrap.value = true
  try {
    const [keyResult, availableGroups, rates] = await Promise.all([
      keysAPI.list(1, 1000),
      userGroupsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates()
    ])
    apiKeys.value = keyResult.items || []
    groups.value = availableGroups || []
    userGroupRates.value = rates || {}
    if (selectedGroupId.value == null && groups.value.length > 0) {
      selectedGroupId.value = groups.value[0].id
    }
  } catch (error) {
    console.error('Failed to load voice bootstrap:', error)
    appStore.showError(t('voiceChat.bootstrapFailed'))
  } finally {
    loadingBootstrap.value = false
  }
}

const applySelectedApiKey = async () => {
  const matched = apiKeys.value.find((item) => item.id === selectedApiKeyId.value)
  if (!matched) return
  apiKeyInput.value = matched.key
  if (matched.group_id != null) {
    selectedGroupId.value = matched.group_id
  }
  await runPreflight()
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

    await room.connect(session.url, session.token)
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
    appStore.showError(error.response?.data?.message || error.message || t('voiceChat.call.startFailed'))
    pushLog(t('voiceChat.log.errorTitle'), error.message || t('voiceChat.call.startFailed'))
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
  void loadBootstrap()
  void checkMicrophoneSupport()
})

onBeforeUnmount(() => {
  void stopConversation()
})
</script>
