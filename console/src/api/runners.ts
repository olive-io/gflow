import { apiClient } from './client';
import type { Runner } from '../types';

export interface ListRunnerParams {
  page?: number;
  size?: number;
  online?: number;
  [key: string]: unknown;
}

export interface ListRunnerResponse {
  runners: Runner[];
  total?: number;
}

interface RawRunner extends Omit<Runner, 'id' | 'online' | 'createAt' | 'heartbeatMs' | 'cpu' | 'memory'> {
  id: string | number;
  online: string | number;
  createAt?: string | number;
  heartbeatMs?: string | number;
  cpu?: string | number;
  memory?: string | number;
}

const normalizeRunner = (runner: RawRunner): Runner => ({
  ...runner,
  id: String(runner.id),
  online: Number(runner.online || 0),
  createAt: runner.createAt !== undefined ? Number(runner.createAt) : undefined,
  heartbeatMs: runner.heartbeatMs !== undefined ? Number(runner.heartbeatMs) : undefined,
  cpu: runner.cpu !== undefined ? Number(runner.cpu) : undefined,
  memory: runner.memory !== undefined ? Number(runner.memory) : undefined,
});

export const runnersApi = {
  list: async (params?: ListRunnerParams): Promise<ListRunnerResponse> => {
    const response = await apiClient.get<{ runners: RawRunner[]; total?: number | string }>('/runners', params);
    return {
      runners: (response.runners || []).map(normalizeRunner),
      total: response.total !== undefined ? Number(response.total) : undefined,
    };
  },

  get: async (id: number): Promise<{ runner: Runner }> => {
    const response = await apiClient.get<{ runner: RawRunner }>(`/runners/${id}`);
    return { runner: normalizeRunner(response.runner) };
  },
};
