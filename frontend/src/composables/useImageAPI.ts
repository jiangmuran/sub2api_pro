/**
 * Image generation API composable
 * Supports Grok image models via OpenAI-compatible endpoint
 */

import { ref } from 'vue'
import type { ImageGenerationRequest, ImageGenerationResponse, GeneratedImage } from '@/types/chat'

export function useImageAPI() {
  const isGenerating = ref(false)

  /**
   * Generate images
   */
  const generateImage = async (
    apiKey: string,
    request: ImageGenerationRequest
  ): Promise<GeneratedImage[]> => {
    if (!apiKey) {
      throw new Error('API key not set')
    }

    isGenerating.value = true

    try {
      const response = await fetch('/v1/images/generations', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`
        },
        body: JSON.stringify(request)
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error?.message || `HTTP ${response.status}`)
      }

      const data: ImageGenerationResponse = await response.json()

      // Convert to GeneratedImage format
      const images: GeneratedImage[] = await Promise.all(
        data.data.map(async (img, index) => {
          // Download and convert to data URL for offline storage
          const dataUrl = await fetchAsDataURL(img.url)

          return {
            id: crypto.randomUUID(),
            prompt: request.prompt,
            url: img.url,
            dataUrl,
            size: request.size || '1024x1024',
            model: request.model,
            timestamp: data.created * 1000 + index,
            cost: estimateImageCost(request.model, request.size),
            revisedPrompt: img.revised_prompt
          }
        })
      )

      return images
    } finally {
      isGenerating.value = false
    }
  }

  /**
   * Fetch image as data URL for offline storage
   */
  const fetchAsDataURL = async (url: string): Promise<string> => {
    try {
      const response = await fetch(url)
      const blob = await response.blob()

      return new Promise((resolve, reject) => {
        const reader = new FileReader()
        reader.onloadend = () => resolve(reader.result as string)
        reader.onerror = reject
        reader.readAsDataURL(blob)
      })
    } catch (error) {
      console.warn('Failed to cache image as data URL:', error)
      return url // Fallback to original URL
    }
  }

  /**
   * Estimate image generation cost
   */
  const estimateImageCost = (model: string, size?: string): number => {
    // Grok image pricing (adjust based on actual pricing)
    const grokPricing: Record<string, number> = {
      '1024x1024': 0.04,
      '1024x1792': 0.06,
      '1792x1024': 0.06,
      '1792x1792': 0.08
    }

    // DALL-E 3 pricing
    const dalleStandardPricing: Record<string, number> = {
      '1024x1024': 0.04,
      '1024x1792': 0.08,
      '1792x1024': 0.08
    }

    if (model.includes('grok')) {
      return grokPricing[size || '1024x1024'] || 0.04
    } else if (model.includes('dall-e-3')) {
      // Assume standard quality by default
      return dalleStandardPricing[size || '1024x1024'] || 0.04
    }

    return 0.04 // Default fallback
  }

  /**
   * Download image to local device
   */
  const downloadImage = async (image: GeneratedImage) => {
    try {
      // Use data URL if available, otherwise fetch original
      const url = image.dataUrl || image.url

      let blob: Blob

      if (url.startsWith('data:')) {
        // Convert data URL to blob
        const response = await fetch(url)
        blob = await response.blob()
      } else {
        // Fetch from URL
        const response = await fetch(url)
        blob = await response.blob()
      }

      // Create download link
      const blobUrl = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = blobUrl
      link.download = `image-${image.timestamp}.png`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(blobUrl)
    } catch (error) {
      console.error('Failed to download image:', error)
      throw error
    }
  }

  return {
    isGenerating,
    generateImage,
    downloadImage,
    estimateImageCost
  }
}
