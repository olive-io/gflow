import { http } from '@/lib/http'
import type { Definitions } from '@/types/api'

export interface ListDefinitionsParams {
  page?: number
  size?: number
  uid?: string
  name?: string
}

export interface ListDefinitionsResponse {
  definitionsList: Definitions[]
  total: number
}

export interface DeployDefinitionsRequest {
  metadata?: Record<string, string>
  content: string
  description?: string
}

export const definitionsApi = {
  list: async (params?: ListDefinitionsParams): Promise<ListDefinitionsResponse> => {
    return http.get<ListDefinitionsResponse>('/v1/definitions', params)
  },

  get: async (uid: string, version?: number): Promise<{ definitions: Definitions }> => {
    const params = version ? { version } : undefined
    return http.get<{ definitions: Definitions }>(`/v1/definitions/${uid}`, params)
  },

  deploy: async (data: DeployDefinitionsRequest): Promise<{ definitions: Definitions }> => {
    return http.post<{ definitions: Definitions }>('/v1/definitions', data)
  },

  remove: async (uid: string, version?: number): Promise<{ definitions: Definitions }> => {
    const params = version ? { version } : undefined
    return http.delete<{ definitions: Definitions }>(`/v1/definitions/${uid}`, params)
  },
}
