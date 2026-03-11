import { describe, expect, it } from 'vitest'

import { normalizeCommunityLink } from '@/utils/community-link'

describe('normalizeCommunityLink', () => {
  it('keeps valid http links', () => {
    expect(normalizeCommunityLink('https://t.me/sub2api')).toBe('https://t.me/sub2api')
  })

  it('rejects non-http protocols', () => {
    expect(normalizeCommunityLink('javascript:alert(1)')).toBe('')
    expect(normalizeCommunityLink('data:text/html,hello')).toBe('')
  })

  it('returns empty for invalid values', () => {
    expect(normalizeCommunityLink('')).toBe('')
    expect(normalizeCommunityLink('qq:123456')).toBe('')
    expect(normalizeCommunityLink('not-a-url')).toBe('')
  })
})
