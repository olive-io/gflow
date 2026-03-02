<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { GitBranch, CheckCircle, XCircle, Clock, TrendingUp, Activity } from 'lucide-vue-next'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'

// Stats data
const stats = ref([
  { title: '流程总数', value: '128', icon: GitBranch, color: 'text-gflow-primary', bgColor: 'bg-gflow-primary/10' },
  { title: '运行中', value: '23', icon: Activity, color: 'text-gflow-info', bgColor: 'bg-blue-100' },
  { title: '成功', value: '98', icon: CheckCircle, color: 'text-gflow-success', bgColor: 'bg-green-100' },
  { title: '失败', value: '7', icon: XCircle, color: 'text-gflow-error', bgColor: 'bg-red-100' },
])

// Recent executions
const recentExecutions = ref([
  { id: 1, name: '订单处理流程', status: 'success', startTime: '2024-01-15 14:30:00', duration: '2m 30s' },
  { id: 2, name: '数据同步任务', status: 'running', startTime: '2024-01-15 14:25:00', duration: '-' },
  { id: 3, name: '报表生成', status: 'success', startTime: '2024-01-15 14:20:00', duration: '5m 12s' },
  { id: 4, name: '用户审批流程', status: 'failed', startTime: '2024-01-15 14:15:00', duration: '1m 45s' },
  { id: 5, name: '库存检查', status: 'success', startTime: '2024-01-15 14:10:00', duration: '30s' },
])

const getStatusBadge = (status: string) => {
  const statusMap: Record<string, { variant: 'success' | 'warning' | 'error' | 'info', label: string }> = {
    success: { variant: 'success', label: '成功' },
    running: { variant: 'info', label: '运行中' },
    failed: { variant: 'error', label: '失败' },
    pending: { variant: 'warning', label: '等待中' },
  }
  return statusMap[status] || { variant: 'info', label: status }
}

// Charts
const trendChartRef = ref<HTMLElement>()
const pieChartRef = ref<HTMLElement>()
let trendChart: ECharts | null = null
let pieChart: ECharts | null = null

const initCharts = () => {
  // Trend chart
  if (trendChartRef.value) {
    trendChart = echarts.init(trendChartRef.value)
    const trendOption = {
      tooltip: {
        trigger: 'axis',
      },
      legend: {
        data: ['成功', '失败', '运行中'],
        bottom: 0,
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '15%',
        top: '10%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          name: '成功',
          type: 'line',
          smooth: true,
          data: [120, 132, 101, 134, 90, 230, 210],
          itemStyle: { color: '#22c55e' },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(34, 197, 94, 0.3)' },
              { offset: 1, color: 'rgba(34, 197, 94, 0.05)' },
            ]),
          },
        },
        {
          name: '失败',
          type: 'line',
          smooth: true,
          data: [5, 8, 3, 6, 2, 10, 8],
          itemStyle: { color: '#ef4444' },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(239, 68, 68, 0.3)' },
              { offset: 1, color: 'rgba(239, 68, 68, 0.05)' },
            ]),
          },
        },
        {
          name: '运行中',
          type: 'line',
          smooth: true,
          data: [20, 18, 25, 22, 15, 28, 23],
          itemStyle: { color: '#3b82f6' },
          areaStyle: {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
              { offset: 1, color: 'rgba(59, 130, 246, 0.05)' },
            ]),
          },
        },
      ],
    }
    trendChart.setOption(trendOption)
  }

  // Pie chart
  if (pieChartRef.value) {
    pieChart = echarts.init(pieChartRef.value)
    const pieOption = {
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'vertical',
        right: '5%',
        top: 'center',
      },
      series: [
        {
          name: '流程状态',
          type: 'pie',
          radius: ['45%', '70%'],
          center: ['35%', '50%'],
          avoidLabelOverlap: false,
          itemStyle: {
            borderRadius: 8,
            borderColor: '#fff',
            borderWidth: 2,
          },
          label: {
            show: false,
            position: 'center',
          },
          emphasis: {
            label: {
              show: true,
              fontSize: 16,
              fontWeight: 'bold',
            },
          },
          labelLine: {
            show: false,
          },
          data: [
            { value: 98, name: '成功', itemStyle: { color: '#22c55e' } },
            { value: 23, name: '运行中', itemStyle: { color: '#3b82f6' } },
            { value: 7, name: '失败', itemStyle: { color: '#ef4444' } },
          ],
        },
      ],
    }
    pieChart.setOption(pieOption)
  }
}

const handleResize = () => {
  trendChart?.resize()
  pieChart?.resize()
}

onMounted(() => {
  initCharts()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  trendChart?.dispose()
  pieChart?.dispose()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Stats cards -->
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card v-for="stat in stats" :key="stat.title">
        <CardContent class="p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-muted-foreground">{{ stat.title }}</p>
              <p class="text-3xl font-bold mt-1">{{ stat.value }}</p>
            </div>
            <div :class="['w-12 h-12 rounded-xl flex items-center justify-center', stat.bgColor]">
              <component :is="stat.icon" :class="['w-6 h-6', stat.color]" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Charts -->
    <div class="grid gap-6 lg:grid-cols-2">
      <!-- Trend chart -->
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <TrendingUp class="w-5 h-5 text-gflow-primary" />
            流程执行趋势
          </CardTitle>
          <CardDescription>最近一周流程执行情况</CardDescription>
        </CardHeader>
        <CardContent>
          <div ref="trendChartRef" class="h-[300px] w-full"></div>
        </CardContent>
      </Card>

      <!-- Pie chart -->
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Clock class="w-5 h-5 text-gflow-primary" />
            流程状态分布
          </CardTitle>
          <CardDescription>当前流程实例状态统计</CardDescription>
        </CardHeader>
        <CardContent>
          <div ref="pieChartRef" class="h-[300px] w-full"></div>
        </CardContent>
      </Card>
    </div>

    <!-- Recent executions -->
    <Card>
      <CardHeader>
        <CardTitle>最近执行</CardTitle>
        <CardDescription>最近执行的流程实例列表</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>流程名称</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>开始时间</TableHead>
              <TableHead>执行时长</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="execution in recentExecutions" :key="execution.id">
              <TableCell class="font-medium">{{ execution.name }}</TableCell>
              <TableCell>
                <Badge :variant="getStatusBadge(execution.status).variant">
                  {{ getStatusBadge(execution.status).label }}
                </Badge>
              </TableCell>
              <TableCell>{{ execution.startTime }}</TableCell>
              <TableCell>{{ execution.duration }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
