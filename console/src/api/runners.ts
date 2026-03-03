import { http } from '@/lib/http'
import type { Runner } from '@/types/api'

export interface ListRunnerParams {
  page?: number
  size?: number
  online?: number
}

export interface ListRunnerResponse {
  runners: Runner[]
  total?: number
}

export const runnersApi = {
  list: async (params?: ListRunnerParams): Promise<ListRunnerResponse> => {
    return http.get<ListRunnerResponse>('/v1/runners', params)
  },

  get: async (id: number): Promise<{ runner: Runner }> => {
    return http.get<{ runner: Runner }>(`/v1/runners/${id}`)
  },
}
