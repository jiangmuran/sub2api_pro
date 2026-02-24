import { apiClient } from '../client'
import type { DistributorAdminSummary, DistributorOffer, DistributorOrder, DistributorProfile, PaginatedResponse } from '@/types'

const distributorsAPI = {
  async listProfiles(page = 1, pageSize = 20, search = ''): Promise<PaginatedResponse<DistributorProfile>> {
    const { data } = await apiClient.get<PaginatedResponse<DistributorProfile>>('/admin/distributors/profiles', {
      params: {
        page,
        page_size: pageSize,
        search: search || undefined
      }
    })
    return data
  },
  async upsertProfile(email: string, enabled: boolean, notes = ''): Promise<DistributorProfile> {
    const { data } = await apiClient.post<DistributorProfile>('/admin/distributors/profiles', { email, enabled, notes })
    return data
  },
  async adjustBalance(userId: number, operation: 'topup' | 'refund', amountCNY: number, notes = ''): Promise<DistributorProfile> {
    const { data } = await apiClient.post<DistributorProfile>(`/admin/distributors/profiles/${userId}/balance`, {
      operation,
      amount_cny: amountCNY,
      notes
    })
    return data
  },
  async listOffers(distributorUserId: number): Promise<DistributorOffer[]> {
    const { data } = await apiClient.get<DistributorOffer[]>('/admin/distributors/offers', {
      params: { distributor_user_id: distributorUserId }
    })
    return data
  },
  async createOffer(payload: {
    distributor_user_id: number
    name: string
    target_group_id: number
    validity_days: number
    cost_cny: number
    enabled: boolean
    notes?: string
  }): Promise<DistributorOffer> {
    const { data } = await apiClient.post<DistributorOffer>('/admin/distributors/offers', payload)
    return data
  },
  async updateOffer(id: number, payload: {
    distributor_user_id: number
    name: string
    target_group_id: number
    validity_days: number
    cost_cny: number
    enabled: boolean
    notes?: string
  }): Promise<DistributorOffer> {
    const { data } = await apiClient.put<DistributorOffer>(`/admin/distributors/offers/${id}`, payload)
    return data
  },
  async deleteOffer(id: number): Promise<{ deleted: boolean }> {
    const { data } = await apiClient.delete<{ deleted: boolean }>(`/admin/distributors/offers/${id}`)
    return data
  },
  async listOrders(page = 1, pageSize = 20, params?: { distributor_user_id?: number; status?: string; search?: string }): Promise<PaginatedResponse<DistributorOrder>> {
    const { data } = await apiClient.get<PaginatedResponse<DistributorOrder>>('/admin/distributors/orders', {
      params: {
        page,
        page_size: pageSize,
        distributor_user_id: params?.distributor_user_id,
        status: params?.status,
        search: params?.search
      }
    })
    return data
  },
  async revokeOrder(id: number, notes = ''): Promise<DistributorOrder> {
    const { data } = await apiClient.post<DistributorOrder>(`/admin/distributors/orders/${id}/revoke`, { notes })
    return data
  },
  async summary(): Promise<DistributorAdminSummary> {
    const { data } = await apiClient.get<DistributorAdminSummary>('/admin/distributors/summary')
    return data
  },
  async settle(notes = ''): Promise<DistributorAdminSummary> {
    const { data } = await apiClient.post<DistributorAdminSummary>('/admin/distributors/summary/settle', { notes })
    return data
  }
}

export default distributorsAPI
