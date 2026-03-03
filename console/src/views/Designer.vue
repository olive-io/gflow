<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import {
  Loader2,
  Save,
  Play,
  Undo,
  Redo,
  ZoomIn,
  ZoomOut,
  Maximize,
  ArrowLeft,
} from 'lucide-vue-next'
import BpmnModeler from 'bpmn-js/lib/Modeler'
import 'bpmn-js/dist/assets/diagram-js.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-codes.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-embedded.css'
import { definitionsApi } from '@/api'
import type { Definitions } from '@/types/api'

const route = useRoute()
const router = useRouter()

const isModelerReady = ref(false)
const canvasRef = ref<HTMLElement | null>(null)
let bpmnModeler: InstanceType<typeof BpmnModeler> | null = null

const editingDefinition = ref<Definitions | null>(null)
const processName = ref('')
const processKey = ref('')
const processDescription = ref('')
const selectedElement = ref<{ id: string; businessObject?: { name?: string; $attrs?: Record<string, string> } } | null>(null)
const saving = ref(false)
const deploying = ref(false)
const zoom = ref(1)

const defaultDiagram = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
  xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI"
  xmlns:dc="http://www.omg.org/spec/DD/20100524/DC"
  xmlns:di="http://www.omg.org/spec/DD/20100524/DI"
  id="Definitions_1"
  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1" name="开始">
      <bpmn:outgoing>Flow_1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:task id="Task_1" name="任务">
      <bpmn:incoming>Flow_1</bpmn:incoming>
      <bpmn:outgoing>Flow_2</bpmn:outgoing>
    </bpmn:task>
    <bpmn:endEvent id="EndEvent_1" name="结束">
      <bpmn:incoming>Flow_2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="StartEvent_1" targetRef="Task_1" />
    <bpmn:sequenceFlow id="Flow_2" sourceRef="Task_1" targetRef="EndEvent_1" />
  </bpmn:process>
  <bpmndi:BPMNDiagram id="BPMNDiagram_1">
    <bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="Process_1">
      <bpmndi:BPMNShape id="StartEvent_1_di" bpmnElement="StartEvent_1">
        <dc:Bounds x="152" y="102" width="36" height="36" />
        <bpmndi:BPMNLabel>
          <dc:Bounds x="159" y="145" width="22" height="14" />
        </bpmndi:BPMNLabel>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNShape id="Task_1_di" bpmnElement="Task_1">
        <dc:Bounds x="240" y="80" width="100" height="80" />
      </bpmndi:BPMNShape>
      <bpmndi:BPMNShape id="EndEvent_1_di" bpmnElement="EndEvent_1">
        <dc:Bounds x="402" y="102" width="36" height="36" />
        <bpmndi:BPMNLabel>
          <dc:Bounds x="409" y="145" width="22" height="14" />
        </bpmndi:BPMNLabel>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNEdge id="Flow_1_di" bpmnElement="Flow_1">
        <di:waypoint x="188" y="120" />
        <di:waypoint x="240" y="120" />
      </bpmndi:BPMNEdge>
      <bpmndi:BPMNEdge id="Flow_2_di" bpmnElement="Flow_2">
        <di:waypoint x="340" y="120" />
        <di:waypoint x="402" y="120" />
      </bpmndi:BPMNEdge>
    </bpmndi:BPMNPlane>
  </bpmndi:BPMNDiagram>
