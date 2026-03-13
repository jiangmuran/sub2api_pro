<template>
  <div class="overflow-hidden rounded-2xl border border-gray-200 bg-white/90 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
    <div class="border-b border-gray-200 px-4 py-4 dark:border-dark-600">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <div class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ t('admin.accounts.openai.workbenchTitle') }}
          </div>
          <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t('admin.accounts.openai.workbenchDesc') }}
          </div>
        </div>
        <button
          type="button"
          class="btn btn-secondary shrink-0"
          :disabled="loadingModels || !canQuery"
          @click="loadModels"
        >
          {{ loadingModels ? t('common.loading') + '...' : t('admin.accounts.openai.refreshWorkbench') }}
        </button>
      </div>
    </div>

    <div class="grid min-h-[420px] grid-cols-1 sm:grid-cols-[96px_minmax(0,1fr)]">
      <div class="border-b border-gray-200 bg-gray-50/80 p-3 dark:border-dark-600 dark:bg-dark-900/30 sm:border-b-0 sm:border-r">
        <div class="flex gap-2 overflow-x-auto pb-1 sm:block sm:overflow-visible sm:pb-0">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="flex min-w-[82px] flex-col items-center rounded-xl px-2 py-3 text-center text-xs font-medium transition-all sm:mb-2 sm:w-full sm:min-w-0"
          :class="activeTab === tab.key
            ? 'bg-primary-100 text-primary-700 shadow-sm dark:bg-primary-900/30 dark:text-primary-300'
            : 'text-gray-500 hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-gray-200'"
          @click="activeTab = tab.key"
        >
          <Icon :name="tab.icon" size="sm" :stroke-width="2" />
          <span class="mt-1.5 leading-4">{{ tab.label }}</span>
        </button>
        </div>
      </div>

      <div class="space-y-4 p-4">
        <div v-if="!canQuery" class="flex min-h-[320px] items-center justify-center rounded-xl border border-dashed border-gray-300 bg-gray-50 text-sm text-gray-500 dark:border-dark-500 dark:bg-dark-900/30 dark:text-gray-400">
          {{ t('admin.accounts.openai.workbenchEmpty') }}
        </div>

        <template v-else>
          <div class="rounded-xl border border-gray-200 bg-gray-50/80 p-3 dark:border-dark-600 dark:bg-dark-900/30">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
              <div class="min-w-0 flex-1">
                <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  {{ t('admin.accounts.openai.previewApiKeyTitle') }}
                </div>
                <div class="mt-2 flex flex-col gap-2 sm:flex-row">
                  <input
                    v-model="previewApiKey"
                    type="password"
                    class="input flex-1 font-mono"
                    :placeholder="t('admin.accounts.openai.previewApiKeyPlaceholder')"
                  />
                  <button type="button" class="btn btn-secondary" @click="useAccountKey">
                    {{ t('admin.accounts.openai.useAccountKey') }}
                  </button>
                </div>
                <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.accounts.openai.previewApiKeyHint') }}
                </div>
              </div>

              <div class="min-w-0 lg:w-[280px]">
                <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  {{ t('admin.accounts.openai.userTestApiKeyTitle') }}
                </div>
                <div class="mt-2 flex flex-col gap-2 sm:flex-row">
                  <input
                    v-model="generatedUserApiKey"
                    type="text"
                    class="input flex-1 font-mono"
                    :placeholder="t('admin.accounts.openai.userTestApiKeyPlaceholder')"
                  />
                  <button type="button" class="btn btn-secondary" :disabled="generatingUserKey" @click="generateUserTestKey">
                    {{ generatingUserKey ? t('common.loading') + '...' : t('admin.accounts.openai.generateUserTestApiKey') }}
                  </button>
                  <button type="button" class="btn btn-secondary" :disabled="!generatedUserApiKey" @click="copyUserTestKey">
                    {{ t('common.copy') }}
                  </button>
                </div>
                <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.accounts.openai.userTestApiKeyHint') }}
                </div>
              </div>
            </div>
          </div>

          <div v-if="activeTab === 'models'" class="space-y-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.accounts.openai.onlineModelsTitle') }}</div>
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.onlineModelsDesc') }}</div>
              </div>
              <button
                type="button"
                class="rounded-lg bg-primary-50 px-3 py-2 text-xs font-medium text-primary-700 transition hover:bg-primary-100 dark:bg-primary-900/30 dark:text-primary-300 dark:hover:bg-primary-900/40"
                :disabled="loadingModels || models.length === 0"
                @click="emit('apply-models', models.map(item => item.id))"
              >
                {{ t('admin.accounts.openai.applyOnlineModels') }}
              </button>
            </div>

            <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_190px]">
              <div class="max-h-[320px] overflow-y-auto rounded-xl border border-gray-200 dark:border-dark-600">
              <div v-if="models.length === 0" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.accounts.openai.noOnlineModels') }}
              </div>
              <div v-for="item in models" :key="item.id" class="flex items-center justify-between gap-3 border-b border-gray-100 px-4 py-3 last:border-b-0 dark:border-dark-700">
                <div class="min-w-0">
                  <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ item.display_name }}</div>
                  <div class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">{{ item.id }}</div>
                </div>
                <button
                  type="button"
                  class="rounded-lg border border-gray-200 px-2.5 py-1.5 text-xs font-medium text-gray-600 hover:bg-gray-50 dark:border-dark-500 dark:text-gray-300 dark:hover:bg-dark-700"
                  @click="selectModel(item.id)"
                >
                  {{ t('common.select') }}
                </button>
              </div>
            </div>
              <div class="rounded-xl border border-gray-200 bg-gray-50/80 p-4 dark:border-dark-600 dark:bg-dark-900/30">
                <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.modelSummary') }}</div>
                <div class="mt-4 space-y-3 text-sm">
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.modelCount') }}</span>
                    <span class="font-semibold text-gray-900 dark:text-white">{{ models.length }}</span>
                  </div>
                  <div class="flex items-center justify-between gap-3">
                    <span class="text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.pricedCount') }}</span>
                    <span class="font-semibold text-gray-900 dark:text-white">{{ pricedModelsCount }}</span>
                  </div>
                  <div class="rounded-lg bg-white px-3 py-2 text-xs text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                    {{ t('admin.accounts.openai.modelSummaryHint') }}
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeTab === 'pricing'" class="space-y-4">
            <div>
              <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.accounts.openai.pricingTitle') }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.pricingDesc') }}</div>
            </div>

            <div class="max-h-[340px] overflow-auto rounded-xl border border-gray-200 dark:border-dark-600">
              <table class="min-w-full divide-y divide-gray-200 text-xs dark:divide-dark-600">
                <thead class="bg-gray-50 dark:bg-dark-900/40">
                  <tr>
                    <th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.name') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.standardInputPrice') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.standardOutputPrice') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.accountInputPrice') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.accountOutputPrice') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                  <tr v-if="models.length === 0">
                    <td colspan="5" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.noPricingRows') }}</td>
                  </tr>
                  <tr v-for="item in models" :key="item.id">
                    <td class="px-3 py-2 text-gray-900 dark:text-white">{{ item.id }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(item.input_price_per_1m, item.pricing_available) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(item.output_price_per_1m, item.pricing_available) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(item.account_input_price_per_1m, item.pricing_available) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(item.account_output_price_per_1m, item.pricing_available) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div v-else class="flex min-h-[360px] flex-col gap-4">
            <div class="flex items-center gap-3">
              <Select
                v-model="chatModel"
                class="flex-1"
                :options="chatModelOptions"
                value-key="value"
                label-key="label"
                :placeholder="t('admin.accounts.openai.selectChatModel')"
              />
              <button
                type="button"
                class="rounded-lg border border-gray-200 px-3 py-2 text-xs font-medium text-gray-600 hover:bg-gray-50 dark:border-dark-500 dark:text-gray-300 dark:hover:bg-dark-700"
                @click="clearChat"
              >
                {{ t('admin.accounts.openai.clearChat') }}
              </button>
            </div>

            <div ref="chatScrollRef" class="flex-1 overflow-y-auto rounded-xl border border-gray-200 bg-gray-50/70 p-3 dark:border-dark-600 dark:bg-dark-900/30">
              <div v-if="chatMessages.length === 0" class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('admin.accounts.openai.chatEmpty') }}
              </div>
              <div v-for="(message, index) in chatMessages" :key="index" class="mb-3 flex" :class="message.role === 'user' ? 'justify-end' : 'justify-start'">
                <div :class="message.role === 'user'
                  ? 'max-w-[85%] rounded-2xl rounded-br-md bg-primary-600 px-3 py-2 text-sm text-white'
                  : 'max-w-[85%] rounded-2xl rounded-bl-md bg-white px-3 py-2 text-sm text-gray-800 shadow-sm dark:bg-dark-700 dark:text-gray-100'">
                  <div class="whitespace-pre-wrap break-words">{{ message.content }}</div>
                </div>
              </div>
              <div v-if="chatSending" class="mb-3 flex justify-start">
                <div class="rounded-2xl rounded-bl-md bg-white px-3 py-2 text-sm text-gray-500 shadow-sm dark:bg-dark-700 dark:text-gray-300">
                  {{ t('admin.accounts.openai.chatSending') }}
                </div>
              </div>
            </div>

            <div class="space-y-3">
              <textarea
                v-model="chatInput"
                rows="3"
                class="input"
                :placeholder="t('admin.accounts.openai.chatPlaceholder')"
                @keydown.enter.exact.prevent="sendChat"
              />
              <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.chatHint') }}</div>
                <button
                  type="button"
                  class="btn btn-primary"
                  :disabled="chatSending || !chatModel || !chatInput.trim() || !previewApiKey.trim()"
                  @click="sendChat"
                >
                  {{ chatSending ? t('admin.accounts.openai.chatSending') : t('admin.accounts.openai.sendChat') }}
                </button>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import keysAPI from '@/api/keys'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores/app'
