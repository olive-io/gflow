export interface User {
  id: number
  uid: string
  username: string
  email: string
  description: string
  metadata: Record<string, string>
  roleId: number
  createAt: number
  updateAt: number
}

export interface Token {
  id: number
  text: string
  expireAt: number
  enable: number
  userId: number
  roleId: number
  createAt: number
}

export interface Role {
  id: number
  type: RoleType
  name: string
  displayName: string
  description: string
  metadata: Record<string, string>
  createAt: number
  updateAt: number
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
  createAt: number
  updateAt: number
}

export interface BpmnArgs {
  headers: Record<string, string>
  properties: Record<string, Value>
  dataObjects: Record<string, Value>
}

export interface Value {
  type: ValueType
  kind: string
  value: string
  default: string
}

export enum ValueType {
  String = 0,
  Integer = 1,
  Float = 2,
  Boolean = 3,
  Array = 4,
  Object = 5,
}

export interface ProcessContext {
  variables: Record<string, Value>
  dataObjects: Record<string, Value>
}

export interface Process {
  id: number
  name: string
  uid: string
  metadata: Record<string, string>
  priority: number
  mode: string
  args: BpmnArgs
  definitionsUid: string
  definitionsVersion: number
  definitionsProcess: string
  context: ProcessContext
  attempts: number
  startAt: number
  endAt: number
  stage: string
  status: string
  errMsg: string
  activities: Activity[]
  createAt: number
  updateAt: number
}

export interface Activity {
  id: number
  name: string
  flowId: string
  flowType: string
  type: string
  target: string
  mode: string
  headers: Record<string, string>
  properties: Record<string, Value>
  dataObjects: Record<string, Value>
  results: Record<string, Value>
  retries: number
  startTime: number
  endTime: number
  processId: number
  processUid: string
  traceId: string
  stage: string
  status: string
  errMsg: string
  createAt: number
}

export enum TransitionMode {
  UnknownMode = 0,
  Simple = 1,
  Transition = 2,
}

export enum ProcessStatus {
  UnknownStatus = 0,
  Waiting = 1,
  Running = 2,
  Success = 3,
  Warn = 4,
  Failed = 5,
}

export enum ProcessStage {
  UnknownStage = 0,
  Prepare = 1,
  Ready = 2,
  Commit = 3,
  Rollback = 4,
  Destroy = 5,
  Finish = 6,
}

export interface Step {
  id: number
  name: string
  flowId: string
  flowType: FlowNodeType
  type: string
  target: string
  mode: TransitionMode
  headers: Record<string, string>
  properties: Record<string, Value>
  dataObjects: Record<string, Value>
  results: Record<string, Value>
  retries: number
  startTime: number
  endTime: number
  processId: number
  processUid: string
  traceId: string
  stage: FlowNodeStage
  status: FlowNodeStatus
  errMsg: string
  createAt: number
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

export enum FlowNodeStage {
  Prepare = 0,
  Ready = 1,
  Commit = 2,
  Destroy = 3,
  Rollback = 4,
  Finish = 5,
}

export enum FlowNodeStatus {
  Waiting = 0,
  Running = 1,
  Success = 2,
  Warn = 3,
  Failed = 4,
}

export interface Runner {
  id: string
  uid: string
  listenUrl: string
  version: string
  heartbeatMs: number
  hostname: string
  metadata: Record<string, string>
  features: Record<string, string>
  transport: string
  cpu: number
  memory: number
  onlineTimestamp: number
  offlineTimestamp: number
  online: number
  state: string
  message: string
  createAt: number
}

export interface EndpointProperty {
  type: string
  kind: string
  value: string
  default: string
}

export interface Endpoint {
  id: number
  taskType: number | string
  type: string
  name: string
  description: string
  mode: number | string
  httpUrl: string
  metadata: Record<string, string>
  targets: string[]
  headers: Record<string, string>
  properties: Record<string, EndpointProperty>
  dataObjects: Record<string, unknown>
  results: Record<string, EndpointProperty>
  createAt: number
  updateAt: number
}

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  size: number
}

export interface ExecuteProcessRequest {
  name: string
  definitionsUid: string
  definitionsVersion?: number
  priority?: number
  mode?: number
  headers?: Record<string, string>
  properties?: Record<string, unknown>
  dataObjects?: Record<string, unknown>
}
