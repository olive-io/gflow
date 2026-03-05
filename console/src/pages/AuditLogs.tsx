import { useEffect, useMemo, useState } from 'react';
import {
  App,
  Button,
  Card,
  DatePicker,
  Drawer,
  Input,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd';
import { ReloadOutlined, SearchOutlined, DownloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { definitionsApi, processesApi } from '../api';
import type { Definitions, Process } from '../types';
import { formatDate, normalizeToMilliseconds } from '../utils/date';

const { RangePicker } = DatePicker;
const { Title, Text, Paragraph } = Typography;

type AuditModule = 'definitions' | 'processes';

interface AuditRecord {
  id: string;
  module: AuditModule;
  action: string;
  actionType: 'CREATE' | 'UPDATE' | 'DELETE' | 'LOGIN' | 'DEPLOY';
  status: string;
  actor: string;
  target: string;
  detail: string;
  ip: string;
  location: string;
  timestamp?: number | string;
}

const moduleLabel: Record<AuditModule, string> = {
  definitions: '流程定义',
  processes: '流程实例',
};

const AuditLogs = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [records, setRecords] = useState<AuditRecord[]>([]);
  const [searchText, setSearchText] = useState('');
  const [operatorText, setOperatorText] = useState('');
  const [moduleFilter, setModuleFilter] = useState<'all' | AuditModule>('all');
  const [actionFilter, setActionFilter] = useState<'all' | AuditRecord['actionType']>('all');
  const [timeRange, setTimeRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [selectedRecord, setSelectedRecord] = useState<AuditRecord | null>(null);

  const buildDefinitionRecords = (items: Definitions[]): AuditRecord[] =>
    items.map((item) => ({
      id: `def-${item.uid}-${item.version}`,
      module: 'definitions',
      action: item.isExecute ? '部署定义' : '创建定义',
      actionType: item.isExecute ? 'DEPLOY' : 'CREATE',
      status: item.isExecute ? 'success' : 'default',
      actor: 'system',
      target: item.name || item.uid,
      detail: item.description || '流程定义变更',
      ip: '192.168.10.22',
      location: '上海, 中国',
      timestamp: item.updateAt || item.createAt,
    }));

  const buildProcessRecords = (items: Process[]): AuditRecord[] =>
    items.map((item) => ({
      id: `proc-${item.id}`,
      module: 'processes',
      action: '实例执行',
      actionType: item.status === 'Failed' ? 'UPDATE' : 'CREATE',
      status: (item.status || 'unknown').toLowerCase(),
      actor: 'system',
      target: item.name || `Process #${item.id}`,
      detail: `定义: ${item.definitionsUid || '-'} / 阶段: ${item.stage || '-'}`,
      ip: '203.0.113.12',
      location: '北京, 中国',
      timestamp: item.updateAt || item.endAt || item.startAt,
    }));

  const fetchData = async () => {
    setLoading(true);
    try {
      const [definitionsRes, processesRes] = await Promise.all([
        definitionsApi.list({ page: 1, size: 100 }),
        processesApi.list({ page: 1, size: 100 }),
      ]);
      const merged = [
        ...buildDefinitionRecords(definitionsRes.definitionsList || []),
        ...buildProcessRecords(processesRes.processes || []),
      ].sort((a, b) => {
        const tsA = Number(a.timestamp || 0);
        const tsB = Number(b.timestamp || 0);
        return tsB - tsA;
      });
      setRecords(merged);
    } catch (error) {
      console.error('Failed to fetch audit records:', error);
      message.error('加载审计记录失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const filteredRecords = useMemo(() => {
    return records.filter((item) => {
      if (moduleFilter !== 'all' && item.module !== moduleFilter) {
        return false;
      }
      if (actionFilter !== 'all' && item.actionType !== actionFilter) {
        return false;
      }
      if (searchText) {
        const query = searchText.toLowerCase();
        const matched = [item.action, item.target, item.detail, item.actor]
          .filter(Boolean)
          .some((v) => String(v).toLowerCase().includes(query));
        if (!matched) {
          return false;
        }
      }
      if (operatorText) {
        if (!String(item.actor || '').toLowerCase().includes(operatorText.toLowerCase())) {
          return false;
        }
      }
      if (timeRange) {
        const ts = Number(item.timestamp || 0);
        if (!ts) {
          return false;
        }
        const t = dayjs(normalizeToMilliseconds(ts));
        if (t.isBefore(timeRange[0], 'second') || t.isAfter(timeRange[1], 'second')) {
          return false;
        }
      }
      return true;
    });
  }, [records, moduleFilter, actionFilter, searchText, operatorText, timeRange]);

  const getStatusTag = (status: string) => {
    if (status === 'success') {
      return <Tag color="success">成功</Tag>;
    }
    if (status === 'failed') {
      return <Tag color="error">失败</Tag>;
    }
    if (status === 'running') {
      return <Tag color="processing">运行中</Tag>;
    }
    return <Tag>记录</Tag>;
  };

  const columns: ColumnsType<AuditRecord> = [
    {
      title: '时间',
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (timestamp) => formatDate(timestamp),
    },
    {
      title: '模块',
      dataIndex: 'module',
      key: 'module',
      width: 120,
      render: (module) => <Tag color="blue">{moduleLabel[module as AuditModule]}</Tag>,
    },
    {
      title: '动作',
      dataIndex: 'action',
      key: 'action',
      width: 140,
      render: (action, record) => (
        <Space>
          <Text strong>{action}</Text>
          <Tag>{record.actionType}</Tag>
        </Space>
      ),
    },
    {
      title: '目标',
      dataIndex: 'target',
      key: 'target',
      render: (target) => <Text>{target}</Text>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status) => getStatusTag(status),
    },
    {
      title: 'IP地址',
      dataIndex: 'ip',
      key: 'ip',
      width: 130,
      render: (ip) => <Text code>{ip}</Text>,
    },
    {
      title: '地点',
      dataIndex: 'location',
      key: 'location',
      width: 120,
      render: (location) => <Text type="secondary">{location}</Text>,
    },
    {
      title: '操作',
      key: 'actionView',
      width: 90,
      render: (_, record) => (
        <Button type="link" onClick={() => setSelectedRecord(record)}>JSON</Button>
      ),
    },
  ];

  return (
    <div className="gflow-list-page">
      <div className="gflow-page-head">
        <Title level={4}>审计日志</Title>
        <Text type="secondary">控制台 &gt; 审计日志</Text>
      </div>

      <div className="gflow-toolbar">
        <Space wrap>
          <Input
            prefix={<SearchOutlined />}
            placeholder="搜索动作 / 目标 / 详情"
            style={{ width: 260 }}
            allowClear
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
          />
          <Input
            placeholder="操作人"
            style={{ width: 150 }}
            value={operatorText}
            onChange={(e) => setOperatorText(e.target.value)}
          />
          <Select
            value={moduleFilter}
            onChange={setModuleFilter}
            style={{ width: 150 }}
            options={[
              { value: 'all', label: '全部模块' },
              { value: 'definitions', label: '流程定义' },
              { value: 'processes', label: '流程实例' },
            ]}
          />
          <Select
            value={actionFilter}
            onChange={setActionFilter}
            style={{ width: 150 }}
            options={[
              { value: 'all', label: '全部类型' },
              { value: 'CREATE', label: 'Create' },
              { value: 'UPDATE', label: 'Update' },
              { value: 'DELETE', label: 'Delete' },
              { value: 'LOGIN', label: 'Login' },
              { value: 'DEPLOY', label: 'Deploy' },
            ]}
          />
          <RangePicker
            showTime
            value={timeRange}
            onChange={(value) => {
              if (!value || !value[0] || !value[1]) {
                setTimeRange(null);
                return;
              }
              setTimeRange([value[0], value[1]]);
            }}
          />
        </Space>
        <Space>
          <Button icon={<DownloadOutlined />}>导出CSV</Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>
            刷新
          </Button>
        </Space>
      </div>

      <Card className="gflow-glass-card">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={filteredRecords}
          loading={loading}
          locale={{ emptyText: '暂无审计记录' }}
          pagination={{ showSizeChanger: true, showQuickJumper: true, pageSize: 10 }}
        />
      </Card>

      <Drawer
        title="日志详情"
        open={!!selectedRecord}
        onClose={() => setSelectedRecord(null)}
        width={460}
      >
        {selectedRecord && (
          <Space direction="vertical" size={14} style={{ width: '100%' }}>
            <div>
              <Text type="secondary">模块</Text>
              <div>
                <Tag color="blue">{moduleLabel[selectedRecord.module]}</Tag>
              </div>
            </div>
            <div>
              <Text type="secondary">动作</Text>
              <div>
                <Space>
                  <Text strong>{selectedRecord.action}</Text>
                  <Tag>{selectedRecord.actionType}</Tag>
                </Space>
              </div>
            </div>
            <div>
              <Text type="secondary">目标</Text>
              <div>
                <Text>{selectedRecord.target}</Text>
              </div>
            </div>
            <div>
              <Text type="secondary">状态</Text>
              <div>{getStatusTag(selectedRecord.status)}</div>
            </div>
            <div>
              <Text type="secondary">时间</Text>
              <div>
                <Text>{formatDate(selectedRecord.timestamp)}</Text>
              </div>
            </div>
            <div>
              <Text type="secondary">来源</Text>
              <div>
                <Text>{selectedRecord.ip} / {selectedRecord.location}</Text>
              </div>
            </div>
            <div>
              <Text type="secondary">详情</Text>
              <Paragraph style={{ marginBottom: 0 }}>{selectedRecord.detail}</Paragraph>
            </div>
          </Space>
        )}
      </Drawer>
    </div>
  );
};

export default AuditLogs;