import type { OpenAICompatiblePreviewModel } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const props = defineProps<{
  baseUrl: string
  apiKey: string
  proxyId?: number | null
  rateMultiplier?: number | null
}>()

const emit = defineEmits<{
  (e: 'apply-models', models: string[]): void
}>()

const activeTab = ref<'models' | 'pricing' | 'chat'>('models')
const loadingModels = ref(false)
const chatSending = ref(false)
const generatingUserKey = ref(false)
const models = ref<OpenAICompatiblePreviewModel[]>([])
const chatModel = ref('')
const chatInput = ref('')
const previewApiKey = ref('')
const generatedUserApiKey = ref('')
const chatMessages = ref<Array<{ role: 'user' | 'assistant'; content: string }>>([])
const chatScrollRef = ref<HTMLElement | null>(null)

const tabs = computed(() => [
  { key: 'models' as const, label: t('admin.accounts.openai.workbenchModels'), icon: 'search' as const },
  { key: 'pricing' as const, label: t('admin.accounts.openai.workbenchPricing'), icon: 'chart' as const },
  { key: 'chat' as const, label: t('admin.accounts.openai.workbenchChat'), icon: 'chatBubble' as const }
])

const canQuery = computed(() => props.baseUrl.trim() !== '' && props.apiKey.trim() !== '')
const pricedModelsCount = computed(() => models.value.filter(item => item.pricing_available).length)

