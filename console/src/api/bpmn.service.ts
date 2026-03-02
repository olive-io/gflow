import { apiClient } from './client'
import type {
  DeployDefinitionsRequest,
  DeployDefinitionsResponse,
  ListDefinitionsRequest,
  ListDefinitionsResponse,
  GetDefinitionsRequest,
  GetDefinitionsResponse,
  ExecuteProcessRequest,
  ExecuteProcessResponse,
  ListProcessRequest,
  ListProcessResponse,
  GetProcessResponse,
} from './bpmn'

export const bpmnApi = {
  deployDefinition: async (content: string, metadata?: Record<string, string>, description?: string): Promise<DeployDefinitionsResponse> => {
    const request: DeployDefinitionsRequest = { content, metadata, description }
    return apiClient.post<DeployDefinitionsResponse>('/definitions', request)
  },

  listDefinitions: async (page: number = 1, size: number = 20): Promise<ListDefinitionsResponse> => {
    const request: ListDefinitionsRequest = { page, size }
    return apiClient.get<ListDefinitionsResponse>('/definitions', request as unknown as Record<string, string | number | boolean>)
  },

  getDefinition: async (uid: string, version?: number): Promise<GetDefinitionsResponse> => {
    const params: GetDefinitionsRequest = { uid, version }
    return apiClient.get<GetDefinitionsResponse>(`/definitions/${uid}`, params as unknown as Record<string, string | number | boolean>)
  },

  removeDefinition: async (uid: string): Promise<void> => {
    await apiClient.delete(`/definitions/${uid}`)
  },

  executeProcess: async (request: ExecuteProcessRequest): Promise<ExecuteProcessResponse> => {
    return apiClient.post<ExecuteProcessResponse>('/processes/execute', request)
  },

  listProcesses: async (request: ListProcessRequest): Promise<ListProcessResponse> => {
    return apiClient.get<ListProcessResponse>('/processes', request as unknown as Record<string, string | number | boolean>)
  },

  getProcess: async (id: number): Promise<GetProcessResponse> => {
    return apiClient.get<GetProcessResponse>(`/processes/${id}`)
  },
}
