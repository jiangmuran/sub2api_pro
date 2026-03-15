<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="rounded-3xl border border-gray-200 bg-gradient-to-br from-white via-slate-50 to-emerald-50 p-6 shadow-sm dark:border-dark-600 dark:from-dark-800 dark:via-dark-800 dark:to-emerald-950/20">
        <div class="flex flex-col gap-6 xl:flex-row xl:items-start xl:justify-between">
          <div class="max-w-2xl">
            <div class="inline-flex items-center gap-2 rounded-full bg-white/80 px-3 py-1 text-xs font-medium text-emerald-700 shadow-sm dark:bg-dark-700/80 dark:text-emerald-300">
              <Icon name="search" size="sm" :stroke-width="2" />
              {{ t('modelTest.badge') }}
            </div>
            <h1 class="mt-4 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ t('modelTest.title') }}
            </h1>
            <p class="mt-3 max-w-xl text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('modelTest.description') }}
            </p>
          </div>

          <div class="grid gap-3 sm:grid-cols-3 xl:w-[360px] xl:grid-cols-1">
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('modelTest.summary.keyCount') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ apiKeys.length }}</div>
            </div>
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('modelTest.summary.modelCount') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ models.length }}</div>
            </div>
            <div class="rounded-2xl border border-white/80 bg-white/80 p-4 shadow-sm dark:border-dark-600 dark:bg-dark-800/90">
              <div class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('modelTest.summary.pricedCount') }}</div>
              <div class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ pricedModelsCount }}</div>
            </div>
          </div>
        </div>
      </section>

      <div class="grid gap-6 xl:grid-cols-[380px_minmax(0,1fr)]">
        <aside class="space-y-6">
          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.keyPanel.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.keyPanel.description') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" :disabled="loadingBootstrap" @click="loadBootstrap">
                {{ t('common.refresh') }}
              </button>
            </div>

            <div class="mt-5 space-y-4">
              <div>
                <label class="input-label">{{ t('modelTest.keyPanel.existingKeys') }}</label>
                <Select
                  v-model="selectedApiKeyId"
                  :options="apiKeyOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="t('modelTest.keyPanel.existingKeysPlaceholder')"
                  @change="applySelectedApiKey"
                />
              </div>

              <div>
                <label class="input-label">{{ t('modelTest.keyPanel.directInput') }}</label>
                <div class="flex gap-2">
                  <input v-model="apiKeyInput" type="password" class="input flex-1 font-mono" :placeholder="t('modelTest.keyPanel.directInputPlaceholder')" />
                  <button type="button" class="btn btn-secondary" :disabled="!apiKeyInput.trim()" @click="loadModels">
                    {{ t('modelTest.keyPanel.verify') }}
                  </button>
                </div>
              </div>

              <div class="rounded-xl border border-emerald-200 bg-emerald-50/70 p-4 dark:border-emerald-900/40 dark:bg-emerald-950/20">
                <div class="text-sm font-medium text-emerald-900 dark:text-emerald-200">{{ t('modelTest.keyPanel.generateTitle') }}</div>
                <div class="mt-1 text-xs text-emerald-700 dark:text-emerald-300">{{ t('modelTest.keyPanel.generateHint') }}</div>

                <div class="mt-4 space-y-3">
                  <div>
                    <label class="input-label">{{ t('modelTest.keyPanel.groupLabel') }}</label>
                    <Select
                      v-model="selectedGroupId"
                      :options="groupOptions"
                      value-key="value"
                      label-key="label"
                      :placeholder="t('modelTest.keyPanel.groupPlaceholder')"
                    />
                  </div>
                  <button type="button" class="btn btn-primary w-full" :disabled="generatingKey || selectedGroupId == null" @click="generateApiKey">
                    {{ generatingKey ? t('common.loading') + '...' : t('modelTest.keyPanel.generateButton') }}
                  </button>
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.models.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.models.description') }}</p>
              </div>
              <button type="button" class="btn btn-secondary" :disabled="loadingModels || !apiKeyInput.trim()" @click="loadModels">
                {{ loadingModels ? t('common.loading') + '...' : t('modelTest.models.fetch') }}
              </button>
            </div>

            <div class="mt-4 max-h-[360px] overflow-y-auto rounded-xl border border-gray-200 dark:border-dark-600">
              <div v-if="models.length === 0" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('modelTest.models.empty') }}
              </div>
              <button
                v-for="model in models"
                :key="model.id"
                type="button"
                class="flex w-full items-center justify-between border-b border-gray-100 px-4 py-3 text-left transition hover:bg-gray-50 last:border-b-0 dark:border-dark-700 dark:hover:bg-dark-700/50"
                @click="selectModel(model.id)"
              >
                <div class="min-w-0">
                  <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ model.display_name || model.id }}</div>
                  <div class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">{{ model.id }}</div>
                </div>
                <span class="rounded-full bg-gray-100 px-2 py-1 text-[11px] font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300">{{ t('common.select') }}</span>
              </button>
            </div>
          </section>
        </aside>

        <section class="space-y-6">
          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex items-center justify-between gap-3">
              <div>
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.pricing.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.description') }}</p>
              </div>
              <div class="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">
                {{ effectiveRateLabel }}
              </div>
            </div>

            <div class="mt-4 overflow-auto rounded-xl border border-gray-200 dark:border-dark-600">
              <table class="min-w-full divide-y divide-gray-200 text-xs dark:divide-dark-600">
                <thead class="bg-gray-50 dark:bg-dark-900/40">
                  <tr>
                    <th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.name') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.standardInput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.standardOutput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.actualInput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.actualOutput') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.imagePrice') }}</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                  <tr v-if="models.length === 0">
                    <td colspan="6" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('modelTest.pricing.empty') }}</td>
                  </tr>
                  <tr v-for="model in pricedModels" :key="model.id">
                    <td class="px-3 py-2 text-gray-900 dark:text-white">{{ model.id }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.standardInputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.standardOutputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.actualInputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.actualOutputPrice, model.pricingAvailable) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(model.imagePricePerImage, model.pricingAvailable && model.imagePricePerImage > 0) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="flex-1">
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.image.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.image.description') }}</p>
              </div>
              <div class="flex w-full gap-3 lg:w-[420px]">
                <Select v-model="imageModel" class="flex-1" :options="imageModelOptions" value-key="value" label-key="label" :placeholder="t('modelTest.image.selectModel')" />
                <Select v-model="imageSize" class="w-[140px]" :options="imageSizeOptions" value-key="value" label-key="label" />
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <textarea v-model="imagePrompt" rows="3" class="input" :placeholder="t('modelTest.image.placeholder')" />
              <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.image.hint') }}</div>
                <button type="button" class="btn btn-primary" :disabled="generatingImage || !apiKeyInput.trim() || !imageModel || !imagePrompt.trim()" @click="generateImage">
                  {{ generatingImage ? t('common.loading') + '...' : t('modelTest.image.generate') }}
                </button>
              </div>
            </div>

            <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <div v-if="generatedImages.length === 0" class="col-span-full rounded-xl border border-dashed border-gray-300 px-4 py-10 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
                {{ t('modelTest.image.empty') }}
              </div>
              <a v-for="(image, index) in generatedImages" :key="`${image}-${index}`" :href="image" target="_blank" rel="noreferrer" class="overflow-hidden rounded-2xl border border-gray-200 bg-gray-50 shadow-sm dark:border-dark-600 dark:bg-dark-900/30">
                <img :src="image" :alt="`generated-${index}`" class="aspect-square w-full object-cover" />
              </a>
            </div>
          </div>

          <div class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
              <div class="flex-1">
                <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('modelTest.chat.title') }}</h2>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.chat.description') }}</p>
              </div>
              <div class="flex w-full gap-3 lg:w-[360px]">
                <Select
                  v-model="chatModel"
                  class="flex-1"
                  :options="modelOptions"
                  value-key="value"
                  label-key="label"
                  :placeholder="t('modelTest.chat.selectModel')"
                />
                <button type="button" class="btn btn-secondary" @click="clearChat">{{ t('modelTest.chat.clear') }}</button>
              </div>
            </div>

            <div ref="chatScrollRef" class="mt-4 min-h-[360px] max-h-[520px] overflow-y-auto rounded-2xl border border-gray-200 bg-gray-50/70 p-4 dark:border-dark-600 dark:bg-dark-900/30">
              <div v-if="chatMessages.length === 0" class="flex h-[320px] items-center justify-center text-sm text-gray-500 dark:text-gray-400">
                {{ t('modelTest.chat.empty') }}
              </div>
              <div v-for="(message, index) in chatMessages" :key="index" class="mb-3 flex" :class="message.role === 'user' ? 'justify-end' : 'justify-start'">
                <div :class="message.role === 'user' ? 'max-w-[85%] rounded-2xl rounded-br-md bg-emerald-600 px-4 py-3 text-sm text-white' : 'max-w-[85%] rounded-2xl rounded-bl-md bg-white px-4 py-3 text-sm text-gray-800 shadow-sm dark:bg-dark-700 dark:text-gray-100'">
                  <div class="whitespace-pre-wrap break-words">{{ message.content }}</div>
                </div>
              </div>
              <div v-if="sending" class="mb-3 flex justify-start">
                <div class="rounded-2xl rounded-bl-md bg-white px-4 py-3 text-sm text-gray-500 shadow-sm dark:bg-dark-700 dark:text-gray-300">
                  {{ t('modelTest.chat.sending') }}
                </div>
              </div>
            </div>

            <div class="mt-4 space-y-3">
              <textarea
                v-model="chatInput"
                rows="4"
                class="input"
                :placeholder="t('modelTest.chat.placeholder')"
                @keydown.enter.exact.prevent="sendMessage"
              />
              <div class="flex items-center justify-between gap-3">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('modelTest.chat.hint') }}</div>
                <button type="button" class="btn btn-primary" :disabled="sending || !apiKeyInput.trim() || !chatModel || !chatInput.trim()" @click="sendMessage">
                  {{ t('modelTest.chat.send') }}
                </button>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { keysAPI, usageAPI, userGroupsAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import type { ApiKey, Group, ModelPricingPreviewItem } from '@/types'

type LiveModel = { id: string; display_name?: string }
type ChatMessage = { role: 'user' | 'assistant'; content: string }
type ImageGenerationResponse = { data?: Array<{ url?: string }> }

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loadingBootstrap = ref(false)
const loadingModels = ref(false)
const generatingKey = ref(false)
const sending = ref(false)
const generatingImage = ref(false)

const apiKeys = ref<ApiKey[]>([])
const groups = ref<Group[]>([])
const userGroupRates = ref<Record<number, number>>({})
const models = ref<LiveModel[]>([])
const pricingMap = ref<Record<string, ModelPricingPreviewItem>>({})
const selectedApiKeyId = ref<number | null>(null)
const selectedGroupId = ref<number | null>(null)
const apiKeyInput = ref('')
const chatModel = ref('')
const chatInput = ref('')
const chatMessages = ref<ChatMessage[]>([])
const chatScrollRef = ref<HTMLElement | null>(null)
const imageModel = ref('')
const imagePrompt = ref('')
const imageSize = ref('1024x1024')
const generatedImages = ref<string[]>([])

const apiKeyOptions = computed(() => apiKeys.value.map((key) => ({ value: key.id, label: `${key.name} · ${maskKey(key.key)}` })))
const groupOptions = computed(() => groups.value.map((group) => ({ value: group.id, label: `${group.name} · ${effectiveRateForGroup(group.id).toFixed(2)}x` })))
const modelOptions = computed(() => models.value.map((model) => ({ value: model.id, label: model.display_name || model.id })))
const imageModelOptions = computed(() =>
  models.value
    .filter((model) => {
      const pricing = pricingMap.value[model.id]
      if ((pricing?.image_price_per_image || 0) > 0) {
        return true
      }
      return /(imagine|image|img|flux|sdxl|dall-e|recraft|canvas|vision-image)/i.test(model.id)
    })
    .map((model) => ({ value: model.id, label: model.display_name || model.id }))
)
const imageSizeOptions = [
  { value: '1024x1024', label: '1024x1024' },
  { value: '1792x1024', label: '1792x1024' },
  { value: '1024x1792', label: '1024x1792' }
]

const effectiveRateForGroup = (groupId?: number | null) => {
  if (!groupId) return 1
  const group = groups.value.find((item) => item.id === groupId) || apiKeys.value.find((item) => item.group_id === groupId)?.group
  const userRate = userGroupRates.value[groupId]
  if (userRate != null && userRate > 0) return userRate
  return group?.rate_multiplier || 1
}

const activeApiKey = computed(() => {
  if (selectedApiKeyId.value == null) return apiKeys.value.find((item) => item.key === apiKeyInput.value.trim()) || null
  return apiKeys.value.find((item) => item.id === selectedApiKeyId.value) || null
})

const effectiveRate = computed(() => effectiveRateForGroup(activeApiKey.value?.group_id ?? selectedGroupId.value))
const effectiveRateLabel = computed(() => t('modelTest.pricing.effectiveRate', { rate: effectiveRate.value.toFixed(2) }))
const pricedModelsCount = computed(() => Object.values(pricingMap.value).filter((item) => item.pricing_available).length)

const pricedModels = computed(() =>
  models.value.map((model) => {
    const pricing = pricingMap.value[model.id]
    const standardInputPrice = pricing?.input_price_per_1m || 0
    const standardOutputPrice = pricing?.output_price_per_1m || 0
    const imagePricePerImage = pricing?.image_price_per_image || 0
    return {
      ...model,
      pricingAvailable: pricing?.pricing_available || false,
      standardInputPrice,
      standardOutputPrice,
      imagePricePerImage,
      actualInputPrice: standardInputPrice * effectiveRate.value,
      actualOutputPrice: standardOutputPrice * effectiveRate.value
    }
  })
)

const maskKey = (value: string) => (value.length <= 12 ? value : `${value.slice(0, 8)}...${value.slice(-4)}`)

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
    console.error('Failed to load model test bootstrap data:', error)
    appStore.showError(t('modelTest.bootstrapFailed'))
  } finally {
    loadingBootstrap.value = false
  }
}

