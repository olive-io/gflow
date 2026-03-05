import { apiClient } from './client';
import type { Endpoint } from '../types';

export interface ListEndpointParams {
  page?: number;
  size?: number;
  [key: string]: unknown;
}

export interface ListEndpointResponse {
  endpoints: Endpoint[];
  total: number;
}

export interface AddEndpointParams {
  endpoints: Partial<Endpoint>[];
  target?: string;
}

export interface AddOpenAPIParams {
  doc: string;
  contentType?: string;
  target?: string;
}

export interface AddGRPCParams {
  target: string;
}

export interface AddOpenAPIResponse {
  runner: { uid: string };
  endpoints: Endpoint[];
}

export interface ConvertEndpointParams {
  id: number;
  format?: 'BPMN';
}

export interface ConvertEndpointResponse {
  contentType: string;
  content: string;
}

export const endpointsApi = {
  list: (params?: ListEndpointParams) =>
    apiClient.get<ListEndpointResponse>('/endpoints', params),

  add: (params: AddEndpointParams) =>
    apiClient.post('/endpoints', params),

  addOpenAPI: (params: AddOpenAPIParams) =>
    apiClient.post<AddOpenAPIResponse>('/endpoints/openapi', params),

  addGRPC: (params: AddGRPCParams) =>
    apiClient.post<AddOpenAPIResponse>('/endpoints/grpc', params),

  convert: (params: ConvertEndpointParams) =>
    apiClient.post<ConvertEndpointResponse>(`/endpoints/${params.id}/convert`, {
      format: params.format || 'BPMN'
    }),
};
