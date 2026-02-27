<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="rounded-2xl border border-gray-200 bg-gradient-to-br from-cyan-50 via-white to-blue-50 p-3.5 shadow-sm dark:border-dark-700 dark:from-dark-900 dark:via-dark-900 dark:to-dark-800">
          <div class="flex flex-wrap items-center gap-3">
            <div class="w-full sm:w-auto sm:flex-1 sm:max-w-72">
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.announcements.searchAnnouncements')"
                class="input"
                @input="handleSearch"
              />
            </div>
            <Select
              v-model="filters.status"
              :options="statusFilterOptions"
              class="w-40"
              @change="handleStatusChange"
            />

            <div class="flex flex-wrap items-center gap-1.5">
              <button
                class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
                :class="filters.status === ''
                  ? 'bg-gray-900 text-white dark:bg-white dark:text-dark-900'
                  : 'bg-white text-gray-600 hover:bg-gray-100 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
                @click="setStatusFilter('')"
              >
                {{ t('admin.announcements.allStatus') }}
              </button>
              <button
                class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
                :class="filters.status === 'active'
                  ? 'bg-emerald-600 text-white'
                  : 'bg-white text-gray-600 hover:bg-gray-100 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
                @click="setStatusFilter('active')"
              >
                {{ t('admin.announcements.statusLabels.active') }}
              </button>
              <button
                class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
                :class="filters.status === 'draft'
                  ? 'bg-amber-600 text-white'
                  : 'bg-white text-gray-600 hover:bg-gray-100 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
                @click="setStatusFilter('draft')"
              >
                {{ t('admin.announcements.statusLabels.draft') }}
              </button>
              <button
                class="rounded-lg px-3 py-1.5 text-xs font-medium transition-colors"
                :class="filters.status === 'archived'
                  ? 'bg-slate-600 text-white'
                  : 'bg-white text-gray-600 hover:bg-gray-100 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'"
                @click="setStatusFilter('archived')"
              >
                {{ t('admin.announcements.statusLabels.archived') }}
              </button>
            </div>

            <div class="ml-auto flex flex-wrap items-center gap-2">
              <button
                @click="loadAnnouncements"
                :disabled="loading"
                class="btn btn-secondary"
                :title="t('common.refresh')"
              >
                <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
              </button>
              <button @click="openCreateDialog" class="btn btn-primary">
                <Icon name="plus" size="md" class="mr-1" />
                {{ t('admin.announcements.createAnnouncement') }}
              </button>
            </div>
          </div>
        </div>
      </template>

      <template #table>
        <div class="mb-4 grid grid-cols-2 gap-3 xl:grid-cols-4">
          <div class="rounded-xl border border-blue-200 bg-white p-3 dark:border-blue-800 dark:bg-dark-900">
            <div class="text-xs text-blue-700 dark:text-blue-300">{{ t('common.total') }}</div>
            <div class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">{{ pagination.total }}</div>
          </div>
          <div class="rounded-xl border border-emerald-200 bg-white p-3 dark:border-emerald-800 dark:bg-dark-900">
            <div class="text-xs text-emerald-700 dark:text-emerald-300">{{ t('admin.announcements.statusLabels.active') }}</div>
            <div class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">{{ activeCount }}</div>
          </div>
          <div class="rounded-xl border border-amber-200 bg-white p-3 dark:border-amber-800 dark:bg-dark-900">
            <div class="text-xs text-amber-700 dark:text-amber-300">{{ t('admin.announcements.statusLabels.draft') }}</div>
            <div class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">{{ draftCount }}</div>
          </div>
          <div class="rounded-xl border border-slate-200 bg-white p-3 dark:border-slate-700 dark:bg-dark-900">
            <div class="text-xs text-slate-700 dark:text-slate-300">{{ t('admin.announcements.statusLabels.archived') }}</div>
            <div class="mt-1 text-xl font-semibold text-gray-900 dark:text-white">{{ archivedCount }}</div>
          </div>
        </div>

        <div class="overflow-hidden rounded-2xl border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <DataTable :columns="columns" :data="announcements" :loading="loading">
          <template #cell-title="{ value, row }">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="truncate font-medium text-gray-900 dark:text-white">{{ value }}</span>
                <span class="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] text-gray-500 dark:bg-dark-700 dark:text-gray-400">#{{ row.id }}</span>
              </div>
              <p class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">{{ contentPreview(row.content) }}</p>
              <div class="mt-1 flex items-center gap-2 text-xs text-gray-500 dark:text-dark-400">
                <span>{{ formatDateTime(row.created_at) }}</span>
              </div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span class="inline-flex rounded-full px-2 py-1 text-xs font-medium" :class="statusClass(value)">
              {{ statusLabel(value) }}
            </span>
          </template>

          <template #cell-targeting="{ row }">
            <span class="inline-flex rounded-lg bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300">
              {{ targetingSummary(row.targeting) }}
            </span>
          </template>

          <template #cell-timeRange="{ row }">
            <div class="space-y-1 text-xs text-gray-600 dark:text-gray-300">
              <div class="rounded bg-gray-50 px-2 py-1 dark:bg-dark-800">
                <span class="font-medium">{{ t('admin.announcements.form.startsAt') }}:</span>
                <span class="ml-1">{{ row.starts_at ? formatDateTime(row.starts_at) : t('admin.announcements.timeImmediate') }}</span>
              </div>
              <div class="rounded bg-gray-50 px-2 py-1 dark:bg-dark-800">
                <span class="font-medium">{{ t('admin.announcements.form.endsAt') }}:</span>
                <span class="ml-1">{{ row.ends_at ? formatDateTime(row.ends_at) : t('admin.announcements.timeNever') }}</span>
              </div>
            </div>
          </template>

          <template #cell-createdAt="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center space-x-1">
              <button
                @click="openReadStatus(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/20 dark:hover:text-blue-400"
                :title="t('admin.announcements.readStatus')"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button
                @click="openEditDialog(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-600 dark:hover:text-gray-300"
                :title="t('common.edit')"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                @click="handleDelete(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('empty.noData')"
              :description="t('admin.announcements.description')"
              :action-text="t('admin.announcements.createAnnouncement')"
              @action="openCreateDialog"
            />
          </template>
          </DataTable>
        </div>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Create/Edit Dialog -->
    <BaseDialog
      :show="showEditDialog"
      :title="isEditing ? t('admin.announcements.editAnnouncement') : t('admin.announcements.createAnnouncement')"
      width="wide"
      @close="closeEdit"
    >
      <form id="announcement-form" @submit.prevent="handleSave" class="space-y-4">
        <div>
          <label class="input-label">{{ t('admin.announcements.form.title') }}</label>
          <input v-model="form.title" type="text" class="input" required />
        </div>

        <div>
          <label class="input-label">{{ t('admin.announcements.form.content') }}</label>
          <textarea v-model="form.content" rows="6" class="input" required></textarea>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.announcements.form.status') }}</label>
            <Select v-model="form.status" :options="statusOptions" />
          </div>
          <div></div>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.announcements.form.startsAt') }}</label>
            <input v-model="form.starts_at_str" type="datetime-local" class="input" />
            <p class="input-hint">{{ t('admin.announcements.form.startsAtHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.announcements.form.endsAt') }}</label>
            <input v-model="form.ends_at_str" type="datetime-local" class="input" />
            <p class="input-hint">{{ t('admin.announcements.form.endsAtHint') }}</p>
          </div>
        </div>

        <AnnouncementTargetingEditor
          v-model="form.targeting"
          :groups="subscriptionGroups"
        />
      </form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" @click="closeEdit" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button type="submit" form="announcement-form" :disabled="saving" class="btn btn-primary">
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.announcements.deleteAnnouncement')"
      :message="t('admin.announcements.deleteConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Read Status Dialog -->
    <AnnouncementReadStatusDialog
      :show="showReadStatusDialog"
      :announcement-id="readStatusAnnouncementId"
      @close="showReadStatusDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { formatDateTime, formatDateTimeLocalInput, parseDateTimeLocalInput } from '@/utils/format'
