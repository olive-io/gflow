import { useEffect, useRef, useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Input, Select, Space, App, theme, Spin, Card, Typography, Divider } from 'antd';
import {
  AppstoreOutlined,
  SaveOutlined,
  PlayCircleOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  FullscreenOutlined,
  EyeOutlined,
  ExportOutlined,
} from '@ant-design/icons';
import BpmnModeler from 'bpmn-js/lib/Modeler';
import 'bpmn-js/dist/assets/diagram-js.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-codes.css';
import 'bpmn-js/dist/assets/bpmn-font/css/bpmn-embedded.css';
import { definitionsApi, endpointsApi } from '../api';
import { useAppStore } from '../stores';
import type { Definitions, Endpoint } from '../types';

const { Text } = Typography;

const defaultDiagram = `<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
  xmlns:bpmndi="http://www.omg.org/spec/BPMN/20100524/DI"
  xmlns:dc="http://www.omg.org/spec/DD/20100524/DC"
  xmlns:di="http://www.omg.org/spec/DD/20100524/DI"
  id="Definitions_1"
  targetNamespace="http://bpmn.io/schema/bpmn">
  <bpmn:process id="Process_1" isExecutable="true">
    <bpmn:startEvent id="StartEvent_1" name="开始">
      <bpmn:outgoing>Flow_1</bpmn:outgoing>
    </bpmn:startEvent>
    <bpmn:task id="Task_1" name="任务">
      <bpmn:incoming>Flow_1</bpmn:incoming>
      <bpmn:outgoing>Flow_2</bpmn:outgoing>
    </bpmn:task>
    <bpmn:endEvent id="EndEvent_1" name="结束">
      <bpmn:incoming>Flow_2</bpmn:incoming>
    </bpmn:endEvent>
    <bpmn:sequenceFlow id="Flow_1" sourceRef="StartEvent_1" targetRef="Task_1" />
    <bpmn:sequenceFlow id="Flow_2" sourceRef="Task_1" targetRef="EndEvent_1" />
  </bpmn:process>
  <bpmndi:BPMNDiagram id="BPMNDiagram_1">
    <bpmndi:BPMNPlane id="BPMNPlane_1" bpmnElement="Process_1">
      <bpmndi:BPMNShape id="StartEvent_1_di" bpmnElement="StartEvent_1">
        <dc:Bounds x="152" y="102" width="36" height="36" />
        <bpmndi:BPMNLabel>
          <dc:Bounds x="159" y="145" width="22" height="14" />
        </bpmndi:BPMNLabel>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNShape id="Task_1_di" bpmnElement="Task_1">
        <dc:Bounds x="240" y="80" width="100" height="80" />
      </bpmndi:BPMNShape>
      <bpmndi:BPMNShape id="EndEvent_1_di" bpmnElement="EndEvent_1">
        <dc:Bounds x="402" y="102" width="36" height="36" />
        <bpmndi:BPMNLabel>
          <dc:Bounds x="409" y="145" width="22" height="14" />
        </bpmndi:BPMNLabel>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNEdge id="Flow_1_di" bpmnElement="Flow_1">
        <di:waypoint x="188" y="120" />
        <di:waypoint x="240" y="120" />
      </bpmndi:BPMNEdge>
      <bpmndi:BPMNEdge id="Flow_2_di" bpmnElement="Flow_2">
        <di:waypoint x="340" y="120" />
        <di:waypoint x="402" y="120" />
      </bpmndi:BPMNEdge>
    </bpmndi:BPMNPlane>
  </bpmndi:BPMNDiagram>
</bpmn:definitions>`;

interface SelectedElement {
  id: string;
  type: string;
  businessObject?: {
    name?: string;
    $attrs?: Record<string, string>;
  };
}

