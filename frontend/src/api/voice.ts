import { apiClient } from './client'
import type { VoicePreflightResponse, VoiceSessionResponse } from '@/types'

export const voiceAPI = {
  async preflight(apiKey: string): Promise<VoicePreflightResponse> {
    const { data } = await apiClient.post<VoicePreflightResponse>('/voice/preflight', { api_key: apiKey })
    if (!data) {
      throw new Error('Voice preflight failed')
    }
    return data
  },

  async createSession(payload: { api_key: string; voice: string; personality: string; speed: number }): Promise<VoiceSessionResponse> {
    const { data } = await apiClient.post<VoiceSessionResponse>('/voice/session', payload)
    if (!data) {
      throw new Error('Voice session failed')
    }
    return data
  }
}

export default voiceAPI
