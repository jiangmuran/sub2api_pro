import { describe, expect, it } from 'vitest'
import { buildModelMappingObject, getModelsByPlatform } from '../useModelWhitelist'

describe('useModelWhitelist', () => {
  it('antigravity 模型列表包含图片模型兼容项', () => {
    const models = getModelsByPlatform('antigravity')

    expect(models).toContain('gemini-3.1-flash-image')
    expect(models).toContain('gemini-3-pro-image')
  })

  it('whitelist 模式会忽略通配符条目', () => {
    const mapping = buildModelMappingObject('whitelist', ['claude-*', 'gemini-3.1-flash-image'], [])
    expect(mapping).toEqual({
      'gemini-3.1-flash-image': 'gemini-3.1-flash-image'
    })
  })

  it('支持白名单与映射并存', () => {
    const mapping = buildModelMappingObject(
      'mapping',
      ['gpt-4o', 'gpt-4o-mini'],
      [
        { from: 'gpt-4o-mini', to: 'gpt-4.1-mini' },
        { from: 'claude-3-5-sonnet-20241022', to: 'claude-3-7-sonnet-20250219' }
      ]
    )

    expect(mapping).toEqual({
      'gpt-4o': 'gpt-4o',
      'gpt-4o-mini': 'gpt-4.1-mini',
      'claude-3-5-sonnet-20241022': 'claude-3-7-sonnet-20250219'
    })
  })
})
