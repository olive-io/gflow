<script setup lang="ts">
import { DialogClose } from 'radix-vue'
import { X } from 'lucide-vue-next'
</script>

<template>
  <DialogClose
    class="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
  >
    <X class="h-4 w-4" />
    <span class="sr-only">Close</span>
  </DialogClose>
</template>
