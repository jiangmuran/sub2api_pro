<template>
  <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
    <div class="flex min-h-screen items-end justify-center px-4 pb-20 pt-4 text-center sm:block sm:p-0">
      <!-- Background overlay -->
      <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity dark:bg-gray-900 dark:bg-opacity-75" aria-hidden="true" @click="handleClose"></div>

      <!-- Center modal -->
      <span class="hidden sm:inline-block sm:h-screen sm:align-middle" aria-hidden="true">&#8203;</span>

      <div class="inline-block transform overflow-hidden rounded-lg bg-white text-left align-bottom shadow-xl transition-all dark:bg-dark-800 sm:my-8 sm:w-full sm:max-w-4xl sm:align-middle">
        <!-- Header -->
        <div class="border-b border-gray-200 bg-white px-6 py-4 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-medium leading-6 text-gray-900 dark:text-white">
              {{ t('admin.accounts.batchTest.title') }}
            </h3>
            <button @click="handleClose" class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300">
              <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Body -->
        <div class="bg-gray-50 px-6 py-4 dark:bg-dark-900">
          <!-- Configuration Section -->
          <div v-if="!testing && results.length === 0" class="space-y-4">
            <div class="rounded-lg bg-white p-4 shadow dark:bg-dark-800">
              <h4 class="mb-3 text-sm font-medium text-gray-900 dark:text-white">
                {{ t('admin.accounts.batchTest.configuration') }}
              </h4>
              
              <!-- Selected Accounts -->
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('admin.accounts.batchTest.selectedAccounts') }}
                </label>
                <div class="mt-1 text-sm text-gray-600 dark:text-gray-400">
                  {{ t('admin.accounts.batchTest.selectedCount', { count: accountIds.length }) }}
                </div>
              </div>

              <!-- Model Selection -->
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('admin.accounts.batchTest.model') }}
                </label>
                <input
                  v-model="config.model"
                  type="text"
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-700 dark:text-white sm:text-sm"
                  :placeholder="t('admin.accounts.batchTest.modelPlaceholder')"
                />
              </div>

              <!-- Delay -->
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('admin.accounts.batchTest.delay') }} (ms)
                </label>
                <input
                  v-model.number="config.delayMs"
                  type="number"
                  min="0"
                  max="10000"
                  step="100"
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-700 dark:text-white sm:text-sm"
                />
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.accounts.batchTest.delayHint') }}
                </p>
              </div>

              <!-- Concurrency -->
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('admin.accounts.batchTest.concurrency') }}
                </label>
                <input
                  v-model.number="config.concurrency"
                  type="number"
                  min="1"
                  max="20"
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-700 dark:text-white sm:text-sm"
                />
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.accounts.batchTest.concurrencyHint') }}
                </p>
              </div>

              <!-- Timeout -->
              <div class="mb-4">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('admin.accounts.batchTest.timeout') }} (s)
                </label>
                <input
                  v-model.number="config.timeoutSeconds"
                  type="number"
                  min="5"
                  max="60"
                  class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-primary-500 focus:outline-none focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-700 dark:text-white sm:text-sm"
                />
              </div>
            </div>
          </div>

          <!-- Progress Section -->
          <div v-if="testing || results.length > 0" class="space-y-4">
            <!-- Progress Bar -->
            <div class="rounded-lg bg-white p-4 shadow dark:bg-dark-800">
              <div class="mb-2 flex items-center justify-between text-sm">
                <span class="font-medium text-gray-900 dark:text-white">
                  {{ t('admin.accounts.batchTest.progress') }}
                </span>
                <span class="text-gray-600 dark:text-gray-400">
                  {{ progress.completed }} / {{ progress.total }}
                </span>
              </div>
              <div class="h-2 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-700">
                <div
                  class="h-full bg-primary-500 transition-all duration-300"
                  :style="{ width: `${progressPercentage}%` }"
                ></div>
              </div>
              <div v-if="testing" class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.accounts.batchTest.testing') }}
              </div>
            </div>

            <!-- Results Summary -->
            <div v-if="results.length > 0" class="rounded-lg bg-white p-4 shadow dark:bg-dark-800">
              <h4 class="mb-3 text-sm font-medium text-gray-900 dark:text-white">
                {{ t('admin.accounts.batchTest.results') }}
              </h4>

              <!-- Status Summary -->
              <div class="space-y-2">
                <div
                  v-for="(group, status) in groupedResults"
                  :key="status"
                  class="cursor-pointer rounded-lg border border-gray-200 p-3 transition-colors hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-700"
                  @click="toggleGroup(status)"
                >
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-3">
                      <!-- Status Icon -->
                      <div
                        class="flex h-8 w-8 items-center justify-center rounded-full"
                        :class="getStatusColor(status)"
                      >
                        <svg v-if="status === 'success'" class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                        </svg>
                        <svg v-else class="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                        </svg>
                      </div>

                      <!-- Status Text -->
                      <div>
                        <div class="font-medium text-gray-900 dark:text-white">
                          {{ getStatusLabel(status) }}
                        </div>
                        <div class="text-sm text-gray-600 dark:text-gray-400">
                          {{ group.length }} {{ t('admin.accounts.batchTest.accounts') }}
                        </div>
                      </div>
                    </div>

                    <!-- Expand Icon -->
                    <svg
                      class="h-5 w-5 transform text-gray-400 transition-transform"
                      :class="{ 'rotate-180': expandedGroups.has(status) }"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>

                  <!-- Account List (Expanded) -->
                  <div
                    v-if="expandedGroups.has(status)"
                    class="mt-3 space-y-2 border-t border-gray-100 pt-3 dark:border-dark-600"
                  >
                    <div
                      v-for="result in group"
                      :key="result.account_id"
                      class="flex items-start gap-2 rounded-md bg-gray-50 p-2 text-sm dark:bg-dark-900"
                    >
                      <div class="flex-1">
                        <div class="font-medium text-gray-900 dark:text-white">
                          ID: {{ result.account_id }} - {{ result.account_name }}
                        </div>
                        <div v-if="result.error" class="mt-1 text-xs text-red-600 dark:text-red-400">
                          {{ result.error }}
                        </div>
                        <div v-if="result.status_code" class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                          {{ t('admin.accounts.batchTest.statusCode') }}: {{ result.status_code }}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="bg-gray-50 px-6 py-4 dark:bg-dark-900 sm:flex sm:flex-row-reverse">
          <button
            v-if="!testing && results.length === 0"
            @click="startTest"
            :disabled="!canStart"
            class="inline-flex w-full justify-center rounded-md border border-transparent bg-primary-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:focus:ring-offset-dark-800 sm:ml-3 sm:w-auto sm:text-sm"
          >
            {{ t('admin.accounts.batchTest.start') }}
          </button>
          <button
            v-if="testing"
            @click="stopTest"
            class="inline-flex w-full justify-center rounded-md border border-transparent bg-red-600 px-4 py-2 text-base font-medium text-white shadow-sm hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 dark:focus:ring-offset-dark-800 sm:ml-3 sm:w-auto sm:text-sm"
          >
            {{ t('admin.accounts.batchTest.stop') }}
          </button>
          <button
            v-if="results.length > 0 && !testing"
            @click="exportResults"
            class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600 dark:focus:ring-offset-dark-800 sm:ml-3 sm:mt-0 sm:w-auto sm:text-sm"
          >
            {{ t('admin.accounts.batchTest.export') }}
          </button>
          <button
            @click="handleClose"
            :disabled="testing"
            class="mt-3 inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600 dark:focus:ring-offset-dark-800 sm:mt-0 sm:w-auto sm:text-sm"
          >
            {{ testing ? t('common.cancel') : t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import api from '@/api'

const { t } = useI18n()
const appStore = useAppStore()

interface Props {
  show: boolean
  accountIds: number[]
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  accountIds: () => []
})

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'complete', results: TestResult[]): void
}>()