import type { AdminGroup, Announcement, AnnouncementTargeting } from '@/types'
import type { Column } from '@/components/common/types'

import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'

import AnnouncementTargetingEditor from '@/components/admin/announcements/AnnouncementTargetingEditor.vue'
import AnnouncementReadStatusDialog from '@/components/admin/announcements/AnnouncementReadStatusDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const announcements = ref<Announcement[]>([])
const loading = ref(false)

const filters = reactive({
  status: '',
})
const searchQuery = ref('')

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const statusFilterOptions = computed(() => [
  { value: '', label: t('admin.announcements.allStatus') },
  { value: 'draft', label: t('admin.announcements.statusLabels.draft') },
  { value: 'active', label: t('admin.announcements.statusLabels.active') },
  { value: 'archived', label: t('admin.announcements.statusLabels.archived') }
])

const statusOptions = computed(() => [
  { value: 'draft', label: t('admin.announcements.statusLabels.draft') },
  { value: 'active', label: t('admin.announcements.statusLabels.active') },
  { value: 'archived', label: t('admin.announcements.statusLabels.archived') }
])

const columns = computed<Column[]>(() => [
  { key: 'title', label: t('admin.announcements.columns.title') },
  { key: 'status', label: t('admin.announcements.columns.status') },
  { key: 'targeting', label: t('admin.announcements.columns.targeting') },
  { key: 'timeRange', label: t('admin.announcements.columns.timeRange') },
  { key: 'createdAt', label: t('admin.announcements.columns.createdAt') },
  { key: 'actions', label: t('admin.announcements.columns.actions') }
])

