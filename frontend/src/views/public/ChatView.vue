<template>
  <div class="h-screen flex flex-col bg-slate-950 text-slate-100">
    <!-- API Key Setup Modal -->
    <div
      v-if="!apiKey"
      class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl max-w-md w-full p-8 border border-gray-200 dark:border-gray-700">
        <div class="text-center mb-6">
          <div class="text-6xl mb-4">💬</div>
          <h1 class="text-2xl font-bold mb-2">欢迎使用 AI 聊天</h1>
          <p class="text-gray-500 dark:text-gray-400">输入您的 API 密钥开始使用</p>
        </div>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium mb-2">API 密钥</label>
            <input
              v-model="keyInput"
              type="password"
              placeholder="sk-..."
              class="w-full bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-cyan-500"
              @keydown.enter="handleSetApiKey"
            />
          </div>

          <button
            @click="handleSetApiKey"
            :disabled="!keyInput.trim() || isValidating"
            class="w-full bg-gradient-to-r from-cyan-500 to-blue-600 text-white rounded-lg py-3 font-medium hover:from-cyan-600 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg"
          >
            {{ isValidating ? '验证中...' : '继续' }}
          </button>

          <p v-if="keyError" class="text-red-500 text-sm text-center">{{ keyError }}</p>
        </div>
      </div>
    </div>

    <!-- Main Chat Interface -->
    <div v-else class="flex-1 flex flex-col md:flex-row overflow-hidden">
      <!-- Sidebar -->
      <transition name="slide-left">
        <div
          v-show="isSidebarOpen || isDesktop"
          class="fixed md:relative inset-y-0 left-0 w-64 md:w-72 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col z-40 md:z-0 shadow-xl md:shadow-none"
        >
          <!-- Sidebar Header -->
          <div class="p-3 border-b border-gray-200 dark:border-gray-700">
            <button
              @click="createNewChat"
              class="w-full bg-gradient-to-r from-cyan-500 to-blue-600 text-white rounded-lg py-2.5 px-4 font-medium hover:from-cyan-600 hover:to-blue-700 transition-all flex items-center justify-center gap-2 shadow-md"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              新对话
            </button>
          </div>

          <!-- Conversation List -->
          <div class="flex-1 overflow-y-auto p-2 space-y-1">
            <div
              v-for="conv in conversations"
              :key="conv.id"
              @click="selectConversation(conv.id)"
              :class="[
                'p-3 rounded-lg cursor-pointer transition-all',
                currentConversationId === conv.id
                  ? 'bg-gray-100 dark:bg-gray-700 border-l-4 border-cyan-500'
                  : 'hover:bg-gray-50 dark:hover:bg-gray-700/50'
              ]"
            >
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <h3 class="font-medium truncate text-sm">{{ conv.title }}</h3>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                    {{ formatDate(conv.updatedAt) }}
                  </p>
                </div>
                <button
                  @click.stop="deleteConv(conv.id)"
                  class="flex-shrink-0 p-1 hover:bg-red-100 dark:hover:bg-red-500/20 rounded text-gray-400 hover:text-red-500 transition-colors"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>

            <div v-if="conversations.length === 0" class="text-center text-gray-400 py-8 text-sm">
              暂无对话记录
            </div>
          </div>

          <!-- Bottom Actions -->
          <div class="p-3 border-t border-gray-200 dark:border-gray-700 space-y-2">
            <button
              @click="activeTab = 'chat'"
              :class="[
                'w-full py-2 px-3 rounded-lg font-medium transition-all text-sm',
                activeTab === 'chat'
                  ? 'bg-cyan-500 text-white shadow-md'
                  : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
              ]"
            >
              💬 聊天
            </button>
            <button
              @click="activeTab = 'images'"
              :class="[
                'w-full py-2 px-3 rounded-lg font-medium transition-all text-sm',
                activeTab === 'images'
                  ? 'bg-cyan-500 text-white shadow-md'
                  : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
              ]"
            >
              🎨 图片
            </button>
            <button
              @click="clearApiKeyHandler"
              class="w-full py-2 px-3 rounded-lg font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-red-50 dark:hover:bg-red-500/20 hover:text-red-500 transition-all text-sm"
            >
              🔑 更换密钥
            </button>
          </div>
        </div>
      </transition>

      <!-- Mobile Sidebar Overlay -->
      <div
        v-if="isSidebarOpen && !isDesktop"
        @click="isSidebarOpen = false"
        class="fixed inset-0 bg-black/30 backdrop-blur-sm z-30"
      />

      <!-- Main Content -->
      <div class="flex-1 flex flex-col overflow-hidden bg-gray-50 dark:bg-gray-900">
        <!-- Header -->
        <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between shadow-sm">
          <div class="flex items-center gap-3">
            <button
              v-if="!isDesktop"
              @click="isSidebarOpen = !isSidebarOpen"
              class="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
            <h2 class="font-semibold text-lg">{{ activeTab === 'chat' ? '对话' : '图片生成' }}</h2>
          </div>
          <div class="flex items-center gap-2 text-sm">
            <span class="text-gray-500 dark:text-gray-400">总费用:</span>
            <span class="font-bold text-cyan-600 dark:text-cyan-400">${{ getTotalCost().toFixed(6) }}</span>
          </div>
        </div>

        <!-- Chat Tab -->
        <div v-if="activeTab === 'chat'" class="flex-1 flex flex-col overflow-hidden">
          <!-- Messages Area -->
          <div ref="messagesContainer" class="flex-1 overflow-y-auto px-4 py-6 space-y-6">
            <div v-if="currentMessages.length === 0" class="flex flex-col items-center justify-center h-full text-center">
              <div class="text-6xl mb-4">👋</div>
              <h3 class="text-xl font-semibold mb-2">开始新对话</h3>
              <p class="text-gray-500 dark:text-gray-400 max-w-md">
                我可以帮您解答问题、编写代码、翻译文本等。请在下方输入您的问题。
              </p>
            </div>
            
            <ChatMessage
              v-for="message in currentMessages"
              :key="message.id"
              :message="message"
              :isStreaming="isLoading && streamingMessageId === message.id"
            />
          </div>

          <!-- Model Selector & Cost Display -->
          <div class="px-4 py-2 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
            <div class="flex items-center justify-between text-sm mb-2">
              <select
                v-model="selectedModel"
                class="flex-1 max-w-xs bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500"
              >
                <option v-for="model in models" :key="model.id" :value="model.id">
                  {{ model.name }}
                </option>
              </select>
              <div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
                <span>本次: ${{ currentCost.toFixed(6) }}</span>
                <span v-if="pricing[selectedModel]" class="hidden sm:inline">
                  (输入: ${{ pricing[selectedModel].prompt }}/M · 输出: ${{ pricing[selectedModel].completion }}/M)
                </span>
              </div>
            </div>
          </div>

          <!-- Input Area -->
          <div class="px-4 py-3 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
            <div class="flex items-end gap-2">
              <textarea
                ref="inputRef"
                v-model="inputText"
                @keydown.enter.exact.prevent="sendMessage"
                @keydown.meta.enter="inputText += '\n'"
                @keydown.ctrl.enter="inputText += '\n'"
                placeholder="输入消息... (Enter 发送, Ctrl/⌘+Enter 换行)"
                rows="1"
                class="flex-1 bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-cyan-500 resize-none max-h-32 text-sm"
                style="min-height: 44px"
              />
              <button
                @click="isLoading ? stopGeneration() : sendMessage()"
                :disabled="!inputText.trim() && !isLoading"
                class="flex-shrink-0 bg-gradient-to-r from-cyan-500 to-blue-600 text-white rounded-lg px-4 py-3 font-medium hover:from-cyan-600 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-md flex items-center gap-2"
              >
                <svg v-if="!isLoading" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                </svg>
                <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
                <span class="hidden sm:inline">{{ isLoading ? '停止' : '发送' }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Images Tab -->
        <div v-else class="flex-1 flex flex-col overflow-hidden">
          <!-- Image Generation Form -->
          <div class="p-6 space-y-4 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <div>
              <label class="block text-sm font-medium mb-2">描述您想要的图片</label>
              <textarea
                v-model="imagePrompt"
                placeholder="例如: 一只可爱的小猫在花园里玩耍..."
                rows="3"
                class="w-full bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-4 py-3 focus:outline-none focus:ring-2 focus:ring-cyan-500 resize-none text-sm"
              />
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label class="block text-sm font-medium mb-2">尺寸</label>
                <select
                  v-model="imageSize"
                  class="w-full bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500"
                >
                  <option value="1024x1024">1024×1024 (正方形)</option>
                  <option value="1792x1024">1792×1024 (横向)</option>
                  <option value="1024x1792">1024×1792 (竖向)</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium mb-2">模型</label>
                <select
                  v-model="selectedImageModel"
                  class="w-full bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500"
                >
                  <option value="grok-2-image-1212">Grok 图片生成</option>
                  <option value="dall-e-3">DALL·E 3</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium mb-2">数量</label>
                <input
                  v-model.number="imageCount"
                  type="number"
                  min="1"
                  max="4"
                  class="w-full bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500"
                />
              </div>
            </div>

            <div class="flex gap-3">
              <button
                @click="generateImage"
                :disabled="!imagePrompt.trim() || isGenerating"
                class="flex-1 bg-gradient-to-r from-cyan-500 to-teal-600 text-white rounded-lg py-3 font-medium hover:from-cyan-600 hover:to-teal-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-md"
              >
                {{ isGenerating ? '生成中...' : '生成图片' }}
              </button>
              <button
                v-if="images.length > 0"
                @click="clearImages"
                class="px-6 bg-red-500/20 text-red-400 border border-red-500/50 rounded-lg py-3 font-medium hover:bg-red-500/30 transition-all"
              >
                清空
              </button>
            </div>
          </div>

          <!-- Generated Images Grid -->
          <div class="flex-1 overflow-y-auto p-6">
            <div v-if="images.length === 0" class="flex flex-col items-center justify-center h-full text-center">
              <div class="text-6xl mb-4">🎨</div>
              <h3 class="text-xl font-semibold mb-2">AI 图片生成</h3>
              <p class="text-gray-500 dark:text-gray-400 max-w-md">
                使用 AI 模型生成精美图片，支持多种尺寸和风格。
              </p>
            </div>

            <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              <div
                v-for="img in images"
                :key="img.id"
                class="group relative bg-white dark:bg-gray-800 rounded-lg overflow-hidden shadow-md hover:shadow-xl transition-all border border-gray-200 dark:border-gray-700"
              >
                <img
                  :src="img.url"
                  :alt="img.prompt"
                  class="w-full aspect-square object-cover cursor-pointer"
                  @click="previewImage = img"
                />
                <div class="p-3">
                  <p class="text-sm text-gray-600 dark:text-gray-400 line-clamp-2">{{ img.prompt }}</p>
                  <div class="flex items-center justify-between mt-2 text-xs text-gray-500 dark:text-gray-400">
                    <span>{{ formatDate(img.timestamp) }}</span>
                    <button
                      @click="downloadImage(img)"
                      class="px-3 py-1 bg-cyan-500 text-white rounded hover:bg-cyan-600 transition-colors"
                    >
                      下载
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Image Preview Modal -->
    <div
      v-if="previewImage"
      @click="previewImage = null"
      class="fixed inset-0 bg-black/90 z-50 flex items-center justify-center p-4"
    >
      <img
        :src="previewImage.url"
        :alt="previewImage.prompt"
        class="max-w-full max-h-full object-contain rounded-lg"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useWindowSize } from '@vueuse/core'
