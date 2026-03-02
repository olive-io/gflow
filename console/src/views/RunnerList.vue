<script setup lang="ts">
import { ref } from 'vue'
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
import { Search, MoreVertical, Power, PowerOff, RefreshCw, Server } from 'lucide-vue-next'

// Mock data
const runners = ref([
  { id: 1, name: 'runner-01', host: '192.168.1.101', port: 8080, status: 'online', lastHeartbeat: '2024-01-15 14:30:00', tasks: 12, version: '1.0.0' },
  { id: 2, name: 'runner-02', host: '192.168.1.102', port: 8080, status: 'online', lastHeartbeat: '2024-01-15 14:29:55', tasks: 8, version: '1.0.0' },
  { id: 3, name: 'runner-03', host: '192.168.1.103', port: 8080, status: 'offline', lastHeartbeat: '2024-01-15 13:45:00', tasks: 0, version: '1.0.0' },
  { id: 4, name: 'runner-04', host: '192.168.1.104', port: 8080, status: 'online', lastHeartbeat: '2024-01-15 14:30:10', tasks: 5, version: '1.0.1' },
  { id: 5, name: 'runner-05', host: '192.168.1.105', port: 8080, status: 'busy', lastHeartbeat: '2024-01-15 14:29:50', tasks: 25, version: '1.0.0' },
])

const searchQuery = ref('')

const getStatusBadge = (status: string) => {
  const statusMap: Record<string, { variant: 'success' | 'warning' | 'error' | 'secondary', label: string }> = {
    online: { variant: 'success', label: '在线' },
    offline: { variant: 'error', label: '离线' },
    busy: { variant: 'warning', label: '繁忙' },
  }
  return statusMap[status] || { variant: 'secondary', label: status }
}

function handleSearch() {
  // TODO: 实现搜索逻辑
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
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-2xl font-bold">Runner 管理</h1>
      <p class="text-muted-foreground mt-1">管理流程执行器节点</p>
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

    <!-- Table -->
    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>名称</TableHead>
              <TableHead>地址</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>任务数</TableHead>
              <TableHead>最后心跳</TableHead>
              <TableHead>版本</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="runner in runners" :key="runner.id">
              <TableCell>
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-muted flex items-center justify-center">
                    <Server class="w-4 h-4" />
                  </div>
                  <span class="font-medium">{{ runner.name }}</span>
                </div>
              </TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">{{ runner.host }}:{{ runner.port }}</code>
              </TableCell>
              <TableCell>
                <Badge :variant="getStatusBadge(runner.status).variant">
                  {{ getStatusBadge(runner.status).label }}
                </Badge>
              </TableCell>
              <TableCell>
                <span class="font-medium">{{ runner.tasks }}</span>
              </TableCell>
              <TableCell>{{ runner.lastHeartbeat }}</TableCell>
              <TableCell>
                <code class="text-xs bg-muted px-2 py-1 rounded">v{{ runner.version }}</code>
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8">
                      <MoreVertical class="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem v-if="runner.status === 'offline'" @click="enableRunner(runner.id)">
                      <Power class="w-4 h-4 mr-2" />
                      启用
                    </DropdownMenuItem>
                    <DropdownMenuItem v-if="runner.status !== 'offline'" @click="disableRunner(runner.id)">
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

    <!-- Stats -->
    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">在线 Runner</p>
              <p class="text-2xl font-bold text-gflow-success">4</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-green-100 flex items-center justify-center">
              <Server class="w-5 h-5 text-gflow-success" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">总任务数</p>
              <p class="text-2xl font-bold">50</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-blue-100 flex items-center justify-center">
              <Server class="w-5 h-5 text-gflow-info" />
            </div>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="p-4">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-muted-foreground">离线 Runner</p>
              <p class="text-2xl font-bold text-gflow-error">1</p>
            </div>
            <div class="w-10 h-10 rounded-lg bg-red-100 flex items-center justify-center">
              <Server class="w-5 h-5 text-gflow-error" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
