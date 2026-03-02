import { apiClient } from './client'
import type { LoginRequest, LoginResponse, GetSelfResponse, UpdateSelfRequest, UpdateSelfResponse } from './auth'

export const authApi = {
  login: async (username: string, password: string): Promise<LoginResponse> => {
    const request: LoginRequest = { username, password }
    return apiClient.post<LoginResponse>('/login', request)
  },

  getSelf: async (): Promise<GetSelfResponse> => {
    return apiClient.get<GetSelfResponse>('/users/self')
  },

  updateSelf: async (data: UpdateSelfRequest): Promise<UpdateSelfResponse> => {
    return apiClient.patch<UpdateSelfResponse>('/users/self', data)
  },
}
