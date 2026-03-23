<template>
  <AppLayout>
    <div class="container mx-auto px-4 py-6 max-w-7xl">
      <!-- 页面标题 -->
      <div class="mb-6">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('activities.title') }}
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('activities.description') }}
        </p>
      </div>

      <!-- 活动分类标签 -->
      <div class="mb-6 flex flex-wrap gap-2">
        <button
          v-for="type in activityTypes"
          :key="type.value"
          @click="selectedType = type.value"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            selectedType === type.value
              ? 'bg-primary-600 text-white dark:bg-primary-500'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'
          ]"
        >
          {{ type.label }}
        </button>
      </div>

      <!-- 加载状态 -->
      <div v-if="loading" class="flex justify-center items-center py-20">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>

      <!-- 空状态 -->
      <div
        v-else-if="filteredActivities.length === 0"
        class="text-center py-20"
      >
        <div class="text-gray-400 dark:text-gray-500 mb-4">
          <svg
            class="mx-auto h-16 w-16"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
            />
          </svg>
        </div>
        <p class="text-gray-500 dark:text-gray-400">
          {{ t('activities.noActivities') }}
        </p>
      </div>

      <!-- 活动卡片网格 -->
      <div
        v-else
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
      >
        <div
          v-for="activity in filteredActivities"
          :key="activity.id"
          class="bg-white dark:bg-dark-800 rounded-lg shadow-sm border border-gray-200 dark:border-dark-700 overflow-hidden hover:shadow-lg transition-shadow"
        >
          <!-- 活动卡片头部 -->
          <div class="p-6">
            <div class="flex items-start justify-between mb-4">
              <div class="flex-1">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-1">
                  {{ activity.name }}
                </h3>
                <p class="text-sm text-gray-500 dark:text-gray-400 line-clamp-2">
                  {{ activity.description }}
                </p>
              </div>
              <div class="ml-4">
                <span
                  :class="[
                    'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                    getActivityTypeColor(activity.type)
                  ]"
                >
                  {{ getActivityTypeLabel(activity.type) }}
                </span>
              </div>
            </div>

            <!-- 奖励预览 -->
            <div v-if="activity.edges?.rewards?.length" class="mb-4">
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="reward in activity.edges.rewards.slice(0, 3)"
                  :key="reward.id"
                  class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300"
                >
                  🎁 {{ reward.name }}
                </span>
                <span
                  v-if="activity.edges.rewards.length > 3"
                  class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400"
                >
                  +{{ activity.edges.rewards.length - 3 }} 更多
                </span>
              </div>
            </div>

            <!-- 活动统计 -->
            <div class="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400 mb-4">
              <span class="flex items-center gap-1">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
                {{ activity.total_participations || 0 }} 次参与
              </span>
            </div>

            <!-- 参与按钮 -->
            <button
              @click="participateInActivity(activity)"
              :disabled="participating[activity.id]"
              class="w-full btn btn-primary"
            >
              <span v-if="participating[activity.id]" class="flex items-center justify-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                {{ t('activities.participating') }}
              </span>
              <span v-else>
                {{ getParticipateButtonText(activity) }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 参与结果弹窗 -->
    <Teleport to="body">
      <div
        v-if="showResultModal"
        class="fixed inset-0 z-50 overflow-y-auto"
        @click.self="showResultModal = false"
      >
        <div class="flex min-h-screen items-center justify-center p-4">
          <div class="fixed inset-0 bg-black/50 transition-opacity"></div>
          
          <div class="relative bg-white dark:bg-dark-800 rounded-lg shadow-xl max-w-md w-full p-6">
            <!-- 成功图标 -->
            <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 dark:bg-green-900/30 mb-4">
              <svg class="h-6 w-6 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>

            <!-- 消息 -->
            <h3 class="text-lg font-medium text-gray-900 dark:text-white text-center mb-2">
              {{ participationResult?.message || t('activities.success') }}
            </h3>

            <!-- 奖励列表 -->
            <div v-if="participationResult?.rewards?.length" class="mt-4">
              <p class="text-sm text-gray-500 dark:text-gray-400 mb-3 text-center">
                {{ t('activities.rewardsReceived') }}
              </p>
              <div class="space-y-2">
                <div
                  v-for="(reward, index) in participationResult.rewards"
                  :key="index"
                  class="flex items-center justify-between p-3 bg-gray-50 dark:bg-dark-700 rounded-lg"
                >
                  <span class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ reward.name }}
                  </span>
                  <span class="text-xs text-gray-500 dark:text-gray-400">
                    {{ getRewardValueText(reward) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- 关闭按钮 -->
            <div class="mt-6">
              <button
                @click="closeResultModal"
                class="w-full btn btn-primary"
              >
                {{ t('common.close') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

interface Activity {
  id: number
  name: string
  description: string
  icon: string
  type: string
  status: string
  total_participations: number
  edges?: {
    rewards?: Array<{
      id: number
      name: string
      description: string
      reward_type: string
      reward_value: string
    }>
  }
}

interface ParticipationResult {
  success: boolean
  message: string
  rewards: Array<{
    reward_id: number
    name: string
    type: string
    value: Record<string, any>
  }>
}

// 状态
const loading = ref(false)
const activities = ref<Activity[]>([])
const selectedType = ref('all')
const participating = ref<Record<number, boolean>>({})
const showResultModal = ref(false)
const participationResult = ref<ParticipationResult | null>(null)

// 活动类型
const activityTypes = [
  { value: 'all', label: t('activities.allActivities') },
  { value: 'check_in', label: t('activities.types.checkIn') },
  { value: 'lottery', label: t('activities.types.lottery') },
  { value: 'redeem', label: t('activities.types.redeem') },
  { value: 'newbie', label: t('activities.types.newbie') },
  { value: 'limited_time', label: t('activities.types.limitedTime') }
]

// 过滤后的活动
const filteredActivities = computed(() => {
  if (selectedType.value === 'all') {
    return activities.value
  }
  return activities.value.filter(a => a.type === selectedType.value)
})

// 获取活动类型标签
const getActivityTypeLabel = (type: string) => {
  const typeMap: Record<string, string> = {
    check_in: t('activities.types.checkIn'),
    lottery: t('activities.types.lottery'),
    redeem: t('activities.types.redeem'),
    task: t('activities.types.task'),
    newbie: t('activities.types.newbie'),
    limited_time: t('activities.types.limitedTime')
  }
  return typeMap[type] || type
}

// 获取活动类型颜色
const getActivityTypeColor = (type: string) => {
  const colorMap: Record<string, string> = {
    check_in: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300',
    lottery: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300',
    redeem: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
    task: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
    newbie: 'bg-pink-100 text-pink-800 dark:bg-pink-900/30 dark:text-pink-300',
    limited_time: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300'
  }
  return colorMap[type] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
}

// 获取参与按钮文本
const getParticipateButtonText = (activity: Activity) => {
  const textMap: Record<string, string> = {
    check_in: t('activities.actions.checkIn'),
    lottery: t('activities.actions.draw'),
    redeem: t('activities.actions.redeem'),
    task: t('activities.actions.complete'),
    newbie: t('activities.actions.claim'),
    limited_time: t('activities.actions.participate')
  }
  return textMap[activity.type] || t('activities.actions.participate')
}

// 获取奖励值文本
const getRewardValueText = (reward: any) => {
  try {
    const value = typeof reward.value === 'string' ? JSON.parse(reward.value) : reward.value
    if (reward.type === 'balance') {
      return `¥${value.amount}`
    } else if (reward.type === 'subscription') {
      return `${value.days} ${t('activities.days')}`
    }
    return ''
  } catch {
    return ''
  }
}

// 加载活动列表
const loadActivities = async () => {
  loading.value = true
  try {
    // TODO: 实现 API 调用
    // const response = await activitiesAPI.list()
    // activities.value = response.activities
    
    // 临时模拟数据
    activities.value = [
      {
        id: 1,
        name: '每日签到',
        description: '每天签到领取奖励，连续签到更多惊喜',
        icon: 'calendar-check',
        type: 'check_in',
        status: 'active',
        total_participations: 1520,
        edges: {
          rewards: [
            {
              id: 1,
              name: '5元余额',
              description: '',
              reward_type: 'balance',
              reward_value: '{"amount": 5.0}'
            }
          ]
        }
      }
    ]
  } catch (error: any) {
    appStore.showToast(error.message || t('activities.loadFailed'), 'error')
  } finally {
    loading.value = false
  }
}

// 参与活动
const participateInActivity = async (activity: Activity) => {
  participating.value[activity.id] = true
  try {
    // TODO: 实现 API 调用
    // const result = await activitiesAPI.participate(activity.id)
    // participationResult.value = result
    
    // 临时模拟
    participationResult.value = {
      success: true,
      message: '签到成功！已连续签到 1 天',
      rewards: [
        {
          reward_id: 1,
          name: '5元余额',
          type: 'balance',
          value: { amount: 5.0 }
        }
      ]
    }
    
    showResultModal.value = true
    
    // 重新加载活动列表
    await loadActivities()
  } catch (error: any) {
    appStore.showToast(error.message || t('activities.participateFailed'), 'error')
  } finally {
    participating.value[activity.id] = false
  }
}

// 关闭结果弹窗
const closeResultModal = () => {
  showResultModal.value = false
  participationResult.value = null
}

// 初始化
onMounted(() => {
  loadActivities()
})
</script>