const activeCount = computed(() => announcements.value.filter((item) => item.status === 'active').length)
const draftCount = computed(() => announcements.value.filter((item) => item.status === 'draft').length)
const archivedCount = computed(() => announcements.value.filter((item) => item.status === 'archived').length)

const statusLabel = (status: string) => {
  if (status === 'draft') return t('admin.announcements.statusLabels.draft')
  if (status === 'active') return t('admin.announcements.statusLabels.active')
  if (status === 'archived') return t('admin.announcements.statusLabels.archived')
  return status
}

const statusClass = (status: string) => {
  if (status === 'active') {
    return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  }
  if (status === 'draft') {
    return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
  }
  return 'bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300'
}

const contentPreview = (content: string) => {
  if (!content) return '-'
  return content
    .replace(/!\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/\[[^\]]*\]\([^)]*\)/g, ' ')
    .replace(/[>#*_`~\-|]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

const setStatusFilter = (status: string) => {
  if (filters.status === status) return
  filters.status = status
  pagination.page = 1
  loadAnnouncements()
}

const targetingSummary = (targeting: AnnouncementTargeting) => {
  const anyOf = targeting?.any_of ?? []
  if (!anyOf || anyOf.length === 0) return t('admin.announcements.targetingSummaryAll')
  return t('admin.announcements.targetingSummaryCustom', { groups: anyOf.length })
}

// ===== CRUD / list =====
let currentController: AbortController | null = null

async function loadAnnouncements() {
  if (currentController) currentController.abort()
  currentController = new AbortController()

  try {
    loading.value = true
    const res = await adminAPI.announcements.list(pagination.page, pagination.page_size, {
      status: filters.status || undefined,
      search: searchQuery.value || undefined
    })

    announcements.value = res.items
    pagination.total = res.total
    pagination.pages = res.pages
    pagination.page = res.page
    pagination.page_size = res.page_size
  } catch (error: any) {
    if (currentController.signal.aborted || error?.name === 'AbortError') return
    console.error('Error loading announcements:', error)
    appStore.showError(error.response?.data?.detail || t('admin.announcements.failedToLoad'))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadAnnouncements()
}

function handlePageSizeChange(pageSize: number) {
  pagination.page_size = pageSize
  pagination.page = 1
  loadAnnouncements()
}

function handleStatusChange() {
  pagination.page = 1
  loadAnnouncements()
}

let searchDebounceTimer: number | null = null
function handleSearch() {
  if (searchDebounceTimer) window.clearTimeout(searchDebounceTimer)
  searchDebounceTimer = window.setTimeout(() => {
    pagination.page = 1
    loadAnnouncements()
  }, 300)
}

// ===== Create/Edit dialog =====
const showEditDialog = ref(false)
const saving = ref(false)
const editingAnnouncement = ref<Announcement | null>(null)

const isEditing = computed(() => !!editingAnnouncement.value)

const form = reactive({
  title: '',
  content: '',
  status: 'draft',
  starts_at_str: '',
  ends_at_str: '',
  targeting: { any_of: [] } as AnnouncementTargeting
})

const subscriptionGroups = ref<AdminGroup[]>([])

async function loadSubscriptionGroups() {
  try {
    const all = await adminAPI.groups.getAll()
    subscriptionGroups.value = (all || []).filter((g) => g.subscription_type === 'subscription')
  } catch (error: any) {
    console.error('Error loading groups:', error)
    // not fatal
  }
}

function resetForm() {
  form.title = ''
  form.content = ''
  form.status = 'draft'
  form.starts_at_str = ''
  form.ends_at_str = ''
  form.targeting = { any_of: [] }
}

function fillFormFromAnnouncement(a: Announcement) {
  form.title = a.title
  form.content = a.content
  form.status = a.status

  // Backend returns RFC3339 strings
  form.starts_at_str = a.starts_at ? formatDateTimeLocalInput(Math.floor(new Date(a.starts_at).getTime() / 1000)) : ''
  form.ends_at_str = a.ends_at ? formatDateTimeLocalInput(Math.floor(new Date(a.ends_at).getTime() / 1000)) : ''

  form.targeting = a.targeting ?? { any_of: [] }
}

function openCreateDialog() {
  editingAnnouncement.value = null
  resetForm()
  showEditDialog.value = true
}

function openEditDialog(row: Announcement) {
  editingAnnouncement.value = row
  fillFormFromAnnouncement(row)
  showEditDialog.value = true
}

function closeEdit() {
  showEditDialog.value = false
  editingAnnouncement.value = null
}

function buildCreatePayload() {
  const startsAt = parseDateTimeLocalInput(form.starts_at_str)
  const endsAt = parseDateTimeLocalInput(form.ends_at_str)

  return {
    title: form.title,
    content: form.content,
    status: form.status as any,
    targeting: form.targeting,
    starts_at: startsAt ?? undefined,
    ends_at: endsAt ?? undefined
  }
}

function buildUpdatePayload(original: Announcement) {
  const payload: any = {}

  if (form.title !== original.title) payload.title = form.title
  if (form.content !== original.content) payload.content = form.content
  if (form.status !== original.status) payload.status = form.status

  // starts_at / ends_at: distinguish unchanged vs clear(0) vs set
  const originalStarts = original.starts_at ? Math.floor(new Date(original.starts_at).getTime() / 1000) : null
  const originalEnds = original.ends_at ? Math.floor(new Date(original.ends_at).getTime() / 1000) : null

  const newStarts = parseDateTimeLocalInput(form.starts_at_str)
  const newEnds = parseDateTimeLocalInput(form.ends_at_str)

  if (newStarts !== originalStarts) {
    payload.starts_at = newStarts === null ? 0 : newStarts
  }
  if (newEnds !== originalEnds) {
    payload.ends_at = newEnds === null ? 0 : newEnds
  }

  // targeting: do shallow compare by JSON
  if (JSON.stringify(form.targeting ?? {}) !== JSON.stringify(original.targeting ?? {})) {
    payload.targeting = form.targeting
  }

  return payload
}

async function handleSave() {
  // Frontend validation for targeting (to avoid ANNOUNCEMENT_INVALID_TARGET)
  const anyOf = form.targeting?.any_of ?? []
  if (anyOf.length > 50) {
    appStore.showError(t('admin.announcements.failedToCreate'))
    return
  }
  for (const g of anyOf) {
    const allOf = g?.all_of ?? []
    if (allOf.length > 50) {
      appStore.showError(t('admin.announcements.failedToCreate'))
      return
    }
  }

  saving.value = true
  try {
    if (!editingAnnouncement.value) {
      const payload = buildCreatePayload()
      await adminAPI.announcements.create(payload)
      appStore.showSuccess(t('common.success'))
      showEditDialog.value = false
      await loadAnnouncements()
      return
    }

    const original = editingAnnouncement.value
    const payload = buildUpdatePayload(original)
    await adminAPI.announcements.update(original.id, payload)
    appStore.showSuccess(t('common.success'))
    showEditDialog.value = false
    editingAnnouncement.value = null
    await loadAnnouncements()
  } catch (error: any) {
    console.error('Failed to save announcement:', error)
    appStore.showError(error.response?.data?.detail || (editingAnnouncement.value ? t('admin.announcements.failedToUpdate') : t('admin.announcements.failedToCreate')))
  } finally {
    saving.value = false
  }
}

// ===== Delete =====
const showDeleteDialog = ref(false)
const deletingAnnouncement = ref<Announcement | null>(null)

function handleDelete(row: Announcement) {
  deletingAnnouncement.value = row
  showDeleteDialog.value = true
}

async function confirmDelete() {
  if (!deletingAnnouncement.value) return

  try {
    await adminAPI.announcements.delete(deletingAnnouncement.value.id)
    appStore.showSuccess(t('common.success'))
    showDeleteDialog.value = false
    deletingAnnouncement.value = null
    await loadAnnouncements()
  } catch (error: any) {
    console.error('Failed to delete announcement:', error)
    appStore.showError(error.response?.data?.detail || t('admin.announcements.failedToDelete'))
  }
}

// ===== Read status =====
const showReadStatusDialog = ref(false)
const readStatusAnnouncementId = ref<number | null>(null)

function openReadStatus(row: Announcement) {
  readStatusAnnouncementId.value = row.id
  showReadStatusDialog.value = true
}

onMounted(async () => {
  await loadSubscriptionGroups()
  await loadAnnouncements()
})
</script>
