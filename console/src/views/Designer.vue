<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Save,
  Play,
  Undo,
  Redo,
  ZoomIn,
  ZoomOut,
  Maximize,
  Hand,
  MousePointer,
  Square,
  Diamond,
  Circle,
  ArrowRight,
} from 'lucide-vue-next'
import BpmnModeler from 'bpmn-js/lib/Modeler'
import 'bpmn-js/dist/assets/diagram-js.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-codes.css'
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-embedded.css'

const canvasRef = ref<HTMLElement>()
let bpmnModeler: InstanceType<typeof BpmnModeler> | null = null

const processName = ref('新建流程')
const processKey = ref('new-process')
const selectedElement = ref<{ id: string; businessObject?: { name?: string } } | null>(null)

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

const tools = [
  { id: 'select', icon: MousePointer, label: '选择' },
  { id: 'hand', icon: Hand, label: '移动' },
  { id: 'task', icon: Square, label: '任务' },
  { id: 'gateway', icon: Diamond, label: '网关' },
  { id: 'event', icon: Circle, label: '事件' },
  { id: 'connection', icon: ArrowRight, label: '连线' },
]

const activeTool = ref('select')
const zoom = ref(1)

function initModeler() {
  if (!canvasRef.value) return

  bpmnModeler = new BpmnModeler({
    container: canvasRef.value,
    keyboard: {
      bindTo: document,
    },
  })

  bpmnModeler.importXML(defaultDiagram).then(() => {
    const canvas = bpmnModeler?.get('canvas') as { zoom: (level: string | number) => void } | undefined
    canvas?.zoom('fit-viewport')
  })

  bpmnModeler.on('selection.changed', (e: { newSelection: Array<{ id: string; businessObject?: { name?: string } }> }) => {
    selectedElement.value = e.newSelection[0] || null
  })
}

async function handleSave() {
  if (!bpmnModeler) return
  
  try {
    const { xml } = await bpmnModeler.saveXML({ format: true })
    console.log('Saved XML:', xml)
  } catch (error) {
    console.error('Save failed:', error)
  }
}

function handleDeploy() {
  console.log('Deploy process')
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

onMounted(() => {
  initModeler()
})

onUnmounted(() => {
  bpmnModeler?.destroy()
})
</script>

<template>
  <div class="h-[calc(100vh-120px)] flex flex-col">
    <!-- Toolbar -->
    <div class="bg-card border-b px-4 py-2 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <Button variant="outline" size="sm" @click="handleSave">
          <Save class="w-4 h-4 mr-1" />
          保存
        </Button>
        <Button size="sm" @click="handleDeploy">
          <Play class="w-4 h-4 mr-1" />
          部署
        </Button>
        <div class="w-px h-6 bg-border mx-2" />
        <Button variant="ghost" size="icon" @click="handleUndo">
          <Undo class="w-4 h-4" />
        </Button>
        <Button variant="ghost" size="icon" @click="handleRedo">
          <Redo class="w-4 h-4" />
        </Button>
        <div class="w-px h-6 bg-border mx-2" />
        <Button variant="ghost" size="icon" @click="handleZoomIn">
          <ZoomIn class="w-4 h-4" />
        </Button>
        <Button variant="ghost" size="icon" @click="handleZoomOut">
          <ZoomOut class="w-4 h-4" />
        </Button>
        <Button variant="ghost" size="icon" @click="handleFitViewport">
          <Maximize class="w-4 h-4" />
        </Button>
      </div>
      <div class="text-sm text-muted-foreground">
        {{ Math.round(zoom * 100) }}%
      </div>
    </div>

    <div class="flex-1 flex overflow-hidden">
      <!-- Left palette -->
      <div class="w-14 bg-card border-r flex flex-col items-center py-2 gap-1">
        <Button
          v-for="tool in tools"
          :key="tool.id"
          :variant="activeTool === tool.id ? 'secondary' : 'ghost'"
          size="icon"
          class="w-10 h-10"
          :title="tool.label"
          @click="activeTool = tool.id"
        >
          <component :is="tool.icon" class="w-4 h-4" />
        </Button>
      </div>

      <!-- Canvas -->
      <div ref="canvasRef" class="flex-1 bg-muted/30"></div>

      <!-- Right properties panel -->
      <div class="w-72 bg-card border-l overflow-y-auto">
        <Card class="border-0 rounded-none shadow-none">
          <CardHeader class="border-b">
            <CardTitle class="text-base">属性</CardTitle>
          </CardHeader>
          <CardContent class="p-4 space-y-4">
            <div class="space-y-2">
              <Label>流程名称</Label>
              <Input v-model="processName" />
            </div>
            <div class="space-y-2">
              <Label>流程 Key</Label>
              <Input v-model="processKey" />
            </div>
            
            <template v-if="selectedElement">
              <div class="pt-4 border-t">
                <h4 class="font-medium mb-3">选中元素</h4>
                <div class="space-y-2">
                  <div class="space-y-2">
                    <Label>ID</Label>
                    <Input :value="selectedElement.id" disabled />
                  </div>
                  <div class="space-y-2">
                    <Label>名称</Label>
                    <Input :value="selectedElement.businessObject?.name || ''" />
                  </div>
                </div>
              </div>
            </template>
            
            <div v-else class="pt-4 border-t">
              <p class="text-sm text-muted-foreground text-center py-4">
                选择一个元素查看属性
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>

<style>
.bpmn-container {
  height: 100%;
  width: 100%;
}

.djs-palette {
  display: none;
}

.djs-container {
  background: transparent !important;
}
</style>