import { useChatAPI } from '@/composables/useChatAPI'
import { useImageAPI } from '@/composables/useImageAPI'
import { useChatStorage } from '@/composables/useChatStorage'
import ChatMessage from '@/components/chat/ChatMessage.vue'
import type { Message } from '@/types/chat'

// Window size for responsive behavior
const { width } = useWindowSize()
const isDesktop = computed(() => width.value >= 768)
const route = useRoute()

// API composables
const {
  apiKey,
  isValidating,
  models,
  pricing,
  setApiKey,
  clearApiKey: clearKey,
  fetchModels,
  fetchPricing,
  sendMessageStreaming
} = useChatAPI()

const {
  isGenerating,
  generateImage: genImage,
  downloadImage: dlImage
} = useImageAPI()

const {
  conversations,
  currentConversationId,
  images,
  getTotalCost,
  loadConversations,
  createConversation,
  deleteConversation,
  loadMessages,
  saveMessage,
  loadImages,
  saveImage,
  clearAllImages,
  updateConversationTitle,
  getDefaultModel,
  setDefaultModel
} = useChatStorage()

// UI State
const isSidebarOpen = ref(false)
const activeTab = ref<'chat' | 'images'>('chat')
const keyInput = ref('')
const keyError = ref('')
const selectedModel = ref('gpt-4o')
const selectedImageModel = ref('grok-2-image-1212')
const inputText = ref('')
const inputRef = ref<HTMLTextAreaElement>()
const messagesContainer = ref<HTMLDivElement>()
const isLoading = ref(false)
const streamingMessageId = ref<string | null>(null)
const abortController = ref<AbortController | null>(null)

