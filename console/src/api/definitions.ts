import { apiClient } from './client';
import type { Definitions } from '../types';

export interface ListDefinitionsParams {
  page?: number;
  size?: number;
  uid?: string;
  name?: string;
  [key: string]: unknown;
}

export interface ListDefinitionsResponse {
  definitionsList: Definitions[];
  total: number | string;
}

export interface DeployDefinitionsRequest {
  metadata?: Record<string, string>;
  content: string;
  description?: string;
}

export const definitionsApi = {
  list: (params?: ListDefinitionsParams) =>
    apiClient.get<ListDefinitionsResponse>('/definitions', params),

  get: (uid: string, version?: number) => {
    const params = version ? { version } : undefined;
    return apiClient.get<{ definitions: Definitions }>(`/definitions/${uid}`, params);
  },

  deploy: (data: DeployDefinitionsRequest) =>
    apiClient.post<{ definitions: Definitions }>('/definitions', data),

  remove: (uid: string, version?: number) => {
    const params = version ? { version } : undefined;
    return apiClient.delete<{ definitions: Definitions }>(`/definitions/${uid}`, params);
  },
};
