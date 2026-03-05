import { apiClient } from './client';
import type { Process } from '../types';

export interface ListProcessParams {
  page?: number;
  size?: number;
  definitionsUid?: string;
  definitionsVersion?: number;
  processStatus?: number;
  processStage?: number;
  [key: string]: unknown;
}

export interface ListProcessResponse {
  processes: Process[];
  total: number | string;
}

interface RawFlowNode {
  id?: number | string;
  name?: string;
  status?: number | string;
  stage?: number | string;
  startTime?: number | string;
  endTime?: number | string;
  target?: string;
  errMsg?: string;
  message?: string;
}

interface RawProcess {
  id: number | string;
  name?: string;
  uid?: string;
  metadata?: Record<string, string>;
  priority?: number | string;
  mode?: number | string;
  args?: Process['args'];
  definitionsUid?: string;
  definitionsVersion?: number | string;
  definitionsProcess?: string;
  context?: Process['context'];
  attempts?: number;
  startAt: number | string;
  endAt: number | string;
  stage: number | string;
  status: number | string;
  errMsg?: string;
  createAt?: number | string;
  updateAt?: number | string;
}

export interface GetProcessResponse {
  process: Process;
  activities: RawFlowNode[];
  tasks?: RawFlowNode[];
}

export interface ExecuteProcessRequest {
  definitionsUid: string;
  definitionsVersion?: number;
  variables?: Record<string, unknown>;
}

const processStatusMap: Record<number, string> = {
  0: 'UnknownStatus',
  1: 'Waiting',
  2: 'Running',
  3: 'Success',
  4: 'Warn',
  5: 'Failed',
};

const processStageMap: Record<number, string> = {
  0: 'UnknownStage',
  1: 'Prepare',
  2: 'Ready',
  3: 'Commit',
  4: 'Rollback',
  5: 'Destroy',
  6: 'Finish',
};

const normalizeStatus = (status: number | string | undefined): string => {
  if (typeof status === 'string' && Number.isNaN(Number(status))) {
    return status;
  }
  return processStatusMap[Number(status ?? 0)] || 'UnknownStatus';
};

const normalizeStage = (stage: number | string | undefined): string => {
  if (typeof stage === 'string' && Number.isNaN(Number(stage))) {
    return stage;
  }
  return processStageMap[Number(stage ?? 0)] || 'UnknownStage';
};

const normalizeProcess = (item: RawProcess): Process => ({
  ...(item as Process),
  id: Number(item.id || 0),
  priority: item.priority !== undefined ? Number(item.priority) : undefined,
  mode: item.mode !== undefined ? String(item.mode) : undefined,
  definitionsVersion: item.definitionsVersion !== undefined ? Number(item.definitionsVersion) : undefined,
  status: normalizeStatus(item.status),
  stage: normalizeStage(item.stage),
  createAt: item.createAt !== undefined ? Number(item.createAt) : undefined,
  updateAt: item.updateAt !== undefined ? Number(item.updateAt) : undefined,
});

const normalizeFlowNode = (node: RawFlowNode): RawFlowNode => ({
  ...node,
  id: node.id !== undefined ? Number(node.id) : undefined,
  status: normalizeStatus(node.status),
  stage: normalizeStage(node.stage),
  message: node.errMsg,
});

export const processesApi = {
  list: async (params?: ListProcessParams): Promise<ListProcessResponse> => {
    const response = await apiClient.get<{ processes: RawProcess[]; total: number | string }>('/processes', params);
    return {
      processes: (response.processes || []).map(normalizeProcess),
      total: response.total,
    };
  },

  get: async (id: number): Promise<GetProcessResponse> => {
    const response = await apiClient.get<{ process: RawProcess; activities: RawFlowNode[]; tasks?: RawFlowNode[] }>(
      `/processes/${id}`,
    );
    return {
      process: normalizeProcess(response.process),
      activities: (response.activities || []).map(normalizeFlowNode),
      tasks: response.tasks?.map(normalizeFlowNode),
    };
  },

  execute: async (data: ExecuteProcessRequest): Promise<{ process: Process }> => {
    const response = await apiClient.post<{ process: RawProcess }>('/processes/execute', data);
    return { process: normalizeProcess(response.process) };
  },
};
