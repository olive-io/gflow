import { useEffect, useState } from 'react';
import { App, Button, Card, Col, Input, Row, Select, Space, Table, Tag, Typography } from 'antd';
import { PlusOutlined, ReloadOutlined, SearchOutlined, CloudServerOutlined, CheckCircleOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { runnersApi } from '../api';
import type { Runner } from '../types';
import { formatDate } from '../utils/date';

const { Title, Text } = Typography;

const Runners = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Runner[]>([]);
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'online' | 'offline'>('all');

  const fetchData = async () => {
    setLoading(true);
    try {
      const response = await runnersApi.list({ page: 1, size: 100 });
      setData(response.runners || []);
    } catch (error) {
      console.error('Failed to fetch runners:', error);
      message.error('加载运行器失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const filteredData = data.filter((item) => {
    if (statusFilter === 'online' && !item.online) {
      return false;
    }
    if (statusFilter === 'offline' && item.online) {
      return false;
    }
    if (!searchText) {
      return true;
    }
    const query = searchText.toLowerCase();
    return [item.uid, item.hostname, item.listenUrl]
      .filter(Boolean)
      .some((v) => String(v).toLowerCase().includes(query));
  });

  const columns: ColumnsType<Runner> = [
    {
      title: 'Runner UID',
      dataIndex: 'uid',
      key: 'uid',
      render: (uid) => <Text code>{uid || '-'}</Text>,
    },
    {
      title: '主机',
      dataIndex: 'hostname',
      key: 'hostname',
      render: (hostname) => hostname || '-',
    },
    {
      title: '监听地址',
      dataIndex: 'listenUrl',
      key: 'listenUrl',
      render: (listenUrl) => <Text type="secondary">{listenUrl || '-'}</Text>,
    },
    {
      title: '状态',
      dataIndex: 'online',
      key: 'online',
      width: 110,
      render: (online) => <Tag color={online ? 'success' : 'default'}>{online ? '在线' : '离线'}</Tag>,
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 120,
      render: (version) => version || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'createAt',
      key: 'createAt',
      width: 180,
      render: (createAt) => formatDate(createAt),
    },
  ];

  return (
    <div className="gflow-list-page">
      <div className="gflow-page-head">
        <Title level={4}>组件管理</Title>
        <Text type="secondary">控制台 &gt; 组件管理</Text>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={24} md={8}>
          <Card className="gflow-stitch-kpi">
            <Text type="secondary">组件总数</Text>
            <div className="gflow-stitch-kpi-content">
              <CloudServerOutlined />
              <Title level={3}>{data.length}</Title>
            </div>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card className="gflow-stitch-kpi">
            <Text type="secondary">在线组件</Text>
            <div className="gflow-stitch-kpi-content">
              <CheckCircleOutlined />
              <Title level={3}>{data.filter((item) => item.online).length}</Title>
            </div>
          </Card>
        </Col>
        <Col xs={24} md={8}>
          <Card className="gflow-stitch-kpi">
            <Text type="secondary">离线/异常</Text>
            <div className="gflow-stitch-kpi-content">
              <ExclamationCircleOutlined />
              <Title level={3}>{data.filter((item) => !item.online).length}</Title>
            </div>
          </Card>
        </Col>
      </Row>

      <div className="gflow-toolbar">
        <Space wrap>
          <Input
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
            prefix={<SearchOutlined />}
            placeholder="搜索 UID / 主机 / 地址"
            style={{ width: 280 }}
          />
          <Select
            value={statusFilter}
            onChange={(value) => setStatusFilter(value)}
            style={{ width: 140 }}
            options={[
              { value: 'all', label: '全部状态' },
              { value: 'online', label: '在线' },
              { value: 'offline', label: '离线' },
            ]}
          />
        </Space>
        <Space>
          <Button icon={<PlusOutlined />} onClick={() => message.info('注册组件功能开发中')}>
            注册新组件
          </Button>
          <Button
            onClick={() => {
              setSearchText('');
              setStatusFilter('all');
            }}
          >
            重置
          </Button>
          <Button icon={<ReloadOutlined />} type="primary" onClick={fetchData}>
            查询
          </Button>
        </Space>
      </div>

      <Card className="gflow-glass-card">
        <Table
          rowKey="id"
          columns={columns}
          dataSource={filteredData}
          loading={loading}
          locale={{ emptyText: '暂无运行器数据' }}
          pagination={{ showSizeChanger: true, showQuickJumper: true, pageSize: 10 }}
        />
      </Card>
    </div>
  );
};

export default Runners;
