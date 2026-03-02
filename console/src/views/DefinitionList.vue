<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
import { Plus, Search, MoreVertical, Pencil, Trash2, Copy, FileText, Loader2 } from 'lucide-vue-next'
import { definitionsApi } from '@/api'
import type { Definitions } from '@/types/api'

const router = useRouter()

const definitions = ref<Definitions[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(20)

const searchQuery = ref('')
const statusFilter = ref('all')

const filteredDefinitions = computed(() => {
  let result = definitions.value
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(d => 
      d.name?.toLowerCase().includes(query) || 
      d.uid?.toLowerCase().includes(query)
    )
  }
  return result
})

async function fetchDefinitions() {
  loading.value = true
  try {
    const response = await definitionsApi.list({ page: page.value, size: size.value })
    definitions.value = response.items || []
    total.value = response.total || 0
  } catch (error) {
    console.error('Failed to fetch definitions:', error)
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchDefinitions()
}

function createDefinition() {
  router.push('/designer')
}

function editDefinition(uid: string) {
  router.push(`/designer/${uid}`)
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

function formatDate(timestamp: number): string {
  if (!timestamp) return '-'
  return new Date(timestamp * 1000).toLocaleDateString('zh-CN')
}

onMounted(() => {
  fetchDefinitions()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">流程定义</h1>
        <p class="text-muted-foreground mt-1">管理和部署工作流定义</p>
      </div>
      <Button @click="createDefinition">
        <Plus class="w-4 h-4 mr-2" />
        新建流程
      </Button>
    </div>

    <!-- Filters -->
    <Card>
      <CardContent class="p-4">
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="搜索流程名称或 UID..."
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
              <SelectItem value="active">已部署</SelectItem>
              <SelectItem value="inactive">已停用</SelectItem>
              <SelectItem value="draft">草稿</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>

    <!-- Loading state -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <!-- Empty state -->
    <div v-else-if="filteredDefinitions.length === 0" class="text-center py-12">
      <FileText class="w-12 h-12 mx-auto text-muted-foreground/50" />
      <p class="mt-4 text-muted-foreground">暂无流程定义</p>
      <Button class="mt-4" @click="createDefinition">
        <Plus class="w-4 h-4 mr-2" />
        创建第一个流程
      </Button>
    </div>

    <!-- Definition cards -->
    <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <Card v-for="def in filteredDefinitions" :key="def.uid" class="hover:shadow-md transition-shadow">
        <CardHeader class="pb-3">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-lg bg-gflow-primary/10 flex items-center justify-center">
                <FileText class="w-5 h-5 text-gflow-primary" />
              </div>
              <div>
                <CardTitle class="text-base">{{ def.name || '未命名流程' }}</CardTitle>
                <CardDescription class="text-xs mt-0.5">{{ def.uid }}</CardDescription>
              </div>
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" class="h-8 w-8">
                  <MoreVertical class="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem @click="editDefinition(def.uid)">
                  <Pencil class="w-4 h-4 mr-2" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Copy class="w-4 h-4 mr-2" />
                  复制
                </DropdownMenuItem>
                <DropdownMenuItem @click="deleteDefinition(def.uid)" class="text-destructive">
                  <Trash2 class="w-4 h-4 mr-2" />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </CardHeader>
        <CardContent>
          <p class="text-sm text-muted-foreground mb-4 line-clamp-2">{{ def.description || '暂无描述' }}</p>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Badge variant="success">
                已部署
              </Badge>
              <span class="text-xs text-muted-foreground">v{{ def.version }}</span>
            </div>
            <span class="text-xs text-muted-foreground">{{ formatDate(def.create_at) }}</span>
          </div>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