</bpmn:definitions>`

async function loadDefinition() {
  const uid = route.params.id as string
  if (uid && uid !== 'new') {
    try {
      const response = await definitionsApi.get(uid)
      editingDefinition.value = response.definitions
      processName.value = response.definitions.name || ''
      processKey.value = response.definitions.uid || ''
      processDescription.value = response.definitions.description || ''
      return response.definitions.content
    } catch (error) {
      console.error('Failed to load definition:', error)
    }
  }
  return null
}

function initModeler(initialContent?: string | null) {
  if (!canvasRef.value) return

  bpmnModeler = new BpmnModeler({
    container: canvasRef.value,
  })

  const diagram = initialContent || editingDefinition.value?.content || defaultDiagram
  
  bpmnModeler.importXML(diagram).then(() => {
    const canvas = bpmnModeler?.get('canvas') as { zoom: (level: string) => void } | undefined
    canvas?.zoom('fit-viewport')
    isModelerReady.value = true
  }).catch((error: Error) => {
    console.error('Failed to import XML:', error)
  })

  bpmnModeler.on('selection.changed', (e: { newSelection: Array<{ id: string; businessObject?: { name?: string; $attrs?: Record<string, string> } }> }) => {
    selectedElement.value = e.newSelection[0] || null
  })
  
  bpmnModeler.on('element.changed', () => {
    // 元素变化时可以更新属性面板
  })
}

function destroyModeler() {
  if (bpmnModeler) {
    bpmnModeler.destroy()
    bpmnModeler = null
    isModelerReady.value = false
  }
}

function goBack() {
  destroyModeler()
  router.push('/definitions')
}

async function handleSave() {
  if (!bpmnModeler) return
  
  saving.value = true
  try {
    const { xml } = await bpmnModeler.saveXML({ format: true })
    if (xml) {
      await definitionsApi.deploy({
        content: xml,
        description: processDescription.value,
      })
      alert('保存成功')
    }
  } catch (error) {
    console.error('Save failed:', error)
    alert('保存失败')
  } finally {
    saving.value = false
  }
}

async function handleDeploy() {
  if (!bpmnModeler) return
  
  deploying.value = true
  try {
    const { xml } = await bpmnModeler.saveXML({ format: true })
    if (xml) {
      await definitionsApi.deploy({
        content: xml,
        description: processDescription.value,
      })
      alert('部署成功')
      goBack()
    }
  } catch (error) {
    console.error('Deploy failed:', error)
    alert('部署失败')
  } finally {
    deploying.value = false
  }
}

function handleUndo() {
  const commandStack = bpmnModeler?.get('commandStack') as { undo: () => void } | undefined
  commandStack?.undo()
}

function handleRedo() {
  const commandStack = bpmnModeler?.get('commandStack') as { redo: () => void } | undefined
  commandStack?.redo()
}

function handleZoomIn() {
  const canvas = bpmnModeler?.get('canvas') as { zoom: (level: number) => void } | undefined
  if (canvas) {
    zoom.value = Math.min(zoom.value * 1.2, 4)
    canvas.zoom(zoom.value)
  }
}

function handleZoomOut() {
  const canvas = bpmnModeler?.get('canvas') as { zoom: (level: number) => void } | undefined
  if (canvas) {
    zoom.value = Math.max(zoom.value / 1.2, 0.2)
    canvas.zoom(zoom.value)
  }
}

function handleFitViewport() {
  const canvas = bpmnModeler?.get('canvas') as { zoom: (level: string) => void } | undefined
  canvas?.zoom('fit-viewport')
  zoom.value = 1
}

onMounted(async () => {
  await nextTick()
  setTimeout(async () => {
    const content = await loadDefinition()
    initModeler(content)
  }, 100)
})

onUnmounted(() => {
  destroyModeler()
})

watch(() => route.params.id, async (newId) => {
  if (newId) {
    destroyModeler()
    isModelerReady.value = false
    await nextTick()
    setTimeout(async () => {
      const content = await loadDefinition()
      initModeler(content)
    }, 100)
  }
})
</script>

<template>
  <div class="fixed inset-0 z-50 bg-background flex flex-col">
    <!-- Toolbar -->
    <div class="bg-card border-b px-4 h-14 flex items-center justify-between shrink-0">
      <div class="flex items-center gap-2">
        <Button variant="ghost" size="sm" @click="goBack">
          <ArrowLeft class="w-4 h-4 mr-1" />
          返回
        </Button>
        <div class="w-px h-6 bg-border mx-2"></div>
        <Button variant="outline" size="sm" @click="handleSave" :disabled="saving || !isModelerReady">
          <Loader2 v-if="saving" class="w-4 h-4 mr-1 animate-spin" />
          <Save v-else class="w-4 h-4 mr-1" />
          保存
        </Button>
        <Button size="sm" @click="handleDeploy" :disabled="deploying || !isModelerReady">
          <Loader2 v-if="deploying" class="w-4 h-4 mr-1 animate-spin" />
          <Play v-else class="w-4 h-4 mr-1" />
          部署
        </Button>
        <div class="w-px h-6 bg-border mx-2"></div>
        <Button variant="ghost" size="sm" @click="handleUndo" :disabled="!isModelerReady">
          <Undo class="w-4 h-4 mr-1" />
          撤销
        </Button>
        <Button variant="ghost" size="sm" @click="handleRedo" :disabled="!isModelerReady">
          <Redo class="w-4 h-4 mr-1" />
          重做
        </Button>
        <div class="w-px h-6 bg-border mx-2"></div>
        <Button variant="ghost" size="sm" @click="handleZoomIn" :disabled="!isModelerReady">
          <ZoomIn class="w-4 h-4 mr-1" />
          放大
        </Button>
        <Button variant="ghost" size="sm" @click="handleZoomOut" :disabled="!isModelerReady">
          <ZoomOut class="w-4 h-4 mr-1" />
          缩小
        </Button>
        <Button variant="ghost" size="sm" @click="handleFitViewport" :disabled="!isModelerReady">
          <Maximize class="w-4 h-4 mr-1" />
          适应
        </Button>
      </div>
      <div class="flex items-center gap-4">
        <span class="text-sm text-muted-foreground">
          {{ Math.round(zoom * 100) }}%
        </span>
        <span class="text-sm font-medium">
          {{ processName || '未命名流程' }}
        </span>
      </div>
    </div>

    <div class="flex-1 flex overflow-hidden min-h-0">
      <!-- Canvas with built-in palette -->
      <div ref="canvasRef" class="flex-1 relative min-h-0">
        <div v-if="!isModelerReady" class="absolute inset-0 flex items-center justify-center bg-muted/30 z-10">
          <div class="flex flex-col items-center gap-2">
            <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
            <span class="text-sm text-muted-foreground">加载中...</span>
          </div>
        </div>
      </div>

      <!-- Right properties panel -->
      <div class="w-80 bg-card border-l overflow-y-auto shrink-0">
        <Card class="border-0 rounded-none shadow-none">
          <CardHeader class="border-b">
            <CardTitle class="text-base">属性面板</CardTitle>
          </CardHeader>
          <CardContent class="p-4 space-y-4">
            <!-- Process Properties -->
            <div class="space-y-4">
              <h4 class="text-sm font-medium text-muted-foreground">流程属性</h4>
              <div class="space-y-2">
                <Label>流程名称</Label>
                <Input v-model="processName" placeholder="请输入流程名称" />
              </div>
              <div class="space-y-2">
                <Label>流程 Key</Label>
                <Input v-model="processKey" placeholder="请输入流程 Key" disabled />
              </div>
              <div class="space-y-2">
                <Label>描述</Label>
                <Textarea v-model="processDescription" placeholder="请输入描述" :rows="3" />
              </div>
            </div>
            
            <div class="h-px w-full bg-border my-4"></div>
            
            <!-- Selected Element Properties -->
            <div v-if="selectedElement" class="space-y-4">
              <h4 class="text-sm font-medium text-muted-foreground">元素属性</h4>
              <div class="space-y-2">
                <Label>元素 ID</Label>
                <Input :value="selectedElement.id" disabled class="font-mono text-xs" />
              </div>
              <div class="space-y-2">
                <Label>元素名称</Label>
                <Input :value="selectedElement.businessObject?.name || ''" placeholder="请输入名称" />
              </div>
            </div>
            
            <div v-else class="py-8 text-center">
              <p class="text-sm text-muted-foreground">
                点击画布中的元素<br>查看和编辑属性
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>

<style>
.bpmn-js-palette {
  background: var(--card) !important;
  border-right: 1px solid var(--border) !important;
}

.bpmn-js-palette .entry {
  border-radius: 4px !important;
}

.bpmn-js-palette .entry:hover {
  background: var(--accent) !important;
}

.bpmn-js-palette .separator {
  border-top: 1px solid var(--border) !important;
  margin: 4px 8px !important;
}

.djs-container {
  background: hsl(var(--muted) / 0.3) !important;
}

.djs-palette {
  background: var(--card) !important;
  border: 1px solid var(--border) !important;
  border-radius: 4px !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1) !important;
}

.djs-palette .entry {
  border-radius: 4px !important;
  transition: background-color 0.15s ease !important;
}

.djs-palette .entry:hover {
  background-color: var(--accent) !important;
}

.djs-palette-separator {
  border-top: 1px solid var(--border) !important;
  margin: 4px 8px !important;
}

.djs-context-pad {
  background: var(--card) !important;
  border: 1px solid var(--border) !important;
  border-radius: 4px !important;
}

.djs-context-pad .entry {
  border-radius: 4px !important;
}

.djs-context-pad .entry:hover {
  background-color: var(--accent) !important;
}
</style>
