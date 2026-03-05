import { useCallback, useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Button,
  Card,
  Col,
  Empty,
  Space,
  Spin,
  Tag,
  Timeline,
  Typography,
  Row,
  message,
} from 'antd';
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  FullscreenOutlined,
} from '@ant-design/icons';
import BpmnViewer from 'bpmn-js/lib/NavigatedViewer';
import 'bpmn-js/dist/assets/diagram-js.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-codes.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-embedded.css';
import { definitionsApi, processesApi } from '../api';
import type { Definitions, Process } from '../types';
import { formatDate, formatDuration } from '../utils/date';

const { Text, Paragraph } = Typography;

interface Task {
  id: number;
  name: string;
  status: string;
  stage: string;
  startTime?: number | string;
  endTime?: number | string;
  message?: string;
}

const ProcessDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const [loading, setLoading] = useState(true);
  const [process, setProcess] = useState<Process | null>(null);
  const [definition, setDefinition] = useState<Definitions | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [bpmnLoading, setBpmnLoading] = useState(false);
  const [zoom, setZoom] = useState(1);

  const bpmnContainerRef = useRef<HTMLDivElement | null>(null);
  const bpmnViewerRef = useRef<BpmnViewer | null>(null);

  useEffect(() => {
    const fetchProcess = async () => {
      if (!id) {
        return;
      }
      setLoading(true);
      try {
        const response = await processesApi.get(Number(id));
        setProcess(response.process || null);

        if (response.process?.definitionsUid) {
          const defResponse = await definitionsApi.get(response.process.definitionsUid);
          setDefinition(defResponse.definitions || null);
        }

        const responseTasks = response.tasks as Task[] | undefined;
        const responseActivities = response.activities as Task[] | undefined;
        setTasks(responseTasks || responseActivities || []);
      } catch (error) {
        console.error('Failed to fetch process:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchProcess();
  }, [id]);

  const initBpmnViewer = useCallback(async () => {
    if (!bpmnContainerRef.current || !definition?.content) {
      return;
    }

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

      const canvas = viewer.get('canvas') as { zoom: (level: string | number) => void };
      canvas.zoom('fit-viewport');
      setZoom(1);
    } catch (error) {
      console.error('Failed to initialize BPMN viewer:', error);
    } finally {
      setBpmnLoading(false);
    }
  }, [definition?.content]);

  useEffect(() => {
    const timer = setTimeout(() => {
      initBpmnViewer();
    }, 120);

    return () => clearTimeout(timer);
  }, [initBpmnViewer]);

  useEffect(() => {
    return () => {
      if (bpmnViewerRef.current) {
        bpmnViewerRef.current.destroy();
        bpmnViewerRef.current = null;
      }
    };
  }, []);

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string; icon: ReactNode }> = {
      Success: { color: 'success', text: '已完成', icon: <CheckCircleOutlined /> },
      Failed: { color: 'error', text: '失败', icon: <CloseCircleOutlined /> },
      Running: { color: 'processing', text: '运行中', icon: <PlayCircleOutlined spin /> },
      Waiting: { color: 'default', text: '等待', icon: <ClockCircleOutlined /> },
    };

    const { color, text, icon } = statusMap[status] || {
      color: 'default',
      text: status || '未知',
      icon: <ClockCircleOutlined />,
    };

    return (
      <Tag color={color} icon={icon}>
        {text}
      </Tag>
    );
  };

  const handleZoomIn = () => {
    const canvas = bpmnViewerRef.current?.get('canvas') as { zoom: (level: number) => void } | undefined;
    if (!canvas) {
      return;
    }
    const next = Math.min(zoom + 0.2, 3);
    canvas.zoom(next);
    setZoom(next);
  };

  const handleZoomOut = () => {
    const canvas = bpmnViewerRef.current?.get('canvas') as { zoom: (level: number) => void } | undefined;
    if (!canvas) {
      return;
    }
    const next = Math.max(zoom - 0.2, 0.3);
    canvas.zoom(next);
    setZoom(next);
  };

  const handleFit = () => {
    const canvas = bpmnViewerRef.current?.get('canvas') as { zoom: (level: string) => void } | undefined;
    if (!canvas) {
      return;
    }
    canvas.zoom('fit-viewport');
    setZoom(1);
  };

  if (loading) {
    return <Card loading />;
  }

  if (!process) {
    return (
      <Card>
        <Empty description="流程实例不存在">
          <Button type="primary" onClick={() => navigate('/processes')}>
            返回列表
          </Button>
        </Empty>
      </Card>
    );
  }

  const timelineItems = (tasks.length > 0 ? tasks : [])
    .slice(0, 12)
    .map((task) => ({
      color:
        task.status?.toLowerCase() === 'success'
          ? 'green'
          : task.status?.toLowerCase() === 'failed'
            ? 'red'
            : task.status?.toLowerCase() === 'running'
              ? 'blue'
              : 'gray',
      children: (
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text strong>{task.name || `Task #${task.id}`}</Text>
            {getStatusTag(task.status)}
          </div>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {formatDate(task.startTime)}
          </Text>
          {task.message ? (
            <Paragraph type="secondary" style={{ marginBottom: 0, marginTop: 6, fontSize: 12 }}>
              {task.message}
            </Paragraph>
          ) : null}
        </div>
      ),
    }));

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <Card>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <Space>
            <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/processes')}>
              返回
            </Button>
            <Text type="secondary">流程实例 &gt; {process.id}</Text>
          </Space>
          <Space>
            <Button icon={<ReloadOutlined />}>刷新数据</Button>
            <Button type="primary">干预实例</Button>
            {process.status?.toLowerCase() === 'running' ? (
              <Button icon={<PauseCircleOutlined />} onClick={() => message.info('停止功能开发中')}>
                停止
              </Button>
            ) : null}
          </Space>
        </div>
      </Card>

      <Card>
        <Row gutter={[20, 12]}>
          <Col xs={24} md={12} xl={6}>
            <Text type="secondary">实例状态</Text>
            <div style={{ marginTop: 8 }}>{getStatusTag(process.status)}</div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              当前节点: {process.stage || '-'}
            </Text>
          </Col>
          <Col xs={24} md={12} xl={6}>
            <Text type="secondary">总执行时长</Text>
            <div style={{ marginTop: 8 }}>
              <Text strong style={{ fontSize: 20 }}>
                {formatDuration(process.startAt, process.endAt)}
              </Text>
            </div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              任务数量: {tasks.length}
            </Text>
          </Col>
          <Col xs={24} md={12} xl={6}>
            <Text type="secondary">启动时间</Text>
            <div style={{ marginTop: 8 }}>
              <Text strong>{formatDate(process.startAt)}</Text>
            </div>
          </Col>
          <Col xs={24} md={12} xl={6}>
            <Text type="secondary">最后更新</Text>
            <div style={{ marginTop: 8 }}>
              <Text strong>{formatDate(process.updateAt || process.endAt || process.startAt)}</Text>
            </div>
          </Col>
        </Row>
      </Card>

      <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0,1fr) 420px', gap: 16, minHeight: 560 }}>
        <Card styles={{ body: { padding: 0, height: '100%', display: 'flex', flexDirection: 'column' }}}>
          <div style={{ padding: 12, borderBottom: '1px solid #f1f5f9', display: 'flex', justifyContent: 'space-between' }}>
            <Text strong>可视化流程图 (Read-only)</Text>
            <Space size={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>已完成</Text>
              <Text type="secondary" style={{ fontSize: 12 }}>当前节点</Text>
              <Text type="secondary" style={{ fontSize: 12 }}>未开始</Text>
            </Space>
          </div>
          <div style={{ position: 'relative', flex: 1, minHeight: 460 }}>
            <div style={{ position: 'absolute', left: 12, top: 12, zIndex: 3 }}>
              <Space direction="vertical" size={6}>
                <Button icon={<ZoomInOutlined />} onClick={handleZoomIn} />
                <Button icon={<ZoomOutOutlined />} onClick={handleZoomOut} />
                <Button icon={<FullscreenOutlined />} onClick={handleFit} />
              </Space>
            </div>
            <div
              ref={bpmnContainerRef}
              style={{
                width: '100%',
                height: '100%',
                backgroundImage:
                  'radial-gradient(circle at 1px 1px, rgba(148,163,184,0.35) 1px, transparent 0)',
                backgroundSize: '20px 20px',
              }}
            />
            {bpmnLoading ? (
              <div
                style={{
                  position: 'absolute',
                  inset: 0,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: 'rgba(255,255,255,0.75)',
                }}
              >
                <Spin />
              </div>
            ) : null}
          </div>
        </Card>

        <Card
          title="执行历史"
          styles={{body: { display: 'flex', flexDirection: 'column', height: '100%', paddingBottom: 0 }}}
        >
          <div style={{ flex: 1, overflow: 'auto', paddingRight: 6 }}>
            {timelineItems.length > 0 ? <Timeline items={timelineItems} /> : <Empty description="暂无执行历史" />}
          </div>
          <div style={{ marginTop: 10, borderTop: '1px solid #f1f5f9', padding: '14px 0' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 10 }}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                实例ID
              </Text>
              <Text code>{process.id}</Text>
            </div>
            <Space style={{ width: '100%' }}>
              <Button block>导出报告</Button>
              <Button block danger>
                强行终止
              </Button>
            </Space>
          </div>
        </Card>
      </div>
    </div>
  );
};

export default ProcessDetail;
