import { apiClient } from '../client'

export interface SecurityChatMessage {
  role: string
  content: string
  source: 'request' | 'response' | string
  index: number
}

export interface SecurityChatSession {
  session_id: string
  user_id?: number
  user_email?: string
  api_key_id?: number
  account_id?: number
  group_id?: number
  platform?: string
  model?: string
  message_preview?: string
  last_at: string
  request_count: number
}

export interface SecurityChatLog {
  id: number
  session_id: string
  request_id?: string
  client_request_id?: string
  user_id?: number
  api_key_id?: number
  account_id?: number
  group_id?: number
  platform?: string
  model?: string
  request_path?: string
  stream: boolean
  status_code?: number
  messages: SecurityChatMessage[]
  created_at: string
}

export interface SecurityChatSessionList {
  items: SecurityChatSession[]
  total: number
  page: number
  page_size: number
}

export interface SecurityChatLogList {
  items: SecurityChatLog[]
  total: number
  page: number
  page_size: number
}

export interface SecurityChatSummary {
  summary: string
  sensitive_findings?: string[]
  risk_level?: string
  recommended_actions?: string[]
}

export interface SecurityApiKey {
  id: number
  name: string
  group_id?: number | null
  status: string
}

export async function listSessions(params: Record<string, any>): Promise<SecurityChatSessionList> {
  const { data } = await apiClient.get<SecurityChatSessionList>('/admin/security/sessions', { params })
  return data
}

export async function listMessages(params: Record<string, any>): Promise<SecurityChatLogList> {
  const { data } = await apiClient.get<SecurityChatLogList>('/admin/security/messages', { params })
  return data
}

export async function summarize(payload: Record<string, any>): Promise<SecurityChatSummary> {
  const { data } = await apiClient.post<SecurityChatSummary>('/admin/security/summarize', payload)
  return data
}

export async function listApiKeys(): Promise<SecurityApiKey[]> {
  const { data } = await apiClient.get<SecurityApiKey[]>('/admin/security/api-keys')
  return data
}

export async function deleteSession(sessionId: string, params?: Record<string, any>): Promise<any> {
  const { data } = await apiClient.delete(`/admin/security/sessions/${sessionId}`, { params })
  return data
}

export async function bulkDeleteSessions(payload: Record<string, any>): Promise<any> {
  const { data } = await apiClient.post('/admin/security/sessions/bulk-delete', payload)
  return data
}

export async function chatWithAI(payload: Record<string, any>): Promise<SecurityChatSummary> {
  const { data } = await apiClient.post<SecurityChatSummary>('/admin/security/ai-chat', payload)
  return data
}

export default {
  listSessions,
  listMessages,
  summarize,
  listApiKeys,
  deleteSession,
  bulkDeleteSessions,
  chatWithAI
}
