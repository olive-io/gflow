import { useEffect, useMemo, useState } from 'react';
import { App, Button, Card, Input, Select, Space, Table, Tag, Typography } from 'antd';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { endpointsApi } from '../api';
import type { Endpoint } from '../types';
import { formatDate } from '../utils/date';

const { Title, Text } = Typography;

const Endpoints = () => {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<Endpoint[]>([]);
  const [searchText, setSearchText] = useState('');
  const [typeFilter, setTypeFilter] = useState('all');

  const fetchData = async () => {
    setLoading(true);
    try {
      const response = await endpointsApi.list({ page: 1, size: 200 });
      setData(response.endpoints || []);
    } catch (error) {
      console.error('Failed to fetch endpoints:', error);
      message.error('加载接口列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const endpointTypes = useMemo(() => {
    const set = new Set<string>();
    data.forEach((item) => {
      if (item.type) {
        set.add(item.type);
      }
    });
    return Array.from(set);
  }, [data]);

  const filteredData = data.filter((item) => {
    if (typeFilter !== 'all' && item.type !== typeFilter) {
      return false;
    }
    if (!searchText) {
      return true;
    }
    const query = searchText.toLowerCase();
    return [item.name, item.type, item.description, item.httpUrl]
      .filter(Boolean)
      .some((v) => String(v).toLowerCase().includes(query));
  });

  const columns: ColumnsType<Endpoint> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name) => <Text strong>{name || '-'}</Text>,
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type) => <Tag>{type || '-'}</Tag>,
    },
    {
      title: '任务类型',
      dataIndex: 'taskType',
      key: 'taskType',
      width: 120,
      render: (taskType) => <Text type="secondary">{String(taskType ?? '-')}</Text>,
    },
    {
      title: '地址',
      dataIndex: 'httpUrl',
      key: 'httpUrl',
      render: (httpUrl) => <Text type="secondary">{httpUrl || '-'}</Text>,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      render: (description) => <Text type="secondary">{description || '-'}</Text>,
    },
    {
      title: '更新时间',
      dataIndex: 'updateAt',
      key: 'updateAt',
      width: 180,
      render: (updateAt) => formatDate(updateAt),
    },
  ];

  return (
    <div className="gflow-list-page">
      <div className="gflow-page-head">
        <Title level={4}>系统配置</Title>
        <Text type="secondary">控制台 &gt; 系统配置</Text>
      </div>

      <div className="gflow-toolbar">
        <Space wrap>
          <Input
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            allowClear
            prefix={<SearchOutlined />}
            placeholder="搜索名称 / 类型 / 描述"
            style={{ width: 280 }}
          />
          <Select
            value={typeFilter}
            onChange={setTypeFilter}
            style={{ width: 160 }}
            options={[
              { value: 'all', label: '全部类型' },
              ...endpointTypes.map((item) => ({ value: item, label: item })),
            ]}
          />
        </Space>
        <Space>
          <Button
            onClick={() => {
              setSearchText('');
              setTypeFilter('all');
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
          locale={{ emptyText: '暂无接口数据' }}
          pagination={{ showSizeChanger: true, showQuickJumper: true, pageSize: 10 }}
        />
      </Card>
    </div>
  );
};

export default Endpoints;
