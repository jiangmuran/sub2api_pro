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
                  <button type="button" class="w-full rounded-lg bg-primary-600 px-3 py-2 text-xs font-medium text-white transition hover:bg-primary-700" @click="activeTab = 'pricing'">
                    {{ t('admin.accounts.openai.viewPricingPreview') }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="activeTab === 'pricing'" class="space-y-4">
            <div>
              <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.accounts.openai.pricingTitle') }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.pricingDesc') }}</div>
              <div class="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-700 dark:bg-amber-900/20 dark:text-amber-300">
                {{ t('admin.accounts.openai.pricingHint') }}
              </div>
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
                  <template v-for="item in models" :key="item.id">
                  <tr>
                    <td class="px-3 py-2 text-gray-900 dark:text-white">{{ item.id }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">
                      <template v-if="item.pricing_available || hasManualPricing(item.id)">
                        {{ formatPrice(resolveInputPrice(item), true) }}
                      </template>
                      <input
                        v-else
                        v-model="manualPricing[item.id].input"
                        type="number"
                        min="0"
                        step="0.01"
                        class="input h-8 w-24 text-right text-xs"
                        :placeholder="t('admin.accounts.openai.manualPricingPlaceholder')"
                      />
                    </td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">
                      <template v-if="item.pricing_available || hasManualPricing(item.id)">
                        {{ formatPrice(resolveOutputPrice(item), true) }}
                      </template>
                      <input
                        v-else
                        v-model="manualPricing[item.id].output"
                        type="number"
                        min="0"
                        step="0.01"
                        class="input h-8 w-24 text-right text-xs"
                        :placeholder="t('admin.accounts.openai.manualPricingPlaceholder')"
                      />
                    </td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(resolveInputPrice(item) * rateMultiplierValue, item.pricing_available || hasManualPricing(item.id)) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(resolveOutputPrice(item) * rateMultiplierValue, item.pricing_available || hasManualPricing(item.id)) }}</td>
                  </tr>
                  <tr class="bg-gray-50/70 dark:bg-dark-900/20">
                    <td colspan="5" class="px-3 py-2 text-[11px] text-gray-500 dark:text-gray-400">
                      {{ pricingSourceLabel(item) }}
                    </td>
                  </tr>
                  </template>
                </tbody>
              </table>
            </div>
            <div class="rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-500 dark:bg-dark-900/30 dark:text-gray-400">
              {{ t('admin.accounts.openai.unavailablePricingHint') }}
            </div>
            <div class="rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-500 dark:bg-dark-900/30 dark:text-slate-400">
              {{ t('admin.accounts.openai.manualPricingHint') }}
            </div>
          </div>

        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import type { OpenAICompatiblePreviewModel } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps<{
  baseUrl: string
  apiKey: string
  proxyId?: number | null
  rateMultiplier?: number | null
}>()

const emit = defineEmits<{
  (e: 'apply-models', models: string[]): void
  (e: 'manual-pricing-change', pricing: Record<string, { input_price_per_1m: number; output_price_per_1m: number }>): void
}>()

const activeTab = ref<'models' | 'pricing'>('models')
const loadingModels = ref(false)
const models = ref<OpenAICompatiblePreviewModel[]>([])
const manualPricing = ref<Record<string, { input: string; output: string }>>({})

const tabs = computed(() => [
  { key: 'models' as const, label: t('admin.accounts.openai.workbenchModels'), icon: 'search' as const },
  { key: 'pricing' as const, label: t('admin.accounts.openai.workbenchPricing'), icon: 'chart' as const }
])

const canQuery = computed(() => props.baseUrl.trim() !== '' && props.apiKey.trim() !== '')
const pricedModelsCount = computed(() => models.value.filter(item => item.pricing_available).length)
const rateMultiplierValue = computed(() => (props.rateMultiplier && props.rateMultiplier > 0 ? props.rateMultiplier : 1))

watch([() => props.baseUrl, () => props.apiKey, () => props.proxyId, () => props.rateMultiplier], () => {
  models.value = []
  manualPricing.value = {}
})

watch(
  manualPricing,
  (value) => {
    const payload: Record<string, { input_price_per_1m: number; output_price_per_1m: number }> = {}
    for (const [model, pricing] of Object.entries(value)) {
      const input = parseManualPrice(pricing.input)
      const output = parseManualPrice(pricing.output)
      if (input > 0 || output > 0) {
        payload[model] = {
          input_price_per_1m: input,
          output_price_per_1m: output
        }
      }
    }
    emit('manual-pricing-change', payload)
  },
  { deep: true }
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
    manualPricing.value = Object.fromEntries(
      models.value.map((item) => [item.id, { input: '', output: '' }])
    )
    if (models.value.length > 0) {
      activeTab.value = 'pricing'
    }
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.accounts.openai.workbenchLoadFailed'))
  } finally {
    loadingModels.value = false
  }
}

const selectModel = (modelId: string) => {
  activeTab.value = 'pricing'
  emit('apply-models', [modelId])
}

const parseManualPrice = (value: string) => {
  const parsed = Number.parseFloat(value)
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0
}

const hasManualPricing = (modelId: string) => {
  const entry = manualPricing.value[modelId]
  return !!entry && (entry.input.trim() !== '' || entry.output.trim() !== '')
}

const resolveInputPrice = (item: OpenAICompatiblePreviewModel) =>
  item.pricing_available ? item.input_price_per_1m : parseManualPrice(manualPricing.value[item.id]?.input || '')

const resolveOutputPrice = (item: OpenAICompatiblePreviewModel) =>
  item.pricing_available ? item.output_price_per_1m : parseManualPrice(manualPricing.value[item.id]?.output || '')

const pricingSourceLabel = (item: OpenAICompatiblePreviewModel) => {
  if (hasManualPricing(item.id)) {
    return t('admin.accounts.openai.pricingSourceManual')
  }
  switch (item.pricing_source) {
    case 'local':
      return t('admin.accounts.openai.pricingSourceLocal')
    case 'builtin':
      return t('admin.accounts.openai.pricingSourceBuiltin')
    case 'openrouter':
      return t('admin.accounts.openai.pricingSourceOpenRouter')
    case 'account':
      return t('admin.accounts.openai.pricingSourceAccount')
    default:
      return t('admin.accounts.openai.pricingSourceUnknown')
  }
}

const formatPrice = (value: number, available: boolean) => {
  if (!available) return '--'
  return `$${value.toFixed(2)}`
}
</script>
