export function normalizeCommunityLink(value?: string | null): string {
  const trimmed = value?.trim() ?? ''
  if (!trimmed) return ''

  try {
    const url = new URL(trimmed)
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return ''
    }
    return url.toString()
  } catch {
    return ''
  }
}
