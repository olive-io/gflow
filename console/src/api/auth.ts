import { apiClient } from './client';
import type { LoginResponse, SelfResponse } from '../types';

export const authApi = {
  login: (username: string, password: string) =>
    apiClient.post<LoginResponse>('/login', { username, password }),

  getSelf: () => apiClient.get<SelfResponse>('/users/self'),

  logout: () => apiClient.post('/logout'),
};
