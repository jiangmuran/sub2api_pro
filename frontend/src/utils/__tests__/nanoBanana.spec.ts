import { describe, expect, it } from 'vitest'
import {
  getNanoBananaAspectRatioOptions,
  getNanoBananaDefaultImageSize,
  getNanoBananaSupportedImageSizes,
  isNanoBananaModel,
  nanoBananaAspectRatioOptions,
  nanoBananaImageSizeOptions,
  nanoBananaModels
} from '@/utils/nanoBanana'

describe('nanoBanana utils', () => {
  it('detects nano banana models by prefix', () => {
    expect(isNanoBananaModel('nano-banana-fast')).toBe(true)
    expect(isNanoBananaModel(' Nano-Banana-Pro ')).toBe(true)
    expect(isNanoBananaModel('gpt-image-1')).toBe(false)
  })

  it('exposes supported selector options', () => {
    expect(nanoBananaModels).toContain('nano-banana-fast')
    expect(nanoBananaAspectRatioOptions).toContain('16:9')
    expect(nanoBananaImageSizeOptions).toEqual(['1K', '2K', '4K'])
  })

  it('resolves model-aware size defaults and aspect ratios', () => {
    expect(getNanoBananaSupportedImageSizes('nano-banana-pro-vip')).toEqual(['1K', '2K'])
    expect(getNanoBananaSupportedImageSizes('nano-banana-2-4k-cl')).toEqual(['4K'])
    expect(getNanoBananaDefaultImageSize('nano-banana-pro')).toBe('2K')
    expect(getNanoBananaDefaultImageSize('nano-banana-2-4k-cl')).toBe('4K')
    expect(getNanoBananaAspectRatioOptions('nano-banana-2')).toContain('1:8')
    expect(getNanoBananaAspectRatioOptions('nano-banana-fast')).not.toContain('1:8')
  })
})
