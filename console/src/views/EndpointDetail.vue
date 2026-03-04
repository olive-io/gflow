<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from '@/components/ui/dialog'
import { ArrowLeft, Globe, Server, FileText, ArrowRight, Loader2 } from 'lucide-vue-next'
import { endpointsApi } from '@/api'
import type { Endpoint, EndpointProperty } from '@/types/api'

const route = useRoute()
const router = useRouter()

const endpoint = ref<Endpoint | null>(null)
const loading = ref(true)

const showBpmnDialog = ref(false)
const bpmnContent = ref('')
const converting = ref(false)

const taskTypeLabel = computed(() => {
  if (!endpoint.value) return ''
  const taskType = endpoint.value.taskType
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
  return types[taskType as number] || 'Unknown'
})

const typeIcon = computed(() => {
  if (!endpoint.value) return Globe
  return endpoint.value.type === 'grpc' ? Server : Globe
})

const modeLabel = computed(() => {
  if (!endpoint.value) return ''
  const mode = endpoint.value.mode
  if (typeof mode === 'string') {
    if (!mode || mode === 'UnknownMode') {
      return 'Simple'
    }
    return mode
  }
  return 'Simple'
})

const propertiesList = computed(() => {
  if (!endpoint.value?.properties) return []
  return Object.entries(endpoint.value.properties).map(([key, value]) => ({
    name: key,
    ...value
  }))
})

const resultsList = computed(() => {
  if (!endpoint.value?.results) return []
  return Object.entries(endpoint.value.results).map(([key, value]) => ({
    name: key,
    ...value
  }))
})

const headersList = computed(() => {
  if (!endpoint.value?.headers) return []
  return Object.entries(endpoint.value.headers)
    .filter(([key]) => key && endpoint.value?.headers[key] !== undefined)
    .map(([key, value]) => ({
      name: key,
      value: value || ''
    }))
})

