<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Search, MoreVertical, Globe, FileCode, Server, Plus, Loader2, ArrowRight, Eye } from 'lucide-vue-next'
import { endpointsApi } from '@/api'
import type { Endpoint } from '@/types/api'
import Pagination from '@/components/Pagination.vue'

const router = useRouter()
const endpoints = ref<Endpoint[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')

const showAddDialog = ref(false)
const endpointType = ref<'openapi' | 'grpc'>('openapi')
const openapiDoc = ref('')
const openapiContentType = ref('json')
const grpcTarget = ref('')
const adding = ref(false)

const showBpmnDialog = ref(false)
const bpmnContent = ref('')
const convertingId = ref<number | null>(null)

const filteredEndpoints = computed(() => {
  if (!searchQuery.value) return endpoints.value
  const query = searchQuery.value.toLowerCase()
  return endpoints.value.filter(e => 
    e.name?.toLowerCase().includes(query) || 
    e.type?.toLowerCase().includes(query) ||
    e.httpUrl?.toLowerCase().includes(query)
  )
})

const getTaskTypeLabel = (taskType: number | string) => {
  if (typeof taskType === 'string') {
    return taskType
  }
  const types: Record<number, string> = {
    11: 'Task',
    12: 'SendTask',
    13: 'ReceiveTask',
    14: 'ServiceTask',
    15: 'UserTask',
    16: 'ScriptTask',
    17: 'ManualTask',
    18: 'CallActivity',
    19: 'BusinessRuleTask',
  }
  return types[taskType] || 'Unknown'
}

const getModeLabel = (mode: number | string) => {
  if (typeof mode === 'string') {
    if (!mode || mode === 'UnknownMode') {
      return 'Simple'
    }
    return mode
  }
  return 'Simple'
}

async function fetchEndpoints() {
  loading.value = true
  try {
    const response = await endpointsApi.list({
      page: currentPage.value,
      size: pageSize.value,
    })
    endpoints.value = response.endpoints || []
    total.value = Number(response.total) || endpoints.value.length
  } catch (error) {
    console.error('Failed to fetch endpoints:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchEndpoints()
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchEndpoints()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  fetchEndpoints()
}

async function addEndpoint() {
  if (endpointType.value === 'openapi' && !openapiDoc.value.trim()) {
    return
  }
  if (endpointType.value === 'grpc' && !grpcTarget.value.trim()) {
    return
  }

  adding.value = true
  try {
    if (endpointType.value === 'openapi') {
      await endpointsApi.addOpenAPI({
        doc: openapiDoc.value,
        contentType: openapiContentType.value,
      })
    } else {
      await endpointsApi.addGRPC({
        target: grpcTarget.value,
      })
    }
    showAddDialog.value = false
    openapiDoc.value = ''
    grpcTarget.value = ''
    fetchEndpoints()
  } catch (error) {
    console.error('Failed to add endpoint:', error)
  } finally {
    adding.value = false
  }
}

function viewDetail(endpoint: Endpoint) {
  router.push(`/endpoints/${endpoint.id}`)
}

function formatXml(xml: string): string {
  try {
    let formatted = ''
    let indent = ''
    const tab = '  '
    xml.split(/>\s*</).forEach(node => {
      if (node.match(/^\/\w/)) {
        indent = indent.substring(tab.length)
      }
      formatted += indent + '<' + node + '>\n'
      if (node.match(/^<?\w[^>]*[^\/]$/) && !node.startsWith('?')) {
        indent += tab
      }
    })
    return formatted.substring(1, formatted.length - 2)
  } catch {
    return xml
  }
}

async function convertToBpmn(endpoint: Endpoint) {
  convertingId.value = endpoint.id
  try {
    const response = await endpointsApi.convert({ id: endpoint.id, format: 'BPMN' })
    bpmnContent.value = formatXml(response.content || '')
    showBpmnDialog.value = true
  } catch (error) {
    console.error('Failed to convert endpoint:', error)
    alert('转换失败')
  } finally {
    convertingId.value = null
  }
}

function copyBpmn() {
  navigator.clipboard.writeText(bpmnContent.value)
  alert('已复制到剪贴板')
}

function formatDateTime(timestamp: number | string): string {
  if (!timestamp) return '-'
  const ts = typeof timestamp === 'string' ? Number(timestamp) : timestamp
  if (ts === 0) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchEndpoints()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold">接口管理</h1>
      <p class="text-muted-foreground mt-1">管理 Swagger 和 gRPC 服务接口</p>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">总接口数</p>
              <p class="text-2xl font-bold">{{ endpoints.length }}</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-purple-100 flex items-center justify-center">
              <Globe class="w-5 h-5 text-purple-600" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">HTTP 接口</p>
              <p class="text-2xl font-bold text-blue-600">{{ endpoints.filter(e => e.type === 'http' || e.type === 'rest').length }}</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center">
              <Globe class="w-5 h-5 text-blue-600" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">gRPC 接口</p>
              <p class="text-2xl font-bold text-green-600">{{ endpoints.filter(e => e.type === 'grpc').length }}</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-green-100 flex items-center justify-center">
              <Server class="w-5 h-5 text-green-600" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardContent class="p-4">
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="搜索接口名称或地址..."
              class="pl-9"
              @keyup.enter="handleSearch"
            />
          </div>
          <Button @click="showAddDialog = true">
            <Plus class="w-4 h-4 mr-2" />
            添加接口
          </Button>
        </div>
      </CardContent>
    </Card>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <Card v-else>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>任务类型</TableHead>
              <TableHead>类型</TableHead>
              <TableHead>执行模式</TableHead>
              <TableHead>URL</TableHead>
              <TableHead>描述</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="filteredEndpoints.length === 0">
              <TableCell colspan="8" class="text-center text-muted-foreground py-8">
                暂无接口，点击上方"添加接口"按钮注册新的服务接口
              </TableCell>
            </TableRow>
            <TableRow 
              v-for="endpoint in filteredEndpoints" 
              :key="endpoint.id"
              class="cursor-pointer hover:bg-muted/50"
              @click="viewDetail(endpoint)"
            >
              <TableCell>
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-purple-100 flex items-center justify-center">
                    <Globe class="w-4 h-4 text-purple-600" />
                  </div>
                  <span class="font-medium">{{ endpoint.name || `Endpoint #${endpoint.id}` }}</span>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="secondary">{{ getTaskTypeLabel(endpoint.taskType) }}</Badge>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{{ endpoint.type || '-' }}</Badge>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{{ getModeLabel(endpoint.mode) }}</Badge>
              </TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">{{ endpoint.httpUrl || '-' }}</code>
              </TableCell>
              <TableCell class="max-w-[200px] truncate">{{ endpoint.description || '-' }}</TableCell>
              <TableCell @click.stop>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8">
                      <MoreVertical class="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="viewDetail(endpoint)">
                      <Eye class="w-4 h-4 mr-2" />
                      查看详情
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="convertToBpmn(endpoint)" :disabled="convertingId === endpoint.id">
                      <Loader2 v-if="convertingId === endpoint.id" class="w-4 h-4 mr-2 animate-spin" />
                      <ArrowRight v-else class="w-4 h-4 mr-2" />
                      转换为BPMN
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Pagination
      :page="currentPage"
      :page-size="pageSize"
      :total="total"
      @update:page="handlePageChange"
      @update:page-size="handlePageSizeChange"
    />

    <Dialog v-model:open="showAddDialog">
      <DialogContent class="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>添加服务接口</DialogTitle>
          <DialogDescription>
            通过 OpenAPI/Swagger 文档或 gRPC 反射注册服务接口
          </DialogDescription>
        </DialogHeader>
        <div class="space-y-4 py-4">
          <div class="space-y-2">
            <Label>接口类型</Label>
            <Tabs v-model="endpointType" class="w-full">
              <TabsList class="grid w-full grid-cols-2">
                <TabsTrigger value="openapi">
                  <FileCode class="w-4 h-4 mr-2" />
                  OpenAPI
                </TabsTrigger>
                <TabsTrigger value="grpc">
                  <Server class="w-4 h-4 mr-2" />
                  gRPC
                </TabsTrigger>
              </TabsList>
            </Tabs>
          </div>

          <div v-if="endpointType === 'openapi'" class="space-y-4">
            <div class="space-y-2">
              <Label>内容类型</Label>
              <Select v-model="openapiContentType">
                <SelectTrigger>
                  <SelectValue placeholder="选择内容类型" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="json">JSON</SelectItem>
                  <SelectItem value="yaml">YAML</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label>OpenAPI 文档</Label>
              <Textarea
                v-model="openapiDoc"
                placeholder="粘贴 OpenAPI/Swagger 文档内容..."
                class="min-h-[200px] font-mono text-sm"
              />
            </div>
          </div>

          <div v-if="endpointType === 'grpc'" class="space-y-4">
            <div class="space-y-2">
              <Label>gRPC 服务地址</Label>
              <Input
                v-model="grpcTarget"
                placeholder="例如: localhost:50051"
              />
            </div>
            <p class="text-sm text-muted-foreground">
              gRPC 服务需要启用 gRPC Server Reflection
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showAddDialog = false">取消</Button>
          <Button @click="addEndpoint" :disabled="adding">
            <Loader2 v-if="adding" class="w-4 h-4 mr-2 animate-spin" />
            添加
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- BPMN Dialog -->
    <Dialog v-model:open="showBpmnDialog">
      <DialogContent class="sm:max-w-[800px]">
        <DialogHeader>
          <DialogTitle>BPMN 转换结果</DialogTitle>
          <DialogDescription>
            接口已转换为 BPMN ServiceTask 元素
          </DialogDescription>
        </DialogHeader>
        <div class="py-4">
          <pre class="bg-muted p-4 rounded-lg overflow-auto max-h-[400px] text-sm font-mono whitespace-pre-wrap break-all border">{{ bpmnContent }}</pre>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showBpmnDialog = false">关闭</Button>
          <Button @click="copyBpmn">复制内容</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
