<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Search, MoreVertical, Eye, StopCircle, RefreshCw, Loader2 } from 'lucide-vue-next'
import { processesApi } from '@/api'
import type { Process } from '@/types/api'
import Pagination from '@/components/Pagination.vue'

const router = useRouter()

const processes = ref<Process[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')
const statusFilter = ref('all')

const filteredProcesses = computed(() => {
  let result = processes.value
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(p => 
      p.name?.toLowerCase().includes(query) || 
      p.definitionsUid?.toLowerCase().includes(query)
    )
  }
  if (statusFilter.value !== 'all') {
    result = result.filter(p => p.status?.toLowerCase() === statusFilter.value)
  }
  return result
})

const getStatusBadge = (status: string) => {
  const statusMap: Record<string, { variant: 'success' | 'warning' | 'error' | 'info' | 'secondary', label: string }> = {
    'unknownstatus': { variant: 'secondary', label: '未知' },
    'waiting': { variant: 'warning', label: '等待中' },
    'running': { variant: 'info', label: '运行中' },
    'success': { variant: 'success', label: '成功' },
    'warn': { variant: 'warning', label: '警告' },
    'failed': { variant: 'error', label: '失败' },
  }
  const key = (status || '').toLowerCase()
  return statusMap[key] || { variant: 'secondary', label: status || '未知' }
}

const getStageBadge = (stage: string) => {
  const stageMap: Record<string, { variant: 'success' | 'warning' | 'error' | 'info' | 'secondary', label: string }> = {
    'unknownstage': { variant: 'secondary', label: '未知' },
    'prepare': { variant: 'warning', label: '准备中' },
    'ready': { variant: 'info', label: '就绪' },
    'commit': { variant: 'success', label: '提交' },
    'rollback': { variant: 'error', label: '回滚' },
    'destroy': { variant: 'secondary', label: '销毁' },
    'finish': { variant: 'success', label: '完成' },
  }
  const key = (stage || '').toLowerCase()
  return stageMap[key] || { variant: 'secondary', label: stage || '未知' }
}

async function fetchProcesses() {
  loading.value = true
  try {
    const response = await processesApi.list({
      page: currentPage.value,
      size: pageSize.value,
    })
    processes.value = response.processes || []
    total.value = Number(response.total) || 0
  } catch (error) {
    console.error('Failed to fetch processes:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchProcesses()
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchProcesses()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  fetchProcesses()
}

function viewInstance(id: number) {
  router.push(`/processes/${id}`)
}

function stopInstance(id: number) {
  console.log('Stop:', id)
}

function retryInstance(id: number) {
  console.log('Retry:', id)
}

function formatDateTime(timestamp: number | string): string {
  if (!timestamp) return '-'
  const ts = typeof timestamp === 'string' ? Number(timestamp) : timestamp
  if (ts === 0) return '-'
  const ms = ts > 10000000000 ? ts : ts * 1000
  return new Date(ms).toLocaleString('zh-CN')
}

function formatDuration(startTime: number | string, endTime: number | string): string {
  const start = typeof startTime === 'string' ? Number(startTime) : startTime
  if (!start || start === 0) return '-'
  const end = endTime ? (typeof endTime === 'string' ? Number(endTime) : endTime) : Date.now()
  const duration = end - start
  if (duration < 0) return '-'
  const durationSec = duration > 100000 ? Math.floor(duration / 1000) : duration
  if (durationSec < 60) return `${Math.floor(durationSec)}s`
  if (durationSec < 3600) return `${Math.floor(durationSec / 60)}m ${Math.floor(durationSec % 60)}s`
  return `${Math.floor(durationSec / 3600)}h ${Math.floor((durationSec % 3600) / 60)}m`
}

onMounted(() => {
  fetchProcesses()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold">流程实例</h1>
      <p class="text-muted-foreground mt-1">查看和管理流程实例执行情况</p>
    </div>

    <!-- Filters -->
    <Card>
      <CardContent class="p-4">
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="搜索实例名称..."
              class="pl-9"
              @keyup.enter="handleSearch"
            />
          </div>
          <Select v-model="statusFilter">
            <SelectTrigger class="w-full md:w-[180px]">
              <SelectValue placeholder="状态筛选" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="waiting">等待中</SelectItem>
              <SelectItem value="running">运行中</SelectItem>
              <SelectItem value="success">成功</SelectItem>
              <SelectItem value="warn">警告</SelectItem>
              <SelectItem value="failed">失败</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>

    <!-- Loading state -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <!-- Table -->
    <Card v-else>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>实例名称</TableHead>
              <TableHead>流程定义</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>阶段</TableHead>
              <TableHead>开始时间</TableHead>
              <TableHead>执行时长</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="filteredProcesses.length === 0">
              <TableCell colspan="7" class="text-center text-muted-foreground py-8">
                暂无流程实例
              </TableCell>
            </TableRow>
            <TableRow v-for="process in filteredProcesses" :key="process.id">
              <TableCell class="font-medium">{{ process.name || `Process #${process.id}` }}</TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">{{ process.definitionsUid || '-' }}</code>
              </TableCell>
              <TableCell>
                <Badge :variant="getStatusBadge(process.status).variant">
                  {{ getStatusBadge(process.status).label }}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge :variant="getStageBadge(process.stage).variant">
                  {{ getStageBadge(process.stage).label }}
                </Badge>
              </TableCell>
              <TableCell>{{ formatDateTime(process.startAt) }}</TableCell>
              <TableCell>{{ formatDuration(process.startAt, process.endAt) }}</TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8">
                      <MoreVertical class="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="viewInstance(process.id)">
                      <Eye class="w-4 h-4 mr-2" />
                      查看详情
                    </DropdownMenuItem>
                    <DropdownMenuItem v-if="process.status?.toLowerCase() === 'running'" @click="stopInstance(process.id)">
                      <StopCircle class="w-4 h-4 mr-2" />
                      停止
                    </DropdownMenuItem>
                    <DropdownMenuItem v-if="process.status?.toLowerCase() === 'failed'" @click="retryInstance(process.id)">
                      <RefreshCw class="w-4 h-4 mr-2" />
                      重试
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <!-- Pagination -->
    <Pagination
      :page="currentPage"
      :page-size="pageSize"
      :total="total"
      @update:page="handlePageChange"
      @update:page-size="handlePageSizeChange"
    />
  </div>
</template>
