<template>
  <div class="card overflow-hidden">
    <div
      class="border-b border-gray-100 bg-gradient-to-r from-primary-500/10 to-primary-600/5 px-6 py-5 dark:border-dark-700 dark:from-primary-500/20 dark:to-primary-600/10"
    >
      <div class="flex items-center gap-4">
        <!-- Avatar -->
        <div
          class="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 text-2xl font-bold text-white shadow-lg shadow-primary-500/20"
        >
          {{ user?.email?.charAt(0).toUpperCase() || 'U' }}
        </div>
        <div class="min-w-0 flex-1">
          <h2 class="truncate text-lg font-semibold text-gray-900 dark:text-white">
            {{ user?.email }}
          </h2>
          <div class="mt-1 flex items-center gap-2">
            <span :class="['badge', user?.role === 'admin' ? 'badge-primary' : 'badge-gray']">
              {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
            </span>
            <span
              :class="['badge', user?.status === 'active' ? 'badge-success' : 'badge-danger']"
            >
              {{ user?.status }}
            </span>
          </div>
        </div>
        <div class="min-w-0 flex-1">
          <div v-if="checkinVisible" class="flex flex-col items-end gap-1">
            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :disabled="checkinButtonDisabled"
              @click="handleDailyCheckin"
            >
              {{ checkinButtonText }}
            </button>
            <p v-if="checkinRewardText" class="text-xs text-gray-500 dark:text-gray-400">
              {{ checkinRewardText }}
            </p>
          </div>
        </div>
      </div>
    </div>
    <div class="px-6 py-4">
      <div class="space-y-3">
        <div class="flex items-center gap-3 text-sm text-gray-600 dark:text-gray-400">
          <Icon name="mail" size="sm" class="text-gray-400 dark:text-gray-500" />
          <span class="truncate">{{ user?.email }}</span>
        </div>
        <div
          v-if="user?.username"
          class="flex items-center gap-3 text-sm text-gray-600 dark:text-gray-400"
        >
          <Icon name="user" size="sm" class="text-gray-400 dark:text-gray-500" />
          <span class="truncate">{{ user.username }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { redeemAPI } from '@/api'
import type { DailyCheckinStatus } from '@/api/redeem'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/types'

const props = defineProps<{
  user: User | null
}>()

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const checkinStatus = ref<DailyCheckinStatus | null>(null)
const loadingCheckinStatus = ref(false)
const checkinSubmitting = ref(false)

const checkinVisible = computed(() => checkinStatus.value?.enabled === true)

const checkinButtonDisabled = computed(
  () =>
    loadingCheckinStatus.value ||
    checkinSubmitting.value ||
    checkinStatus.value?.checked_in_today === true
)

const checkinButtonText = computed(() => {
  if (loadingCheckinStatus.value) {
    return t('profile.checkin.loading')
  }
  if (checkinSubmitting.value) {
    return t('profile.checkin.checkingIn')
  }
  if (checkinStatus.value?.checked_in_today) {
    return t('profile.checkin.checkedInToday')
  }
  return t('profile.checkin.checkInNow')
})

const checkinRewardText = computed(() => {
  if (!checkinStatus.value?.checked_in_today || checkinStatus.value.reward_amount == null) {
    return ''
  }
  return t('profile.checkin.rewardAmount', {
    amount: formatAmount(checkinStatus.value.reward_amount)
  })
})

function formatAmount(value: number): string {
  return value.toFixed(2)
}

async function loadDailyCheckinStatus() {
  if (!props.user) {
    checkinStatus.value = null
    return
  }

  loadingCheckinStatus.value = true
  try {
    checkinStatus.value = await redeemAPI.getDailyCheckinStatus()
  } catch (error) {
    console.error('Failed to load daily check-in status:', error)
    checkinStatus.value = null
  } finally {
    loadingCheckinStatus.value = false
  }
}

async function handleDailyCheckin() {
  if (!checkinStatus.value || checkinButtonDisabled.value) {
    return
  }

  checkinSubmitting.value = true
  try {
    const result = await redeemAPI.dailyCheckin()
    checkinStatus.value = {
      ...checkinStatus.value,
      checked_in_today: true,
      reward_amount: result.reward_amount
    }
    appStore.showSuccess(
      t('profile.checkin.success', {
        amount: formatAmount(result.reward_amount)
      })
    )
    await authStore.refreshUser()
    await loadDailyCheckinStatus()
  } catch (error: any) {
    const statusCode = Number(error?.status || 0)
    if (statusCode === 409) {
      if (checkinStatus.value) {
        checkinStatus.value.checked_in_today = true
      }
      appStore.showInfo(t('profile.checkin.checkedInToday'))
      return
    }
    if (statusCode === 403) {
      appStore.showError(t('profile.checkin.unavailable'))
      await loadDailyCheckinStatus()
      return
    }
    appStore.showError(
      t('profile.checkin.failed') + ': ' + (error?.message || t('common.unknownError'))
    )
  } finally {
    checkinSubmitting.value = false
  }
}

watch(
  () => props.user?.id,
  (newUserID, oldUserID) => {
    if (!newUserID) {
      checkinStatus.value = null
      return
    }
    if (newUserID !== oldUserID) {
      loadDailyCheckinStatus()
    }
  }
)

onMounted(() => {
  loadDailyCheckinStatus()
})
</script>
