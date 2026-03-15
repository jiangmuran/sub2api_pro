<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="rounded-3xl border border-gray-200 bg-gradient-to-br from-white via-slate-50 to-amber-50 p-6 shadow-sm dark:border-dark-600 dark:from-dark-800 dark:via-dark-800 dark:to-amber-950/20">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <div class="inline-flex items-center gap-2 rounded-full bg-white/80 px-3 py-1 text-xs font-medium text-amber-700 shadow-sm dark:bg-dark-700/80 dark:text-amber-300">
              <Icon name="chart" size="sm" :stroke-width="2" />
              {{ t('admin.pricingManagement.badge') }}
            </div>
            <h1 class="mt-4 text-3xl font-semibold tracking-tight text-gray-900 dark:text-white">
              {{ t('admin.pricingManagement.title') }}
            </h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-300">
              {{ t('admin.pricingManagement.description') }}
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-3">
            <button type="button" class="btn btn-secondary" :disabled="loadingAccounts || loadingModels" @click="reloadAll">
              <Icon name="refresh" size="sm" class="mr-1.5" />
              {{ t('common.refresh') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="saving || !selectedAccountId || !hasManualChanges" @click="savePricing">
              <Icon name="check" size="sm" class="mr-1.5" />
              {{ saving ? t('common.saving') : t('admin.pricingManagement.savePricing') }}
            </button>
          </div>
        </div>
      </section>

      <div class="grid gap-6 xl:grid-cols-[340px_minmax(0,1fr)]">
        <aside class="space-y-6">
          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.pricingManagement.accountSelector') }}</div>
            <div class="mt-4 space-y-4">
              <input v-model="search" type="text" class="input" :placeholder="t('admin.pricingManagement.searchPlaceholder')" />
              <div class="max-h-[520px] overflow-y-auto rounded-xl border border-gray-200 dark:border-dark-600">
                <button
                  v-for="account in filteredAccounts"
                  :key="account.id"
                  type="button"
                  class="flex w-full items-center justify-between border-b border-gray-100 px-4 py-3 text-left transition hover:bg-gray-50 last:border-b-0 dark:border-dark-700 dark:hover:bg-dark-700/50"
                  :class="selectedAccountId === account.id ? 'bg-amber-50 dark:bg-amber-900/10' : ''"
                  @click="selectAccount(account.id)"
                >
                  <div class="min-w-0">
                    <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ account.name }}</div>
                    <div class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">{{ account.credentials?.base_url || '-' }}</div>
                  </div>
                  <span class="rounded-full bg-gray-100 px-2 py-1 text-[11px] font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">P{{ account.priority }}</span>
                </button>
                <div v-if="filteredAccounts.length === 0" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.pricingManagement.noAccounts') }}
                </div>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
            <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.pricingManagement.summaryTitle') }}</div>
            <div class="mt-4 space-y-3 text-sm">
              <div class="flex items-center justify-between"><span class="text-gray-500 dark:text-gray-400">{{ t('admin.pricingManagement.totalModels') }}</span><span class="font-semibold text-gray-900 dark:text-white">{{ previewModels.length }}</span></div>
              <div class="flex items-center justify-between"><span class="text-gray-500 dark:text-gray-400">{{ t('admin.pricingManagement.missingModels') }}</span><span class="font-semibold text-red-600 dark:text-red-400">{{ missingCount }}</span></div>
              <div class="flex items-center justify-between"><span class="text-gray-500 dark:text-gray-400">{{ t('admin.pricingManagement.manualCount') }}</span><span class="font-semibold text-amber-600 dark:text-amber-400">{{ manualCount }}</span></div>
              <label class="flex items-center gap-2 rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-900/30 dark:text-gray-400">
                <input v-model="showMissingOnly" type="checkbox" class="rounded border-gray-300 text-primary-600 focus:ring-primary-500" />
                {{ t('admin.pricingManagement.showMissingOnly') }}
              </label>
            </div>
          </section>
        </aside>

        <section class="rounded-2xl border border-gray-200 bg-white p-5 shadow-sm dark:border-dark-600 dark:bg-dark-800">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.pricingManagement.tableTitle') }}</div>
              <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.pricingManagement.tableDesc') }}</div>
            </div>
            <div v-if="selectedAccount" class="rounded-full bg-gray-100 px-3 py-1 text-xs font-medium text-gray-600 dark:bg-dark-700 dark:text-gray-300">
              {{ selectedAccount.name }}
            </div>
          </div>

          <div class="mt-4 overflow-auto rounded-xl border border-gray-200 dark:border-dark-600">
            <table class="min-w-full divide-y divide-gray-200 text-xs dark:divide-dark-600">
              <thead class="bg-gray-50 dark:bg-dark-900/40">
                <tr>
                  <th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-400">{{ t('common.name') }}</th>
                  <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.standardInputPrice') }}</th>
                  <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.standardOutputPrice') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.accountInputPrice') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.accounts.openai.accountOutputPrice') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.pricingManagement.imagePrice') }}</th>
                    <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">{{ t('admin.pricingManagement.accountImagePrice') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-if="visibleModels.length === 0">
                  <td colspan="7" class="px-4 py-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('admin.pricingManagement.noModels') }}</td>
                </tr>
                <template v-for="item in visibleModels" :key="item.id">
                  <tr>
                    <td class="px-3 py-2 text-gray-900 dark:text-white">{{ item.id }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">
                      <template v-if="item.pricing_available && !hasManualPricing(item.id)">{{ formatPrice(resolveInputPrice(item), true) }}</template>
                      <input v-else v-model="manualPricing[item.id].input" type="number" min="0" step="0.01" class="input h-8 w-24 text-right text-xs" />
                    </td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">
                      <template v-if="item.pricing_available && !hasManualPricing(item.id)">{{ formatPrice(resolveOutputPrice(item), true) }}</template>
                      <input v-else v-model="manualPricing[item.id].output" type="number" min="0" step="0.01" class="input h-8 w-24 text-right text-xs" />
                    </td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(resolveInputPrice(item) * rateMultiplierValue, item.pricing_available || hasManualPricing(item.id)) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(resolveOutputPrice(item) * rateMultiplierValue, item.pricing_available || hasManualPricing(item.id)) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">
                      <template v-if="item.image_price_per_image > 0 && !hasManualPricing(item.id)">{{ formatPrice(resolveImagePrice(item), true) }}</template>
                      <input v-else v-model="manualPricing[item.id].image" type="number" min="0" step="0.01" class="input h-8 w-24 text-right text-xs" />
                    </td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300">{{ formatPrice(resolveImagePrice(item) * rateMultiplierValue, item.image_price_per_image > 0 || hasManualPricing(item.id)) }}</td>
                  </tr>
                  <tr class="bg-gray-50/70 dark:bg-dark-900/20">
                    <td colspan="7" class="px-3 py-2 text-[11px] text-gray-500 dark:text-gray-400">{{ pricingSourceLabel(item) }}</td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { Account, OpenAICompatiblePreviewModel } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const loadingAccounts = ref(false)
const loadingModels = ref(false)
const saving = ref(false)
const search = ref('')
const showMissingOnly = ref(false)
const accounts = ref<Account[]>([])
const selectedAccountId = ref<number | null>(null)
const selectedAccount = ref<Account | null>(null)
const previewModels = ref<OpenAICompatiblePreviewModel[]>([])
const manualPricing = ref<Record<string, { input: string; output: string; image: string }>>({})

const filteredAccounts = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  return accounts.value.filter((account) => {
    if (account.platform !== 'openai' || account.type !== 'apikey') return false
    if (!keyword) return true
    return account.name.toLowerCase().includes(keyword) || String(account.credentials?.base_url || '').toLowerCase().includes(keyword)
  })
})

