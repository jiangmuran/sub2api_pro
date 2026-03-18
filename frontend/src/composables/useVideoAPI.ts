import { ref } from 'vue'

export interface VideoGenerationParams {
  apiKey: string
  model: string
  prompt: string
  size?: string // '1280x720', '720x1280', '1792x1024', '1024x1792', '1024x1024'
  seconds?: number // 6-30 seconds
  quality?: 'standard' | 'high'
  image_reference?: {
    image_url: string
  }
}

export interface GeneratedVideo {
  url: string
  cost?: number
}

export interface VideoGenerationResponse {
  created: number
  data: GeneratedVideo[]
}

export function useVideoAPI() {
  const loading = ref(false)
  const error = ref<string | null>(null)
  const progress = ref<string>('')

  const generateVideo = async (params: VideoGenerationParams): Promise<VideoGenerationResponse> => {
    loading.value = true
    error.value = null
    progress.value = 'Sending request...'

    try {
      // 视频生成时间：6秒视频约30-60秒，30秒视频需要多轮可能需要3-5分钟
      // 考虑网络延迟，设置超时时间为5分钟（300秒）
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 300000) // 300 seconds (5 minutes)

      const response = await fetch('/v1/videos', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${params.apiKey}`
        },
        body: JSON.stringify({
          model: params.model,
          prompt: params.prompt,
          size: params.size || '1280x720',
          seconds: params.seconds || 6,
          quality: params.quality || 'standard',
          ...(params.image_reference && { image_reference: params.image_reference })
        }),
        signal: controller.signal
      })

      clearTimeout(timeoutId)

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: { message: response.statusText } }))
        throw new Error(errorData.error?.message || `HTTP ${response.status}`)
      }

      progress.value = 'Processing video...'
      const data = await response.json()
      
      // 估算成本
      const cost = estimateVideoCost(params.model, params.quality || 'standard')
      if (data.data && Array.isArray(data.data)) {
        data.data = data.data.map((video: GeneratedVideo) => ({ ...video, cost }))
      }

      progress.value = 'Completed'
      return data
    } catch (err: unknown) {
      let errorMessage = 'Failed to generate video'
      
      if (err instanceof Error) {
        if (err.name === 'AbortError') {
          errorMessage = 'Request timeout (5 minutes). Video generation is taking longer than expected. For 30s videos, this may happen during peak hours. Please try again later or try a shorter duration.'
        } else {
          errorMessage = err.message
        }
      }
      
      error.value = errorMessage
      progress.value = ''
      throw new Error(errorMessage)
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    progress,
    generateVideo
  }
}

function estimateVideoCost(model: string, quality: string): number {
  const modelLower = model.toLowerCase()
  
  if (modelLower.includes('grok') && modelLower.includes('video')) {
    return quality === 'high' ? 0.20 : 0.10
  }
  
  // 默认价格
  return quality === 'high' ? 0.15 : 0.08
}
