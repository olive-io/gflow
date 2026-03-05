import { apiClient } from './client';
import type { User, Role } from '../types';

export interface ListUsersParams {
  page?: number;
  size?: number;
  [key: string]: unknown;
}

export interface ListUsersResponse {
  users: User[];
  total: number | string;
}

export interface CreateUserRequest {
  username: string;
  password: string;
  email?: string;
  description?: string;
  roleId?: number;
}

export interface UpdateUserRequest {
  email?: string;
  description?: string;
  password?: string;
  init?: number;
}

export interface ListRolesResponse {
  roles: Role[];
  total: number | string;
}

export const usersApi = {
  list: (params?: ListUsersParams) =>
    apiClient.get<ListUsersResponse>('/admin/users', params),

  create: (data: CreateUserRequest) =>
    apiClient.post<{ user: User }>('/admin/users', data),

  update: (id: number | string, data: UpdateUserRequest) =>
    apiClient.patch<{ user: User }>(`/admin/users/${id}`, data),

  remove: (id: number | string) =>
    apiClient.delete<{ user: User }>(`/admin/users/${id}`),

  listRoles: () =>
    apiClient.get<ListRolesResponse>('/admin/roles'),
};