const visibleModels = computed(() =>
  showMissingOnly.value
    ? previewModels.value.filter((item) => !item.pricing_available && !hasManualPricing(item.id))
    : previewModels.value
)
const missingCount = computed(() => previewModels.value.filter((item) => !item.pricing_available && !hasManualPricing(item.id)).length)
const manualCount = computed(() => Object.keys(buildManualPricingPayload()).length)
const hasManualChanges = computed(() => manualCount.value > 0)
const rateMultiplierValue = computed(() => (selectedAccount.value?.rate_multiplier && selectedAccount.value.rate_multiplier > 0 ? selectedAccount.value.rate_multiplier : 1))

const formatPrice = (value: number, available: boolean) => (available ? `$${value.toFixed(2)}` : '--')
const normalizeManualValue = (value: unknown) => String(value ?? '')
const parseManualPrice = (value: unknown) => {
  const parsed = Number.parseFloat(normalizeManualValue(value))
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : 0
}
const hasManualPricing = (modelId: string) => {
  const entry = manualPricing.value[modelId]
  return !!entry && (
    normalizeManualValue(entry.input).trim() !== '' ||
    normalizeManualValue(entry.output).trim() !== '' ||
    normalizeManualValue(entry.image).trim() !== ''
  )
}
const resolveInputPrice = (item: OpenAICompatiblePreviewModel) => item.pricing_available ? item.input_price_per_1m : parseManualPrice(manualPricing.value[item.id]?.input)
const resolveOutputPrice = (item: OpenAICompatiblePreviewModel) => item.pricing_available ? item.output_price_per_1m : parseManualPrice(manualPricing.value[item.id]?.output)
const resolveImagePrice = (item: OpenAICompatiblePreviewModel) => item.image_price_per_image > 0 ? item.image_price_per_image : parseManualPrice(manualPricing.value[item.id]?.image)

