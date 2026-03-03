<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ArrowLeft, Loader2, Clock, FileText, AlertCircle, CheckCircle, XCircle, Timer, ChevronRight, Play } from 'lucide-vue-next'
import BpmnModeler from 'bpmn-js/lib/Modeler'
import 'bpmn-js/dist/assets/diagram-js.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-codes.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-embedded.css'
import { http } from '@/lib/http'
import type { Process, Definitions, Activity } from '@/types/api'

const route = useRoute()
const router = useRouter()

const process = ref<Process | null>(null)
const definition = ref<Definitions | null>(null)
const loading = ref(true)
const viewerReady = ref(false)
const canvasRef = ref<HTMLElement | null>(null)
const selectedActivity = ref<Activity | null>(null)
let bpmnModeler: InstanceType<typeof BpmnModeler> | null = null

const statusConfig = computed(() => {
  const status = process.value?.status
  const configs: Record<string, { variant: 'success' | 'warning' | 'error' | 'info' | 'secondary', label: string, icon: typeof CheckCircle }> = {
    'Waiting': { variant: 'warning', label: '等待中', icon: Clock },
    'Running': { variant: 'info', label: '运行中', icon: Loader2 },
    'Success': { variant: 'success', label: '成功', icon: CheckCircle },
    'Warn': { variant: 'warning', label: '警告', icon: AlertCircle },
    'Failed': { variant: 'error', label: '失败', icon: XCircle },
  }
  return configs[status || ''] || { variant: 'secondary', label: status || '未知', icon: AlertCircle }
})

const stageConfig = computed(() => {
  const stage = process.value?.stage
  const configs: Record<string, { variant: 'success' | 'warning' | 'error' | 'info' | 'secondary', label: string }> = {
    'Prepare': { variant: 'warning', label: '准备中' },
    'Ready': { variant: 'info', label: '就绪' },
    'Commit': { variant: 'success', label: '提交' },
    'Rollback': { variant: 'error', label: '回滚' },
    'Destroy': { variant: 'secondary', label: '销毁' },
    'Finish': { variant: 'success', label: '完成' },
  }
  return configs[stage || ''] || { variant: 'secondary', label: stage || '未知' }
})

const activityStatusConfig = (status: string | number) => {
  const statusStr = typeof status === 'number' ? String(status) : status
  const configs: Record<string, { variant: 'success' | 'warning' | 'error' | 'info' | 'secondary', label: string, color: string }> = {
    '0': { variant: 'warning', label: '等待', color: 'bg-yellow-500' },
    '1': { variant: 'info', label: '运行中', color: 'bg-blue-500' },
    '2': { variant: 'success', label: '成功', color: 'bg-green-500' },
    '3': { variant: 'warning', label: '警告', color: 'bg-yellow-500' },
    '4': { variant: 'error', label: '失败', color: 'bg-red-500' },
    'Waiting': { variant: 'warning', label: '等待', color: 'bg-yellow-500' },
    'Running': { variant: 'info', label: '运行中', color: 'bg-blue-500' },
    'Success': { variant: 'success', label: '成功', color: 'bg-green-500' },
    'Warn': { variant: 'warning', label: '警告', color: 'bg-yellow-500' },
    'Failed': { variant: 'error', label: '失败', color: 'bg-red-500' },
  }
  return configs[statusStr] || { variant: 'secondary', label: '未知', color: 'bg-gray-500' }
}

const activityStageConfig = (stage: string | number) => {
  const stageStr = typeof stage === 'number' ? String(stage) : stage
  const configs: Record<string, { variant: 'success' | 'warning' | 'error' | 'info' | 'secondary', label: string }> = {
    '0': { variant: 'warning', label: '准备' },
    '1': { variant: 'info', label: '就绪' },
    '2': { variant: 'success', label: '提交' },
    '3': { variant: 'secondary', label: '销毁' },
    '4': { variant: 'error', label: '回滚' },
    '5': { variant: 'success', label: '完成' },
    'Prepare': { variant: 'warning', label: '准备' },
    'Ready': { variant: 'info', label: '就绪' },
    'Commit': { variant: 'success', label: '提交' },
    'Destroy': { variant: 'secondary', label: '销毁' },
    'Rollback': { variant: 'error', label: '回滚' },
    'Finish': { variant: 'success', label: '完成' },
  }
  return configs[stageStr] || { variant: 'secondary', label: '未知' }
}