const chatModelOptions = computed(() => models.value.map(item => ({ value: item.id, label: item.display_name })))

watch([() => props.baseUrl, () => props.apiKey, () => props.proxyId, () => props.rateMultiplier], () => {
  models.value = []
  chatModel.value = ''
  chatMessages.value = []
  previewApiKey.value = props.apiKey
})

watch(
  () => props.apiKey,
  (value) => {
    if (!previewApiKey.value) {
      previewApiKey.value = value
    }
  },
  { immediate: true }
)

watch(
  canQuery,
  (value) => {
    if (value && models.value.length === 0 && !loadingModels.value) {
      void loadModels()
    }
  },
  { immediate: true }
)

const loadModels = async () => {
  if (!canQuery.value) return
  loadingModels.value = true
  try {
    const result = await adminAPI.accounts.previewOpenAICompatibleModels({
      base_url: props.baseUrl.trim(),
      api_key: props.apiKey.trim(),
      proxy_id: props.proxyId ?? null,
      rate_multiplier: props.rateMultiplier ?? 1
    })
    models.value = result.models || []
    if (!chatModel.value && models.value.length > 0) {
      chatModel.value = models.value[0].id
    }
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.accounts.openai.workbenchLoadFailed'))
  } finally {
    loadingModels.value = false
  }
}

const selectModel = (modelId: string) => {
  chatModel.value = modelId
  activeTab.value = 'chat'
}

const clearChat = () => {
  chatMessages.value = []
  chatInput.value = ''
}

const useAccountKey = () => {
  previewApiKey.value = props.apiKey
  appStore.showSuccess(t('admin.accounts.openai.accountKeyApplied'))
}

const generateUserTestKey = async () => {
  generatingUserKey.value = true
  try {
    const created = await keysAPI.create(`model-test-${Date.now()}`)
    generatedUserApiKey.value = created.key
    await copyToClipboard(created.key)
    appStore.showSuccess(t('admin.accounts.openai.userTestApiKeyGenerated'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.accounts.openai.userTestApiKeyGenerateFailed'))
  } finally {
    generatingUserKey.value = false
  }
}

const copyUserTestKey = async () => {
  if (!generatedUserApiKey.value) return
  await copyToClipboard(generatedUserApiKey.value)
  appStore.showSuccess(t('admin.accounts.openai.userTestApiKeyCopied'))
}

const scrollChatToBottom = async () => {
  await nextTick()
  if (chatScrollRef.value) {
    chatScrollRef.value.scrollTop = chatScrollRef.value.scrollHeight
  }
}

const sendChat = async () => {
  const prompt = chatInput.value.trim()
  if (!prompt || !chatModel.value || !previewApiKey.value.trim()) return
  chatMessages.value.push({ role: 'user', content: prompt })
  chatInput.value = ''
  await scrollChatToBottom()
  chatSending.value = true
  try {
    const result = await adminAPI.accounts.previewOpenAICompatibleChat({
      base_url: props.baseUrl.trim(),
      api_key: previewApiKey.value.trim(),
      proxy_id: props.proxyId ?? null,
      model: chatModel.value,
      messages: chatMessages.value.map(item => ({ role: item.role, content: item.content }))
    })
    chatMessages.value.push({ role: 'assistant', content: result.reply || t('admin.accounts.openai.emptyReply') })
    await scrollChatToBottom()
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.accounts.openai.chatFailed'))
  } finally {
    chatSending.value = false
  }
}

const formatPrice = (value: number, available: boolean) => {
  if (!available) return '--'
  return `$${value.toFixed(2)}`
}
</script>