const pricingSourceLabel = (item: OpenAICompatiblePreviewModel) => {
  if (hasManualPricing(item.id)) return t('admin.accounts.openai.pricingSourceManual')
  switch (item.pricing_source) {
    case 'local': return t('admin.accounts.openai.pricingSourceLocal')
    case 'builtin': return t('admin.accounts.openai.pricingSourceBuiltin')
    case 'openrouter': return t('admin.accounts.openai.pricingSourceOpenRouter')
    case 'account': return t('admin.accounts.openai.pricingSourceAccount')
    default: return t('admin.accounts.openai.pricingSourceUnknown')
  }
}

const buildManualPricingPayload = () => {
  const payload: Record<string, { input_price_per_1m: number; output_price_per_1m: number; image_price_per_image?: number }> = {}
  for (const [model, pricing] of Object.entries(manualPricing.value)) {
    const input = parseManualPrice(pricing.input)
    const output = parseManualPrice(pricing.output)
    const image = parseManualPrice(pricing.image)
    if (input > 0 || output > 0 || image > 0) payload[model] = { input_price_per_1m: input, output_price_per_1m: output, ...(image > 0 ? { image_price_per_image: image } : {}) }
  }
  return payload
}

const extractAccountModels = (account: Account): string[] => {
  const raw = account.credentials?.model_mapping as Record<string, string> | undefined
  return raw ? Object.keys(raw) : []
}

const loadAccounts = async () => {
  loadingAccounts.value = true
  try {
    const result = await adminAPI.accounts.list(1, 1000, { platform: 'openai', type: 'apikey' })
    accounts.value = result.items || []
    if (!selectedAccountId.value && accounts.value.length > 0) {
      await selectAccount(accounts.value[0].id)
    }
  } finally {
    loadingAccounts.value = false
  }
}

const selectAccount = async (accountID: number) => {
  selectedAccountId.value = accountID
  loadingModels.value = true
  try {
    const account = await adminAPI.accounts.getById(accountID)
    selectedAccount.value = account
    const result = await adminAPI.accounts.previewOpenAICompatibleModels({
      base_url: String(account.credentials?.base_url || ''),
      api_key: String(account.credentials?.api_key || ''),
      proxy_id: account.proxy_id ?? null,
      rate_multiplier: account.rate_multiplier ?? 1,
      models: extractAccountModels(account)
    })
    previewModels.value = result.models || []
    const saved = ((account.extra as Record<string, any> | undefined)?.openai_manual_model_pricing || {}) as Record<string, { input_price_per_1m: number; output_price_per_1m: number; image_price_per_image?: number }>
    manualPricing.value = Object.fromEntries(
        previewModels.value.map((item) => [item.id, { input: saved[item.id]?.input_price_per_1m?.toString?.() || '', output: saved[item.id]?.output_price_per_1m?.toString?.() || '', image: saved[item.id]?.image_price_per_image?.toString?.() || '' }])
      )
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.pricingManagement.loadFailed'))
  } finally {
    loadingModels.value = false
  }
}

const reloadAll = async () => {
  if (selectedAccountId.value) {
    await selectAccount(selectedAccountId.value)
  } else {
    await loadAccounts()
  }
}

const savePricing = async () => {
  if (!selectedAccount.value) return
  saving.value = true
  try {
    const currentExtra = (selectedAccount.value.extra as Record<string, unknown>) || {}
    await adminAPI.accounts.update(selectedAccount.value.id, {
      extra: {
        ...currentExtra,
        openai_manual_model_pricing: buildManualPricingPayload()
      }
    })
    appStore.showSuccess(t('admin.pricingManagement.saveSuccess'))
    await selectAccount(selectedAccount.value.id)
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('admin.pricingManagement.saveFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  void loadAccounts()
})
</script>