// Image state
const imagePrompt = ref('')
const imageSize = ref('1024x1024')
const imageCount = ref(1)
const previewImage = ref<any>(null)

// Current messages
const currentMessages = ref<Message[]>([])
const currentCost = computed(() => {
  return currentMessages.value.reduce((sum, msg) => sum + (msg.cost || 0), 0)
})

// Watch for model changes and save to storage
watch(selectedModel, (newModel) => {
  if (newModel) {
    setDefaultModel(newModel)
  }
})

// Initialize
onMounted(async () => {
  // Check for URL parameters
  const urlKey = route.query.key as string
  const urlModel = route.query.model as string

  if (urlKey) {
    keyInput.value = urlKey
    await handleSetApiKey()
  }

  if (apiKey.value) {
    await Promise.all([
      loadConversations(),
      fetchModels(),
      fetchPricing(),
      loadImages()
    ])
    
    // Set model from URL or use saved default
    if (urlModel && models.value.some(m => m.id === urlModel)) {
      selectedModel.value = urlModel
      setDefaultModel(urlModel)
    } else {
      const savedModel = await getDefaultModel()
      if (savedModel && models.value.some(m => m.id === savedModel)) {
        selectedModel.value = savedModel
      } else if (models.value.length > 0) {
        selectedModel.value = models.value[0].id
      }
    }
  }
})

