<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'
import {
  Plus,
  Search,
  LayoutGrid,
  List,
  Loader2,
  FileText,
  Receipt,
  UserPlus,
  Package,
  FileSignature,
  Megaphone,
  MoreHorizontal,
  Eye,
  Pencil,
  Trash2,
} from 'lucide-vue-next'
import { definitionsApi } from '@/api'
import type { Definitions } from '@/types/api'
import Pagination from '@/components/Pagination.vue'

const router = useRouter()

const definitions = ref<Definitions[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(12)
const searchQuery = ref('')
const statusFilter = ref('all')
const viewMode = ref<'grid' | 'list'>('list')

function getIconForDefinition(def: Definitions) {
  const name = def.name?.toLowerCase() || ''
  if (name.includes('报销') || name.includes('费用')) return Receipt
  if (name.includes('入职') || name.includes('培训')) return UserPlus
  if (name.includes('物资') || name.includes('领用')) return Package
  if (name.includes('合同')) return FileSignature
  if (name.includes('营销') || name.includes('活动')) return Megaphone
  return FileText
}

async function fetchDefinitions() {
  loading.value = true
  try {
    const response = await definitionsApi.list({ page: currentPage.value, size: pageSize.value })
    definitions.value = response.definitionsList || []
    total.value = Number(response.total) || 0
  } catch (error) {
    console.error('Failed to fetch definitions:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchDefinitions()
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchDefinitions()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  fetchDefinitions()
}

function createDefinition() {
  router.push('/designer/new')
}

function editDefinition(uid: string) {
  router.push(`/designer/${uid}`)
}

function viewDefinition(uid: string) {
  router.push(`/definitions/${uid}`)
}

async function deleteDefinition(uid: string) {
  if (!confirm('确定要删除此流程定义吗？')) {
    return
  }
  try {
    await definitionsApi.remove(uid)
    await fetchDefinitions()
  } catch (error) {
    console.error('Failed to delete definition:', error)
  }
}

function formatDate(timestamp: number | string): string {
  if (!timestamp) return '-'
  const ts = typeof timestamp === 'string' ? Number(timestamp) : timestamp
  if (ts === 0) return '-'
  return new Date(ts * 1000).toLocaleDateString('zh-CN')
}

onMounted(() => {
  fetchDefinitions()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold">流程定义</h1>
        <p class="text-muted-foreground mt-1">管理和设计您的业务工作流定义</p>
      </div>
      <Button @click="createDefinition">
        <Plus class="w-4 h-4 mr-2" />
        新建流程
      </Button>
    </div>

    <div class="flex flex-wrap items-center gap-4">
      <div class="flex-1 min-w-[240px]">
        <div class="relative">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            v-model="searchQuery"
            placeholder="搜索流程名称..."
            class="pl-10"
            @keyup.enter="handleSearch"
          />
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Select v-model="statusFilter">
          <SelectTrigger class="w-[140px]">
            <SelectValue>
              {{ statusFilter === 'all' ? '所有状态' : statusFilter === 'active' ? '已部署' : '未部署' }}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">所有状态</SelectItem>
            <SelectItem value="active">已部署</SelectItem>
            <SelectItem value="inactive">未部署</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div class="h-10 w-px bg-border mx-2 hidden sm:block" />
      <div class="flex bg-muted p-1 rounded-lg">
        <Button
          variant="ghost"
          size="icon"
          :class="viewMode === 'grid' ? 'bg-background shadow-sm text-primary' : 'text-muted-foreground'"
          @click="viewMode = 'grid'"
        >
          <LayoutGrid class="w-4 h-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          :class="viewMode === 'list' ? 'bg-background shadow-sm text-primary' : 'text-muted-foreground'"
          @click="viewMode = 'list'"
        >
          <List class="w-4 h-4" />
        </Button>
      </div>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <div v-else-if="definitions.length === 0" class="text-center py-12">
      <FileText class="w-12 h-12 mx-auto text-muted-foreground/50" />
      <p class="mt-4 text-muted-foreground">暂无流程定义</p>
      <Button class="mt-4" @click="createDefinition">
        <Plus class="w-4 h-4 mr-2" />
        创建第一个流程
      </Button>
    </div>

    <template v-else>
      <div v-if="viewMode === 'grid'" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        <div
          v-for="def in definitions"
          :key="def.uid"
          class="bg-card border rounded-xl overflow-hidden hover:shadow-xl hover:shadow-primary/5 transition-all group"
          :class="def.isExecute ? 'border-b-4 border-b-primary/50' : 'border-b-4 border-b-muted'"
        >
          <div class="p-5">
            <div class="flex items-start justify-between mb-4">
              <div
                class="w-12 h-12 rounded-xl flex items-center justify-center"
                :class="def.isExecute ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'"
              >
                <component :is="getIconForDefinition(def)" class="w-6 h-6" />
              </div>
              <span
                class="px-2.5 py-1 rounded-full text-xs font-bold flex items-center gap-1"
                :class="def.isExecute 
                  ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' 
                  : 'bg-muted text-muted-foreground'"
              >
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :class="def.isExecute ? 'bg-green-500' : 'bg-muted-foreground/50'"
                />
                {{ def.isExecute ? '已部署' : '未部署' }}
              </span>
            </div>
            <h3 class="text-lg font-bold truncate mb-1">{{ def.name || def.uid || '未命名流程' }}</h3>
            <p class="text-muted-foreground text-sm line-clamp-2 min-h-[40px] mb-4">
              {{ def.description || '暂无描述' }}
            </p>
            <div class="grid grid-cols-2 gap-y-3 pt-4 border-t">
              <div>
                <p class="text-[10px] text-muted-foreground uppercase font-bold tracking-wider">版本</p>
                <p class="text-sm font-semibold">v{{ def.version }}</p>
              </div>
              <div>
                <p class="text-[10px] text-muted-foreground uppercase font-bold tracking-wider">创建时间</p>
                <p class="text-sm font-semibold">{{ formatDate(def.createAt) }}</p>
              </div>
            </div>
          </div>
          <div class="flex divide-x divide-border border-t bg-muted/30">
            <button
              class="flex-1 py-3 text-xs font-bold text-muted-foreground hover:text-primary hover:bg-primary/5 transition-colors"
              @click.stop="viewDefinition(def.uid)"
            >
              查看
            </button>
            <button
              class="flex-1 py-3 text-xs font-bold text-muted-foreground hover:text-primary hover:bg-primary/5 transition-colors"
              @click.stop="editDefinition(def.uid)"
            >
              编辑
            </button>
            <button
              class="flex-1 py-3 text-xs font-bold text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
              @click.stop="deleteDefinition(def.uid)"
            >
              删除
            </button>
          </div>
        </div>

        <div
          class="border-2 border-dashed rounded-xl flex flex-col items-center justify-center p-8 hover:border-primary/50 hover:bg-primary/5 transition-all cursor-pointer group"
          @click="createDefinition"
        >
          <div class="w-16 h-16 bg-muted rounded-full flex items-center justify-center text-muted-foreground group-hover:bg-primary/10 group-hover:text-primary transition-colors mb-4">
            <Plus class="w-8 h-8" />
          </div>
          <p class="font-bold">创建新流程</p>
          <p class="text-sm text-muted-foreground text-center mt-2 px-4">
            点击此处开始设计您的下一个自动化工作流
          </p>
        </div>
      </div>

      <div v-else class="bg-card border rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>流程名称</TableHead>
              <TableHead class="hidden md:table-cell">描述</TableHead>
              <TableHead>版本</TableHead>
              <TableHead>状态</TableHead>
              <TableHead class="hidden sm:table-cell">创建时间</TableHead>
              <TableHead class="text-right">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="def in definitions"
              :key="def.uid"
              class="cursor-pointer hover:bg-muted/50"
              @click="viewDefinition(def.uid)"
            >
              <TableCell>
                <div class="flex items-center gap-3">
                  <div
                    class="w-8 h-8 rounded-lg flex items-center justify-center shrink-0"
                    :class="def.isExecute ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground'"
                  >
                    <component :is="getIconForDefinition(def)" class="w-4 h-4" />
                  </div>
                  <span class="font-medium truncate">{{ def.name || def.uid || '未命名流程' }}</span>
                </div>
              </TableCell>
              <TableCell class="hidden md:table-cell">
                <span class="text-muted-foreground text-sm truncate block max-w-[200px]">
                  {{ def.description || '暂无描述' }}
                </span>
              </TableCell>
              <TableCell>
                <span class="font-mono text-sm">v{{ def.version }}</span>
              </TableCell>
              <TableCell>
                <Badge :variant="def.isExecute ? 'success' : 'secondary'">
                  {{ def.isExecute ? '已部署' : '未部署' }}
                </Badge>
              </TableCell>
              <TableCell class="hidden sm:table-cell">
                <span class="text-muted-foreground text-sm">{{ formatDate(def.createAt) }}</span>
              </TableCell>
              <TableCell class="text-right">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8" @click.stop>
                      <MoreHorizontal class="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click.stop="viewDefinition(def.uid)">
                      <Eye class="w-4 h-4 mr-2" />
                      查看
                    </DropdownMenuItem>
                    <DropdownMenuItem @click.stop="editDefinition(def.uid)">
                      <Pencil class="w-4 h-4 mr-2" />
                      编辑
                    </DropdownMenuItem>
                    <DropdownMenuItem 
                      class="text-destructive focus:text-destructive"
                      @click.stop="deleteDefinition(def.uid)"
                    >
                      <Trash2 class="w-4 h-4 mr-2" />
                      删除
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </template>

    <Pagination
      :page="currentPage"
      :page-size="pageSize"
      :total="total"
      @update:page="handlePageChange"
      @update:page-size="handlePageSizeChange"
    />
  </div>
</template>