interface TestConfig {
  model: string
  delayMs: number
  concurrency: number
  timeoutSeconds: number
}

interface TestResult {
  account_id: number
  account_name: string
  status: 'success' | 'error'
  status_code?: number
  error?: string
  duration_ms?: number
}

const config = ref<TestConfig>({
  model: 'claude-3-5-sonnet-20241022',
  delayMs: 500,
  concurrency: 5,
  timeoutSeconds: 30
})

const testing = ref(false)
const results = ref<TestResult[]>([])
const progress = ref({ completed: 0, total: 0 })
const expandedGroups = ref<Set<string>>(new Set())
const abortController = ref<AbortController | null>(null)

const canStart = computed(() => {
  return props.accountIds.length > 0 && config.value.model.trim() !== ''
})

const progressPercentage = computed(() => {
  if (progress.value.total === 0) return 0
  return Math.round((progress.value.completed / progress.value.total) * 100)
})

const groupedResults = computed(() => {
  const groups: Record<string, TestResult[]> = {}
  
  for (const result of results.value) {
    let key: string = result.status
    if (result.status === 'error' && result.error) {
      // Group by error message
      key = `error:${result.error}` as string
    } else if (result.status_code) {
      // Group by status code
      key = `status:${result.status_code}` as string
    }
    
    if (!groups[key]) {
      groups[key] = []
    }
    groups[key].push(result)
  }
  
  return groups
})