const flowNodeTypeLabel = (flowType: string | number) => {
  const flowTypeStr = typeof flowType === 'number' ? String(flowType) : flowType
  const labels: Record<string, string> = {
    '1': '开始事件',
    '2': '结束事件',
    '3': '边界事件',
    '4': '中间捕获事件',
    '11': '任务',
    '12': '发送任务',
    '13': '接收任务',
    '14': '服务任务',
    '15': '用户任务',
    '16': '脚本任务',
    '17': '手工任务',
    '18': '调用活动',
    '19': '业务规则任务',
    '20': '子流程',
    '31': '事件网关',
    '32': '排他网关',
    '33': '包含网关',
    '34': '并行网关',
    'StartEvent': '开始事件',
    'EndEvent': '结束事件',
    'BoundaryEvent': '边界事件',
    'IntermediateCatchEvent': '中间捕获事件',
    'Task': '任务',
    'SendTask': '发送任务',
    'ReceiveTask': '接收任务',
    'ServiceTask': '服务任务',
    'UserTask': '用户任务',
    'ScriptTask': '脚本任务',
    'ManualTask': '手工任务',
    'CallActivity': '调用活动',
    'BusinessRuleTask': '业务规则任务',
    'SubProcess': '子流程',
    'EventBasedGateway': '事件网关',
    'ExclusiveGateway': '排他网关',
    'InclusiveGateway': '包含网关',
    'ParallelGateway': '并行网关',
  }
  return labels[flowTypeStr] || '未知类型'
}

async function fetchProcess() {
  const id = route.params.id as string
  if (!id) return
  
  loading.value = true
  try {
    const response = await http.get<{ process: Process; activities?: Activity[] }>(`/v1/processes/${id}`)
    process.value = response.process
    
    // activities 在响应的顶层，不在 process 对象内
    if (response.activities && response.activities.length > 0) {
      process.value.activities = response.activities
    }
    
    if (response.process.definitionsUid) {
      await fetchDefinition(response.process.definitionsUid)
    }
  } catch (error) {
    console.error('Failed to fetch process:', error)
  } finally {
    loading.value = false
  }
}

async function fetchDefinition(uid: string) {
  try {
    const response = await http.get<{ definitions: Definitions }>(`/v1/definitions/${uid}`)
    definition.value = response.definitions
  } catch (error) {
    console.error('Failed to fetch definition:', error)
  }
}

async function initModeler(xml: string) {
  if (!canvasRef.value) {
    console.warn('Canvas ref not ready')
    return
  }
  
  destroyModeler()
  
  try {
    bpmnModeler = new BpmnModeler({
      container: canvasRef.value,
    })
    
    await bpmnModeler.importXML(xml)
    const canvas = bpmnModeler.get('canvas') as { zoom: (level: string) => void }
    canvas.zoom('fit-viewport')
    viewerReady.value = true
    
    highlightActivities()
    setupClickHandler()
  } catch (error) {
    console.error('Failed to import XML:', error)
  }
}

function highlightActivities() {
  if (!bpmnModeler || !process.value?.activities) return
  
  const canvas = bpmnModeler.get('canvas') as { 
    addMarker: (elementId: string, marker: string) => void 
  }
  
  process.value.activities.forEach(activity => {
    const statusStr = typeof activity.status === 'number' ? String(activity.status) : activity.status
    let marker = ''
    
    if (statusStr === '1' || statusStr === 'Running') {
      marker = 'highlight-running'
    } else if (statusStr === '2' || statusStr === 'Success') {
      marker = 'highlight-success'
    } else if (statusStr === '3' || statusStr === 'Warn') {
      marker = 'highlight-warning'
    } else if (statusStr === '4' || statusStr === 'Failed') {
      marker = 'highlight-error'
    }
    
    if (marker) {
      canvas.addMarker(activity.flowId, marker)
    }
  })
}

