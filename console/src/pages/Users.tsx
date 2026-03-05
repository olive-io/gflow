import { useEffect, useMemo, useState } from 'react';
import {
  App,
  Avatar,
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Switch,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd';
import { DeleteOutlined, EditOutlined, LockOutlined, PlusOutlined, SearchOutlined, UserOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { usersApi } from '../api';
import type { User, Role } from '../types';
import { formatDate } from '../utils/date';

const { Title, Text } = Typography;

interface CreateForm {
  username: string;
  password: string;
  email?: string;
  description?: string;
  roleId?: number;
}

interface EditForm {
  email?: string;
  description?: string;
  password?: string;
}

const Users = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [searchText, setSearchText] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');

  const [createOpen, setCreateOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);

  const [createForm] = Form.useForm<CreateForm>();
  const [editForm] = Form.useForm<EditForm>();

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const response = await usersApi.list({ page, size: pageSize });
      setUsers(response.users || []);
      setTotal(Number(response.total) || 0);
    } catch (error) {
      console.error('Failed to fetch users:', error);
      message.error('加载用户失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchRoles = async () => {
    try {
      const response = await usersApi.listRoles();
      setRoles(response.roles || []);
    } catch (error) {
      console.error('Failed to fetch roles:', error);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [page, pageSize]);

  useEffect(() => {
    fetchRoles();
  }, []);

  const filteredUsers = useMemo(() => {
    if (!searchText) {
      return users;
    }
    const query = searchText.toLowerCase();
    return users.filter((user) => {
      if (roleFilter !== 'all' && String(user.roleId || '') !== roleFilter) {
        return false;
      }
      const username = user.username?.toLowerCase() || '';
      const email = user.email?.toLowerCase() || '';
      const uid = user.uid?.toLowerCase() || '';
      return username.includes(query) || email.includes(query) || uid.includes(query);
    });
  }, [users, searchText, roleFilter]);

  const roleMap = useMemo(() => {
    const map = new Map<string, Role>();
    roles.forEach((item) => {
      map.set(String(item.id), item);
    });
    return map;
  }, [roles]);

  const roleOptions = useMemo(
    () =>
      roles.map((role) => ({
        value: Number(role.id),
        label: role.displayName || role.name,
      })),
    [roles],
  );

  const handleCreate = async () => {
    try {
      const values = await createForm.validateFields();
      setSubmitting(true);
      await usersApi.create(values);
      message.success('创建用户成功');
      setCreateOpen(false);
      createForm.resetFields();
      fetchUsers();
    } catch (error) {
      if (error instanceof Error && error.message) {
        console.error('Create user failed:', error);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const openEditModal = (user: User) => {
    setEditingUser(user);
    editForm.setFieldsValue({
      email: user.email,
      description: user.description,
      password: undefined,
    });
    setEditOpen(true);
  };

  const handleUpdate = async () => {
    if (!editingUser) {
      return;
    }
    try {
      const values = await editForm.validateFields();
      setSubmitting(true);
      const payload = {
        email: values.email,
        description: values.description,
        ...(values.password ? { password: values.password } : {}),
      };
      await usersApi.update(editingUser.id, payload);
      message.success('更新用户成功');
      setEditOpen(false);
      setEditingUser(null);
      fetchUsers();
    } catch (error) {
      if (error instanceof Error && error.message) {
        console.error('Update user failed:', error);
      }
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: User['id']) => {
    try {
      await usersApi.remove(id);
      message.success('删除用户成功');
      fetchUsers();
    } catch (error) {
      console.error('Delete user failed:', error);
      message.error('删除用户失败');
    }
  };

  const columns: ColumnsType<User> = [
    {
      title: '用户',
      dataIndex: 'username',
      key: 'username',
      render: (_, record) => (
        <Space>
          <Avatar icon={<UserOutlined />} />
          <div>
            <Text strong>{record.username || '-'}</Text>
            <div>
              <Text type="secondary" style={{ fontSize: 12 }}>
                {record.uid || '-'}
              </Text>
            </div>
          </div>
        </Space>
      ),
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      key: 'email',
      render: (email) => <Text type="secondary">{email || '-'}</Text>,
    },
    {
      title: '角色',
      dataIndex: 'roleId',
      key: 'roleId',
      width: 140,
      render: (roleId) => {
        const role = roleMap.get(String(roleId));
        const roleText = (role?.displayName || role?.name || 'USER').toUpperCase();
        return <Tag color={roleText.includes('ADMIN') ? 'purple' : 'blue'}>{roleText}</Tag>;
      },
    },
    {
      title: '状态',
      key: 'status',
      width: 110,
      render: () => <Switch size="small" defaultChecked />,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      render: (description) => <Text type="secondary">{description || '-'}</Text>,
    },
    {
      title: '创建时间',
      dataIndex: 'createAt',
      key: 'createAt',
      width: 180,
      render: (timestamp) => formatDate(timestamp),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Space>
          <Button type="text" icon={<EditOutlined />} onClick={() => openEditModal(record)} />
          <Popconfirm title="确定删除该用户吗？" okText="确定" cancelText="取消" onConfirm={() => handleDelete(record.id)}>
            <Button type="text" danger icon={<DeleteOutlined />} />
          </Popconfirm>
          <Button type="text" icon={<LockOutlined />} onClick={() => message.info('重置密码功能开发中')} />
        </Space>
      ),
    },
  ];

  return (
    <div className="gflow-list-page">
      <div className="gflow-page-head">
        <Title level={4}>用户管理</Title>
        <Text type="secondary">控制台 &gt; 用户管理</Text>
      </div>

      <div className="gflow-toolbar">
        <Space wrap>
            <Input
              prefix={<SearchOutlined />}
              placeholder="搜索姓名、邮箱、账号..."
            style={{ width: 280 }}
            allowClear
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
          />
          <Select
            value={roleFilter}
            onChange={setRoleFilter}
            style={{ width: 160 }}
            options={[
              { value: 'all', label: '全部角色' },
              ...roles.map((role) => ({ value: String(role.id), label: role.displayName || role.name })),
            ]}
          />
        </Space>
        <Space>
          <Button
            onClick={() => {
              setSearchText('');
              setRoleFilter('all');
            }}
          >
            重置
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              createForm.resetFields();
              setCreateOpen(true);
            }}
          >
            创建用户
          </Button>
        </Space>
      </div>

      <Card className="gflow-glass-card">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={filteredUsers}
          loading={loading}
          locale={{ emptyText: '暂无用户数据' }}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (value) => `共 ${value} 条`,
            onChange: (nextPage, nextSize) => {
              setPage(nextPage);
              setPageSize(nextSize);
            },
          }}
        />
      </Card>

      <Modal
        title="创建用户"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        confirmLoading={submitting}
        okText="创建"
        cancelText="取消"
      >
        <Form form={createForm} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
            <Input placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="roleId" label="角色">
            <Select placeholder="请选择角色" options={roleOptions} allowClear />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="请输入描述" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑用户"
        open={editOpen}
        onCancel={() => {
          setEditOpen(false);
          setEditingUser(null);
        }}
        onOk={handleUpdate}
        confirmLoading={submitting}
        okText="保存"
        cancelText="取消"
      >
        <Form form={editForm} layout="vertical">
          <Form.Item label="用户名">
            <Input value={editingUser?.username} disabled />
          </Form.Item>
          <Form.Item name="email" label="邮箱">
            <Input placeholder="请输入邮箱" />
          </Form.Item>
          <Form.Item name="password" label="重置密码">
            <Input.Password placeholder="不修改请留空" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="请输入描述" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Users;
