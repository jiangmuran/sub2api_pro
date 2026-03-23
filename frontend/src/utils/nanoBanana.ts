export const nanoBananaModels = [
  'nano-banana-2',
  'nano-banana-2-cl',
  'nano-banana-2-4k-cl',
  'nano-banana-fast',
  'nano-banana',
  'nano-banana-pro',
  'nano-banana-pro-vt',
  'nano-banana-pro-cl',
  'nano-banana-pro-vip',
  'nano-banana-pro-4k-vip'
]

export const nanoBananaAspectRatioOptions = [
  'auto',
  '1:1',
  '16:9',
  '9:16',
  '4:3',
  '3:4',
  '3:2',
  '2:3',
  '5:4',
  '4:5',
  '21:9'
]

export const nanoBananaExtendedAspectRatioOptions = [
  'auto',
  ...nanoBananaAspectRatioOptions.filter((value) => value !== 'auto'),
  '1:4',
  '4:1',
  '1:8',
  '8:1'
]

export const nanoBananaImageSizeOptions = ['1K', '2K', '4K']

export function isNanoBananaModel(model?: string | null): boolean {
  return (model || '').trim().toLowerCase().startsWith('nano-banana')
}

export function getNanoBananaSupportedImageSizes(model?: string | null): string[] {
  const normalized = (model || '').trim().toLowerCase()
  switch (normalized) {
    case 'nano-banana-2-cl':
    case 'nano-banana-pro-vip':
      return ['1K', '2K']
    case 'nano-banana-2-4k-cl':
    case 'nano-banana-pro-4k-vip':
      return ['4K']
    case 'nano-banana-2':
    case 'nano-banana-pro':
    case 'nano-banana-pro-vt':
    case 'nano-banana-pro-cl':
      return ['1K', '2K', '4K']
    default:
      return ['1K']
  }
}

export function getNanoBananaDefaultImageSize(model?: string | null): string {
  const supported = getNanoBananaSupportedImageSizes(model)
  if (supported.includes('2K')) return '2K'
  return supported[0] || '1K'
}

export function getNanoBananaAspectRatioOptions(model?: string | null): string[] {
  const normalized = (model || '').trim().toLowerCase()
  if (normalized === 'nano-banana-2' || normalized === 'nano-banana-2-cl' || normalized === 'nano-banana-2-4k-cl') {
    return nanoBananaExtendedAspectRatioOptions
  }
  return nanoBananaAspectRatioOptions
}
