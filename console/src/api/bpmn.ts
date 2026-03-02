export interface Definitions {
  id: number
  uid: string
  create_at: number
  update_at: number
  version: number
  name: string
  xml: string
  description: string
  metadata: Record<string, string>
}

export interface Process {
  id: number
  create_at: number
  update_at: number
  definitions_id: number
  definitions_uid: string
  definitions_version: number
  name: string
  status: ProcessStatus
  stage: ProcessStage
  mode: TransitionMode
  priority: number
  headers: Record<string, string>
  properties: Record<string, Value>
  data_objects: Record<string, Value>
  result: Record<string, Value>
  start_time: number
  end_time: number
  error_message: string
}

export interface FlowNode {
  id: number
  create_at: number
  update_at: number
  process_id: number
  flow_type: FlowNodeType
  name: string
  type: string
  target: string
  status: FlowNodeStatus
  stage: CallTaskStage
  headers: Record<string, string>
  properties: Record<string, Value>
  data_objects: Record<string, Value>
  result: Record<string, Value>
  start_time: number
  end_time: number
  error_message: string
}

export interface Value {
  type: ValueType
  value: string
}

export enum ValueType {
  String = 0,
  Integer = 1,
  Float = 2,
  Boolean = 3,
  Array = 4,
  Object = 5,
}

export enum ProcessStatus {
  Unknown = 0,
  Ready = 1,
  Running = 2,
  Completed = 3,
  Failed = 4,
  Cancelled = 5,
  Compensated = 6,
}

export enum ProcessStage {
  Unknown = 0,
  Ready = 1,
  Commit = 2,
  Rollback = 3,
  Destroy = 4,
  Done = 5,
}

export enum TransitionMode {
  Unknown = 0,
  Simple = 1,
  Transition = 2,
}

export enum FlowNodeType {
  UnknownNode = 0,
  StartEvent = 1,
  EndEvent = 2,
  BoundaryEvent = 3,
  IntermediateCatchEvent = 4,
  Task = 11,
  SendTask = 12,
  ReceiveTask = 13,
  ServiceTask = 14,
  UserTask = 15,
  ScriptTask = 16,
  ManualTask = 17,
  CallActivity = 18,
  BusinessRuleTask = 19,
  SubProcess = 20,
  EventBasedGateway = 31,
  ExclusiveGateway = 32,
  InclusiveGateway = 33,
  ParallelGateway = 34,
}

export enum FlowNodeStatus {
  UnknownStatus = 0,
  Ready = 1,
  Running = 2,
  Completed = 3,
  Failed = 4,
  Cancelled = 5,
}

export enum CallTaskStage {
  Echo = 0,
  Commit = 1,
  Rollback = 2,
  Destroy = 3,
}

export interface DeployDefinitionsRequest {
  metadata?: Record<string, string>
  content: string
  description?: string
}

export interface DeployDefinitionsResponse {
  definitions: Definitions
}

export interface ListDefinitionsRequest {
  page: number
  size: number
}

export interface ListDefinitionsResponse {
  definitions_list: Definitions[]
  total: number
}

export interface GetDefinitionsRequest {
  uid: string
  version?: number
}

export interface GetDefinitionsResponse {
  definitions: Definitions
}

export interface RemoveDefinitionsRequest {
  uid: string
  version?: number
}

export interface RemoveDefinitionsResponse {
  definitions: Definitions
}

export interface ExecuteProcessRequest {
  name?: string
  definitions_uid: string
  definitions_version?: number
  priority?: number
  mode?: TransitionMode
  headers?: Record<string, string>
  properties?: Record<string, Value>
  dataObjects?: Record<string, Value>
}

export interface ExecuteProcessResponse {
  process: Process
}

export interface ListProcessRequest {
  page?: number
  size: number
  definitions_uid?: string
  definitions_version?: number
  process_status?: ProcessStatus
  process_stage?: ProcessStage
}

export interface ListProcessResponse {
  processes: Process[]
  total: number
}

export interface GetProcessRequest {
  id: number
}

export interface GetProcessResponse {
  process: Process
  activities: FlowNode[]
}