// Handle API key setup
const handleSetApiKey = async () => {
  keyError.value = ''
  const success = await setApiKey(keyInput.value)
  
  if (success) {
    await Promise.all([
      loadConversations(),
      fetchModels(),
      fetchPricing(),
      loadImages()
    ])
    
    // Check for model parameter
    const urlModel = route.query.model as string
    if (urlModel && models.value.some(m => m.id === urlModel)) {
      selectedModel.value = urlModel
      setDefaultModel(urlModel)
    } else {
      const savedModel = await getDefaultModel()
      if (savedModel && models.value.some(m => m.id === savedModel)) {
        selectedModel.value = savedModel
      } else if (models.value.length > 0) {
        selectedModel.value = models.value[0].id
        setDefaultModel(models.value[0].id)
      }
    }
  } else {
    keyError.value = 'API 密钥无效，请检查后重试'
  }
}

const clearApiKeyHandler = () => {
  if (confirm('确定要更换 API 密钥吗？这将清除所有本地数据。')) {
    clearKey()
    currentMessages.value = []
    keyInput.value = ''
    keyError.value = ''
  }
}

const createNewChat = async () => {
  await createConversation()
  currentMessages.value = []
  if (!isDesktop.value) {
    isSidebarOpen.value = false
  }
}

const selectConversation = async (id: string) => {
  const messages = await loadMessages(id)
  currentMessages.value = messages
  if (!isDesktop.value) {
    isSidebarOpen.value = false
  }
  await nextTick()
  scrollToBottom()
}

const deleteConv = async (id: string) => {
  if (confirm('确定要删除这个对话吗？')) {
    await deleteConversation(id)
    if (currentConversationId.value === id) {
      currentMessages.value = []
    }
  }
}

