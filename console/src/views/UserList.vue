<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogClose,
} from '@/components/ui/dialog'
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
import { Plus, Search, MoreVertical, Pencil, Trash2 } from 'lucide-vue-next'

const users = ref([
  { id: 1, username: 'admin', email: 'admin@example.com', role: 'admin', status: 'active', createTime: '2024-01-01', lastLogin: '2024-01-15 14:30:00' },
  { id: 2, username: 'operator1', email: 'operator1@example.com', role: 'operator', status: 'active', createTime: '2024-01-05', lastLogin: '2024-01-15 10:20:00' },
  { id: 3, username: 'operator2', email: 'operator2@example.com', role: 'operator', status: 'active', createTime: '2024-01-05', lastLogin: '2024-01-14 16:45:00' },
  { id: 4, username: 'viewer1', email: 'viewer1@example.com', role: 'viewer', status: 'active', createTime: '2024-01-10', lastLogin: '2024-01-13 09:00:00' },
  { id: 5, username: 'viewer2', email: 'viewer2@example.com', role: 'viewer', status: 'inactive', createTime: '2024-01-10', lastLogin: '-' },
])

const searchQuery = ref('')
const roleFilter = ref('all')
const showCreateDialog = ref(false)

const newUser = ref({
  username: '',
  email: '',
  password: '',
  role: 'viewer',
})

const getRoleBadge = (role: string) => {
  const roleMap: Record<string, { variant: 'default' | 'secondary' | 'outline', label: string }> = {
    admin: { variant: 'default', label: '管理员' },
    operator: { variant: 'secondary', label: '操作员' },
    viewer: { variant: 'outline', label: '查看者' },
  }
  return roleMap[role] || { variant: 'outline', label: role }
}

const getStatusBadge = (status: string) => {
  const statusMap: Record<string, { variant: 'success' | 'secondary', label: string }> = {
    active: { variant: 'success', label: '正常' },
    inactive: { variant: 'secondary', label: '禁用' },
  }
  return statusMap[status] || { variant: 'secondary', label: status }
}

function handleSearch() {
}

function handleCreateUser() {
  console.log('Create user:', newUser.value)
  showCreateDialog.value = false
  newUser.value = {
    username: '',
    email: '',
    password: '',
    role: 'viewer',
  }
}

function editUser(id: number) {
  console.log('Edit:', id)
}

function deleteUser(id: number) {
  console.log('Delete:', id)
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">用户管理</h1>
        <p class="text-muted-foreground mt-1">管理系统用户和权限</p>
      </div>
      <Dialog v-model:open="showCreateDialog">
        <DialogTrigger as-child>
          <Button>
            <Plus class="w-4 h-4 mr-2" />
            创建用户
          </Button>
        </DialogTrigger>
        <DialogContent class="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>创建用户</DialogTitle>
            <DialogDescription>
              填写以下信息创建新用户
            </DialogDescription>
          </DialogHeader>
          <div class="grid gap-4 py-4">
            <div class="grid gap-2">
              <Label for="username">用户名</Label>
              <Input id="username" v-model="newUser.username" placeholder="请输入用户名" />
            </div>
            <div class="grid gap-2">
              <Label for="email">邮箱</Label>
              <Input id="email" v-model="newUser.email" type="email" placeholder="请输入邮箱" />
            </div>
            <div class="grid gap-2">
              <Label for="password">密码</Label>
              <Input id="password" v-model="newUser.password" type="password" placeholder="请输入密码" />
            </div>
            <div class="grid gap-2">
              <Label for="role">角色</Label>
              <Select v-model="newUser.role">
                <SelectTrigger>
                  <SelectValue placeholder="选择角色" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">管理员</SelectItem>
                  <SelectItem value="operator">操作员</SelectItem>
                  <SelectItem value="viewer">查看者</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <DialogClose as-child>
              <Button variant="outline">取消</Button>
            </DialogClose>
            <Button @click="handleCreateUser">创建</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>

    <!-- Filters -->
    <Card>
      <CardContent class="p-4">
        <div class="flex flex-col md:flex-row gap-4">
          <div class="flex-1 relative">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
            <Input
              v-model="searchQuery"
              placeholder="搜索用户名或邮箱..."
              class="pl-9"
              @keyup.enter="handleSearch"
            />
          </div>
          <Select v-model="roleFilter">
            <SelectTrigger class="w-full md:w-[180px]">
              <SelectValue placeholder="角色筛选" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部角色</SelectItem>
              <SelectItem value="admin">管理员</SelectItem>
              <SelectItem value="operator">操作员</SelectItem>
              <SelectItem value="viewer">查看者</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>

    <!-- Table -->
    <Card>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>角色</TableHead>
              <TableHead>状态</TableHead>
              <TableHead>创建时间</TableHead>
              <TableHead>最后登录</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in users" :key="user.id">
              <TableCell class="font-medium">{{ user.username }}</TableCell>
              <TableCell>{{ user.email }}</TableCell>
              <TableCell>
                <Badge :variant="getRoleBadge(user.role).variant">
                  {{ getRoleBadge(user.role).label }}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge :variant="getStatusBadge(user.status).variant">
                  {{ getStatusBadge(user.status).label }}
                </Badge>
              </TableCell>
              <TableCell>{{ user.createTime }}</TableCell>
              <TableCell>{{ user.lastLogin }}</TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8">
                      <MoreVertical class="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="editUser(user.id)">
                      <Pencil class="w-4 h-4 mr-2" />
                      编辑
                    </DropdownMenuItem>
                    <DropdownMenuItem @click="deleteUser(user.id)" class="text-destructive">
                      <Trash2 class="w-4 h-4 mr-2" />
                      删除
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