const applySelectedApiKey = () => {
  const matched = apiKeys.value.find((item) => item.id === selectedApiKeyId.value)
  if (!matched) return
  apiKeyInput.value = matched.key
  if (matched.group_id != null) {
    selectedGroupId.value = matched.group_id
  }
}

const fetchGatewayJSON = async (path: string, payload?: unknown) => {
  const response = await fetch(path, {
    method: payload ? 'POST' : 'GET',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${apiKeyInput.value.trim()}`
    },
    body: payload ? JSON.stringify(payload) : undefined
  })
  const text = await response.text()
  const contentType = response.headers.get('content-type') || ''
  if (text && !contentType.includes('application/json')) {
    throw new Error(t('modelTest.models.nonJsonResponse'))
  }
  let parsed: any = null
  try {
    parsed = text ? JSON.parse(text) : null
  } catch {
    throw new Error(t('modelTest.models.nonJsonResponse'))
  }
  if (!response.ok) {
    throw new Error(parsed?.error?.message || parsed?.message || `HTTP ${response.status}`)
  }
  return parsed
}

const loadModels = async () => {
  if (!apiKeyInput.value.trim()) {
    appStore.showError(t('modelTest.keyPanel.directInputRequired'))
    return
  }
  loadingModels.value = true
  try {
    const data = await fetchGatewayJSON('/v1/models')
    models.value = Array.isArray(data?.data) ? data.data : []
    if (!chatModel.value && models.value.length > 0) {
      chatModel.value = models.value[0].id
    }
    if (!imageModel.value && imageModelOptions.value.length > 0) {
      imageModel.value = imageModelOptions.value[0].value
    }
    const pricing = await usageAPI.getModelPricingPreview(models.value.map((item) => item.id), apiKeyInput.value.trim())
    pricingMap.value = Object.fromEntries((pricing.models || []).map((item) => [item.model, item]))
  } catch (error: any) {
    appStore.showError(error.message || t('modelTest.models.fetchFailed'))
  } finally {
    loadingModels.value = false
  }
}

const selectModel = (modelId: string) => {
  chatModel.value = modelId
}

const generateApiKey = async () => {
  if (selectedGroupId.value == null) {
    appStore.showError(t('modelTest.keyPanel.groupRequired'))
    return
  }
  generatingKey.value = true
  try {
    const created = await keysAPI.create(`model-test-${Date.now()}`, selectedGroupId.value)
    apiKeys.value = [created, ...apiKeys.value]
    selectedApiKeyId.value = created.id
    apiKeyInput.value = created.key
    await copyToClipboard(created.key)
    appStore.showSuccess(t('modelTest.keyPanel.generatedAndCopied'))
    await loadModels()
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('modelTest.keyPanel.generateFailed'))
  } finally {
    generatingKey.value = false
  }
}

const clearChat = () => {
  chatMessages.value = []
  chatInput.value = ''
}

const scrollChatToBottom = async () => {
  await nextTick()
  if (chatScrollRef.value) {
    chatScrollRef.value.scrollTop = chatScrollRef.value.scrollHeight
  }
}

const sendMessage = async () => {
  const prompt = chatInput.value.trim()
  if (!prompt || !chatModel.value || !apiKeyInput.value.trim()) return
  chatMessages.value.push({ role: 'user', content: prompt })
  chatInput.value = ''
  await scrollChatToBottom()
  sending.value = true
  try {
    const input = chatMessages.value.map((message) => ({
      role: message.role,
      content: [{ type: 'input_text', text: message.content }]
    }))
    const result = await fetchGatewayJSON('/v1/responses', { model: chatModel.value, input })
    const reply =
      result?.output_text ||
      result?.output?.find?.((item: any) => item.type === 'message')?.content?.find?.((item: any) => item.type === 'output_text')?.text ||
      t('modelTest.chat.emptyReply')
    chatMessages.value.push({ role: 'assistant', content: reply })
    await scrollChatToBottom()
  } catch (error: any) {
    appStore.showError(error.message || t('modelTest.chat.failed'))
  } finally {
    sending.value = false
  }
}

const generateImage = async () => {
  const prompt = imagePrompt.value.trim()
  if (!prompt || !imageModel.value || !apiKeyInput.value.trim()) return
  generatingImage.value = true
  try {
    const result = await fetchGatewayJSON('/v1/images/generations', {
      model: imageModel.value,
      prompt,
      n: 1,
      size: imageSize.value,
      response_format: 'url'
    }) as ImageGenerationResponse
    generatedImages.value = (result.data || []).map((item) => item.url || '').filter(Boolean)
    if (generatedImages.value.length === 0) {
      throw new Error(t('modelTest.image.failed'))
    }
  } catch (error: any) {
    appStore.showError(error.message || t('modelTest.image.failed'))
  } finally {
    generatingImage.value = false
  }
}

const formatPrice = (value: number, available: boolean) => (available ? `$${value.toFixed(2)}` : '--')

watch(selectedGroupId, () => {
  if (models.value.length > 0) {
    void loadModels()
  }
})

watch(imageModelOptions, (options) => {
  if (!imageModel.value && options.length > 0) {
    imageModel.value = options[0].value
  }
})

onMounted(() => {
  void loadBootstrap()
})
</script>
