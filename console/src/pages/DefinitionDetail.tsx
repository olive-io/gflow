import { useEffect, useState, useRef, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Tabs, Table, Tag, Typography, Space, Row, Col, Spin, Segmented } from 'antd';
import { ArrowLeftOutlined, EditOutlined, FileTextOutlined, BranchesOutlined, ClockCircleOutlined, InfoCircleOutlined, CodeOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import BpmnViewer from 'bpmn-js/lib/NavigatedViewer';
import 'bpmn-js/dist/assets/diagram-js.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-codes.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-embedded.css';
import { definitionsApi, processesApi } from '../api';
import type { Definitions, Process } from '../types';
import { formatDate, formatDateShort } from '../utils/date';

const { Title, Text } = Typography;

const DefinitionDetail = () => {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [definition, setDefinition] = useState<Definitions | null>(null);
  const [processes, setProcesses] = useState<Process[]>([]);
  const [processesTotal, setProcessesTotal] = useState(0);
  const [processesLoading, setProcessesLoading] = useState(false);
  const [activeTab, setActiveTab] = useState('info');
  const [bpmnViewMode, setBpmnViewMode] = useState<'diagram' | 'xml'>('diagram');
  const [bpmnLoading, setBpmnLoading] = useState(false);

  const bpmnContainerRef = useRef<HTMLDivElement | null>(null);
  const bpmnViewerRef = useRef<BpmnViewer | null>(null);

  useEffect(() => {
    const fetchDefinition = async () => {
      if (!uid) return;
      setLoading(true);
      try {
        const response = await definitionsApi.get(uid);
        setDefinition(response.definitions || null);
      } catch (error) {
        console.error('Failed to fetch definition:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchDefinition();
  }, [uid]);

  useEffect(() => {
    const fetchProcesses = async () => {
      if (!definition?.uid) return;
      setProcessesLoading(true);
      try {
        const response = await processesApi.list({
          page: 1,
          size: 10,
          definitionsUid: definition.uid,
        });
        setProcesses(response.processes || []);
        setProcessesTotal(Number(response.total) || 0);
      } catch (error) {
        console.error('Failed to fetch processes:', error);
      } finally {
        setProcessesLoading(false);
      }
    };

    fetchProcesses();
  }, [definition?.uid]);

  const initBpmnViewer = useCallback(async () => {
    if (!bpmnContainerRef.current || !definition?.content) return;

    setBpmnLoading(true);
    try {
      if (bpmnViewerRef.current) {
        bpmnViewerRef.current.destroy();
      }

      const viewer = new BpmnViewer({
        container: bpmnContainerRef.current,
      });

      bpmnViewerRef.current = viewer;

      await viewer.importXML(definition.content);

      const canvas = viewer.get('canvas') as { zoom: (level: string) => void };
      canvas.zoom('fit-viewport');
    } catch (error) {
      console.error('Failed to initialize BPMN viewer:', error);
    } finally {
      setBpmnLoading(false);
    }
  }, [definition?.content]);

  useEffect(() => {
    if (activeTab === 'bpmn' && bpmnViewMode === 'diagram' && definition?.content) {
      const timer = setTimeout(() => {
        initBpmnViewer();
      }, 100);
      return () => clearTimeout(timer);
    }
  }, [activeTab, bpmnViewMode, definition?.content, initBpmnViewer]);

  useEffect(() => {
    return () => {
      if (bpmnViewerRef.current) {
        bpmnViewerRef.current.destroy();
        bpmnViewerRef.current = null;
      }
    };
  }, []);

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

  const getStageLabel = (stage: string) => {
    const stageMap: Record<string, string> = {
      Prepare: '准备中',
      Ready: '就绪',
      Commit: '提交中',
      Rollback: '回滚中',
      Destroy: '销毁中',
      Finish: '已完成',
    };
    return stageMap[stage] || stage;
  };

  const processColumns: ColumnsType<Process> = [
    {
      title: '实例 ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text) => text || '-',
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
      render: (stage) => <Tag>{getStageLabel(stage)}</Tag>,
    },
    {
      title: '开始时间',
      dataIndex: 'startAt',
      key: 'startAt',
      width: 180,
      render: (timestamp) => formatDate(timestamp),
    },
    {
      title: '结束时间',
      dataIndex: 'endAt',
      key: 'endAt',
      width: 180,
      render: (timestamp) => formatDate(timestamp),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, record) => (
        <Button type="link" onClick={() => navigate(`/processes/${record.id}`)}>
          查看
        </Button>
      ),
    },
  ];

  if (loading) {
    return <Card loading />;
  }

  if (!definition) {
    return (
      <Card>
        <Text>流程定义不存在</Text>
        <Button type="link" onClick={() => navigate('/definitions')}>
          返回列表
        </Button>
      </Card>
    );
  }

  const metadataList = definition.metadata
    ? Object.entries(definition.metadata)
        .filter(([key]) => key)
        .map(([key, value]) => ({ key, value }))
    : [];

  const tabItems = [
    {
      key: 'info',
      label: (
        <span>
          <InfoCircleOutlined style={{ marginRight: 8 }} />
          基本信息
        </span>
      ),
      children: (
        <Card>
          <Descriptions bordered column={2}>
            <Descriptions.Item label="流程 UID">
              <Text code>{definition.uid}</Text>
            </Descriptions.Item>
            <Descriptions.Item label="流程名称">
              {definition.name || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="版本号">
              <Tag>v{definition.version}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="部署状态">
              <Tag color={definition.isExecute ? 'success' : 'default'}>
                {definition.isExecute ? '已部署' : '未部署'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">{formatDate(definition.createAt)}</Descriptions.Item>
            <Descriptions.Item label="更新时间">{formatDate(definition.updateAt)}</Descriptions.Item>
            <Descriptions.Item label="描述" span={2}>
              {definition.description || '-'}
            </Descriptions.Item>
            {metadataList.length > 0 && (
              <Descriptions.Item label="元数据" span={2}>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                  {metadataList.map((item) => (
                    <Tag key={item.key}>
                      {item.key}: {String(item.value)}
                    </Tag>
                  ))}
                </div>
              </Descriptions.Item>
            )}
          </Descriptions>
        </Card>
      ),
    },
    {
      key: 'bpmn',
      label: (
        <span>
          <CodeOutlined style={{ marginRight: 8 }} />
          BPMN 内容
        </span>
      ),
      children: (
        <Card>
          <div style={{ marginBottom: 16 }}>
            <Segmented
              value={bpmnViewMode}
              onChange={(value) => setBpmnViewMode(value as 'diagram' | 'xml')}
              options={[
                { value: 'diagram', label: '图形视图' },
                { value: 'xml', label: 'XML 视图' },
              ]}
            />
          </div>
          {definition.content ? (
            <div
              style={{
                background: '#f5f5f5',
                borderRadius: 8,
                border: '1px solid #d9d9d9',
                overflow: 'hidden',
              }}
            >
              {bpmnViewMode === 'diagram' ? (
                <div style={{ position: 'relative', height: 500 }}>
                  {bpmnLoading && (
                    <div
                      style={{
                        position: 'absolute',
                        inset: 0,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        background: 'rgba(255,255,255,0.8)',
                        zIndex: 10,
                      }}
                    >
                      <Spin />
                    </div>
                  )}
                  <div ref={bpmnContainerRef} style={{ width: '100%', height: '100%', background: '#fff' }} />
                </div>
              ) : (
                <pre
                  style={{
                    padding: 16,
                    overflow: 'auto',
                    maxHeight: 500,
                    fontSize: 12,
                    fontFamily: 'monospace',
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-all',
                    margin: 0,
                  }}
                >
                  {definition.content}
                </pre>
              )}
            </div>
          ) : (
            <div style={{ textAlign: 'center', padding: 48, color: '#999' }}>
              暂无 BPMN 内容
            </div>
          )}
        </Card>
      ),
    },
    {
      key: 'processes',
      label: (
        <span>
          <BranchesOutlined style={{ marginRight: 8 }} />
          运行实例
          {processesTotal > 0 && (
            <Tag color="blue" style={{ marginLeft: 8 }}>
              {processesTotal}
            </Tag>
          )}
        </span>
      ),
      children: (
        <Card>
          <Table
            columns={processColumns}
            dataSource={processes}
            rowKey="id"
            loading={processesLoading}
            pagination={{
              total: processesTotal,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            locale={{ emptyText: '暂无运行实例' }}
          />
        </Card>
      ),
    },
  ];

  return (
    <div className="gflow-detail-page">
      <div className="gflow-detail-hero">
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/definitions')}>
            返回
          </Button>
          <div style={{ width: 1, height: 24, background: '#d9d9d9' }} />
          <div
            style={{
              width: 40,
              height: 40,
              borderRadius: 8,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: definition.isExecute ? 'rgba(22, 119, 255, 0.1)' : '#f5f5f5',
              color: definition.isExecute ? '#1677ff' : '#999',
            }}
          >
            <FileTextOutlined style={{ fontSize: 20 }} />
          </div>
          <div>
            <Title level={4} style={{ margin: 0 }}>
              {definition.name || definition.uid || '未命名流程'}
            </Title>
            <Text type="secondary" style={{ fontSize: 13 }}>
              {definition.description || '暂无描述'}
            </Text>
          </div>
        </Space>
        <Button type="primary" icon={<EditOutlined />} onClick={() => navigate(`/designer/${uid}`)}>
          编辑流程
        </Button>
      </div>

      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={12} sm={6}>
          <Card size="small">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  状态
                </Text>
                <div style={{ marginTop: 4 }}>
                  <Tag color={definition.isExecute ? 'success' : 'default'}>
                    {definition.isExecute ? '已部署' : '未部署'}
                  </Tag>
                </div>
              </div>
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: 8,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: definition.isExecute ? 'rgba(82, 196, 26, 0.1)' : '#f5f5f5',
                }}
              >
                <div
                  style={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    background: definition.isExecute ? '#52c41a' : '#d9d9d9',
                  }}
                />
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  版本
                </Text>
                <div style={{ marginTop: 4 }}>
                  <Text strong style={{ fontSize: 18 }}>
                    v{definition.version}
                  </Text>
                </div>
              </div>
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: 8,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'rgba(22, 119, 255, 0.1)',
                }}
              >
                <Text style={{ color: '#1677ff', fontWeight: 'bold' }}>v</Text>
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  运行实例
                </Text>
                <div style={{ marginTop: 4 }}>
                  <Text strong style={{ fontSize: 18 }}>
                    {processesTotal}
                  </Text>
                </div>
              </div>
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: 8,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'rgba(250, 140, 22, 0.1)',
                }}
              >
                <BranchesOutlined style={{ color: '#fa8c16' }} />
              </div>
            </div>
          </Card>
        </Col>
        <Col xs={12} sm={6}>
          <Card size="small">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>
                  更新时间
                </Text>
                <div style={{ marginTop: 4 }}>
                  <Text strong style={{ fontSize: 14 }}>
                    {formatDateShort(definition.updateAt)}
                  </Text>
                </div>
              </div>
              <div
                style={{
                  width: 32,
                  height: 32,
                  borderRadius: 8,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'rgba(114, 46, 209, 0.1)',
                }}
              >
                <ClockCircleOutlined style={{ color: '#722ed1' }} />
              </div>
            </div>
          </Card>
        </Col>
      </Row>

      <Card className="gflow-glass-card">
        <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />
      </Card>
    </div>
  );
};

export default DefinitionDetail;
