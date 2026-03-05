export interface User {
  id: string | number;
  uid: string;
  username: string;
  email: string;
  description?: string;
  avatar?: string;
  metadata?: Record<string, string>;
  roleId?: string | number;
  createAt?: number | string;
  updateAt?: number | string;
}

export interface Role {
  id: string | number;
  type: RoleType;
  name: string;
  displayName?: string;
  description?: string;
  metadata?: Record<string, string>;
  createAt?: number;
  updateAt?: number;
}

export interface Policy {
  subject: string;
  object: string;
  action: string;
}

export const RoleTypes = {
  Unknown: 0,
  Admin: 1,
  System: 2,
  Operator: 3,
} as const;

export type RoleType = (typeof RoleTypes)[keyof typeof RoleTypes];

export interface Token {
  id: string;
  text: string;
  expireAt: string;
  createAt: string;
  enable: number;
  userId: string;
  roleId: string;
}

export interface Definitions {
  id: number;
  uid: string;
  name: string;
  description?: string;
  version: number;
  content?: string;
  isExecute: boolean;
  metadata?: Record<string, string>;
  createAt: number | string;
  updateAt: number | string;
}

export interface BpmnArgs {
  headers: Record<string, string>;
  properties: Record<string, Value>;
  dataObjects: Record<string, Value>;
}

export interface Value {
  type: ValueType;
  kind: string;
  value: string;
  default: string;
}

export const ValueTypes = {
  String: 0,
  Integer: 1,
  Float: 2,
  Boolean: 3,
  Array: 4,
  Object: 5,
} as const;

export type ValueType = (typeof ValueTypes)[keyof typeof ValueTypes];

export interface ProcessContext {
  variables: Record<string, Value>;
  dataObjects: Record<string, Value>;
}

export interface Process {
  id: number;
  name?: string;
  uid?: string;
  metadata?: Record<string, string>;
  priority?: number;
  mode?: string;
  args?: BpmnArgs;
  definitionsUid?: string;
  definitionsVersion?: number;
  definitionsProcess?: string;
  context?: ProcessContext;
  attempts?: number;
  startAt: number | string;
  endAt: number | string;
  stage: string;
  status: string;
  errMsg?: string;
  activities?: Activity[];
  createAt?: number;
  updateAt?: number;
}

export interface Activity {
  id: number;
  name: string;
  flowId: string;
  flowType: string;
  type: string;
  target: string;
  mode: string;
  headers: Record<string, string>;
  properties: Record<string, Value>;
  dataObjects: Record<string, Value>;
  results: Record<string, Value>;
  retries: number;
  startTime: number;
  endTime: number;
  processId: number;
  processUid: string;
  traceId: string;
  stage: string;
  status: string;
  errMsg: string;
  createAt: number;
}

export const TransitionModes = {
  UnknownMode: 0,
  Simple: 1,
  Transition: 2,
} as const;

export type TransitionMode = (typeof TransitionModes)[keyof typeof TransitionModes];

export const ProcessStatuses = {
  UnknownStatus: 0,
  Waiting: 1,
  Running: 2,
  Success: 3,
  Warn: 4,
  Failed: 5,
} as const;

export type ProcessStatus = (typeof ProcessStatuses)[keyof typeof ProcessStatuses];

export const ProcessStages = {
  UnknownStage: 0,
  Prepare: 1,
  Ready: 2,
  Commit: 3,
  Rollback: 4,
  Destroy: 5,
  Finish: 6,
} as const;

export type ProcessStage = (typeof ProcessStages)[keyof typeof ProcessStages];

export interface Runner {
  id: string;
  uid: string;
  listenUrl: string;
  version?: string;
  heartbeatMs?: number;
  hostname?: string;
  metadata?: Record<string, string>;
  features?: Record<string, string>;
  transport?: string;
  cpu?: number;
  memory?: number;
  onlineTimestamp?: number;
  offlineTimestamp?: number;
  online: number;
  state: string;
  message?: string;
  createAt?: number;
}

export interface EndpointProperty {
  type: string;
  kind: string;
  value: string;
  default: string;
}

export interface Endpoint {
  id: number;
  taskType: number | string;
  type: string;
  name: string;
  description?: string;
  mode: number | string;
  httpUrl?: string;
  metadata?: Record<string, string>;
  targets?: string[];
  headers?: Record<string, string>;
  properties?: Record<string, EndpointProperty>;
  dataObjects?: Record<string, unknown>;
  results?: Record<string, EndpointProperty>;
  createAt?: number;
  updateAt?: number;
}

export interface ListResponse<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
}

export interface ExecuteProcessRequest {
  name: string;
  definitionsUid: string;
  definitionsVersion?: number;
  priority?: number;
  mode?: number;
  headers?: Record<string, string>;
  properties?: Record<string, unknown>;
  dataObjects?: Record<string, unknown>;
}

export interface LoginResponse {
  token: Token;
}

export interface SelfResponse {
  user: User;
  role: Role;
  policies: Policy[];
}
