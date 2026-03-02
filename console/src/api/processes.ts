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

export const processesApi = {
  list: async (params?: ListProcessParams): Promise<ListResponse<Process>> => {
    return http.get<ListResponse<Process>>('/v1/processes', params)
  },

  get: async (id: number): Promise<Process> => {
    return http.get<Process>(`/v1/processes/${id}`)
  },

  execute: async (data: ExecuteProcessRequest): Promise<Process> => {
    return http.post<Process>('/v1/processes/execute', data)
  },
}
