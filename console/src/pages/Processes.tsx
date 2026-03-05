import { useEffect, useState } from 'react';
import { Table, Card, Button, Space, Tag, Typography, Input, Dropdown, Empty, App, Select, DatePicker } from 'antd';
import { EyeOutlined, SearchOutlined, ReloadOutlined, MoreOutlined, StopOutlined, SyncOutlined, ExportOutlined, PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import dayjs from 'dayjs';
import type { ColumnsType } from 'antd/es/table';
import type { MenuProps } from 'antd';
import { processesApi } from '../api';
import type { Process } from '../types';
import { formatDate, formatDuration, normalizeToMilliseconds } from '../utils/date';

const { Title, Text } = Typography;

const ProcessesPage = () => {
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Process[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [definitionFilter, setDefinitionFilter] = useState<string>('all');
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);

  const fetchData = async () => {
    setLoading(true);
    try {
      const response = await processesApi.list({ page, size: pageSize });
      setData(response.processes || []);
      setTotal(Number(response.total) || 0);
    } catch (error) {
      console.error('Failed to fetch processes:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [page, pageSize]);

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      Success: { color: 'success', text: '成功' },
      Failed: { color: 'error', text: '失败' },
      Running: { color: 'processing', text: '运行中' },
      Waiting: { color: 'default', text: '等待' },
    };
    const { color, text } = statusMap[status] || { color: 'default', text: status };
    return <Tag color={color}>{text}</Tag>;
  };

  const getStageTag = (stage: string) => {
    const stageMap: Record<string, { color: string; text: string }> = {
      Prepare: { color: 'warning', text: '准备中' },
      Ready: { color: 'processing', text: '就绪' },
      Commit: { color: 'success', text: '提交中' },
      Rollback: { color: 'error', text: '回滚中' },
      Destroy: { color: 'default', text: '销毁中' },
      Finish: { color: 'success', text: '已完成' },
    };
    const { color, text } = stageMap[stage] || { color: 'default', text: stage };
    return <Tag color={color}>{text}</Tag>;
  };

  const handleStop = async () => {
    message.info('停止功能开发中');
  };

  const handleRetry = async () => {
    message.info('重试功能开发中');
  };

  const filteredData = data.filter(item => {
    if (statusFilter !== 'all') {
      const status = item.status?.toLowerCase() || '';
      if (statusFilter === 'running' && status !== 'running') return false;
      if (statusFilter === 'success' && status !== 'success') return false;
      if (statusFilter === 'failed' && status !== 'failed') return false;
      if (statusFilter === 'waiting' && status !== 'waiting') return false;
    }
    if (searchText) {
      const name = item.name?.toLowerCase() || '';
      const uid = item.definitionsUid?.toLowerCase() || '';
      const id = String(item.id);
      const query = searchText.toLowerCase();
      if (!name.includes(query) && !uid.includes(query) && !id.includes(query)) return false;
    }
    if (definitionFilter !== 'all' && item.definitionsUid !== definitionFilter) {
      return false;
    }
    if (dateRange) {
      const ts = Number(item.startAt || 0);
      if (!ts) return false;
      const time = dayjs(normalizeToMilliseconds(ts));
      if (time.isBefore(dateRange[0], 'second') || time.isAfter(dateRange[1], 'second')) return false;
    }
    return true;
  });

  const definitionOptions = Array.from(new Set(data.map((item) => item.definitionsUid).filter((item): item is string => Boolean(item)))).map((item) => ({
    label: item,
    value: item,
  }));

  const getActionMenu = (record: Process): MenuProps => ({
    items: [
      {
        key: 'view',
        label: '查看详情',
        icon: <EyeOutlined />,
        onClick: () => navigate(`/processes/${record.id}`),
      },
      ...(record.status?.toLowerCase() === 'running' ? [{
        key: 'stop',
        label: '停止',
        icon: <StopOutlined />,
        onClick: () => handleStop(),
      }] : []),
      ...(record.status?.toLowerCase() === 'failed' ? [{
        key: 'retry',
        label: '重试',
        icon: <SyncOutlined />,
        onClick: () => handleRetry(),
      }] : []),
    ],
  });

  const columns: ColumnsType<Process> = [
    {
      title: '实例名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, record) => (
        <a onClick={() => navigate(`/processes/${record.id}`)}>
          {text || `Process #${record.id}`}
        </a>
      ),
    },
    {
      title: '流程定义',
      dataIndex: 'definitionsUid',
      key: 'definitionsUid',
      width: 200,
      render: (uid) => (
        <Text code style={{ fontSize: 12 }}>{uid || '-'}</Text>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status) => getStatusTag(status),
    },
    {
      title: '阶段',
      dataIndex: 'stage',
      key: 'stage',
      width: 100,
      render: (stage) => getStageTag(stage),
    },
    {
      title: '当前节点',
      key: 'node',
      width: 140,
      render: (_, record) => <Text>{record.stage || '任务执行中'}</Text>,
    },
    {
      title: '发起人',
      key: 'starter',
      width: 110,
      render: (_, record) => <Text type="secondary">{record.metadata?.creator || 'admin'}</Text>,
    },
    {
      title: '开始时间',
      dataIndex: 'startAt',
      key: 'startAt',
      width: 180,
      render: (timestamp) => <Text type="secondary">{formatDate(timestamp)}</Text>,
    },
    {
      title: '执行时长',
      key: 'duration',
      width: 100,
      render: (_, record) => (
        <Text type="secondary">{formatDuration(record.startAt, record.endAt)}</Text>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Dropdown menu={getActionMenu(record)}>
          <Button type="text" icon={<MoreOutlined />} />
        </Dropdown>
      ),
    },
  ];

  return (
    <div className="gflow-list-page">
      <div className="gflow-page-head">
        <Title level={4}>流程实例管理</Title>
        <Text type="secondary">控制台 &gt; 流程实例管理</Text>
      </div>

      <div className="gflow-toolbar">
        <Space>
          <Select
            style={{ width: 220 }}
            value={definitionFilter}
            onChange={setDefinitionFilter}
            options={[{ value: 'all', label: '选择流程定义' }, ...definitionOptions]}
          />
            <Input
              placeholder="流程名称 / 实例 ID"
              prefix={<SearchOutlined />}
              value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 240 }}
            allowClear
          />
          <DatePicker.RangePicker
            showTime
            value={dateRange}
            onChange={(value) => {
              if (!value || !value[0] || !value[1]) {
                setDateRange(null);
                return;
              }
              setDateRange([value[0], value[1]]);
            }}
          />
          <Dropdown
            menu={{
              items: [
                { key: 'all', label: '全部状态', onClick: () => setStatusFilter('all') },
                { key: 'running', label: '运行中', onClick: () => setStatusFilter('running') },
                { key: 'success', label: '成功', onClick: () => setStatusFilter('success') },
                { key: 'failed', label: '失败', onClick: () => setStatusFilter('failed') },
                { key: 'waiting', label: '等待', onClick: () => setStatusFilter('waiting') },
              ],
            }}
          >
            <Button>
              {statusFilter === 'all' ? '全部状态' : 
               statusFilter === 'running' ? '运行中' : 
               statusFilter === 'success' ? '成功' : 
               statusFilter === 'failed' ? '失败' : '等待'}
            </Button>
          </Dropdown>
        </Space>
        <Space>
          <Button icon={<ExportOutlined />}>导出数据</Button>
          <Button type="primary" icon={<PlusOutlined />}>发起流程</Button>
          <Button onClick={() => {
            setSearchText('');
            setStatusFilter('all');
            setDefinitionFilter('all');
            setDateRange(null);
          }}>
            重置
          </Button>
            <Button type="primary" icon={<ReloadOutlined />} onClick={fetchData}>
              查询
            </Button>
          </Space>
        </div>

      {loading ? (
        <Card loading className="gflow-glass-card" />
      ) : filteredData.length === 0 ? (
        <Card className="gflow-glass-card">
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="暂无流程实例"
          />
        </Card>
      ) : (
        <Card className="gflow-glass-card">
          <Table
            columns={columns}
            dataSource={filteredData}
            rowKey="id"
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

export default ProcessesPage;
