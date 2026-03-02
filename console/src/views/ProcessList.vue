<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
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
import { Search, MoreVertical, Eye, StopCircle, RefreshCw, ChevronLeft, ChevronRight, Loader2 } from 'lucide-vue-next'
import { processesApi } from '@/api'
import type { Process } from '@/types/api'
import { ProcessStatus } from '@/types/api'

const processes = ref<Process[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const searchQuery = ref('')
const statusFilter = ref('all')

const filteredProcesses = computed(() => {
  let result = processes.value
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(p => 
      p.name?.toLowerCase().includes(query) || 
      p.definitions_uid?.toLowerCase().includes(query)
    )
  }
  if (statusFilter.value !== 'all') {
    const statusMap: Record<string, number> = {
      running: ProcessStatus.Executing,
      completed: ProcessStatus.Commit,
      failed: ProcessStatus.Failed,
      ready: ProcessStatus.Ready,
      cancelled: ProcessStatus.Destroy,
    }
    result = result.filter(p => p.status === statusMap[statusFilter.value])
  }
  return result
})

const getStatusBadge = (status: ProcessStatus) => {
  const statusMap: Record<number, { variant: 'success' | 'warning' | 'error' | 'info' | 'secondary', label: string }> = {
    [ProcessStatus.Ready]: { variant: 'warning', label: '准备中' },
    [ProcessStatus.Executing]: { variant: 'info', label: '运行中' },
    [ProcessStatus.Commit]: { variant: 'success', label: '已完成' },
    [ProcessStatus.Failed]: { variant: 'error', label: '失败' },
    [ProcessStatus.Destroy]: { variant: 'secondary', label: '已取消' },
    [ProcessStatus.Rollback]: { variant: 'secondary', label: '已回滚' },
  }
  return statusMap[status] || { variant: 'secondary', label: '未知' }
}

async function fetchProcesses() {
  loading.value = true
  try {
    const response = await processesApi.list({
      page: currentPage.value,
      size: pageSize.value,
    })
    processes.value = response.items || []
    total.value = response.total || 0
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

function viewInstance(id: number) {
  console.log('View:', id)
}

function stopInstance(id: number) {
  console.log('Stop:', id)
}

function retryInstance(id: number) {
  console.log('Retry:', id)
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    fetchProcesses()
  }
}

function nextPage() {
  if (currentPage.value < Math.ceil(total.value / pageSize.value)) {
    currentPage.value++
    fetchProcesses()
  }
}

function formatDateTime(timestamp: number): string {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

function formatDuration(startTime: number, endTime: number): string {
  if (!startTime) return '-'
  const end = endTime || Date.now() / 1000
  const duration = end - startTime
  if (duration < 60) return `${Math.floor(duration)}s`
  if (duration < 3600) return `${Math.floor(duration / 60)}m ${Math.floor(duration % 60)}s`
  return `${Math.floor(duration / 3600)}h ${Math.floor((duration % 3600) / 60)}m`
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
              <SelectItem value="running">运行中</SelectItem>
              <SelectItem value="completed">已完成</SelectItem>
              <SelectItem value="failed">失败</SelectItem>
              <SelectItem value="ready">准备中</SelectItem>
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
              <TableHead>开始时间</TableHead>
              <TableHead>执行时长</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="filteredProcesses.length === 0">
              <TableCell colspan="6" class="text-center text-muted-foreground py-8">
                暂无流程实例
              </TableCell>
            </TableRow>
            <TableRow v-for="process in filteredProcesses" :key="process.id">
              <TableCell class="font-medium">{{ process.name || `Process #${process.id}` }}</TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">{{ process.definitions_uid }}</code>
              </TableCell>
              <TableCell>
                <Badge :variant="getStatusBadge(process.status).variant">
                  {{ getStatusBadge(process.status).label }}
                </Badge>
              </TableCell>
              <TableCell>{{ formatDateTime(process.start_time) }}</TableCell>
              <TableCell>{{ formatDuration(process.start_time, process.end_time) }}</TableCell>
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
                    <DropdownMenuItem v-if="process.status === ProcessStatus.Running" @click="stopInstance(process.id)">
                      <StopCircle class="w-4 h-4 mr-2" />
                      停止
                    </DropdownMenuItem>
                    <DropdownMenuItem v-if="process.status === ProcessStatus.Failed" @click="retryInstance(process.id)">
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
    <div class="flex items-center justify-between">
      <p class="text-sm text-muted-foreground">
        显示 {{ (currentPage - 1) * pageSize + 1 }} - {{ Math.min(currentPage * pageSize, total) }} 条，共 {{ total }} 条
      </p>
      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" @click="prevPage" :disabled="currentPage === 1">
          <ChevronLeft class="w-4 h-4" />
          上一页
        </Button>
        <Button variant="outline" size="sm" @click="nextPage" :disabled="currentPage >= Math.ceil(total / pageSize)">
          下一页
          <ChevronRight class="w-4 h-4" />
        </Button>
      </div>
    </div>
  </div>
</template>