const sendMessage = async () => {
  if (!inputText.value.trim() || isLoading.value) return

  const userMessage: Message = {
    id: Date.now().toString(),
    role: 'user',
    content: inputText.value.trim(),
    timestamp: Date.now()
  }

  currentMessages.value.push(userMessage)
  
  // Save user message
  if (currentConversationId.value) {
    await saveMessage(currentConversationId.value, userMessage)
  }

  const prompt = inputText.value.trim()
  inputText.value = ''
  isLoading.value = true

  const assistantMessage: Message = {
    id: (Date.now() + 1).toString(),
    role: 'assistant',
    content: '',
    timestamp: Date.now()
  }

  currentMessages.value.push(assistantMessage)
  streamingMessageId.value = assistantMessage.id

  await nextTick()
  scrollToBottom()

  try {
    abortController.value = new AbortController()
    
    // Prepare request
    const request = {
      model: selectedModel.value,
      messages: currentMessages.value
        .filter(m => m.id !== assistantMessage.id)
        .map(m => ({
          role: m.role,
          content: m.content
        }))
    }

    await sendMessageStreaming(
      request,
      (chunk) => {
        assistantMessage.content += chunk
        scrollToBottom()
      },
      (response) => {
        // Calculate cost
        if (response.usage && pricing.value[selectedModel.value]) {
          const prices = pricing.value[selectedModel.value]
          const cost = 
            (response.usage.prompt_tokens / 1_000_000) * prices.prompt +
            (response.usage.completion_tokens / 1_000_000) * prices.completion
          assistantMessage.cost = cost
        }
      },
      (error) => {
        console.error('Stream error:', error)
      },
      abortController.value.signal
    )

    // Save assistant message
    if (currentConversationId.value) {
      await saveMessage(currentConversationId.value, assistantMessage)
      
      // Update conversation title if it's the first message
      if (currentMessages.value.filter(m => m.role === 'user').length === 1) {
        const title = prompt.slice(0, 30) + (prompt.length > 30 ? '...' : '')
        await updateConversationTitle(currentConversationId.value, title)
      }
    }
  } catch (error: any) {
    if (error.name !== 'AbortError') {
      console.error('发送消息失败:', error)
      assistantMessage.content = '抱歉，发送消息时出现错误。请稍后重试。'
    }
  } finally {
    isLoading.value = false
    streamingMessageId.value = null
    abortController.value = null
  }
}

const stopGeneration = () => {
  if (abortController.value) {
    abortController.value.abort()
    abortController.value = null
  }
  isLoading.value = false
  streamingMessageId.value = null
}

const generateImage = async () => {
  if (!imagePrompt.value.trim() || !apiKey.value) return

  try {
    const result = await genImage(apiKey.value, {
      model: selectedImageModel.value,
      prompt: imagePrompt.value,
      size: imageSize.value,
      n: imageCount.value
    })
    
    if (result && result.length > 0) {
      for (const imgData of result) {
        await saveImage({
          id: Date.now().toString() + Math.random(),
          url: imgData.url,
          prompt: imagePrompt.value,
          model: selectedImageModel.value,
          size: imageSize.value,
          timestamp: Date.now(),
          cost: imgData.cost || 0
        })
      }
      imagePrompt.value = ''
    }
  } catch (error) {
    console.error('生成图片失败:', error)
    alert('生成图片失败，请稍后重试')
  }
}

const downloadImage = async (img: any) => {
  await dlImage(img.url)
}

const clearImages = async () => {
  if (confirm('确定要清空所有生成的图片吗？此操作无法撤销。')) {
    await clearAllImages()
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const formatDate = (timestamp: number) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`
  
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}
</script>

<style scoped>
.slide-left-enter-active,
.slide-left-leave-active {
  transition: transform 0.3s ease;
}

.slide-left-enter-from {
  transform: translateX(-100%);
}

.slide-left-leave-to {
  transform: translateX(-100%);
}

/* Auto-resize textarea */
textarea {
  field-sizing: content;
}

/* Scrollbar styling */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.3);
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.5);
}

/* Dark mode scrollbar */
.dark ::-webkit-scrollbar-thumb {
  background: rgba(75, 85, 99, 0.3);
}

.dark ::-webkit-scrollbar-thumb:hover {
  background: rgba(75, 85, 99, 0.5);
}
</style>
