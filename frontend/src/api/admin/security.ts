import { apiClient } from '../client'

export interface SecurityPlatformShare {
  count: number
  ratio: number
}

export interface SecurityChatStats {
  start_time: string
  end_time: string
  request_count: number
  session_count: number
  avg_requests_per_day: number
  avg_sessions_per_day: number
  estimated_bytes: number
  table_bytes: number
  platform_share: {
    opencode: SecurityPlatformShare
    codex: SecurityPlatformShare
    other: SecurityPlatformShare
  }
  platform_share_basis: string
}

export interface SecurityChatDeleteResult {
  logs_deleted: number
  sessions_deleted: number
}

export interface SecurityChatMessage {
  role: string
  content: string
  source?: string
  index?: number
}

export interface SecurityChatLog {
  id: number
  session_id: string
  request_id?: string
  client_request_id?: string
  user_id?: number
  user_email?: string
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

export interface SecurityChatLogList {
  items: SecurityChatLog[]
  total: number
  page: number
  page_size: number
}

export async function getStats(params: Record<string, any>): Promise<SecurityChatStats> {
  const { data } = await apiClient.get<SecurityChatStats>('/admin/security/stats', { params })
  return data
}

export async function exportLogs(params: Record<string, any>): Promise<Blob> {
  const { data } = await apiClient.get('/admin/security/logs/export', {
    params,
    responseType: 'blob',
    timeout: 0
  })
  return data
}

export async function listLogs(params: Record<string, any>): Promise<SecurityChatLogList> {
  const { data } = await apiClient.get<SecurityChatLogList>('/admin/security/logs', {
    params,
    timeout: 0
  })
  return data
}

export async function deleteLogs(payload: Record<string, any>): Promise<SecurityChatDeleteResult> {
  const { data } = await apiClient.post<SecurityChatDeleteResult>('/admin/security/logs/delete', payload)
  return data
}

export default {
  getStats,
  exportLogs,
  listLogs,
  deleteLogs
}
