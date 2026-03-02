export interface User {
  id: number
  uid: string
  username: string
  email: string
  description: string
  metadata: Record<string, string>
  role_id: number
  create_at: number
  update_at: number
}

export interface Token {
  id: number
  text: string
  expire_at: number
  enable: number
  user_id: number
  role_id: number
  create_at: number
}

export interface Role {
  id: number
  type: RoleType
  name: string
  display_name: string
  description: string
  metadata: Record<string, string>
  create_at: number
  update_at: number
}

export interface Policy {
  subject: string
  object: string
  action: string
}

export enum RoleType {
  Unknown = 0,
  Admin = 1,
  System = 2,
  Operator = 3,
}

export interface Definitions {
  id: number
  name: string
  uid: string
  description: string
  metadata: Record<string, string>
  content: string
  version: number
  isExecute: boolean
  create_at: number
  update_at: number
}

export interface Process {
  id: number
  name: string
  definitions_uid: string
  definitions_version: number
  status: ProcessStatus
  stage: ProcessStage
  headers: Record<string, string>
  properties: Record<string, unknown>
  dataObjects: Record<string, unknown>
  results: Record<string, unknown>
  error?: string
  create_at: number
  update_at: number
}

export enum ProcessStatus {
  Created = 0,
  Ready = 1,
  Executing = 2,
  Commit = 3,
  Rollback = 4,
  Destroy = 5,
  Failed = 6,
}

export enum ProcessStage {
  None = 0,
  Commit = 1,
  Rollback = 2,
  Destroy = 3,
}

export interface Runner {
  id: number
  uid: string
  listen_url: string
  version: string
  heartbeat_ms: number
  hostname: string
  metadata: Record<string, string>
  features: Record<string, string>
  transport: string
  cpu: number
  memory: number
  online_timestamp: number
  offline_timestamp: number
  online: number
  state: number
}

export interface Endpoint {
  id: number
  task_type: number
  type: string
  name: string
  description: string
  mode: number
  http_url: string
  metadata: Record<string, string>
  targets: string[]
  headers: Record<string, string>
  properties: Record<string, unknown>
  dataObjects: Record<string, unknown>
  results: Record<string, unknown>
}

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  size: number
}

export interface ExecuteProcessRequest {
  name: string
  definitions_uid: string
  definitions_version?: number
  priority?: number
  mode?: number
  headers?: Record<string, string>
  properties?: Record<string, unknown>
  dataObjects?: Record<string, unknown>
}