async function fetchEndpoint() {
  loading.value = true
  try {
    const id = route.params.id as string
    const response = await endpointsApi.list({ size: 100 })
    const found = response.endpoints?.find(e => String(e.id) === id)
    if (found) {
      endpoint.value = found
    }
  } catch (error) {
    console.error('Failed to fetch endpoint:', error)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.push('/endpoints')
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

async function convertToBpmn() {
  if (!endpoint.value) return
  converting.value = true
  try {
    const response = await endpointsApi.convert({ id: endpoint.value.id, format: 'BPMN' })
    bpmnContent.value = formatXml(response.content || '')
    showBpmnDialog.value = true
  } catch (error) {
    console.error('Failed to convert endpoint:', error)
    alert('转换失败')
  } finally {
    converting.value = false
  }
}

function copyBpmn() {
  navigator.clipboard.writeText(bpmnContent.value)
  alert('已复制到剪贴板')
}

onMounted(() => {
  fetchEndpoint()
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
        <div class="w-px h-6 bg-border"></div>
        <div v-if="endpoint">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg bg-purple-100 flex items-center justify-center">
              <component :is="typeIcon" class="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <h1 class="text-2xl font-bold">{{ endpoint.name || `Endpoint #${endpoint.id}` }}</h1>
              <p class="text-muted-foreground text-sm">{{ endpoint.description || '暂无描述' }}</p>
            </div>
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Button variant="outline" @click="convertToBpmn" :disabled="converting">
          <Loader2 v-if="converting" class="w-4 h-4 mr-2 animate-spin" />
          <ArrowRight v-else class="w-4 h-4 mr-2" />
          转换为BPMN
        </Button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <template v-else-if="endpoint">
      <!-- Overview Cards -->
      <div class="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">接口类型</p>
                <p class="text-lg font-semibold">{{ endpoint.type || '-' }}</p>
              </div>
              <div class="w-8 h-8 rounded-lg bg-blue-100 flex items-center justify-center">
                <Globe class="w-4 h-4 text-blue-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">任务类型</p>
                <p class="text-lg font-semibold">{{ taskTypeLabel }}</p>
              </div>
              <div class="w-8 h-8 rounded-lg bg-green-100 flex items-center justify-center">
                <FileText class="w-4 h-4 text-green-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">执行模式</p>
                <p class="text-lg font-semibold">{{ modeLabel }}</p>
              </div>
              <div class="w-8 h-8 rounded-lg bg-orange-100 flex items-center justify-center">
                <Server class="w-4 h-4 text-orange-600" />
              </div>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-muted-foreground">目标节点</p>
                <p class="text-lg font-semibold">{{ endpoint.targets?.length || 0 }}</p>
              </div>
              <div class="w-8 h-8 rounded-lg bg-purple-100 flex items-center justify-center">
                <Server class="w-4 h-4 text-purple-600" />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Basic Info -->
      <Card>
        <CardHeader>
          <CardTitle>基本信息</CardTitle>
        </CardHeader>
        <CardContent>
          <div class="grid gap-4 md:grid-cols-2">
            <div class="space-y-1">
              <p class="text-sm text-muted-foreground">接口 ID</p>
              <p class="font-mono">{{ endpoint.id }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-sm text-muted-foreground">接口名称</p>
              <p>{{ endpoint.name || '-' }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-sm text-muted-foreground">接口类型</p>
              <Badge variant="outline">{{ endpoint.type }}</Badge>
            </div>
            <div class="space-y-1">
              <p class="text-sm text-muted-foreground">任务类型</p>
              <Badge variant="secondary">{{ taskTypeLabel }}</Badge>
            </div>
            <div class="space-y-1">
              <p class="text-sm text-muted-foreground">执行模式</p>
              <Badge variant="outline">{{ modeLabel }}</Badge>
            </div>
            <div class="space-y-1">
              <p class="text-sm text-muted-foreground">目标节点</p>
              <div class="flex flex-wrap gap-2">
                <Badge v-for="target in endpoint.targets" :key="target" variant="outline">
                  {{ target }}
                </Badge>
                <span v-if="!endpoint.targets?.length" class="text-muted-foreground">-</span>
              </div>
            </div>
            <div class="space-y-1 md:col-span-2">
              <p class="text-sm text-muted-foreground">HTTP URL</p>
              <code class="text-sm bg-muted px-2 py-1 rounded block">{{ endpoint.httpUrl || '-' }}</code>
            </div>
            <div class="space-y-1 md:col-span-2">
              <p class="text-sm text-muted-foreground">描述</p>
              <p>{{ endpoint.description || '-' }}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Request Headers -->
      <Card v-if="headersList.length > 0">
        <CardHeader>
          <CardTitle>请求头</CardTitle>
          <CardDescription>调用接口时需要设置的 HTTP 请求头</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Header 名称</TableHead>
                <TableHead>值</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="header in headersList" :key="header.name">
                <TableCell class="font-mono">{{ header.name }}</TableCell>
                <TableCell>{{ header.value || '-' }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <!-- Input Parameters -->
      <Card>
        <CardHeader>
          <CardTitle>输入参数</CardTitle>
          <CardDescription>调用该接口时需要传入的参数</CardDescription>
        </CardHeader>
        <CardContent>
          <Table v-if="propertiesList.length > 0">
            <TableHeader>
              <TableRow>
                <TableHead>参数名</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>Kind</TableHead>
                <TableHead>默认值</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="prop in propertiesList" :key="prop.name">
                <TableCell class="font-mono">{{ prop.name }}</TableCell>
                <TableCell>
                  <Badge variant="outline">{{ prop.type }}</Badge>
                </TableCell>
                <TableCell>{{ prop.kind || '-' }}</TableCell>
                <TableCell>{{ prop.default || '-' }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <div v-else class="py-8 text-center text-muted-foreground">
            暂无输入参数
          </div>
        </CardContent>
      </Card>

      <!-- Output Results -->
      <Card>
        <CardHeader>
          <CardTitle>输出结果</CardTitle>
          <CardDescription>接口调用返回的结果字段</CardDescription>
        </CardHeader>
        <CardContent>
          <Table v-if="resultsList.length > 0">
            <TableHeader>
              <TableRow>
                <TableHead>字段名</TableHead>
                <TableHead>类型</TableHead>
                <TableHead>Kind</TableHead>
                <TableHead>默认值</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="result in resultsList" :key="result.name">
                <TableCell class="font-mono">{{ result.name }}</TableCell>
                <TableCell>
                  <Badge variant="outline">{{ result.type }}</Badge>
                </TableCell>
                <TableCell>{{ result.kind || '-' }}</TableCell>
                <TableCell>{{ result.default || '-' }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <div v-else class="py-8 text-center text-muted-foreground">
            暂无输出结果
          </div>
        </CardContent>
      </Card>

      <!-- Metadata -->
      <Card v-if="endpoint.metadata && Object.keys(endpoint.metadata).length > 0">
        <CardHeader>
          <CardTitle>元数据</CardTitle>
          <CardDescription>接口的附加元数据信息</CardDescription>
        </CardHeader>
        <CardContent>
          <div class="space-y-3">
            <div v-for="(value, key) in endpoint.metadata" :key="key" class="flex items-center justify-between py-2 border-b last:border-0">
              <span class="font-mono text-sm">{{ key }}</span>
              <span class="text-sm text-muted-foreground">{{ value }}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </template>

    <!-- Not Found -->
    <div v-else class="py-12 text-center">
      <p class="text-muted-foreground">接口不存在</p>
      <Button variant="link" @click="goBack">返回列表</Button>
    </div>

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
