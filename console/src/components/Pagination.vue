<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-vue-next'

interface Props {
  page: number
  pageSize: number
  total: number
  pageSizes?: number[]
}

const props = withDefaults(defineProps<Props>(), {
  pageSizes: () => [10, 20, 50, 100],
})

const emit = defineEmits<{
  'update:page': [value: number]
  'update:pageSize': [value: number]
}>()

const totalPages = computed(() => Math.ceil(props.total / props.pageSize))

const startItem = computed(() => (props.page - 1) * props.pageSize + 1)

const endItem = computed(() => Math.min(props.page * props.pageSize, props.total))

function goToPage(newPage: number) {
  if (newPage >= 1 && newPage <= totalPages.value) {
    emit('update:page', newPage)
  }
}

function handlePageSizeChange(newSize: string) {
  emit('update:pageSize', Number(newSize))
  emit('update:page', 1)
}

function getVisiblePages(): (number | string)[] {
  const pages: (number | string)[] = []
  const total = totalPages.value
  const current = props.page

  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    pages.push(1)
    
    if (current > 3) {
      pages.push('...')
    }
    
    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)
    
    for (let i = start; i <= end; i++) {
      pages.push(i)
    }
    
    if (current < total - 2) {
      pages.push('...')
    }
    
    pages.push(total)
  }
  
  return pages
}
</script>

<template>
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-4">
      <p class="text-sm text-muted-foreground">
        显示 {{ startItem }} - {{ endItem }} 条，共 {{ total }} 条
      </p>
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">每页</span>
        <Select :model-value="String(pageSize)" @update:model-value="handlePageSizeChange">
          <SelectTrigger class="w-[70px] h-8">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="size in pageSizes" :key="size" :value="String(size)">
              {{ size }}
            </SelectItem>
          </SelectContent>
        </Select>
        <span class="text-sm text-muted-foreground">条</span>
      </div>
    </div>
    
    <div class="flex items-center gap-1">
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="page === 1"
        @click="goToPage(1)"
      >
        <ChevronsLeft class="h-4 w-4" />
      </Button>
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="page === 1"
        @click="goToPage(page - 1)"
      >
        <ChevronLeft class="h-4 w-4" />
      </Button>
      
      <div class="flex items-center gap-1 mx-2">
        <template v-for="(p, index) in getVisiblePages()" :key="index">
          <span v-if="p === '...'" class="px-2 text-muted-foreground">...</span>
          <Button
            v-else
            :variant="p === page ? 'default' : 'outline'"
            size="icon"
            class="h-8 w-8"
            @click="goToPage(p as number)"
          >
            {{ p }}
          </Button>
        </template>
      </div>
      
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="page >= totalPages"
        @click="goToPage(page + 1)"
      >
        <ChevronRight class="h-4 w-4" />
      </Button>
      <Button
        variant="outline"
        size="icon"
        class="h-8 w-8"
        :disabled="page >= totalPages"
        @click="goToPage(totalPages)"
      >
        <ChevronsRight class="h-4 w-4" />
      </Button>
    </div>
  </div>
</template>
