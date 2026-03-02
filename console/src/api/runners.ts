import { http } from '@/lib/http'
import type { Runner, ListResponse } from '@/types/api'

export interface ListRunnerParams {
  page?: number
  size?: number
  online?: number
}

export const runnersApi = {
  list: async (params?: ListRunnerParams): Promise<ListResponse<Runner>> => {
    return http.get<ListResponse<Runner>>('/v1/runners', params)
  },

  get: async (id: number): Promise<Runner> => {
    return http.get<Runner>(`/v1/runners/${id}`)
  },
}
