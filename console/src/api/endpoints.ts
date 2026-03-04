import { http } from '@/lib/http'
import type { Endpoint } from '@/types/api'

export interface ListEndpointParams {
  page?: number
  size?: number
  [key: string]: unknown
}

export interface ListEndpointResponse {
  endpoints: Endpoint[]
  total: number
}

export interface AddEndpointParams {
  endpoints: Partial<Endpoint>[]
  target?: string
}

export interface AddOpenAPIParams {
  doc: string
  contentType?: string
  target?: string
}

export interface AddGRPCParams {
  target: string
}

export interface AddOpenAPIResponse {
  runner: { uid: string }
  endpoints: Endpoint[]
}

export interface ConvertEndpointParams {
  id: number
  format?: 'BPMN'
}

export interface ConvertEndpointResponse {
  contentType: string
  content: string
}

export const endpointsApi = {
  list: async (params?: ListEndpointParams): Promise<ListEndpointResponse> => {
    return http.get<ListEndpointResponse>('/v1/endpoints', params)
  },

  add: async (params: AddEndpointParams): Promise<void> => {
    return http.post('/v1/endpoints', params)
  },

  addOpenAPI: async (params: AddOpenAPIParams): Promise<AddOpenAPIResponse> => {
    return http.post<AddOpenAPIResponse>('/v1/endpoints/openapi', params)
  },

  addGRPC: async (params: AddGRPCParams): Promise<AddOpenAPIResponse> => {
    return http.post<AddOpenAPIResponse>('/v1/endpoints/grpc', params)
  },

  convert: async (params: ConvertEndpointParams): Promise<ConvertEndpointResponse> => {
    return http.post<ConvertEndpointResponse>(`/v1/endpoints/${params.id}/convert`, {
      format: params.format || 'BPMN'
    })
  }
}
