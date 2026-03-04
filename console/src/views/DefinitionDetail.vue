<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ArrowLeft, FileText, Pencil, Loader2, GitBranch, Info, Code } from 'lucide-vue-next'
import { definitionsApi, processesApi } from '@/api'
import type { Definitions, Process } from '@/types/api'
import Pagination from '@/components/Pagination.vue'
import BpmnViewer from 'bpmn-js/lib/NavigatedViewer'
import 'bpmn-js/dist/assets/diagram-js.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-codes.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-embedded.css'

const route = useRoute()
const router = useRouter()

const definition = ref<Definitions | null>(null)
const loading = ref(true)

const processes = ref<Process[]>([])
const processesLoading = ref(false)
const processesTotal = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)

const activeTab = ref('info')

const bpmnViewMode = ref<'diagram' | 'xml'>('diagram')
const bpmnContainer = ref<HTMLElement | null>(null)
const bpmnViewerLoading = ref(false)
let bpmnViewer: InstanceType<typeof BpmnViewer> | null = null

const statusLabel = computed(() => {
  if (!definition.value) return ''
  return definition.value.isExecute ? '已部署' : '未部署'
})

async function fetchDefinition() {
  loading.value = true
  try {
    const uid = route.params.uid as string
    const response = await definitionsApi.get(uid)
    definition.value = response.definitions || null
  } catch (error) {
    console.error('Failed to fetch definition:', error)
  } finally {
    loading.value = false
  }
}

async function fetchProcesses() {
  if (!definition.value) return
  processesLoading.value = true
  try {
    const response = await processesApi.list({
      page: currentPage.value,
      size: pageSize.value,
      definitions_uid: definition.value.uid,
    })
    processes.value = response.processes || []
    processesTotal.value = Number(response.total) || 0
  } catch (error) {
    console.error('Failed to fetch processes:', error)
  } finally {
    processesLoading.value = false
  }
}

async function initBpmnViewer() {
  if (!bpmnContainer.value || !definition.value?.content) return

  bpmnViewerLoading.value = true
  try {
    if (bpmnViewer) {
      bpmnViewer.destroy()
    }

    bpmnViewer = new BpmnViewer({
      container: bpmnContainer.value,
    })

    await bpmnViewer.importXML(definition.value.content)
    
    const canvas = bpmnViewer.get('canvas')
    canvas.zoom('fit-viewport')
  } catch (error) {
    console.error('Failed to initialize BPMN viewer:', error)
  } finally {
    bpmnViewerLoading.value = false
  }
}

function destroyBpmnViewer() {
  if (bpmnViewer) {
    bpmnViewer.destroy()
    bpmnViewer = null
  }
}

watch([activeTab, bpmnViewMode], async ([newTab, newMode]) => {
  if (newTab === 'bpmn' && newMode === 'diagram' && definition.value?.content) {
    await nextTick()
    initBpmnViewer()
  }
})

function handlePageChange(page: number) {
  currentPage.value = page
  fetchProcesses()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  fetchProcesses()
}

function goBack() {
  router.push('/definitions')
}

function editDefinition() {
  if (definition.value) {
    router.push(`/designer/${definition.value.uid}`)
  }
}

function viewProcess(id: number) {
  router.push(`/processes/${id}`)
}

