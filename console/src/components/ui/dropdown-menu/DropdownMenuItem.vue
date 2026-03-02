<script setup lang="ts">
import { DropdownMenuItem, type DropdownMenuItemEmits, type DropdownMenuItemProps, useForwardPropsEmits } from 'radix-vue'
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

interface Props extends DropdownMenuItemProps {
  class?: HTMLAttributes['class']
  inset?: boolean
}

const props = defineProps<Props>()
const emits = defineEmits<DropdownMenuItemEmits>()

const forwarded = useForwardPropsEmits(props, emits)
</script>

<template>
  <DropdownMenuItem
    v-bind="forwarded"
    :class="cn(
      'relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
      inset && 'pl-8',
      props.class,
    )"
  >
    <slot />
  </DropdownMenuItem>
</template>