function setupClickHandler() {
  if (!bpmnModeler) return
  
  const eventBus = bpmnModeler.get('eventBus') as { 
    on: (event: string, callback: (e: { element: { id: string } }) => void) => void 
  }
  
  eventBus.on('element.click', (e: { element: { id: string } }) => {
    const flowId = e.element.id
    const activity = process.value?.activities?.find(a => a.flowId === flowId)
    if (activity) {
      selectActivity(activity)
    }
  })
}

function selectActivity(activity: Activity) {
  selectedActivity.value = activity
  
  if (bpmnModeler) {
    const canvas = bpmnModeler.get('canvas') as { 
      removeMarker: (elementId: string, marker: string) => void,
      addMarker: (elementId: string, marker: string) => void 
    }
    
    if (process.value?.activities) {
      process.value.activities.forEach(a => {
        canvas.removeMarker(a.flowId, 'selected')
      })
    }
    
    canvas.addMarker(activity.flowId, 'selected')
  }
  
  const element = document.getElementById(`activity-${activity.id}`)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }
}

function destroyModeler() {
  if (bpmnModeler) {
    bpmnModeler.destroy()
    bpmnModeler = null
    viewerReady.value = false
  }
}

function goBack() {
  router.push('/processes')
}

function formatDateTime(timestamp: number | string): string {
  if (!timestamp) return '-'
  const ts = typeof timestamp === 'string' ? Number(timestamp) : timestamp
  if (ts === 0) return '-'
  const ms = ts > 10000000000 ? ts : ts * 1000
  return new Date(ms).toLocaleString('zh-CN')
}

function formatDuration(start: number | string, end: number | string): string {
  const startTs = typeof start === 'string' ? Number(start) : start
  if (!startTs || startTs === 0) return '-'
  
  const endTs = end ? (typeof end === 'string' ? Number(end) : end) : Date.now()
  const startMs = startTs > 10000000000 ? startTs : startTs * 1000
  const endMs = endTs > 10000000000 ? endTs : endTs * 1000
  
  const duration = (endMs - startMs) / 1000
  if (duration < 0) return '-'
  if (duration < 60) return `${Math.floor(duration)}秒`
  if (duration < 3600) return `${Math.floor(duration / 60)}分 ${Math.floor(duration % 60)}秒`
  return `${Math.floor(duration / 3600)}小时 ${Math.floor((duration % 3600) / 60)}分`
}

function valueToJson(value: unknown): unknown {
  if (!value) return value
  if (typeof value !== 'object') return value
  
  const v = value as { type?: string; value?: string; [key: string]: unknown }
  if (v.type && v.value !== undefined) {
    // Value 结构体：提取 value 字段并尝试解析
    try {
      if (v.type === 'Object' || v.type === 'Array') {
        return JSON.parse(v.value)
      }
      return v.value
    } catch {
      return v.value
    }
  }
  
  // 递归处理对象
  const result: Record<string, unknown> = {}
  for (const [key, val] of Object.entries(value as Record<string, unknown>)) {
    result[key] = valueToJson(val)
  }
  return result
}

function formatValueObject(obj: Record<string, unknown>): string {
  const formatted = valueToJson(obj) as Record<string, unknown>
  return JSON.stringify(formatted, null, 2)
}

onMounted(async () => {
  await fetchProcess()
})

watch([canvasRef, definition], async ([newCanvasRef, newDefinition], [oldCanvasRef]) => {
  if (newCanvasRef && newDefinition?.content) {
    if (oldCanvasRef !== newCanvasRef) {
      viewerReady.value = false
    }
    if (!viewerReady.value) {
      await nextTick()
      setTimeout(() => {
        initModeler(newDefinition.content!)
      }, 100)
    }
  }
}, { immediate: true })

