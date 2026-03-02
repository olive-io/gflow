<script setup lang="ts">
import { AvatarImage } from 'radix-vue'
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  src?: string
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <AvatarImage
    v-if="src"
    :src="src"
    :class="cn('aspect-square h-full w-full', props.class)"
  />
</template>
