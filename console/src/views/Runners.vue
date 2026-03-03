<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Search, MoreVertical, Server, Power, PowerOff, RefreshCw, Loader2 } from 'lucide-vue-next'
import { runnersApi } from '@/api'
import type { Runner } from '@/types/api'
import Pagination from '@/components/Pagination.vue'

const runners = ref<Runner[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')

const filteredRunners = computed(() => {
  if (!searchQuery.value) return runners.value
  const query = searchQuery.value.toLowerCase()
  return runners.value.filter(r => 
    r.uid?.toLowerCase().includes(query) || 
    r.hostname?.toLowerCase().includes(query) ||
    r.listenUrl?.toLowerCase().includes(query)
  )
})

const onlineCount = computed(() => runners.value.filter(r => r.online === 1).length)
const offlineCount = computed(() => runners.value.filter(r => r.online !== 1).length)

const getStatusBadge = (online: number) => {
  if (online === 1) {
    return { variant: 'success' as const, label: '在线' }
  }
  return { variant: 'error' as const, label: '离线' }
}

async function fetchRunners() {
  loading.value = true
  try {
    const response = await runnersApi.list({
      page: currentPage.value,
      size: pageSize.value,
    })
    runners.value = response.runners || []
    total.value = Number(response.total) || runners.value.length
  } catch (error) {
    console.error('Failed to fetch runners:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchRunners()
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchRunners()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  fetchRunners()
}

function enableRunner(id: number) {
  console.log('Enable:', id)
}

function disableRunner(id: number) {
  console.log('Disable:', id)
}

function restartRunner(id: number) {
  console.log('Restart:', id)
}

function formatDateTime(timestamp: number | string): string {
  if (!timestamp) return '-'
  const ts = typeof timestamp === 'string' ? Number(timestamp) : timestamp
  if (ts === 0) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchRunners()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold">Runner 管理</h1>
      <p class="text-muted-foreground mt-1">管理流程执行器节点</p>
    </div>

    <!-- Stats -->
    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">在线 Runner</p>
              <p class="text-2xl font-bold text-green-600">{{ onlineCount }}</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-green-100 flex items-center justify-center">
              <Server class="w-5 h-5 text-green-600" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">总 Runner 数</p>
              <p class="text-2xl font-bold">{{ runners.length }}</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center">
              <Server class="w-5 h-5 text-blue-600" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">离线 Runner</p>
              <p class="text-2xl font-bold text-red-600">{{ offlineCount }}</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-red-100 flex items-center justify-center">
              <Server class="w-5 h-5 text-red-600" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Filters -->
    <Card>
      <CardContent class="p-4">
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="搜索 Runner 名称或地址..."
              class="pl-9"
              @keyup.enter="handleSearch"
            />
          </div>
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
              <TableHead>名称</TableHead>
              <TableHead>地址</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>主机名</TableHead>
              <TableHead>最后心跳</TableHead>
              <TableHead>版本</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="filteredRunners.length === 0">
              <TableCell colspan="7" class="text-center text-muted-foreground py-8">
                暂无 Runner
              </TableCell>
            </TableRow>
            <TableRow v-for="runner in filteredRunners" :key="runner.id">
              <TableCell>
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-muted flex items-center justify-center">
                    <Server class="w-4 h-4" />
                  </div>
                  <span class="font-medium">{{ runner.uid || `Runner #${runner.id}` }}</span>
                </div>
              </TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">{{ runner.listenUrl || '-' }}</code>
              </TableCell>
              <TableCell>
                <Badge :variant="getStatusBadge(runner.online).variant">
                  {{ getStatusBadge(runner.online).label }}
                </Badge>
              </TableCell>
              <TableCell>{{ runner.hostname || '-' }}</TableCell>
              <TableCell>{{ formatDateTime(runner.onlineTimestamp) }}</TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">v{{ runner.version || 'N/A' }}</code>
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8">
                      <MoreVertical class="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem v-if="runner.online !== 1" @click="enableRunner(runner.id)">
                      <Power class="w-4 h-4 mr-2" />
                      启用
                    </DropdownMenuItem>
                    <DropdownMenuItem v-if="runner.online === 1" @click="disableRunner(runner.id)">
                      <PowerOff class="w-4 h-4 mr-2" />
                      停用
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="restartRunner(runner.id)">
                      <RefreshCw class="w-4 h-4 mr-2" />
                      重启
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