const Designer = () => {
  const { uid } = useParams<{ uid: string }>();
  const navigate = useNavigate();
  const { token } = theme.useToken();
  const { message: messageApi } = App.useApp();
  const { theme: appTheme } = useAppStore();

  const canvasRef = useRef<HTMLDivElement>(null);
  const modelerRef = useRef<BpmnModeler | null>(null);
  const selectedElementIdRef = useRef<string | null>(null);

  const isDark = useMemo(() => {
    if (appTheme === 'system') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return appTheme === 'dark';
  }, [appTheme]);

  const [isModelerReady, setIsModelerReady] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [deploying, setDeploying] = useState(false);
  const [zoom, setZoom] = useState(1);

  const [definition, setDefinition] = useState<Definitions | null>(null);
  const [processName, setProcessName] = useState('');
  const [processKey, setProcessKey] = useState('');
  const [processDescription, setProcessDescription] = useState('');

  const [selectedElement, setSelectedElement] = useState<SelectedElement | null>(null);
  const [selectedElementName, setSelectedElementName] = useState('');
  const [endpoints, setEndpoints] = useState<Endpoint[]>([]);
  const [loadingEndpoints, setLoadingEndpoints] = useState(false);
  const [selectedEndpointId, setSelectedEndpointId] = useState<string | null>(null);

  const isNew = !uid || uid === 'new';

  const loadDefinition = useCallback(async () => {
    if (isNew) {
      return null;
    }

    try {
      const response = await definitionsApi.get(uid!);
      setDefinition(response.definitions);
      setProcessName(response.definitions.name || '');
      setProcessKey(response.definitions.uid || '');
      setProcessDescription(response.definitions.description || '');
      return response.definitions.content;
    } catch (error) {
      messageApi.error('加载流程定义失败');
      return null;
    }
  }, [uid, isNew, messageApi]);

  const loadEndpoints = useCallback(async () => {
    setLoadingEndpoints(true);
    try {
      const response = await endpointsApi.list({ size: 100 });
      setEndpoints(response.endpoints || []);
    } catch (error) {
      console.error('Failed to load endpoints:', error);
    } finally {
      setLoadingEndpoints(false);
    }
  }, []);

  const destroyModeler = useCallback(() => {
    if (modelerRef.current) {
      modelerRef.current.destroy();
      modelerRef.current = null;
    }
    selectedElementIdRef.current = null;
    setSelectedElement(null);
    setSelectedElementName('');
    setSelectedEndpointId(null);
    setIsModelerReady(false);
  }, []);

  const initModeler = useCallback((content?: string | null) => {
    if (!canvasRef.current) {
      setLoading(false);
      return;
    }

    destroyModeler();

    const modeler = new BpmnModeler({
      container: canvasRef.current,
    });

    modelerRef.current = modeler;

    const diagram = content || defaultDiagram;

    modeler.importXML(diagram).then(() => {
      const canvas = modeler.get('canvas') as { zoom: (level: string) => void };
      canvas.zoom('fit-viewport');
      setIsModelerReady(true);
    }).catch((error: Error) => {
      console.error('Failed to import XML:', error);
      messageApi.error('加载流程图失败');
    }).finally(() => {
      setLoading(false);
    });

    modeler.on('selection.changed', (e: { newSelection: SelectedElement[] }) => {
      const element = e.newSelection[0];
      selectedElementIdRef.current = element?.id || null;
      setSelectedElement(element || null);
      setSelectedElementName(element?.businessObject?.name || '');

      if (element?.businessObject?.$attrs?.['gflow:endpoint']) {
        setSelectedEndpointId(element.businessObject.$attrs['gflow:endpoint']);
      } else {
        setSelectedEndpointId(null);
      }
    });

    modeler.on('element.changed', (e: { element: SelectedElement }) => {
      const { element } = e;
      if (!element) {
        return;
      }
      if (selectedElementIdRef.current !== element.id) {
        return;
      }

      setSelectedElement(element);
      setSelectedElementName(element.businessObject?.name || '');

      const endpointId = element.businessObject?.$attrs?.['gflow:endpoint'];
      setSelectedEndpointId(endpointId || null);
    });
  }, [destroyModeler, messageApi]);

  useEffect(() => {
    let active = true;

    const init = async () => {
      setLoading(true);
      const content = await loadDefinition();
      if (!active) {
        return;
      }
      loadEndpoints();
      setTimeout(() => {
        if (active) {
          initModeler(content);
        }
      }, 100);
    };

    init();

    return () => {
      active = false;
      destroyModeler();
    };
  }, [loadDefinition, initModeler, loadEndpoints, destroyModeler]);

  const handleSave = async () => {
    if (!modelerRef.current) return;

    setSaving(true);
    try {
      const { xml } = await modelerRef.current.saveXML({ format: true });
      if (xml) {
        await definitionsApi.deploy({
          content: xml,
          description: processDescription,
        });
        messageApi.success('保存成功');
      }
    } catch (error) {
      messageApi.error('保存失败');
    } finally {
      setSaving(false);
    }
  };

  const handleDeploy = async () => {
    if (!modelerRef.current) return;

    setDeploying(true);
    try {
      const { xml } = await modelerRef.current.saveXML({ format: true });
      if (xml) {
        await definitionsApi.deploy({
          content: xml,
          description: processDescription,
        });
        messageApi.success('部署成功');
        navigate('/definitions');
      }
    } catch (error) {
      messageApi.error('部署失败');
    } finally {
      setDeploying(false);
    }
  };

  const handleZoomIn = () => {
    const canvas = modelerRef.current?.get('canvas') as { zoom: (level: number) => void } | undefined;
    if (canvas) {
      const newZoom = Math.min(zoom * 1.2, 4);
      setZoom(newZoom);
      canvas.zoom(newZoom);
    }
  };

  const handleZoomOut = () => {
    const canvas = modelerRef.current?.get('canvas') as { zoom: (level: number) => void } | undefined;
    if (canvas) {
      const newZoom = Math.max(zoom / 1.2, 0.2);
      setZoom(newZoom);
      canvas.zoom(newZoom);
    }
  };

  const handleFitViewport = () => {
    const canvas = modelerRef.current?.get('canvas') as { zoom: (level: string) => void } | undefined;
    canvas?.zoom('fit-viewport');
    setZoom(1);
  };

  const handleEndpointChange = (endpointId?: string) => {
    if (!selectedElement || !modelerRef.current) return;

    const modeling = modelerRef.current.get('modeling') as {
      updateProperties: (element: unknown, props: Record<string, string>) => void;
    };

    if (modeling) {
      if (endpointId) {
        const endpoint = endpoints.find(e => String(e.id) === endpointId);
        if (endpoint) {
          modeling.updateProperties(selectedElement, {
            'gflow:endpoint': endpointId,
            name: endpoint.name || selectedElement.businessObject?.name || '',
          });
          setSelectedEndpointId(endpointId);
          setSelectedElementName(endpoint.name || selectedElement.businessObject?.name || '');
        }
      } else {
        modeling.updateProperties(selectedElement, {
          'gflow:endpoint': '',
        });
        setSelectedEndpointId(null);
      }
    }
  };

  const handleElementNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!selectedElement || !modelerRef.current) return;

    const name = e.target.value;
    setSelectedElementName(name);

    const modeling = modelerRef.current.get('modeling') as {
      updateProperties: (element: unknown, props: Record<string, string>) => void;
    };

    if (modeling) {
      modeling.updateProperties(selectedElement, {
        name,
      });
    }
  };

  const isTaskElement = selectedElement && [
    'bpmn:task',
    'bpmn:serviceTask',
    'bpmn:sendTask',
    'bpmn:receiveTask',
    'bpmn:userTask',
    'bpmn:scriptTask',
    'bpmn:manualTask',
    'bpmn:businessRuleTask',
    'bpmn:callActivity',
  ].includes(selectedElement.type);

  const getTaskTypeLabel = (taskType: number | string) => {
    if (typeof taskType === 'string') return taskType;
    const types: Record<number, string> = {
      11: 'Task',
      12: 'SendTask',
      13: 'ReceiveTask',
      14: 'ServiceTask',
      15: 'UserTask',
      16: 'ScriptTask',
      17: 'ManualTask',
      18: 'CallActivity',
      19: 'BusinessRuleTask',
    };
    return types[taskType] || 'Unknown';
  };

  const toolbarBtnStyle = {
    borderColor: token.colorBorder,
    background: token.colorBgContainer,
    color: token.colorText,
  };

  const actionDisabled = !isModelerReady || loading;

  return (
    <div
      className={`gflow-designer-page ${isDark ? 'dark' : ''}`}
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 50,
        background: token.colorBgLayout,
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <div className="gflow-designer-toolbar" style={{
        background: token.colorBgContainer,
        borderBottom: `1px solid ${token.colorBorder}`,
        padding: '0 16px',
        height: 56,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        flexShrink: 0,
      }}>
        <Space>
          <div
            style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}
            onClick={() => navigate('/definitions')}
          >
            <div style={{ width: 32, height: 32, borderRadius: 6, background: '#137fec', color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <AppstoreOutlined />
            </div>
            <Text style={{ fontWeight: 700, fontSize: 16 }}>GFlow Designer</Text>
          </div>
          <Divider type="vertical" style={{ height: 24, borderColor: token.colorBorder }} />
          <Text type="secondary">{processName || '未命名流程'}</Text>
        </Space>
        <Space>
          <Button icon={<EyeOutlined />} style={toolbarBtnStyle}>
            预览
          </Button>
          <Button icon={<ExportOutlined />} style={toolbarBtnStyle}>
            导出
          </Button>
          <Divider type="vertical" style={{ height: 24, borderColor: token.colorBorder }} />
          <Button
            icon={<SaveOutlined />}
            onClick={handleSave}
            loading={saving}
            disabled={actionDisabled}
            style={toolbarBtnStyle}
          >
            保存
          </Button>
          <Button
            type="primary"
            icon={<PlayCircleOutlined />}
            onClick={handleDeploy}
            loading={deploying}
            disabled={actionDisabled}
          >
            部署
          </Button>
          <Text type="secondary" style={{ fontSize: 12 }}>{Math.round(zoom * 100)}%</Text>
        </Space>
      </div>

      <div className="gflow-designer-body" style={{ flex: 1, display: 'flex', overflow: 'hidden', minHeight: 0 }}>
        <aside style={{ width: 256, background: token.colorBgContainer, borderRight: `1px solid ${token.colorBorder}`, overflowY: 'auto', flexShrink: 0 }}>
          <div style={{ padding: 14, borderBottom: `1px solid ${token.colorBorderSecondary || token.colorBorder}` }}>
            <Text type="secondary" style={{ fontSize: 12, fontWeight: 700 }}>元件库</Text>
          </div>
          <div style={{ padding: 12, display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <Text strong style={{ fontSize: 12 }}>事件 (Events)</Text>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginTop: 8 }}>
                <div style={{ border: `1px solid ${token.colorBorder}`, borderRadius: 8, padding: 8, textAlign: 'center' }}>
                  <div style={{ width: 28, height: 28, borderRadius: 28, border: '2px solid #64748b', margin: '0 auto 4px' }} />
                  <Text type="secondary" style={{ fontSize: 11 }}>开始事件</Text>
                </div>
                <div style={{ border: `1px solid ${token.colorBorder}`, borderRadius: 8, padding: 8, textAlign: 'center' }}>
                  <div style={{ width: 28, height: 28, borderRadius: 28, border: '4px solid #334155', margin: '0 auto 4px' }} />
                  <Text type="secondary" style={{ fontSize: 11 }}>结束事件</Text>
                </div>
              </div>
            </div>
            <div>
              <Text strong style={{ fontSize: 12 }}>任务 (Tasks)</Text>
              <Space direction="vertical" style={{ width: '100%', marginTop: 8 }} size={8}>
                <div style={{ border: `1px solid ${token.colorBorder}`, borderRadius: 8, padding: '8px 10px' }}><Text type="secondary" style={{ fontSize: 12 }}>用户任务</Text></div>
                <div style={{ border: `1px solid ${token.colorBorder}`, borderRadius: 8, padding: '8px 10px' }}><Text type="secondary" style={{ fontSize: 12 }}>服务任务</Text></div>
                <div style={{ border: `1px solid ${token.colorBorder}`, borderRadius: 8, padding: '8px 10px' }}><Text type="secondary" style={{ fontSize: 12 }}>消息任务</Text></div>
              </Space>
            </div>
            <div>
              <Text strong style={{ fontSize: 12 }}>网关 (Gateways)</Text>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginTop: 8 }}>
                <div style={{ border: `1px solid ${token.colorBorder}`, borderRadius: 8, padding: 8, textAlign: 'center' }}><Text type="secondary" style={{ fontSize: 11 }}>排他网关</Text></div>
                <div style={{ border: `1px solid ${token.colorBorder}`, borderRadius: 8, padding: 8, textAlign: 'center' }}><Text type="secondary" style={{ fontSize: 11 }}>并行网关</Text></div>
              </div>
            </div>
          </div>
        </aside>

        <div className="gflow-designer-canvas-wrap" style={{ flex: 1, position: 'relative', overflow: 'hidden', minHeight: 0 }}>
          <div
            ref={canvasRef}
            style={{
              width: '100%',
              height: '100%',
              background: token.colorBgLayout,
            }}
              className="bpmn-designer-canvas"
          />
          <div style={{ position: 'absolute', left: 16, bottom: 16, background: token.colorBgContainer, border: `1px solid ${token.colorBorder}`, borderRadius: 10, padding: 4, display: 'flex', alignItems: 'center', gap: 4, zIndex: 8 }}>
            <Button type="text" icon={<ZoomInOutlined />} onClick={handleZoomIn} disabled={actionDisabled} />
            <Text type="secondary" style={{ fontSize: 12, minWidth: 42, textAlign: 'center' }}>{Math.round(zoom * 100)}%</Text>
            <Button type="text" icon={<ZoomOutOutlined />} onClick={handleZoomOut} disabled={actionDisabled} />
            <Button type="text" icon={<FullscreenOutlined />} onClick={handleFitViewport} disabled={actionDisabled} />
          </div>
          {!isModelerReady && loading && (
            <div style={{
              position: 'absolute',
              inset: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: token.colorBgMask,
            }}>
              <Spin size="large" tip="加载中..." />
            </div>
          )}
        </div>

        <Card
          className="gflow-designer-panel"
          style={{
            width: 320,
            borderLeft: `1px solid ${token.colorBorder}`,
            borderRadius: 0,
            background: token.colorBgContainer,
            overflowY: 'auto',
            flexShrink: 0,
          }}
          styles={{
            header: {
              borderBottom: `1px solid ${token.colorBorder}`,
            },
            body: { padding: 16 },
          }}
          title={<Text style={{ color: token.colorText, fontSize: 16, fontWeight: 500 }}>属性面板</Text>}
        >
          <div style={{ marginBottom: 24 }}>
            <Text type="secondary" style={{ fontSize: 12, marginBottom: 8, display: 'block', fontWeight: 500 }}>
              流程属性
            </Text>
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <div>
                <Text style={{ fontSize: 12, marginBottom: 4, display: 'block', color: token.colorTextSecondary }}>流程名称</Text>
                <Input
                  value={processName}
                  onChange={(e) => setProcessName(e.target.value)}
                  placeholder="请输入流程名称"
                />
              </div>
              <div>
                <Text style={{ fontSize: 12, marginBottom: 4, display: 'block', color: token.colorTextSecondary }}>流程 Key</Text>
                <Input value={processKey || definition?.uid || ''} disabled placeholder="自动生成" />
              </div>
              <div>
                <Text style={{ fontSize: 12, marginBottom: 4, display: 'block', color: token.colorTextSecondary }}>描述</Text>
                <Input.TextArea
                  value={processDescription}
                  onChange={(e) => setProcessDescription(e.target.value)}
                  placeholder="请输入描述"
                  rows={3}
                />
              </div>
            </Space>
          </div>

          <Divider style={{ borderColor: token.colorBorder }} />

          {selectedElement ? (
            <div>
              <Text type="secondary" style={{ fontSize: 12, marginBottom: 8, display: 'block', fontWeight: 500 }}>
                元素属性
              </Text>
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <div>
                  <Text style={{ fontSize: 12, marginBottom: 4, display: 'block', color: token.colorTextSecondary }}>元素 ID</Text>
                  <Input value={selectedElement.id} disabled style={{ fontFamily: 'monospace', fontSize: 12 }} />
                </div>
                <div>
                  <Text style={{ fontSize: 12, marginBottom: 4, display: 'block', color: token.colorTextSecondary }}>元素名称</Text>
                  <Input
                    value={selectedElementName}
                    onChange={handleElementNameChange}
                    placeholder="请输入名称"
                  />
                </div>

                {isTaskElement && (
                  <div>
                    <Text style={{ fontSize: 12, marginBottom: 4, display: 'block', color: token.colorTextSecondary }}>关联接口</Text>
                    <Select
                      value={selectedEndpointId || undefined}
                      onChange={handleEndpointChange}
                      placeholder="选择要调用的接口"
                      style={{ width: '100%' }}
                      disabled={loadingEndpoints}
                      allowClear
                      options={endpoints.map(e => ({
                        value: String(e.id),
                        label: `${e.name || `Endpoint #${e.id}`} (${getTaskTypeLabel(e.taskType)})`,
                      }))}
                    />
                    <Text type="secondary" style={{ fontSize: 11, marginTop: 4, display: 'block' }}>
                      选择接口后，该任务将调用对应的服务
                    </Text>
                  </div>
                )}
              </Space>
            </div>
          ) : (
            <div style={{ textAlign: 'center', padding: '24px 0' }}>
              <Text type="secondary">
                点击画布中的元素
                <br />
                查看和编辑属性
              </Text>
            </div>
          )}
        </Card>
      </div>
    </div>
  );
};

export default Designer;
