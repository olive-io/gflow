import { http } from '@/lib/http'
import type { Process, ListResponse, ExecuteProcessRequest } from '@/types/api'

export interface ListProcessParams {
  page?: number
  size?: number
  definitions_uid?: string
  definitions_version?: number
  process_status?: number
  process_stage?: number
}

export interface ListProcessResponse {
  processes: Process[]
  total: number
}

export interface GetProcessResponse {
  process: Process
  activities: unknown[]
}

export const processesApi = {
  list: async (params?: ListProcessParams): Promise<ListProcessResponse> => {
    return http.get<ListProcessResponse>('/v1/processes', params)
  },

  get: async (id: number): Promise<GetProcessResponse> => {
    return http.get<GetProcessResponse>(`/v1/processes/${id}`)
  },

  execute: async (data: ExecuteProcessRequest): Promise<{ process: Process }> => {
    return http.post<{ process: Process }>('/v1/processes/execute', data)
  },
}
