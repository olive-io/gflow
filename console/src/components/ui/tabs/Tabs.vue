<script setup lang="ts">
import { TabsRoot, type TabsRootProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<TabsRootProps & { class?: string }>()
const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <TabsRoot 
    v-bind="props" 
    :class="cn('w-full', props.class)"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <slot />
  </TabsRoot>
</template>
