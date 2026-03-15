import { apiClient } from './client'
import type { VoicePreflightResponse, VoiceSessionResponse } from '@/types'

export const voiceAPI = {
  async preflight(apiKey: string): Promise<VoicePreflightResponse> {
    const { data } = await apiClient.post<{ data: VoicePreflightResponse }>('/voice/preflight', { api_key: apiKey })
    return data.data
  },

  async createSession(payload: { api_key: string; voice: string; personality: string; speed: number }): Promise<VoiceSessionResponse> {
    const { data } = await apiClient.post<{ data: VoiceSessionResponse }>('/voice/session', payload)
    return data.data
  }
}

export default voiceAPI