function getStatusColor(status: string): string {
  if (status === 'success') {
    return 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400'
  }
  return 'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400'
}

function getStatusLabel(status: string): string {
  if (status === 'success') {
    return t('admin.accounts.batchTest.statusSuccess')
  }
  if (status.startsWith('error:')) {
    return status.substring(6)
  }
  if (status.startsWith('status:')) {
    return t('admin.accounts.batchTest.statusCode') + ': ' + status.substring(7)
  }
  return status
}

function toggleGroup(status: string) {
  if (expandedGroups.value.has(status)) {
    expandedGroups.value.delete(status)
  } else {
    expandedGroups.value.add(status)
  }
}

async function startTest() {
  if (!canStart.value) return

  testing.value = true
  results.value = []
  progress.value = { completed: 0, total: props.accountIds.length }
  abortController.value = new AbortController()

  try {
    // Calculate reasonable timeout: 
    // (number of accounts / concurrency) * (test timeout + delay) + buffer
    const estimatedTime = Math.ceil(props.accountIds.length / config.value.concurrency) * 
                          (config.value.timeoutSeconds * 1000 + config.value.delayMs)
    const requestTimeout = estimatedTime + 30000 // Add 30s buffer
    
    const response = await api.post(
      '/admin/accounts/batch-test',
      {
        account_ids: props.accountIds,
        model: config.value.model,
        delay_ms: config.value.delayMs,
        concurrency: config.value.concurrency,
        timeout_seconds: config.value.timeoutSeconds
      },
      {
        timeout: requestTimeout, // Custom timeout for this request
        signal: abortController.value.signal,
        // Enable streaming
        onDownloadProgress: (progressEvent) => {
          // Parse SSE stream
          const text = progressEvent.event.target.responseText
          const lines = text.split('\\n')
          
          for (const line of lines) {
            if (line.startsWith('data: ')) {
              try {
                const data = JSON.parse(line.substring(6))
                if (data.type === 'progress') {
                  progress.value.completed = data.completed
                } else if (data.type === 'result') {
                  results.value.push(data.result)
                }
              } catch (e) {
                // Ignore parse errors
              }
            }
          }
        }
      }
    )

    // Handle final response if not using SSE
    if (response.data && Array.isArray(response.data.results)) {
      results.value = response.data.results
      progress.value.completed = results.value.length
    }

    appStore.showToast(
      'success',
      t('admin.accounts.batchTest.completeMessage', { count: results.value.length })
    )
    
    emit('complete', results.value)
  } catch (error: any) {
    if (error.name === 'CanceledError') {
      appStore.showToast('info', t('admin.accounts.batchTest.cancelled'))
    } else {
      console.error('Batch test error:', error)
      appStore.showToast(
        'error',
        error.response?.data?.error || t('admin.accounts.batchTest.error')
      )
    }
  } finally {
    testing.value = false
    abortController.value = null
  }
}

function stopTest() {
  if (abortController.value) {
    abortController.value.abort()
  }
}

function exportResults() {
  const csv = [
    ['Account ID', 'Account Name', 'Status', 'Status Code', 'Error', 'Duration (ms)'].join(','),
    ...results.value.map(r =>
      [
        r.account_id,
        `"${r.account_name}"`,
        r.status,
        r.status_code || '',
        r.error ? `"${r.error.replace(/"/g, '""')}"` : '',
        r.duration_ms || ''
      ].join(',')
    )
  ].join('\n')

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const link = document.createElement('a')
  const url = URL.createObjectURL(blob)
  link.setAttribute('href', url)
  link.setAttribute('download', `batch-test-results-${Date.now()}.csv`)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function handleClose() {
  if (testing.value) {
    if (confirm(t('admin.accounts.batchTest.confirmStop'))) {
      stopTest()
      emit('close')
    }
  } else {
    emit('close')
  }
}

// Reset state when modal closes
watch(() => props.show, (newShow) => {
  if (!newShow) {
    testing.value = false
    results.value = []
    progress.value = { completed: 0, total: 0 }
    expandedGroups.value.clear()
  }
})
</script>
