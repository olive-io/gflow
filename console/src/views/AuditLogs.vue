<script setup lang="ts">
import { ref } from 'vue'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Search, Eye, ChevronLeft, ChevronRight } from 'lucide-vue-next'

// Mock data
const auditLogs = ref([
  { id: 1, action: '创建流程', module: '流程定义', user: 'admin', ip: '192.168.1.100', status: 'success', time: '2024-01-15 14:30:00', details: '创建流程定义: order-process' },
  { id: 2, action: '部署流程', module: '流程定义', user: 'admin', ip: '192.168.1.100', status: 'success', time: '2024-01-15 14:25:00', details: '部署流程定义: user-approval v2' },
  { id: 3, action: '启动实例', module: '流程实例', user: 'operator1', ip: '192.168.1.101', status: 'success', time: '2024-01-15 14:20:00', details: '启动流程实例: order-process #123' },
  { id: 4, action: '删除流程', module: '流程定义', user: 'admin', ip: '192.168.1.100', status: 'failed', time: '2024-01-15 14:15:00', details: '删除流程定义失败: data-sync (存在运行中实例)' },
  { id: 5, action: '用户登录', module: '认证', user: 'operator2', ip: '192.168.1.102', status: 'success', time: '2024-01-15 14:10:00', details: '用户登录成功' },
  { id: 6, action: '修改权限', module: '用户管理', user: 'admin', ip: '192.168.1.100', status: 'success', time: '2024-01-15 14:05:00', details: '修改用户角色: viewer1 -> operator' },
  { id: 7, action: '停止实例', module: '流程实例', user: 'operator1', ip: '192.168.1.101', status: 'success', time: '2024-01-15 14:00:00', details: '停止流程实例: report-gen #118' },
  { id: 8, action: '创建用户', module: '用户管理', user: 'admin', ip: '192.168.1.100', status: 'success', time: '2024-01-15 13:55:00', details: '创建新用户: viewer3' },
])

const searchQuery = ref('')
const moduleFilter = ref('all')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(50)

const showDetailDialog = ref(false)
const selectedLog = ref<any>(null)

const getStatusBadge = (status: string) => {
  const statusMap: Record<string, { variant: 'success' | 'error', label: string }> = {
    success: { variant: 'success', label: '成功' },
    failed: { variant: 'error', label: '失败' },
  }
  return statusMap[status] || { variant: 'success', label: status }
}

const getModuleBadge = (module: string) => {
  const moduleColors: Record<string, string> = {
    '流程定义': 'bg-blue-100 text-blue-800',
    '流程实例': 'bg-green-100 text-green-800',
    '用户管理': 'bg-purple-100 text-purple-800',
    '认证': 'bg-orange-100 text-orange-800',
  }
  return moduleColors[module] || 'bg-gray-100 text-gray-800'
}

function handleSearch() {
  // TODO: 实现搜索逻辑
}

function viewDetail(log: any) {
  selectedLog.value = log
  showDetailDialog.value = true
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

function nextPage() {
  if (currentPage.value < Math.ceil(total.value / pageSize.value)) {
    currentPage.value++
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold">审计日志</h1>
      <p class="text-muted-foreground mt-1">查看系统操作审计记录</p>
    </div>

    <!-- Filters -->
    <Card>
      <CardContent class="p-4">
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="搜索操作内容..."
              class="pl-9"
              @keyup.enter="handleSearch"
            />
          </div>
          <Select v-model="moduleFilter">
            <SelectTrigger class="w-full md:w-[180px]">
              <SelectValue placeholder="模块筛选" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部模块</SelectItem>
              <SelectItem value="definition">流程定义</SelectItem>
              <SelectItem value="instance">流程实例</SelectItem>
              <SelectItem value="user">用户管理</SelectItem>
              <SelectItem value="auth">认证</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>

    <!-- Table -->
    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>操作</TableHead>
              <TableHead>模块</TableHead>
              <TableHead>操作人</TableHead>
              <TableHead>IP 地址</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>时间</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="log in auditLogs" :key="log.id">
              <TableCell class="font-medium">{{ log.action }}</TableCell>
              <TableCell>
                <span :class="['text-xs px-2 py-1 rounded', getModuleBadge(log.module)]">
                  {{ log.module }}
                </span>
              </TableCell>
              <TableCell>{{ log.user }}</TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">{{ log.ip }}</code>
              </TableCell>
              <TableCell>
                <Badge :variant="getStatusBadge(log.status).variant">
                  {{ getStatusBadge(log.status).label }}
                </Badge>
              </TableCell>
              <TableCell>{{ log.time }}</TableCell>
              <TableCell>
                <Button variant="ghost" size="icon" class="h-8 w-8" @click="viewDetail(log)">
                  <Eye class="w-4 h-4" />
                </Button>
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

    <!-- Detail Dialog -->
    <Dialog v-model:open="showDetailDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>日志详情</DialogTitle>
          <DialogDescription>查看操作详细信息</DialogDescription>
        </DialogHeader>
        <div v-if="selectedLog" class="space-y-4">
          <div class="grid grid-cols-3 gap-4">
            <div>
              <p class="text-sm text-muted-foreground">操作</p>
              <p class="font-medium">{{ selectedLog.action }}</p>
            </div>
            <div>
              <p class="text-sm text-muted-foreground">模块</p>
              <p class="font-medium">{{ selectedLog.module }}</p>
            </div>
            <div>
              <p class="text-sm text-muted-foreground">状态</p>
              <Badge :variant="getStatusBadge(selectedLog.status).variant">
                {{ getStatusBadge(selectedLog.status).label }}
              </Badge>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-muted-foreground">操作人</p>
              <p class="font-medium">{{ selectedLog.user }}</p>
            </div>
            <div>
              <p class="text-sm text-muted-foreground">IP 地址</p>
              <p class="font-medium">{{ selectedLog.ip }}</p>
            </div>
          </div>
          <div>
            <p class="text-sm text-muted-foreground">时间</p>
            <p class="font-medium">{{ selectedLog.time }}</p>
          </div>
          <div>
            <p class="text-sm text-muted-foreground">详情</p>
            <p class="font-medium">{{ selectedLog.details }}</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
