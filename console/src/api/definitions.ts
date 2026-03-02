import { http } from '@/lib/http'
import type { Definitions, ListResponse } from '@/types/api'

export interface ListDefinitionsParams {
  page?: number
  size?: number
  uid?: string
  name?: string
}

export interface DeployDefinitionsRequest {
  metadata?: Record<string, string>
  content: string
  description?: string
}

export const definitionsApi = {
  list: async (params?: ListDefinitionsParams): Promise<ListResponse<Definitions>> => {
    return http.get<ListResponse<Definitions>>('/v1/definitions', params)
  },

  get: async (uid: string, version?: number): Promise<Definitions> => {
    const params = version ? { version } : undefined
    return http.get<Definitions>(`/v1/definitions/${uid}`, params)
  },

  deploy: async (data: DeployDefinitionsRequest): Promise<Definitions> => {
    return http.post<Definitions>('/v1/definitions', data)
  },

  remove: async (uid: string, version?: number): Promise<void> => {
    const params = version ? { version } : undefined
    return http.delete(`/v1/definitions/${uid}`, params)
  },
}
