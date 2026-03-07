<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <!-- Left: Search + Filters -->
          <div class="flex-1 sm:max-w-64">
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.redeem.searchCodes')"
              class="input"
              @input="handleSearch"
            />
          </div>
          <Select
            v-model="filters.type"
            :options="filterTypeOptions"
            class="w-36"
            @change="loadCodes"
          />
          <Select
            v-model="filters.status"
            :options="filterStatusOptions"
            class="w-36"
            @change="loadCodes"
          />
          <div class="w-40">
            <input
              v-model="filters.category"
              type="text"
              :placeholder="t('admin.redeem.categoryFilter')"
              class="input"
              @input="handleSearch"
            />
          </div>
          <button
            class="btn btn-secondary"
            @click="toggleDistributorFilter"
          >
            {{ filters.category === 'distributor' ? t('admin.redeem.clearDistributorFilter') : t('admin.redeem.onlyDistributor') }}
          </button>

          <!-- Right: Action buttons -->
          <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
            <button @click="toggleCategoryStats" class="btn btn-secondary">
              {{ showCategoryStats ? t('admin.redeem.hideCategoryStats') : t('admin.redeem.showCategoryStats') }}
            </button>
            <button
              @click="loadCodes"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button @click="handleExportCodes" class="btn btn-secondary">
              {{ t('admin.redeem.exportCsv') }}
            </button>
            <button @click="showGenerateDialog = true" class="btn btn-primary">
              {{ t('admin.redeem.generateCodes') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <div
          v-if="showCategoryStats"
          class="mb-4 rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-800"
        >
          <div class="mb-2 text-sm font-medium text-gray-900 dark:text-white">
            {{ t('admin.redeem.categoryStats') }}
          </div>
          <div v-if="categoryStats.length === 0" class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.redeem.noCategoryStats') }}
          </div>
          <div v-else class="overflow-x-auto">
            <table class="min-w-full text-sm">
              <thead>
                <tr class="text-left text-gray-500 dark:text-gray-400">
                  <th class="px-2 py-1">{{ t('admin.redeem.columns.category') }}</th>
                  <th class="px-2 py-1">{{ t('admin.redeem.columns.total') }}</th>
                  <th class="px-2 py-1">{{ t('admin.redeem.used') }}</th>
                  <th class="px-2 py-1">{{ t('admin.redeem.unused') }}</th>
                  <th class="px-2 py-1">{{ t('admin.redeem.status.expired') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in categoryStats" :key="item.category" class="border-t border-gray-100 dark:border-dark-700">
                  <td class="px-2 py-1 text-gray-900 dark:text-gray-100">
                    {{ item.category || t('admin.redeem.ungrouped') }}
                  </td>
                  <td class="px-2 py-1">{{ item.total }}</td>
                  <td class="px-2 py-1">{{ item.used }}</td>
                  <td class="px-2 py-1">{{ item.unused }}</td>
                  <td class="px-2 py-1">{{ item.expired }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div
          v-if="invitationStats"
          class="mb-4 rounded-lg border border-indigo-100 bg-indigo-50 p-4 text-xs text-gray-700 dark:border-indigo-900/40 dark:bg-indigo-900/20 dark:text-indigo-100"
        >
          <div class="mb-1 font-medium text-indigo-700 dark:text-indigo-100">
            {{ t('admin.redeem.invitationStatsHint') }}
          </div>
          <div class="flex flex-wrap gap-4">
            <div>
              <span class="font-semibold">{{ t('admin.redeem.invitationTotal') }}:</span>
              <span class="ml-1">{{ invitationStats.total }}</span>
            </div>
            <div>
              <span class="font-semibold">{{ t('admin.redeem.invitationUsed') }}:</span>
              <span class="ml-1">{{ invitationStats.used }}</span>
            </div>
            <div>
              <span class="font-semibold">{{ t('admin.redeem.invitationUnused') }}:</span>
              <span class="ml-1">{{ invitationStats.unused }}</span>
            </div>
          </div>
        </div>

        <DataTable :columns="columns" :data="codes" :loading="loading">
          <template #cell-code="{ value }">
            <div class="flex items-center space-x-2">
              <code class="font-mono text-sm text-gray-900 dark:text-gray-100">{{ value }}</code>
              <button
                @click="copyToClipboard(value)"
                :class="[
                  'flex items-center transition-colors',
                  copiedCode === value
                    ? 'text-green-500'
                    : 'text-gray-400 hover:text-gray-600 dark:hover:text-gray-300'
                ]"
                :title="copiedCode === value ? t('admin.redeem.copied') : t('keys.copyToClipboard')"
              >
                <Icon v-if="copiedCode !== value" name="copy" size="sm" :stroke-width="2" />
                <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </button>
            </div>
          </template>

          <template #cell-type="{ value }">
            <span
              :class="[
                'badge',
                value === 'balance'
                  ? 'badge-success'
                  : value === 'subscription'
                    ? 'badge-warning'
                    : 'badge-primary'
              ]"
            >
              {{ t('admin.redeem.types.' + value) }}
            </span>
          </template>

          <template #cell-value="{ value, row }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">
              <template v-if="row.type === 'balance'">${{ value.toFixed(2) }}</template>
              <template v-else-if="row.type === 'subscription'">
                {{ row.validity_days || 30 }} {{ t('admin.redeem.days') }}
                <span v-if="row.group" class="ml-1 text-xs text-gray-500 dark:text-gray-400"
                  >({{ row.group.name }})</span
                >
              </template>
              <template v-else>{{ value }}</template>
            </span>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="[
                'badge',
                value === 'unused'
                  ? 'badge-success'
                  : value === 'used'
                    ? 'badge-gray'
                    : 'badge-danger'
              ]"
            >
              {{ t('admin.redeem.status.' + value) }}
            </span>
          </template>

          <template #cell-category="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ value || t('admin.redeem.ungrouped') }}</span>
          </template>

          <template #cell-notes="{ value }">
            <span class="text-sm text-gray-500 dark:text-gray-400">{{ value || '-' }}</span>
          </template>

          <template #cell-used_by="{ value, row }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ row.user?.email || (value ? t('admin.redeem.userPrefix', { id: value }) : '-') }}
            </span>
          </template>

          <template #cell-used_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{
              value ? formatDateTime(value) : '-'
            }}</span>
          </template>

      <template #cell-actions="{ row }">
        <div class="flex items-center space-x-2">
          <button
            v-if="row.type === 'invitation' && row.status === 'used'"
            @click="handleViewInvitationImpact(row)
            "
            class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-indigo-50 hover:text-indigo-600 dark:hover:bg-indigo-900/20 dark:hover:text-indigo-300"
          >
            <Icon name="chartBar" size="sm" />
            <span class="text-xs">{{ t('admin.redeem.viewInvitationImpact') }}</span>
          </button>
          <button
            v-if="row.status === 'unused'"
            @click="handleDelete(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              >
                <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
                <span class="text-xs">{{ t('common.delete') }}</span>
              </button>
              <span v-else class="text-gray-400 dark:text-dark-500">-</span>
            </div>
          </template>
        </DataTable>
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

        <!-- Batch Actions -->
        <div v-if="filters.status === 'unused'" class="flex justify-end">
          <button @click="showDeleteUnusedDialog = true" class="btn btn-danger">
            {{ t('admin.redeem.deleteAllUnused') }}
          </button>
        </div>
      </template>
    </TablePageLayout>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.redeem.deleteCode')"
      :message="t('admin.redeem.deleteCodeConfirm')"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Delete Unused Codes Dialog -->
    <ConfirmDialog
      :show="showDeleteUnusedDialog"
      :title="t('admin.redeem.deleteAllUnused')"
      :message="t('admin.redeem.deleteAllUnusedConfirm')"
      :confirm-text="t('admin.redeem.deleteAll')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmDeleteUnused"
      @cancel="showDeleteUnusedDialog = false"
    />

    <!-- Generate Codes Dialog -->
    <Teleport to="body">
      <div v-if="showGenerateDialog" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="fixed inset-0 bg-black/50" @click="showGenerateDialog = false"></div>
        <div
          class="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-xl dark:bg-dark-800"
        >
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.redeem.generateCodesTitle') }}
          </h2>
          <form @submit.prevent="handleGenerateCodes" class="space-y-4">
            <div>
              <label class="input-label">{{ t('admin.redeem.codeType') }}</label>
              <Select v-model="generateForm.type" :options="typeOptions" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.redeem.category') }}</label>
              <input v-model="generateForm.category" type="text" maxlength="64" class="input" />
            </div>
            <div>
              <label class="input-label">{{ t('admin.redeem.notes') }}</label>
              <textarea v-model="generateForm.notes" maxlength="500" class="input min-h-20"></textarea>
            </div>
            <!-- 余额/并发类型：显示数值输入 -->
            <div v-if="generateForm.type !== 'subscription' && generateForm.type !== 'invitation'">
              <label class="input-label">
                {{
                  generateForm.type === 'balance'
                    ? t('admin.redeem.amount')
                    : t('admin.redeem.columns.value')
                }}
              </label>
              <input
                v-model.number="generateForm.value"
                type="number"
                :step="generateForm.type === 'balance' ? '0.01' : '1'"
                :min="generateForm.type === 'balance' ? '0.01' : '1'"
                required
                class="input"
              />
            </div>
            <!-- 邀请码类型：显示提示信息 -->
            <div v-if="generateForm.type === 'invitation'" class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/20">
              <p class="text-sm text-blue-700 dark:text-blue-300">
                {{ t('admin.redeem.invitationHint') }}
              </p>
            </div>
            <!-- 订阅类型：显示分组选择和有效天数 -->
            <template v-if="generateForm.type === 'subscription'">
              <div>
                <label class="input-label">{{ t('admin.redeem.selectGroup') }}</label>
                <Select
                  v-model="generateForm.group_id"
                  :options="subscriptionGroupOptions"
                  :placeholder="t('admin.redeem.selectGroupPlaceholder')"
                >
                  <template #selected="{ option }">
                    <GroupBadge
                      v-if="option"
                      :name="(option as unknown as GroupOption).label"
                      :platform="(option as unknown as GroupOption).platform"
                      :subscription-type="(option as unknown as GroupOption).subscriptionType"
                      :rate-multiplier="(option as unknown as GroupOption).rate"
                    />
                    <span v-else class="text-gray-400">{{
                      t('admin.redeem.selectGroupPlaceholder')
                    }}</span>
                  </template>
                  <template #option="{ option, selected }">
                    <GroupOptionItem
                      :name="(option as unknown as GroupOption).label"
                      :platform="(option as unknown as GroupOption).platform"
                      :subscription-type="(option as unknown as GroupOption).subscriptionType"
                      :rate-multiplier="(option as unknown as GroupOption).rate"
                      :description="(option as unknown as GroupOption).description"
                      :selected="selected"
                    />
                  </template>
                </Select>
              </div>
              <div>
                <label class="input-label">{{ t('admin.redeem.validityDays') }}</label>
                <input
                  v-model.number="generateForm.validity_days"
                  type="number"
                  min="1"
                  max="365"
                  required
                  class="input"
                />
              </div>
            </template>
            <div>
              <label class="input-label">{{ t('admin.redeem.count') }}</label>
              <input
                v-model.number="generateForm.count"
                type="number"
                min="1"
                max="100"
                required
                class="input"
              />
            </div>
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" @click="showGenerateDialog = false" class="btn btn-secondary">
                {{ t('common.cancel') }}
              </button>
              <button type="submit" :disabled="generating" class="btn btn-primary">
                {{ generating ? t('admin.redeem.generating') : t('admin.redeem.generate') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Invitation Impact Dialog -->
    <Teleport to="body">
      <div
        v-if="showInvitationImpactDialog"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
      >
        <div class="fixed inset-0 bg-black/50" @click="closeInvitationImpact"></div>
        <div
          class="relative z-10 w-full max-w-3xl rounded-xl bg-white shadow-xl dark:bg-dark-800"
        >
          <div
            class="flex items-center justify-between border-b border-gray-200 px-5 py-4 dark:border-dark-600"
          >
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                {{ t('admin.redeem.invitationImpactTitle') }}
              </h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                <span v-if="currentInvitationCode">
                  {{ t('admin.redeem.invitationCodeLabel') }}
                  <code class="ml-1 font-mono text-xs">
                    {{ currentInvitationCode.code }}
                  </code>
                </span>
              </p>
            </div>
            <button
              @click="closeInvitationImpact"
              class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-300"
            >
              <Icon name="x" size="md" :stroke-width="2" />
            </button>
          </div>

          <div class="max-h-[70vh] overflow-y-auto px-5 py-4 text-sm">
            <div v-if="invitationImpactLoading" class="flex items-center justify-center py-8">
              <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
            </div>
            <div v-else-if="!invitationImpact" class="py-8 text-center text-gray-500 dark:text-gray-400">
              {{ t('admin.redeem.invitationImpactEmpty') }}
            </div>
            <div v-else class="space-y-4">
              <div class="grid grid-cols-1 gap-3 md:grid-cols-3">
                <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700/60">
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.redeem.registeredUser') }}
                  </p>
                  <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                    {{ invitationImpact.used_by_email || '-' }}
                  </p>
                </div>
                <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700/60">
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.redeem.subscriptionRedeemsTotal') }}
                  </p>
                  <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                    {{ invitationImpact.subscription_redeems_total }}
                  </p>
                </div>
                <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700/60">
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.redeem.invitationCategory') }}
                  </p>
                  <p class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                    {{ invitationImpact.category || t('admin.redeem.ungrouped') }}
                  </p>
                </div>
              </div>

              <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div>
                  <h3 class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('admin.redeem.byRedeemCategory') }}
                  </h3>
                  <div
                    v-if="!invitationImpact.by_category || invitationImpact.by_category.length === 0"
                    class="rounded-lg border border-dashed border-gray-200 p-3 text-xs text-gray-400 dark:border-dark-600 dark:text-gray-500"
                  >
                    {{ t('admin.redeem.noSubscriptionRedeems') }}
                  </div>
                  <table
                    v-else
                    class="min-w-full divide-y divide-gray-200 overflow-hidden rounded-lg border border-gray-200 text-xs dark:divide-dark-700 dark:border-dark-700"
                  >
                    <thead class="bg-gray-50 dark:bg-dark-800">
                      <tr>
                        <th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-400">
                          {{ t('admin.redeem.columns.category') }}
                        </th>
                        <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">
                          {{ t('admin.redeem.redeemCount') }}
                        </th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
                      <tr v-for="item in invitationImpact.by_category" :key="item.category">
                        <td class="px-3 py-1.5 text-gray-900 dark:text-gray-100">
                          {{ item.category || t('admin.redeem.ungrouped') }}
                        </td>
                        <td class="px-3 py-1.5 text-right text-gray-700 dark:text-gray-300">
                          {{ item.redeem_count }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <div>
                  <h3 class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                    {{ t('admin.redeem.bySubscriptionGroup') }}
                  </h3>
                  <div
                    v-if="!invitationImpact.by_group || invitationImpact.by_group.length === 0"
                    class="rounded-lg border border-dashed border-gray-200 p-3 text-xs text-gray-400 dark:border-dark-600 dark:text-gray-500"
                  >
                    {{ t('admin.redeem.noSubscriptionRedeems') }}
                  </div>
                  <table
                    v-else
                    class="min-w-full divide-y divide-gray-200 overflow-hidden rounded-lg border border-gray-200 text-xs dark:divide-dark-700 dark:border-dark-700"
                  >
                    <thead class="bg-gray-50 dark:bg-dark-800">
                      <tr>
                        <th class="px-3 py-2 text-left font-medium text-gray-500 dark:text-gray-400">
                          {{ t('keys.group') }}
                        </th>
                        <th class="px-3 py-2 text-right font-medium text-gray-500 dark:text-gray-400">
                          {{ t('admin.redeem.redeemCount') }}
                        </th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-gray-100 bg-white dark:divide-dark-700 dark:bg-dark-900">
                      <tr v-for="item in invitationImpact.by_group" :key="item.group_id">
                        <td class="px-3 py-1.5 text-gray-900 dark:text-gray-100">
                          {{ item.group_name || ('#' + item.group_id) }}
                        </td>
                        <td class="px-3 py-1.5 text-right text-gray-700 dark:text-gray-300">
                          {{ item.redeem_count }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>

              <div>
                <h3 class="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  {{ t('admin.redeem.redeemDetails') }}
                </h3>
                <div
                  v-if="!invitationImpact.redeems || invitationImpact.redeems.length === 0"
                  class="rounded-lg border border-dashed border-gray-200 p-3 text-xs text-gray-400 dark:border-dark-600 dark:text-gray-500"
                >
                  {{ t('admin.redeem.noSubscriptionRedeems') }}
                </div>
                <div v-else class="space-y-2">
                  <div
                    v-for="item in invitationImpact.redeems"
                    :key="item.redeem_code_id"
                    class="flex items-center justify-between rounded-lg border border-gray-200 p-3 text-xs dark:border-dark-700"
                  >
                    <div>
                      <p class="font-mono text-gray-900 dark:text-gray-100">
                        {{ item.code }}
                      </p>
                      <p class="mt-0.5 text-[11px] text-gray-500 dark:text-gray-400">
                        <span v-if="item.category">
                          {{ t('admin.redeem.columns.category') }}: {{ item.category }}
                        </span>
                        <span v-if="item.group_name" class="ml-2">
                          {{ t('keys.group') }}: {{ item.group_name }}
                        </span>
                      </p>
                    </div>
                    <div class="text-right">
                      <p class="text-[11px] text-gray-500 dark:text-gray-400">
                        {{ item.used_at ? formatDateTime(item.used_at) : '-' }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            class="flex justify-end gap-2 rounded-b-xl border-t border-gray-200 bg-gray-50 px-5 py-3 text-xs dark:border-dark-600 dark:bg-dark-700/50"
          >
            <button type="button" @click="closeInvitationImpact" class="btn btn-secondary btn-sm">
              {{ t('common.close') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Generated Codes Result Dialog -->
    <Teleport to="body">
      <div v-if="showResultDialog" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="fixed inset-0 bg-black/50" @click="closeResultDialog"></div>
        <div class="relative z-10 w-full max-w-lg rounded-xl bg-white shadow-xl dark:bg-dark-800">
          <!-- Header -->
          <div
            class="flex items-center justify-between border-b border-gray-200 px-5 py-4 dark:border-dark-600"
          >
            <div class="flex items-center gap-3">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/30"
              >
                <svg
                  class="h-5 w-5 text-green-600 dark:text-green-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
              <div>
                <h2 class="text-base font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.redeem.generatedSuccessfully') }}
                </h2>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.redeem.codesCreated', { count: generatedCodes.length }) }}
                </p>
              </div>
            </div>
            <button
              @click="closeResultDialog"
              class="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-300"
            >
              <Icon name="x" size="md" :stroke-width="2" />
            </button>
          </div>
          <!-- Content -->
          <div class="p-5">
            <div class="relative">
              <textarea
                readonly
                :value="generatedCodesText"
                :style="{ height: textareaHeight }"
                class="w-full resize-none rounded-lg border border-gray-200 bg-gray-50 p-3 font-mono text-sm text-gray-800 focus:outline-none dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200"
              ></textarea>
            </div>
          </div>
          <!-- Footer -->
          <div
            class="flex justify-end gap-2 rounded-b-xl border-t border-gray-200 bg-gray-50 px-5 py-4 dark:border-dark-600 dark:bg-dark-700/50"
          >
            <button
              @click="copyGeneratedCodes"
              :class="[
                'btn flex items-center gap-2 transition-all',
                copiedAll ? 'btn-success' : 'btn-secondary'
              ]"
            >
              <Icon v-if="!copiedAll" name="copy" size="sm" :stroke-width="2" />
              <svg v-else class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M5 13l4 4L19 7"
                />
              </svg>
              {{ copiedAll ? t('admin.redeem.copied') : t('admin.redeem.copyAll') }}
            </button>
            <button @click="downloadGeneratedCodes" class="btn btn-primary flex items-center gap-2">
              <Icon name="download" size="sm" :stroke-width="2" />
              {{ t('admin.redeem.download') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { adminAPI } from '@/api/admin'
import { formatDateTime } from '@/utils/format'
import type {
  RedeemCode,
  RedeemCodeType,
  Group,
  GroupPlatform,
  SubscriptionType,
  InvitationRedeemImpactStats
} from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import GroupOptionItem from '@/components/common/GroupOptionItem.vue'
import Icon from '@/components/icons/Icon.vue'

interface RedeemCategoryStat {
  category: string
  total: number
  unused: number
  used: number
  expired: number
  total_value: number
  used_value: number
}

interface InvitationStats {
  total: number
  used: number
  unused: number
}

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard: clipboardCopy } = useClipboard()
const CATEGORY_STATS_VISIBLE_KEY = 'redeem_category_stats_visible'

interface GroupOption {
  value: number
  label: string
  description: string | null
  platform: GroupPlatform
  subscriptionType: SubscriptionType
  rate: number
}

const showGenerateDialog = ref(false)
const showResultDialog = ref(false)
const generatedCodes = ref<RedeemCode[]>([])
const subscriptionGroups = ref<Group[]>([])

// 订阅类型分组选项
const subscriptionGroupOptions = computed(() => {
  return subscriptionGroups.value
    .filter((g) => g.subscription_type === 'subscription')
    .map((g) => ({
      value: g.id,
      label: g.name,
      description: g.description,
      platform: g.platform,
      subscriptionType: g.subscription_type,
      rate: g.rate_multiplier
    }))
})

const generatedCodesText = computed(() => {
  return generatedCodes.value.map((code) => code.code).join('\n')
})

const textareaHeight = computed(() => {
  const lineCount = generatedCodes.value.length
  const lineHeight = 24 // approximate line height in px
  const padding = 24 // top + bottom padding
  const minHeight = 60
  const maxHeight = 240
  const calculatedHeight = Math.min(
    Math.max(lineCount * lineHeight + padding, minHeight),
    maxHeight
  )
  return `${calculatedHeight}px`
})

const copiedAll = ref(false)
const showCategoryStats = ref(true)

const closeResultDialog = () => {
  showResultDialog.value = false
  generatedCodes.value = []
  copiedAll.value = false
}

const copyGeneratedCodes = async () => {
  try {
    await navigator.clipboard.writeText(generatedCodesText.value)
    copiedAll.value = true
    setTimeout(() => {
      copiedAll.value = false
    }, 2000)
  } catch (error) {
    appStore.showError(t('admin.redeem.failedToCopy'))
  }
}

const downloadGeneratedCodes = () => {
  const blob = new Blob([generatedCodesText.value], { type: 'text/plain' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `redeem-codes-${new Date().toISOString().split('T')[0]}.txt`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

const columns = computed<Column[]>(() => [
  { key: 'code', label: t('admin.redeem.columns.code') },
  { key: 'type', label: t('admin.redeem.columns.type'), sortable: true },
  { key: 'value', label: t('admin.redeem.columns.value'), sortable: true },
  { key: 'status', label: t('admin.redeem.columns.status'), sortable: true },
  { key: 'category', label: t('admin.redeem.columns.category') },
  { key: 'notes', label: t('admin.redeem.columns.notes') },
  { key: 'used_by', label: t('admin.redeem.columns.usedBy') },
  { key: 'used_at', label: t('admin.redeem.columns.usedAt'), sortable: true },
  { key: 'actions', label: t('admin.redeem.columns.actions') }
])

const typeOptions = computed(() => [
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') },
  { value: 'subscription', label: t('admin.redeem.subscription') },
  { value: 'invitation', label: t('admin.redeem.invitation') }
])

const filterTypeOptions = computed(() => [
  { value: '', label: t('admin.redeem.allTypes') },
  { value: 'balance', label: t('admin.redeem.balance') },
  { value: 'concurrency', label: t('admin.redeem.concurrency') },
  { value: 'subscription', label: t('admin.redeem.subscription') },
  { value: 'invitation', label: t('admin.redeem.invitation') }
])

const filterStatusOptions = computed(() => [
  { value: '', label: t('admin.redeem.allStatus') },
  { value: 'unused', label: t('admin.redeem.unused') },
  { value: 'used', label: t('admin.redeem.used') },
  { value: 'expired', label: t('admin.redeem.status.expired') },
  { value: 'revoked', label: t('admin.redeem.status.revoked') }
])

const codes = ref<RedeemCode[]>([])
const loading = ref(false)
const generating = ref(false)
const searchQuery = ref('')
const filters = reactive({
  type: '',
  status: '',
  category: ''
})
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

let abortController: AbortController | null = null

const showDeleteDialog = ref(false)
const showDeleteUnusedDialog = ref(false)
const deletingCode = ref<RedeemCode | null>(null)
const copiedCode = ref<string | null>(null)
const categoryStats = ref<RedeemCategoryStat[]>([])
const invitationStats = ref<InvitationStats | null>(null)
const showInvitationImpactDialog = ref(false)
const invitationImpactLoading = ref(false)
const currentInvitationCode = ref<RedeemCode | null>(null)
const invitationImpact = ref<InvitationRedeemImpactStats | null>(null)

const generateForm = reactive({
  type: 'balance' as RedeemCodeType,
  value: 10,
  count: 1,
  category: '',
  notes: '',
  group_id: null as number | null,
  validity_days: 30
})

// 监听类型变化，邀请码类型时自动设置 value 为 0
watch(
  () => generateForm.type,
  (newType) => {
    if (newType === 'invitation') {
      generateForm.value = 0
    } else if (generateForm.value === 0) {
      generateForm.value = 10
    }
  }
)

const loadCodes = async () => {
  if (abortController) {
    abortController.abort()
  }
  const currentController = new AbortController()
  abortController = currentController
  loading.value = true
  try {
    const response = await adminAPI.redeem.list(
      pagination.page,
      pagination.page_size,
      {
        type: filters.type as RedeemCodeType,
        status: filters.status as any,
        search: searchQuery.value || undefined,
        category: filters.category || undefined
      },
      {
        signal: currentController.signal
      }
    )
    if (currentController.signal.aborted) {
      return
    }
    codes.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error: any) {
    if (
      currentController.signal.aborted ||
      error?.name === 'AbortError' ||
      error?.code === 'ERR_CANCELED'
    ) {
      return
    }
    appStore.showError(t('admin.redeem.failedToLoad'))
    console.error('Error loading redeem codes:', error)
  } finally {
    if (abortController === currentController && !currentController.signal.aborted) {
      loading.value = false
      abortController = null
    }
  }
}

let searchTimeout: ReturnType<typeof setTimeout>
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.page = 1
    loadCodes()
  }, 300)
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadCodes()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadCodes()
}

const handleGenerateCodes = async () => {
  // 订阅类型必须选择分组
  if (generateForm.type === 'subscription' && !generateForm.group_id) {
    appStore.showError(t('admin.redeem.groupRequired'))
    return
  }

  generating.value = true
  try {
    const result = await adminAPI.redeem.generate(
      generateForm.count,
      generateForm.type,
      generateForm.value,
      generateForm.notes,
      generateForm.category,
      generateForm.type === 'subscription' ? generateForm.group_id : undefined,
      generateForm.type === 'subscription' ? generateForm.validity_days : undefined
    )
    showGenerateDialog.value = false
    generatedCodes.value = result
    showResultDialog.value = true
    // 重置表单
    generateForm.notes = ''
    generateForm.category = ''
    generateForm.group_id = null
    generateForm.validity_days = 30
    loadCodes()
    loadStats()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToGenerate'))
    console.error('Error generating codes:', error)
  } finally {
    generating.value = false
  }
}

const copyToClipboard = async (text: string) => {
  const success = await clipboardCopy(text, t('admin.redeem.copied'))
  if (success) {
    copiedCode.value = text
    setTimeout(() => {
      copiedCode.value = null
    }, 2000)
  }
}

const handleExportCodes = async () => {
  try {
    const blob = await adminAPI.redeem.exportCodes({
      type: filters.type as RedeemCodeType,
      status: filters.status as any,
      category: filters.category || undefined
    })

    // Create download link
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `redeem-codes-${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    appStore.showSuccess(t('admin.redeem.codesExported'))
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToExport'))
    console.error('Error exporting codes:', error)
  }
}

const handleDelete = (code: RedeemCode) => {
  deletingCode.value = code
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingCode.value) return

  try {
    await adminAPI.redeem.delete(deletingCode.value.id)
    appStore.showSuccess(t('admin.redeem.codeDeleted'))
    showDeleteDialog.value = false
    deletingCode.value = null
    loadCodes()
    loadStats()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToDelete'))
    console.error('Error deleting code:', error)
  }
}

const confirmDeleteUnused = async () => {
  try {
    // Get all unused codes and delete them
    const unusedCodesResponse = await adminAPI.redeem.list(1, 1000, { status: 'unused' })
    const unusedCodeIds = unusedCodesResponse.items.map((code) => code.id)

    if (unusedCodeIds.length === 0) {
      appStore.showInfo(t('admin.redeem.noUnusedCodes'))
      showDeleteUnusedDialog.value = false
      return
    }

    const result = await adminAPI.redeem.batchDelete(unusedCodeIds)
    appStore.showSuccess(t('admin.redeem.codesDeleted', { count: result.deleted }))
    showDeleteUnusedDialog.value = false
    loadCodes()
    loadStats()
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToDeleteUnused'))
    console.error('Error deleting unused codes:', error)
  }
}

const loadStats = async () => {
  try {
    const stats = await adminAPI.redeem.getStats()
    categoryStats.value = stats.by_category || []

    const invitationCategory = stats.by_category.find((item) => item.category === 'invitation')
    if (invitationCategory) {
      invitationStats.value = {
        total: invitationCategory.total,
        used: invitationCategory.used,
        unused: invitationCategory.unused
      }
    } else {
      invitationStats.value = null
    }
  } catch (error) {
    console.error('Error loading redeem stats:', error)
  }
}

const handleViewInvitationImpact = async (code: RedeemCode) => {
  currentInvitationCode.value = code
  showInvitationImpactDialog.value = true
  invitationImpactLoading.value = true
  invitationImpact.value = null
  const requestId = code.id

  try {
    const data = await adminAPI.redeem.getInvitationImpact(code.id)
    if (currentInvitationCode.value?.id !== requestId) {
      return
    }
    invitationImpact.value = data
  } catch (error: any) {
    if (currentInvitationCode.value?.id !== requestId) {
      return
    }
    appStore.showError(error.response?.data?.detail || t('admin.redeem.failedToLoad'))
    console.error('Error loading invitation impact:', error)
  } finally {
    if (currentInvitationCode.value?.id === requestId) {
      invitationImpactLoading.value = false
    }
  }
}

const closeInvitationImpact = () => {
  showInvitationImpactDialog.value = false
  invitationImpactLoading.value = false
  invitationImpact.value = null
  currentInvitationCode.value = null
}

const toggleCategoryStats = () => {
  showCategoryStats.value = !showCategoryStats.value
  localStorage.setItem(CATEGORY_STATS_VISIBLE_KEY, showCategoryStats.value ? '1' : '0')
}

const toggleDistributorFilter = () => {
  filters.category = filters.category === 'distributor' ? '' : 'distributor'
  loadCodes()
}

// 加载订阅类型分组
const loadSubscriptionGroups = async () => {
  try {
    const groups = await adminAPI.groups.getAll()
    subscriptionGroups.value = groups
  } catch (error) {
    console.error('Error loading subscription groups:', error)
  }
}

onMounted(() => {
  const saved = localStorage.getItem(CATEGORY_STATS_VISIBLE_KEY)
  if (saved === '0') {
    showCategoryStats.value = false
  }
  loadCodes()
  loadSubscriptionGroups()
  loadStats()
})

onUnmounted(() => {
  clearTimeout(searchTimeout)
  abortController?.abort()
})
</script>