function formatDate(timestamp: number | string): string {
  if (!timestamp) return '-'
  const ts = typeof timestamp === 'string' ? Number(timestamp) : timestamp
  if (ts === 0) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

function getStatusVariant(status: string) {
  const variants: Record<string, 'success' | 'destructive' | 'warning' | 'secondary' | 'outline'> = {
    'Success': 'success',
    '成功': 'success',
    '完成': 'success',
    'Failed': 'destructive',
    '失败': 'destructive',
    'Running': 'warning',
    '运行中': 'warning',
    'Waiting': 'secondary',
    '等待': 'secondary',
  }
  return variants[status] || 'secondary'
}

function getStatusLabel(status: string) {
  const labels: Record<string, string> = {
    'Success': '成功',
    '成功': '成功',
    '完成': '完成',
    'Failed': '失败',
    '失败': '失败',
    'Running': '运行中',
    '运行中': '运行中',
    'Waiting': '等待',
    '等待': '等待',
  }
  return labels[status] || status
}

function getStageLabel(stage: string) {
  const labels: Record<string, string> = {
    'Prepare': '准备中',
    'Ready': '就绪',
    'Commit': '提交中',
    'Rollback': '回滚中',
    'Destroy': '销毁中',
    'Finish': '已完成',
  }
  return labels[stage] || stage
}

const metadataList = computed(() => {
  if (!definition.value?.metadata) return []
  return Object.entries(definition.value.metadata)
    .filter(([key]) => key)
    .map(([key, value]) => ({ key, value }))
})

onMounted(async () => {
  await fetchDefinition()
  if (definition.value) {
    fetchProcesses()
  }
})

onUnmounted(() => {
  destroyBpmnViewer()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-4">
        <Button variant="ghost" size="sm" @click="goBack">
          <ArrowLeft class="w-4 h-4 mr-1" />
          返回
        </Button>
        <div class="w-px h-6 bg-border"></div>
        <div v-if="definition">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
              <FileText class="w-5 h-5 text-primary" />
            </div>
            <div>
              <h1 class="text-2xl font-bold">{{ definition.name || definition.uid || '未命名流程' }}</h1>
              <p class="text-muted-foreground text-sm">{{ definition.description || '暂无描述' }}</p>
            </div>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Button @click="editDefinition">
          <Pencil class="w-4 h-4 mr-2" />
          编辑流程
        </Button>
      </div>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <template v-else-if="definition">
      <div class="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">状态</p>
                <p class="text-lg font-semibold">{{ statusLabel }}</p>
              </div>
              <div 
                class="w-8 h-8 rounded-lg flex items-center justify-center"
                :class="definition.isExecute ? 'bg-green-100' : 'bg-muted'"
              >
                <div 
                  class="w-3 h-3 rounded-full"
                  :class="definition.isExecute ? 'bg-green-500' : 'bg-muted-foreground'"
                />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">版本</p>
                <p class="text-lg font-semibold">v{{ definition.version }}</p>
              </div>
              <div class="w-8 h-8 rounded-lg bg-blue-100 flex items-center justify-center">
                <span class="text-blue-600 font-bold text-sm">v</span>
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">运行实例</p>
                <p class="text-lg font-semibold">{{ processesTotal }}</p>
              </div>
              <div class="w-8 h-8 rounded-lg bg-orange-100 flex items-center justify-center">
                <GitBranch class="w-4 h-4 text-orange-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">更新时间</p>
                <p class="text-sm font-semibold">{{ formatDate(definition.updateAt) }}</p>
              </div>
              <div class="w-8 h-8 rounded-lg bg-purple-100 flex items-center justify-center">
                <FileText class="w-4 h-4 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs :model-value="activeTab" @update:model-value="activeTab = $event" class="w-full">
        <TabsList class="grid w-full grid-cols-3">
          <TabsTrigger value="info" class="flex items-center gap-2">
            <Info class="w-4 h-4" />
            基本信息
          </TabsTrigger>
          <TabsTrigger value="bpmn" class="flex items-center gap-2">
            <Code class="w-4 h-4" />
            BPMN 内容
          </TabsTrigger>
          <TabsTrigger value="processes" class="flex items-center gap-2">
            <GitBranch class="w-4 h-4" />
            运行实例
            <Badge v-if="processesTotal > 0" variant="secondary" class="ml-1">{{ processesTotal }}</Badge>
          </TabsTrigger>
        </TabsList>

        <TabsContent value="info" class="mt-4">
          <Card>
            <CardContent class="p-6">
              <div class="grid gap-4 md:grid-cols-2">
                <div class="space-y-1">
                  <p class="text-sm text-muted-foreground">流程 UID</p>
                  <p class="font-mono">{{ definition.uid }}</p>
                </div>
                <div class="space-y-1">
                  <p class="text-sm text-muted-foreground">流程名称</p>
                  <p>{{ definition.name || '-' }}</p>
                </div>
                <div class="space-y-1">
                  <p class="text-sm text-muted-foreground">版本号</p>
                  <Badge variant="outline">v{{ definition.version }}</Badge>
                </div>
                <div class="space-y-1">
                  <p class="text-sm text-muted-foreground">部署状态</p>
                  <Badge :variant="definition.isExecute ? 'success' : 'secondary'">
                    {{ statusLabel }}
                  </Badge>
                </div>
                <div class="space-y-1">
                  <p class="text-sm text-muted-foreground">创建时间</p>
                  <p>{{ formatDate(definition.createAt) }}</p>
                </div>
                <div class="space-y-1">
                  <p class="text-sm text-muted-foreground">更新时间</p>
                  <p>{{ formatDate(definition.updateAt) }}</p>
                </div>
                <div class="space-y-1 md:col-span-2">
                  <p class="text-sm text-muted-foreground">描述</p>
                  <p>{{ definition.description || '-' }}</p>
                </div>
              </div>

              <div v-if="metadataList.length > 0" class="mt-6 pt-6 border-t">
                <h4 class="text-sm font-semibold mb-4">元数据</h4>
                <div class="space-y-3">
                  <div 
                    v-for="item in metadataList" 
                    :key="item.key" 
                    class="flex items-center justify-between py-2 border-b last:border-0"
                  >
                    <span class="font-mono text-sm">{{ item.key }}</span>
                    <span class="text-sm text-muted-foreground">{{ item.value }}</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="bpmn" class="mt-4">
          <Card>
            <CardContent class="p-6">
              <div class="flex items-center gap-2 mb-4">
                <Button 
                  :variant="bpmnViewMode === 'diagram' ? 'default' : 'outline'" 
                  size="sm"
                  @click="bpmnViewMode = 'diagram'"
                >
                  图形视图
                </Button>
                <Button 
                  :variant="bpmnViewMode === 'xml' ? 'default' : 'outline'" 
                  size="sm"
                  @click="bpmnViewMode = 'xml'"
                >
                  XML 视图
                </Button>
              </div>
              <div v-if="definition.content" class="bg-muted rounded-lg overflow-hidden border">
                <div v-if="bpmnViewMode === 'diagram'" class="relative h-[500px]">
                  <div v-if="bpmnViewerLoading" class="absolute inset-0 flex items-center justify-center bg-background/80">
                    <Loader2 class="w-6 h-6 animate-spin text-muted-foreground" />
                  </div>
                  <div ref="bpmnContainer" class="w-full h-full bg-white"></div>
                </div>
                <pre v-else class="p-4 overflow-auto max-h-[500px] text-sm font-mono whitespace-pre-wrap break-all">{{ definition.content }}</pre>
              </div>
              <div v-else class="py-8 text-center text-muted-foreground">
                暂无 BPMN 内容
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="processes" class="mt-4">
          <Card>
            <CardContent class="p-6">
              <div v-if="processesLoading" class="flex items-center justify-center py-8">
                <Loader2 class="w-6 h-6 animate-spin text-muted-foreground" />
              </div>
              <Table v-else-if="processes.length > 0">
                <TableHeader>
                  <TableRow>
                    <TableHead>实例 ID</TableHead>
                    <TableHead>名称</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>阶段</TableHead>
                    <TableHead>开始时间</TableHead>
                    <TableHead>结束时间</TableHead>
                    <TableHead class="w-[80px]">操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow 
                    v-for="process in processes" 
                    :key="process.id"
                    class="cursor-pointer hover:bg-muted/50"
                    @click="viewProcess(process.id)"
                  >
                    <TableCell class="font-mono">{{ process.id }}</TableCell>
                    <TableCell>{{ process.name || '-' }}</TableCell>
                    <TableCell>
                      <Badge :variant="getStatusVariant(process.status)">
                        {{ getStatusLabel(process.status) }}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{{ getStageLabel(process.stage) }}</Badge>
                    </TableCell>
                    <TableCell>{{ formatDate(process.startAt) }}</TableCell>
                    <TableCell>{{ formatDate(process.endAt) }}</TableCell>
                    <TableCell @click.stop>
                      <Button variant="ghost" size="sm" @click="viewProcess(process.id)">
                        查看
                      </Button>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
              <div v-else class="py-8 text-center text-muted-foreground">
                暂无运行实例
              </div>
              <Pagination
                v-if="processesTotal > 0"
                :page="currentPage"
                :page-size="pageSize"
                :total="processesTotal"
                @update:page="handlePageChange"
                @update:page-size="handlePageSizeChange"
              />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </template>

    <div v-else class="py-12 text-center">
      <p class="text-muted-foreground">流程定义不存在</p>
      <Button variant="link" @click="goBack">返回列表</Button>
    </div>
  </div>
</template>
