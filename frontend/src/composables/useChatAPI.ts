/**
 * Chat API composable
 * Handles OpenAI-compatible chat completions
 */

import { ref } from 'vue'
import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  ChatCompletionChunk,
  Model,
  ModelPricing
} from '@/types/chat'

export function useChatAPI() {
  const apiKey = ref<string | null>(localStorage.getItem('sub2api_chat_key'))
  const isValidating = ref(false)
  const models = ref<Model[]>([])
  const pricing = ref<ModelPricing>({})

  /**
   * Validate API key by fetching models
   */
  const validateKey = async (key: string): Promise<boolean> => {
    try {
      isValidating.value = true
      const response = await fetch('/v1/models', {
        headers: {
          Authorization: `Bearer ${key}`
        }
      })

      if (!response.ok) {
        return false
      }

      const data = await response.json()
      // OpenAI format: { data: [...models], object: "list" }
      if (data.data && Array.isArray(data.data)) {
        models.value = data.data.map((m: any) => ({
          id: m.id,
          name: m.id,
          context_length: m.context_length
        }))
        return true
      }

      return false
    } catch (error) {
      console.error('API key validation failed:', error)
      return false
    } finally {
      isValidating.value = false
    }
  }

  /**
   * Set and save API key
   */
  const setApiKey = async (key: string): Promise<boolean> => {
    const valid = await validateKey(key)
    if (valid) {
      apiKey.value = key
      localStorage.setItem('sub2api_chat_key', key)
      return true
    }
    return false
  }

  /**
   * Clear API key
   */
  const clearApiKey = () => {
    apiKey.value = null
    localStorage.removeItem('sub2api_chat_key')
    models.value = []
  }

  /**
   * Fetch available models
   */
  const fetchModels = async (): Promise<Model[]> => {
    if (!apiKey.value) return []

    try {
      const response = await fetch('/v1/models', {
        headers: {
          Authorization: `Bearer ${apiKey.value}`
        }
      })

      if (!response.ok) {
        throw new Error('Failed to fetch models')
      }

      const data = await response.json()
      if (data.data && Array.isArray(data.data)) {
        models.value = data.data.map((m: any) => ({
          id: m.id,
          name: m.id,
          context_length: m.context_length
        }))
      }

      return models.value
    } catch (error) {
      console.error('Failed to fetch models:', error)
      return []
    }
  }

  /**
   * Fetch model pricing from backend
   */
  const fetchPricing = async (): Promise<ModelPricing> => {
    console.log('[Pricing] 开始获取价格，apiKey存在:', !!apiKey.value)
    
    if (!apiKey.value) {
      return {}
    }

    try {
      // Get all model IDs
      const modelIds = models.value.map(m => m.id)
      console.log('[Pricing] 当前模型列表:', modelIds)
      
      if (modelIds.length === 0) {
        console.log('[Pricing] 模型列表为空，跳过')
        return {}
      }

      // Fetch pricing from backend
      const response = await fetch('/v1/models', {
        headers: {
          Authorization: `Bearer ${apiKey.value}`
        }
      })

      if (!response.ok) {
        throw new Error('Failed to fetch pricing')
      }

      const data = await response.json()
      console.log('[Pricing] API响应:', data)
      
      const pricingMap: ModelPricing = {}

      // Extract pricing from model data if available
      if (data.data && Array.isArray(data.data)) {
        for (const model of data.data) {
          // Check if model has pricing info in metadata
          if (model.pricing) {
            pricingMap[model.id] = {
              prompt: model.pricing.input || 0,
              completion: model.pricing.output || 0
            }
          }
        }
      }

      console.log('[Pricing] 从API提取的价格:', pricingMap)

      // If no pricing from API, use hardcoded fallback
      if (Object.keys(pricingMap).length === 0) {
        console.log('[Pricing] API未返回价格，使用硬编码fallback')
        const hardcodedPricing: ModelPricing = {
          // OpenAI
          'gpt-4o': { prompt: 2.5, completion: 10 },
          'gpt-4o-mini': { prompt: 0.15, completion: 0.6 },
          'gpt-4-turbo': { prompt: 10, completion: 30 },
          'gpt-3.5-turbo': { prompt: 0.5, completion: 1.5 },
          // Grok
          'grok-4': { prompt: 5, completion: 15 },
          'grok-4-thinking': { prompt: 5, completion: 15 },
          'grok-4.1-expert': { prompt: 5, completion: 15 },
          'grok-4.1-fast': { prompt: 3, completion: 10 },
          'grok-4.20-beta': { prompt: 5, completion: 15 },
          'grok-2-1212': { prompt: 2, completion: 10 },
          'grok-beta': { prompt: 5, completion: 15 },
          'grok-2': { prompt: 2, completion: 10 },
          'grok-livechat': { prompt: 2, completion: 10 },
          // Claude
          'claude-opus-4': { prompt: 15, completion: 75 },
          'claude-sonnet-4': { prompt: 3, completion: 15 },
          'claude-3-5-sonnet-20241022': { prompt: 3, completion: 15 },
          'claude-3-5-haiku-20241022': { prompt: 1, completion: 5 },
          'claude-haiku-3-5': { prompt: 1, completion: 5 },
          // Gemini
          'gemini-2.0-flash-exp': { prompt: 0, completion: 0 },
          'gemini-exp-1206': { prompt: 0, completion: 0 },
          // DeepSeek
          'deepseek-v3.2': { prompt: 0.27, completion: 1.1 },
          'deepseek-v3': { prompt: 0.27, completion: 1.1 },
          'deepseek-chat': { prompt: 0.14, completion: 0.28 },
          // Doubao
          'doubao-seed-2.0-code': { prompt: 0.3, completion: 0.6 },
          'doubao-seed-2.0-lite': { prompt: 0.3, completion: 0.6 },
          'doubao-seed-2.0-pro': { prompt: 0.8, completion: 2 },
          'doubao-seed-code': { prompt: 0.3, completion: 0.6 },
          // GLM
          'glm-4.7': { prompt: 1, completion: 1 },
          'glm-4': { prompt: 1, completion: 1 },
          // Kimi
          'kimi-k2.5': { prompt: 0.3, completion: 0.3 },
          'kimi-k1.5': { prompt: 0.3, completion: 0.3 },
          // Minimax
          'minimax-m2.5': { prompt: 1.5, completion: 1.5 },
          'minimax-m1': { prompt: 1.5, completion: 1.5 }
        }

        pricing.value = hardcodedPricing
        console.log('[Pricing] 设置价格完成:', pricing.value)
        return hardcodedPricing
      }

      pricing.value = pricingMap
      console.log('[Pricing] 设置价格完成:', pricing.value)
      return pricingMap
    } catch (error) {
      console.error('[Pricing] 获取失败:', error)
      // Return hardcoded fallback on error
      const hardcodedPricing: ModelPricing = {
        'gpt-4o': { prompt: 2.5, completion: 10 },
        'gpt-4o-mini': { prompt: 0.15, completion: 0.6 },
        'gpt-4-turbo': { prompt: 10, completion: 30 },
        'gpt-3.5-turbo': { prompt: 0.5, completion: 1.5 },
        'grok-4': { prompt: 5, completion: 15 },
        'grok-4-thinking': { prompt: 5, completion: 15 },
        'grok-4.1-expert': { prompt: 5, completion: 15 },
        'grok-4.1-fast': { prompt: 3, completion: 10 },
        'grok-4.20-beta': { prompt: 5, completion: 15 },
        'grok-2-1212': { prompt: 2, completion: 10 },
        'grok-beta': { prompt: 5, completion: 15 },
        'grok-2': { prompt: 2, completion: 10 },
        'grok-livechat': { prompt: 2, completion: 10 },
        'claude-opus-4': { prompt: 15, completion: 75 },
        'claude-sonnet-4': { prompt: 3, completion: 15 },
        'claude-3-5-sonnet-20241022': { prompt: 3, completion: 15 },
        'claude-3-5-haiku-20241022': { prompt: 1, completion: 5 },
        'claude-haiku-3-5': { prompt: 1, completion: 5 },
        'gemini-2.0-flash-exp': { prompt: 0, completion: 0 },
        'deepseek-v3.2': { prompt: 0.27, completion: 1.1 },
        'deepseek-v3': { prompt: 0.27, completion: 1.1 },
        'deepseek-chat': { prompt: 0.14, completion: 0.28 },
        'doubao-seed-2.0-code': { prompt: 0.3, completion: 0.6 },
        'doubao-seed-2.0-lite': { prompt: 0.3, completion: 0.6 },
        'doubao-seed-2.0-pro': { prompt: 0.8, completion: 2 },
        'doubao-seed-code': { prompt: 0.3, completion: 0.6 },
        'glm-4.7': { prompt: 1, completion: 1 },
        'glm-4': { prompt: 1, completion: 1 },
        'kimi-k2.5': { prompt: 0.3, completion: 0.3 },
        'minimax-m2.5': { prompt: 1.5, completion: 1.5 }
      }

      pricing.value = hardcodedPricing
      console.log('[Pricing] Fallback价格设置完成:', pricing.value)
      return hardcodedPricing
    }
  }

  /**
   * Send chat completion request (streaming)
   */
  const sendMessageStreaming = async (
    request: ChatCompletionRequest,
    onChunk: (content: string) => void,
    onDone: (response: ChatCompletionResponse) => void,
    onError: (error: Error) => void,
    abortSignal?: AbortSignal
  ) => {
    if (!apiKey.value) {
      onError(new Error('API key not set'))
      return
    }

    try {
      const response = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey.value}`
        },
        body: JSON.stringify({
          ...request,
          stream: true
        }),
        signal: abortSignal
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error?.message || `HTTP ${response.status}`)
      }

      const reader = response.body!.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let fullContent = ''
      let completionId = ''
      let modelName = request.model
      let finishReason = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (!line.trim() || !line.startsWith('data: ')) continue
          if (line === 'data: [DONE]') continue

          try {
            const jsonStr = line.slice(6) // Remove 'data: '
            const chunk: ChatCompletionChunk = JSON.parse(jsonStr)

            if (!completionId && chunk.id) {
              completionId = chunk.id
            }
            if (chunk.model) {
              modelName = chunk.model
            }

            const choice = chunk.choices?.[0]
            if (choice?.delta?.content) {
              const content = choice.delta.content
              fullContent += content
              onChunk(content)
            }

            if (choice?.finish_reason) {
              finishReason = choice.finish_reason
            }
          } catch (err) {
            console.warn('Failed to parse SSE chunk:', line, err)
          }
        }
      }

      // Construct final response (usage info may not be available in streaming)
      onDone({
        id: completionId,
        object: 'chat.completion',
        created: Math.floor(Date.now() / 1000),
        model: modelName,
        choices: [
          {
            index: 0,
            message: {
              role: 'assistant',
              content: fullContent
            },
            finish_reason: finishReason || 'stop'
          }
        ],
        usage: {
          prompt_tokens: 0, // Will be estimated
          completion_tokens: 0, // Will be estimated
          total_tokens: 0
        }
      })
    } catch (error: any) {
      if (error.name === 'AbortError') {
        return // User cancelled
      }
      onError(error)
    }
  }

  /**
   * Send chat completion request (non-streaming)
   */
  const sendMessage = async (
    request: ChatCompletionRequest
  ): Promise<ChatCompletionResponse> => {
    if (!apiKey.value) {
      throw new Error('API key not set')
    }

    const response = await fetch('/v1/chat/completions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${apiKey.value}`
      },
      body: JSON.stringify({
        ...request,
        stream: false
      })
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.error?.message || `HTTP ${response.status}`)
    }

    return response.json()
  }

  return {
    apiKey,
    isValidating,
    models,
    pricing,
    validateKey,
    setApiKey,
    clearApiKey,
    fetchModels,
    fetchPricing,
    sendMessageStreaming,
    sendMessage
  }
}
