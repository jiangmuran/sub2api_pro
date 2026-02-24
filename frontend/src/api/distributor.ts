import { apiClient } from './client'
import type { DistributorOffer, DistributorOrder, DistributorProfile, PaginatedResponse } from '@/types'

const distributorAPI = {
  async profile(): Promise<DistributorProfile> {
    const { data } = await apiClient.get<DistributorProfile>('/user/distributor/profile')
    return data
  },
  async offers(): Promise<DistributorOffer[]> {
    const { data } = await apiClient.get<DistributorOffer[]>('/user/distributor/offers')
    return data
  },
  async orders(page = 1, pageSize = 20, params?: { status?: string; search?: string }): Promise<PaginatedResponse<DistributorOrder>> {
    const { data } = await apiClient.get<PaginatedResponse<DistributorOrder>>('/user/distributor/orders', {
      params: {
        page,
        page_size: pageSize,
        status: params?.status,
        search: params?.search
      }
    })
    return data
  },
  async createOrder(offerId: number, sellPriceCNY: number, memo?: string): Promise<DistributorOrder> {
    const { data } = await apiClient.post<DistributorOrder>('/user/distributor/orders', {
      offer_id: offerId,
      sell_price_cny: sellPriceCNY,
      memo
    })
    return data
  },
  async revokeOrder(orderId: number, notes?: string): Promise<DistributorOrder> {
    const { data } = await apiClient.post<DistributorOrder>(`/user/distributor/orders/${orderId}/revoke`, { notes })
    return data
  }
}

export default distributorAPI
