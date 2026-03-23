import { ref } from 'vue'

export interface ImageEditParams {
  apiKey: string
  model: string
  prompt: string
  image: File
  n?: number
  size?: string
  aspect_ratio?: string
  image_size?: string
  response_format?: 'url' | 'b64_json'
}

export interface EditedImage {
  url?: string
  b64_json?: string
  cost?: number
}

export interface ImageEditResponse {
  created: number
  data: EditedImage[]
}

export function useImageEditAPI() {
  const loading = ref(false)
  const error = ref<string | null>(null)

  const editImage = async (params: ImageEditParams): Promise<ImageEditResponse> => {
    loading.value = true
    error.value = null

    try {
      const formData = new FormData()
      formData.append('model', params.model)
      formData.append('prompt', params.prompt)
      formData.append('image', params.image)
      if (params.n) formData.append('n', params.n.toString())
      if (params.size) formData.append('size', params.size)
      if (params.aspect_ratio) formData.append('aspect_ratio', params.aspect_ratio)
      if (params.image_size) formData.append('image_size', params.image_size)
      if (params.response_format) formData.append('response_format', params.response_format)

      // 图片编辑可能需要30-60秒，设置较长的超时时间（90秒）
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), 90000) // 90 seconds

      const response = await fetch('/v1/images/edits', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${params.apiKey}`
        },
        body: formData,
        signal: controller.signal
      })

      clearTimeout(timeoutId)

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: { message: response.statusText } }))
        throw new Error(errorData.error?.message || `HTTP ${response.status}`)
      }

      const data = await response.json()
      
      // 估算成本（简化版，实际成本从后端记录）
      const cost = estimateCost(params.model, params.size, params.n || 1)
      if (data.data && Array.isArray(data.data)) {
        data.data = data.data.map((img: EditedImage) => ({ ...img, cost }))
      }

      return data
    } catch (err: unknown) {
      let errorMessage = 'Failed to edit image'
      
      if (err instanceof Error) {
        if (err.name === 'AbortError') {
          errorMessage = 'Request timeout (90s). Image editing may take longer, please try again.'
        } else {
          errorMessage = err.message
        }
      }
      
      error.value = errorMessage
      throw new Error(errorMessage)
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    editImage
  }
}

function estimateCost(model: string, size?: string, count = 1): number {
  const modelLower = model.toLowerCase()
  let basePrice = 0.04

  if (modelLower.includes('grok')) {
    if (size && (size.includes('1792') || size.includes('1280'))) {
      basePrice = 0.08
    } else {
      basePrice = 0.04
    }
  }

  return basePrice * count
}
