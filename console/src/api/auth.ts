import { http } from '@/lib/http'
import type { User, Token, Role, Policy } from '@/types/api'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: Token
}

export interface GetSelfResponse {
  user: User
  role: Role
  policies: Policy[]
}

export interface UpdateSelfRequest {
  oldPassword?: string
  password?: string
  email?: string
  description?: string
  metadata?: Record<string, string>
}

export const authApi = {
  login: async (username: string, password: string): Promise<LoginResponse> => {
    return http.post<LoginResponse>('/v1/login', {
      username,
      password,
    })
  },

  getSelf: async (): Promise<GetSelfResponse> => {
    return http.get<GetSelfResponse>('/v1/users/self')
  },

  updateSelf: async (data: UpdateSelfRequest): Promise<GetSelfResponse> => {
    return http.patch<GetSelfResponse>('/v1/users/self', data)
  },
}
