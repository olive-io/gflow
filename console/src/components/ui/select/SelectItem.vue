<script setup lang="ts">
import { SelectItem, type SelectItemProps, SelectItemIndicator, SelectItemText, useForwardProps } from 'radix-vue'
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'
import { Check } from 'lucide-vue-next'

interface Props extends SelectItemProps {
  class?: HTMLAttributes['class']
}

const props = defineProps<Props>()

const forwarded = useForwardProps(props)
</script>

<template>
  <SelectItem
    v-bind="forwarded"
    :class="cn(
      'relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50',
      props.class,
    )"
  >
    <span class="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
      <SelectItemIndicator>
        <Check class="h-4 w-4" />
      </SelectItemIndicator>
    </span>
    <SelectItemText>
      <slot />
    </SelectItemText>
  </SelectItem>
</template>