onUnmounted(() => {
  destroyModeler()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <Button variant="ghost" size="sm" @click="goBack">
          <ArrowLeft class="w-4 h-4 mr-1" />
          返回
        </Button>
        <div>
          <h1 class="text-2xl font-bold">{{ process?.name || `流程实例 #${route.params.id}` }}</h1>
          <p class="text-muted-foreground mt-1">{{ process?.uid }}</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Badge :variant="statusConfig.variant" class="text-sm">
          <component :is="statusConfig.icon" class="w-4 h-4 mr-1" :class="{ 'animate-spin': process?.status === 'Running' }" />
          {{ statusConfig.label }}
        </Badge>
        <Badge :variant="stageConfig.variant" class="text-sm">
          {{ stageConfig.label }}
        </Badge>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <template v-else-if="process">
      <!-- Overview Cards -->
      <div class="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>流程定义</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center gap-2">
              <FileText class="w-4 h-4 text-muted-foreground" />
              <span class="font-medium">{{ process.definitionsUid || '-' }}</span>
            </div>
            <p class="text-xs text-muted-foreground mt-1">版本: v{{ process.definitionsVersion }}</p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>开始时间</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center gap-2">
              <Clock class="w-4 h-4 text-muted-foreground" />
              <span class="font-medium">{{ formatDateTime(process.startAt) }}</span>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>结束时间</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center gap-2">
              <Clock class="w-4 h-4 text-muted-foreground" />
              <span class="font-medium">{{ formatDateTime(process.endAt) }}</span>
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader class="pb-2">
            <CardDescription>执行时长</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center gap-2">
              <Timer class="w-4 h-4 text-muted-foreground" />
              <span class="font-medium">{{ formatDuration(process.startAt, process.endAt) }}</span>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Error Message -->
      <Card v-if="process.errMsg" class="border-destructive">
        <CardHeader class="pb-2">
          <CardTitle class="text-destructive flex items-center gap-2">
            <AlertCircle class="w-5 h-5" />
            错误信息
          </CardTitle>
        </CardHeader>
        <CardContent>
          <pre class="text-sm text-destructive whitespace-pre-wrap bg-destructive/10 p-3 rounded-md overflow-auto">{{ process.errMsg }}</pre>
        </CardContent>
      </Card>

      <!-- Main Content: Diagram + Activities Side by Side -->
      <div class="grid gap-6 lg:grid-cols-3">
        <!-- Diagram (2/3 width) -->
        <Card class="lg:col-span-2">
          <CardHeader>
            <div class="flex items-center justify-between">
              <div>
                <CardTitle>流程图</CardTitle>
                <CardDescription>流程执行的可视化展示，点击节点查看详情</CardDescription>
              </div>
              <!-- Legend -->
              <div class="flex items-center gap-3 text-xs">
                <div class="flex items-center gap-1">
                  <div class="w-3 h-3 rounded-full bg-blue-500"></div>
                  <span>运行中</span>
                </div>
                <div class="flex items-center gap-1">
                  <div class="w-3 h-3 rounded-full bg-green-500"></div>
                  <span>成功</span>
                </div>
                <div class="flex items-center gap-1">
                  <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
                  <span>警告</span>
                </div>
                <div class="flex items-center gap-1">
                  <div class="w-3 h-3 rounded-full bg-red-500"></div>
                  <span>失败</span>
                </div>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <div ref="canvasRef" class="h-[600px] bg-muted/30 rounded-lg border relative">
              <div v-if="!definition" class="h-full flex items-center justify-center">
                <p class="text-muted-foreground">无法加载流程定义</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Activities Sidebar (1/3 width) -->
        <Card class="lg:col-span-1">
          <CardHeader>
            <CardTitle class="flex items-center gap-2">
              <Play class="w-4 h-4" />
              执行步骤
            </CardTitle>
            <CardDescription>
              {{ process.activities?.length || 0 }} 个步骤
            </CardDescription>
          </CardHeader>
          <CardContent class="p-0">
            <div v-if="!process.activities || process.activities.length === 0" class="text-center py-8">
              <p class="text-muted-foreground text-sm">暂无执行步骤</p>
            </div>
            <div v-else class="max-h-[600px] overflow-auto">
              <div class="divide-y">
                <div 
                  v-for="(activity, index) in process.activities" 
                  :key="activity.id"
                  :id="`activity-${activity.id}`"
                  :class="[
                    'p-4 cursor-pointer transition-all hover:bg-muted/50',
                    selectedActivity?.id === activity.id ? 'bg-primary/10 border-l-2 border-l-primary' : 'border-l-2 border-l-transparent'
                  ]"
                  @click="selectActivity(activity)"
                >
                  <div class="flex items-start gap-3">
                    <!-- Status Indicator -->
                    <div class="flex flex-col items-center">
                      <div :class="['w-3 h-3 rounded-full', activityStatusConfig(activity.status).color]"></div>
                      <div v-if="index < (process.activities?.length || 0) - 1" class="w-0.5 h-full bg-border mt-1"></div>
                    </div>
                    
                    <!-- Content -->
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2">
                        <span class="font-medium text-sm truncate">{{ activity.name || activity.flowId }}</span>
                      </div>
                      <div class="flex items-center gap-2 mt-1">
                        <Badge :variant="activityStatusConfig(activity.status).variant" class="text-xs py-0">
                          {{ activityStatusConfig(activity.status).label }}
                        </Badge>
                        <span class="text-xs text-muted-foreground">{{ flowNodeTypeLabel(activity.flowType) }}</span>
                      </div>
                      <div class="flex items-center gap-2 mt-2 text-xs text-muted-foreground">
                        <Clock class="w-3 h-3" />
                        <span>{{ formatDuration(activity.startTime, activity.endTime) }}</span>
                      </div>
                      
                      <!-- Error Message -->
                      <div v-if="activity.errMsg" class="mt-2 p-2 bg-destructive/10 rounded text-xs text-destructive line-clamp-2">
                        {{ activity.errMsg }}
                      </div>
                      
                      <!-- Expanded Details -->
                      <div v-if="selectedActivity?.id === activity.id" class="mt-3 pt-3 border-t space-y-2">
                        <div class="text-xs">
                          <span class="text-muted-foreground">开始:</span>
                          <span class="ml-1">{{ formatDateTime(activity.startTime) }}</span>
                        </div>
                        <div class="text-xs">
                          <span class="text-muted-foreground">结束:</span>
                          <span class="ml-1">{{ formatDateTime(activity.endTime) }}</span>
                        </div>
                        <div v-if="activity.retries > 0" class="text-xs">
                          <span class="text-muted-foreground">重试:</span>
                          <span class="ml-1">{{ activity.retries }}次</span>
                        </div>
                        <div v-if="Object.keys(activity.results || {}).length > 0">
                          <div class="text-xs font-medium mb-1">执行结果:</div>
                          <pre class="text-xs bg-muted/50 p-2 rounded overflow-auto max-h-32">{{ formatValueObject(activity.results) }}</pre>
                        </div>
                        <div v-if="Object.keys(activity.properties || {}).length > 0">
                          <div class="text-xs font-medium mb-1">属性参数:</div>
                          <pre class="text-xs bg-muted/50 p-2 rounded overflow-auto max-h-32">{{ formatValueObject(activity.properties) }}</pre>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Other Tabs -->
      <Tabs default-value="input" class="space-y-4">
        <TabsList>
          <TabsTrigger value="input">输入参数</TabsTrigger>
          <TabsTrigger value="output">输出数据</TabsTrigger>
          <TabsTrigger value="context">上下文</TabsTrigger>
          <TabsTrigger value="metadata">元数据</TabsTrigger>
        </TabsList>

        <!-- Input Tab -->
        <TabsContent value="input">
          <Card>
            <CardHeader>
              <CardTitle>输入参数</CardTitle>
              <CardDescription>流程启动时传入的参数</CardDescription>
            </CardHeader>
            <CardContent class="space-y-6">
              <div>
                <h4 class="font-medium mb-3">请求头</h4>
                <div v-if="process.args?.headers && Object.keys(process.args.headers).length > 0" class="bg-muted/50 rounded-lg p-4">
                  <pre class="text-sm">{{ JSON.stringify(process.args.headers, null, 2) }}</pre>
                </div>
                <p v-else class="text-muted-foreground text-sm">无请求头</p>
              </div>

              <div>
                <h4 class="font-medium mb-3">属性参数</h4>
                <div v-if="process.args?.properties && Object.keys(process.args.properties).length > 0" class="bg-muted/50 rounded-lg p-4">
                  <pre class="text-sm">{{ formatValueObject(process.args.properties) }}</pre>
                </div>
                <p v-else class="text-muted-foreground text-sm">无属性参数</p>
              </div>

              <div>
                <h4 class="font-medium mb-3">数据对象</h4>
                <div v-if="process.args?.dataObjects && Object.keys(process.args.dataObjects).length > 0" class="bg-muted/50 rounded-lg p-4">
                  <pre class="text-sm">{{ formatValueObject(process.args.dataObjects) }}</pre>
                </div>
                <p v-else class="text-muted-foreground text-sm">无数据对象</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <!-- Output Tab -->
        <TabsContent value="output">
          <Card>
            <CardHeader>
              <CardTitle>输出数据</CardTitle>
              <CardDescription>流程执行产生的输出</CardDescription>
            </CardHeader>
            <CardContent>
              <div v-if="process.context?.variables && Object.keys(process.context.variables).length > 0" class="bg-muted/50 rounded-lg p-4">
                <pre class="text-sm">{{ formatValueObject(process.context.variables) }}</pre>
              </div>
              <p v-else class="text-muted-foreground text-sm">无输出数据</p>
            </CardContent>
          </Card>
        </TabsContent>

        <!-- Context Tab -->
        <TabsContent value="context">
          <Card>
            <CardHeader>
              <CardTitle>执行上下文</CardTitle>
              <CardDescription>流程运行时的上下文数据</CardDescription>
            </CardHeader>
            <CardContent class="space-y-6">
              <div>
                <h4 class="font-medium mb-3">变量</h4>
                <div v-if="process.context?.variables && Object.keys(process.context.variables).length > 0" class="bg-muted/50 rounded-lg p-4">
                  <pre class="text-sm">{{ formatValueObject(process.context.variables) }}</pre>
                </div>
                <p v-else class="text-muted-foreground text-sm">无变量</p>
              </div>

              <div>
                <h4 class="font-medium mb-3">数据对象</h4>
                <div v-if="process.context?.dataObjects && Object.keys(process.context.dataObjects).length > 0" class="bg-muted/50 rounded-lg p-4">
                  <pre class="text-sm">{{ formatValueObject(process.context.dataObjects) }}</pre>
                </div>
                <p v-else class="text-muted-foreground text-sm">无数据对象</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <!-- Metadata Tab -->
        <TabsContent value="metadata">
          <Card>
            <CardHeader>
              <CardTitle>元数据</CardTitle>
            </CardHeader>
            <CardContent>
              <div class="grid gap-4 md:grid-cols-3 text-sm">
                <div>
                  <span class="text-muted-foreground">创建时间:</span>
                  <span class="ml-2">{{ formatDateTime(process.createAt) }}</span>
                </div>
                <div>
                  <span class="text-muted-foreground">更新时间:</span>
                  <span class="ml-2">{{ formatDateTime(process.updateAt) }}</span>
                </div>
                <div>
                  <span class="text-muted-foreground">优先级:</span>
                  <span class="ml-2">{{ process.priority }}</span>
                </div>
                <div>
                  <span class="text-muted-foreground">重试次数:</span>
                  <span class="ml-2">{{ process.attempts }}</span>
                </div>
                <div>
                  <span class="text-muted-foreground">执行模式:</span>
                  <span class="ml-2">{{ process.mode }}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </template>
  </div>
</template>

<style>
.djs-container {
  background: transparent !important;
}

.djs-palette,
.djs-context-pad {
  display: none !important;
}

.highlight-running .djs-visual > :not(.djs-label) {
  stroke: #3b82f6 !important;
  stroke-width: 2px !important;
}

.highlight-success .djs-visual > :not(.djs-label) {
  stroke: #22c55e !important;
  stroke-width: 2px !important;
}

.highlight-warning .djs-visual > :not(.djs-label) {
  stroke: #eab308 !important;
  stroke-width: 2px !important;
}

.highlight-error .djs-visual > :not(.djs-label) {
  stroke: #ef4444 !important;
  stroke-width: 2px !important;
}

.selected .djs-outline {
  stroke: #6366f1 !important;
  stroke-width: 3px !important;
  visibility: visible !important;
}

.selected .djs-visual > :not(.djs-label) {
  stroke: #6366f1 !important;
  stroke-width: 3px !important;
}
</style>
