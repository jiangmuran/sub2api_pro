import { beforeEach, describe, expect, it } from 'vitest'

import { buildEmbeddedUrl } from '@/utils/embedded-url'

describe('buildEmbeddedUrl', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/account/purchase?foo=bar')
  })

  it('appends invitation source for invited users', () => {
    const url = buildEmbeddedUrl('https://pay.example.com/checkout', 12, 'token-123', 'dark', 'INV123')
    const parsed = new URL(url)

    expect(parsed.searchParams.get('user_id')).toBe('12')
    expect(parsed.searchParams.get('token')).toBe('token-123')
    expect(parsed.searchParams.get('theme')).toBe('dark')
    expect(parsed.searchParams.get('ui_mode')).toBe('embedded')
    expect(parsed.searchParams.get('src_host')).toBe(window.location.origin)
    expect(parsed.searchParams.get('source')).toBe('api_INV123')
  })

  it('falls back to api source when invitation code is missing', () => {
    const url = buildEmbeddedUrl('https://pay.example.com/checkout?source=legacy', 12, 'token-123', 'light')
    const parsed = new URL(url)

    expect(parsed.searchParams.get('source')).toBe('api')
  })
})
