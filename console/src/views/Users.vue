<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Search, MoreVertical, Pencil, Trash2, Plus, Users, Loader2 } from 'lucide-vue-next'
import { http } from '@/lib/http'
import type { User, Role } from '@/types/api'
import Pagination from '@/components/Pagination.vue'

const users = ref<User[]>([])
const roles = ref<Role[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const searchQuery = ref('')

const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const editingUser = ref<User | null>(null)

const newUser = ref({
  username: '',
  password: '',
  email: '',
  description: '',
  roleId: 1,
})

const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value
  const query = searchQuery.value.toLowerCase()
  return users.value.filter(u => 
    u.username?.toLowerCase().includes(query) || 
    u.email?.toLowerCase().includes(query)
  )
})

const getRoleBadge = (roleId: number) => {
  const role = roles.value.find(r => r.id === roleId)
  if (roleId === 1) return { variant: 'success' as const, label: role?.displayName || '管理员' }
  if (roleId === 2) return { variant: 'info' as const, label: role?.displayName || '系统' }
  return { variant: 'secondary' as const, label: role?.displayName || '操作员' }
}

async function fetchUsers() {
  loading.value = true
  try {
    const response = await http.get<{ users: User[], total: number }>('/v1/admin/users', {
      page: currentPage.value,
      size: pageSize.value,
    })
    users.value = response.users || []
    total.value = Number(response.total) || 0
  } catch (error) {
    console.error('Failed to fetch users:', error)
  } finally {
    loading.value = false
  }
}

async function fetchRoles() {
  try {
    const response = await http.get<{ roles: Role[] }>('/v1/admin/roles')
    roles.value = response.roles || []
  } catch (error) {
    console.error('Failed to fetch roles:', error)
  }
}

function handleSearch() {
  currentPage.value = 1
  fetchUsers()
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchUsers()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  fetchUsers()
}

function openCreateDialog() {
  newUser.value = { username: '', password: '', email: '', description: '', roleId: 1 }
  showCreateDialog.value = true
}

function openEditDialog(user: User) {
  editingUser.value = { ...user }
  showEditDialog.value = true
}

async function createUser() {
  try {
    await http.post('/v1/admin/users', newUser.value)
    showCreateDialog.value = false
    await fetchUsers()
  } catch (error) {
    console.error('Failed to create user:', error)
  }
}

async function updateUser() {
  if (!editingUser.value) return
  try {
    await http.patch(`/v1/admin/users/${editingUser.value.id}`, {
      email: editingUser.value.email,
      description: editingUser.value.description,
      roleId: editingUser.value.roleId,
    })
    showEditDialog.value = false
    await fetchUsers()
  } catch (error) {
    console.error('Failed to update user:', error)
  }
}

async function deleteUser(id: number) {
  if (!confirm('确定要删除此用户吗？')) return
  try {
    await http.delete(`/v1/admin/users/${id}`)
    await fetchUsers()
  } catch (error) {
    console.error('Failed to delete user:', error)
  }
}

function formatDateTime(timestamp: number | string): string {
  if (!timestamp) return '-'
  const ts = typeof timestamp === 'string' ? Number(timestamp) : timestamp
  if (ts === 0) return '-'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchUsers()
  fetchRoles()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">用户管理</h1>
        <p class="text-muted-foreground mt-1">管理系统用户和权限</p>
      </div>
      <Button @click="openCreateDialog">
        <Plus class="w-4 h-4 mr-2" />
        创建用户
      </Button>
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
        </div>
      </CardContent>
    </Card>

    <!-- Loading state -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <Loader2 class="w-8 h-8 animate-spin text-muted-foreground" />
    </div>

    <!-- Table -->
    <Card v-else>
      <CardContent class="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>用户名</TableHead>
              <TableHead>邮箱</TableHead>
              <TableHead>角色</TableHead>
              <TableHead>描述</TableHead>
              <TableHead>创建时间</TableHead>
              <TableHead class="w-[80px]">操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="filteredUsers.length === 0">
              <TableCell colspan="6" class="text-center text-muted-foreground py-8">
                暂无用户
              </TableCell>
            </TableRow>
            <TableRow v-for="user in filteredUsers" :key="user.id">
              <TableCell>
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-muted flex items-center justify-center">
                    <Users class="w-4 h-4" />
                  </div>
                  <span class="font-medium">{{ user.username }}</span>
                </div>
              </TableCell>
              <TableCell>{{ user.email || '-' }}</TableCell>
              <TableCell>
                <Badge :variant="getRoleBadge(user.roleId).variant">
                  {{ getRoleBadge(user.roleId).label }}
                </Badge>
              </TableCell>
              <TableCell>{{ user.description || '-' }}</TableCell>
              <TableCell>{{ formatDateTime(user.createAt) }}</TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" size="icon" class="h-8 w-8">
                      <MoreVertical class="w-4 h-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem @click="openEditDialog(user)">
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

    <!-- Pagination -->
    <Pagination
      :page="currentPage"
      :page-size="pageSize"
      :total="total"
      @update:page="handlePageChange"
      @update:page-size="handlePageSizeChange"
    />

    <!-- Create Dialog -->
    <Dialog v-model:open="showCreateDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>创建用户</DialogTitle>
          <DialogDescription>创建新的系统用户</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="username">用户名</Label>
            <Input id="username" v-model="newUser.username" placeholder="请输入用户名" />
          </div>
          <div class="space-y-2">
            <Label for="password">密码</Label>
            <Input id="password" v-model="newUser.password" type="password" placeholder="请输入密码" />
          </div>
          <div class="space-y-2">
            <Label for="email">邮箱</Label>
            <Input id="email" v-model="newUser.email" type="email" placeholder="请输入邮箱" />
          </div>
          <div class="space-y-2">
            <Label for="description">描述</Label>
            <Input id="description" v-model="newUser.description" placeholder="请输入描述" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showCreateDialog = false">取消</Button>
          <Button @click="createUser">创建</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Edit Dialog -->
    <Dialog v-model:open="showEditDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑用户</DialogTitle>
          <DialogDescription>修改用户信息</DialogDescription>
        </DialogHeader>
        <div v-if="editingUser" class="space-y-4">
          <div class="space-y-2">
            <Label for="edit-email">邮箱</Label>
            <Input id="edit-email" v-model="editingUser.email" type="email" placeholder="请输入邮箱" />
          </div>
          <div class="space-y-2">
            <Label for="edit-description">描述</Label>
            <Input id="edit-description" v-model="editingUser.description" placeholder="请输入描述" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showEditDialog = false">取消</Button>
          <Button @click="updateUser">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
