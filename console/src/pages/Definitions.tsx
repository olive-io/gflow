import { useEffect, useState } from 'react';
import { Table, Button, Space, Tag, Card, Typography, Popconfirm, Input, Tooltip, Select, Segmented, Empty, Row, Col, App } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, EyeOutlined, SearchOutlined, AppstoreOutlined, UnorderedListOutlined, FileTextOutlined, TeamOutlined, ShoppingOutlined, SolutionOutlined, NotificationOutlined, DollarOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { definitionsApi } from '../api';
import type { Definitions } from '../types';
import { formatDate } from '../utils/date';

const { Title, Text } = Typography;

const getIconForDefinition = (def: Definitions) => {
  const name = def.name?.toLowerCase() || '';
  if (name.includes('报销') || name.includes('费用')) return DollarOutlined;
  if (name.includes('入职') || name.includes('培训')) return TeamOutlined;
  if (name.includes('物资') || name.includes('领用')) return ShoppingOutlined;
  if (name.includes('合同')) return SolutionOutlined;
  if (name.includes('营销') || name.includes('活动')) return NotificationOutlined;
  return FileTextOutlined;
};

const DefinitionsPage = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Definitions[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(12);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('grid');

  const fetchData = async () => {
    setLoading(true);
    try {
      const response = await definitionsApi.list({ page, size: pageSize });
      setData(response.definitionsList || []);
      setTotal(Number(response.total) || 0);
    } catch (error) {
      console.error('Failed to fetch definitions:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize]);

  const handleDelete = async (uid: string) => {
    try {
      await definitionsApi.remove(uid);
      message.success('删除成功');
      fetchData();
    } catch {
      message.error('删除失败');
    }
  };

  const filteredData = data.filter(item => {
    if (statusFilter !== 'all') {
      if (statusFilter === 'active' && !item.isExecute) return false;
      if (statusFilter === 'inactive' && item.isExecute) return false;
    }
    if (searchText) {
      const name = item.name?.toLowerCase() || '';
      const uid = item.uid?.toLowerCase() || '';
      const query = searchText.toLowerCase();
      if (!name.includes(query) && !uid.includes(query)) return false;
    }
    if (categoryFilter !== 'all') {
      const text = `${item.name || ''}${item.description || ''}`.toLowerCase();
      if (categoryFilter === 'finance' && !text.includes('报销') && !text.includes('费用')) return false;
      if (categoryFilter === 'hr' && !text.includes('入职') && !text.includes('培训')) return false;
      if (categoryFilter === 'contract' && !text.includes('合同')) return false;
      if (categoryFilter === 'ops' && !text.includes('物资') && !text.includes('运维')) return false;
    }
    return true;
  });

  const columns: ColumnsType<Definitions> = [
    {
      title: '流程名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => {
        const IconComponent = getIconForDefinition(record);
        return (
          <Space>
            <div
              style={{
                width: 32,
                height: 32,
                borderRadius: 8,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                background: record.isExecute ? 'rgba(22, 119, 255, 0.1)' : '#f5f5f5',
                color: record.isExecute ? '#1677ff' : '#999',
              }}
            >
              <IconComponent />
            </div>
            <a onClick={() => navigate(`/definitions/${record.uid}`)}>
              {text || record.uid || '未命名流程'}
            </a>
          </Space>
        );
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text) => <Text type="secondary">{text || '暂无描述'}</Text>,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 80,
      render: (version) => <Tag>v{version}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'isExecute',
      key: 'isExecute',
      width: 100,
      render: (isExecute) => (
          <Tag color={isExecute ? 'success' : 'default'}>
            {isExecute ? '已发布' : '草稿'}
          </Tag>
        ),
    },
    {
      title: '创建时间',
      dataIndex: 'createAt',
      key: 'createAt',
      width: 180,
      render: (timestamp) => <Text type="secondary">{formatDate(timestamp)}</Text>,
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_, record) => (
        <Space>
          <Tooltip title="查看">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => navigate(`/definitions/${record.uid}`)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => navigate(`/designer/${record.uid}`)}
            />
          </Tooltip>
          <Popconfirm
            title="确定要删除此流程定义吗？"
            onConfirm={() => handleDelete(record.uid)}
            okText="确定"
            cancelText="取消"
          >
            <Tooltip title="删除">
              <Button type="text" danger icon={<DeleteOutlined />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const renderGridView = () => (
    <Row gutter={[16, 16]}>
      {filteredData.map((def) => {
        const IconComponent = getIconForDefinition(def);
        return (
          <Col xs={24} sm={12} md={8} lg={6} key={def.uid}>
            <Card
              hoverable
              style={{
                borderLeft: `4px solid ${def.isExecute ? '#1677ff' : '#d9d9d9'}`,
              }}
              styles={{ body: { padding: 16 } }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                <div
                  style={{
                    width: 48,
                    height: 48,
                    borderRadius: 12,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    background: def.isExecute ? 'rgba(22, 119, 255, 0.1)' : '#f5f5f5',
                    color: def.isExecute ? '#1677ff' : '#999',
                  }}
                >
                  <IconComponent style={{ fontSize: 24 }} />
                </div>
                 <Tag
                   color={def.isExecute ? 'success' : 'default'}
                   style={{ borderRadius: 12 }}
                 >
                   {def.isExecute ? '已发布' : '草稿'}
                 </Tag>
              </div>
              <Text strong style={{ fontSize: 16, display: 'block', marginBottom: 4 }}>
                {def.name || def.uid || '未命名流程'}
              </Text>
              <Text type="secondary" style={{ fontSize: 13, display: 'block', marginBottom: 12, minHeight: 40 }}>
                {def.description || '暂无描述'}
              </Text>
              <Row gutter={16} style={{ borderTop: '1px solid #f0f0f0', paddingTop: 12 }}>
                <Col span={12}>
                  <Text type="secondary" style={{ fontSize: 12 }}>版本</Text>
                  <div><Text strong>v{def.version}</Text></div>
                </Col>
                <Col span={12}>
                  <Text type="secondary" style={{ fontSize: 12 }}>创建时间</Text>
                  <div><Text strong>{formatDate(def.createAt)}</Text></div>
                </Col>
              </Row>
              <div style={{ display: 'flex', borderTop: '1px solid #f0f0f0', marginTop: 12 }}>
                <Button type="text" style={{ flex: 1 }} onClick={() => navigate(`/definitions/${def.uid}`)}>
                  发布
                </Button>
                <Button type="text" style={{ flex: 1 }} onClick={() => navigate(`/designer/${def.uid}`)}>
                  编辑
                </Button>
                <Popconfirm
                  title="确定要删除此流程定义吗？"
                  onConfirm={() => handleDelete(def.uid)}
                  okText="确定"
                  cancelText="取消"
                >
                  <Button type="text" danger style={{ flex: 1 }}>
                    删除
                  </Button>
                </Popconfirm>
              </div>
            </Card>
          </Col>
        );
      })}
      <Col xs={24} sm={12} md={8} lg={6}>
        <Card
          hoverable
          style={{
            border: '2px dashed #d9d9d9',
            height: '100%',
            minHeight: 220,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
          styles={{ body: { textAlign: 'center', width: '100%' } }}
          onClick={() => navigate('/designer/new')}
        >
          <div
            style={{
              width: 64,
              height: 64,
              borderRadius: '50%',
              background: '#f5f5f5',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 16px',
            }}
          >
            <PlusOutlined style={{ fontSize: 24, color: '#999' }} />
          </div>
          <Text strong>创建新流程</Text>
          <br />
          <Text type="secondary" style={{ fontSize: 12 }}>
            点击此处开始设计您的下一个自动化工作流
          </Text>
        </Card>
      </Col>
    </Row>
  );

  return (
    <div className="gflow-list-page">
      <div className="gflow-page-head">
        <Title level={4}>流程定义</Title>
        <Text type="secondary">管理和设计您的业务工作流定义</Text>
      </div>

      <div className="gflow-toolbar">
        <Space>
          <Input
            placeholder="搜索流程名称、编号..."
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 240 }}
            allowClear
          />
          <Select
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 140 }}
            options={[
              { value: 'all', label: '所有状态' },
               { value: 'active', label: '已发布' },
               { value: 'inactive', label: '草稿' },
            ]}
          />
          <Select
            value={categoryFilter}
            onChange={setCategoryFilter}
            style={{ width: 140 }}
            options={[
              { value: 'all', label: '流程分类' },
              { value: 'finance', label: '财务审批' },
              { value: 'hr', label: '人力资源' },
              { value: 'contract', label: '合同法务' },
              { value: 'ops', label: '行政事务' },
            ]}
          />
        </Space>
        <Space>
          <Segmented
            value={viewMode}
            onChange={(value) => setViewMode(value as 'list' | 'grid')}
            options={[
              { value: 'grid', icon: <AppstoreOutlined /> },
              { value: 'list', icon: <UnorderedListOutlined /> },
            ]}
            />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/designer/new')}>
            新建流程
          </Button>
        </Space>
      </div>

      {loading ? (
        <Card loading className="gflow-glass-card" />
      ) : filteredData.length === 0 ? (
        <Card className="gflow-glass-card">
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="暂无流程定义"
          >
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/designer/new')}>
              创建第一个流程
            </Button>
          </Empty>
        </Card>
      ) : viewMode === 'grid' ? (
        renderGridView()
      ) : (
        <Card className="gflow-glass-card">
          <Table
            columns={columns}
            dataSource={filteredData}
            rowKey="uid"
            loading={loading}
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条`,
              onChange: (page, pageSize) => {
                setPage(page);
                setPageSize(pageSize);
              },
            }}
          />
        </Card>
      )}
    </div>
  );
};

export default DefinitionsPage;
