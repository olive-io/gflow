import { useEffect, useMemo, useState } from "react";
import { Card, Col, Row, Space, Table, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import {
  AppstoreOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  RiseOutlined,
  SyncOutlined,
  ArrowRightOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import { definitionsApi, processesApi } from "../api";
import type { Process } from "../types";
import { formatDate, formatDuration } from "../utils/date";

const { Title, Text } = Typography;

interface DashboardStats {
  definitions: number;
  running: number;
  success: number;
  failed: number;
  totalProcesses: number;
}

interface TrendPoint {
  day: string;
  value: number;
}

const weekTrendData: TrendPoint[] = [
  { day: "Mon", value: 40 },
  { day: "Tue", value: 60 },
  { day: "Wed", value: 55 },
  { day: "Thu", value: 85 },
  { day: "Fri", value: 70 },
  { day: "Sat", value: 95 },
  { day: "Sun", value: 80 },
];

const monthTrendData: TrendPoint[] = Array.from({ length: 30 }, (_, index) => {
  const day = index + 1;
  const wave = Math.sin(day / 4.5) * 16;
  const value = Math.max(
    28,
    Math.min(96, Math.round(62 + wave + (day % 6) * 2)),
  );
  return {
    day: `${day}`,
    value,
  };
});

const Dashboard = () => {
  const navigate = useNavigate();
  const [trendMode, setTrendMode] = useState<"week" | "month">("week");
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats>({
    definitions: 0,
    running: 0,
    success: 0,
    failed: 0,
    totalProcesses: 0,
  });
  const [recentProcesses, setRecentProcesses] = useState<Process[]>([]);

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        const [definitionsRes, processesRes] = await Promise.all([
          definitionsApi.list({ page: 1, size: 1 }),
          processesApi.list({ page: 1, size: 100 }),
        ]);

        const processList = processesRes.processes || [];
        const running = processList.filter(
          (item) => item.status === "Running",
        ).length;
        const success = processList.filter(
          (item) => item.status === "Success",
        ).length;
        const failed = processList.filter(
          (item) => item.status === "Failed",
        ).length;

        setStats({
          definitions: Number(definitionsRes.total) || 0,
          running,
          success,
          failed,
          totalProcesses: Number(processesRes.total) || processList.length,
        });

        setRecentProcesses(
          [...processList]
            .sort((a, b) => Number(b.startAt || 0) - Number(a.startAt || 0))
            .slice(0, 4),
        );
      } catch (error) {
        console.error("Failed to fetch dashboard data:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []);

  const statusRatio = useMemo(() => {
    if (!stats.totalProcesses) {
      return { success: 71, running: 27, failed: 2 };
    }
    const success = Math.max(
      0,
      Math.round((stats.success / stats.totalProcesses) * 100),
    );
    const running = Math.max(
      0,
      Math.round((stats.running / stats.totalProcesses) * 100),
    );
    const failed = Math.max(0, 100 - success - running);
    return { success, running, failed };
  }, [stats]);

  const trendData = trendMode === "week" ? weekTrendData : monthTrendData;
  const trendTitle = trendMode === "week" ? "近7天执行趋势" : "近30天执行趋势";

  const recentColumns: ColumnsType<Process> = [
    {
      title: "流程名称",
      dataIndex: "name",
      key: "name",
      width: 320,
      render: (name, record) => (
        <div style={{ display: "flex", flexDirection: "column", minWidth: 0, margin: "0 16px" }}>
          <a
            onClick={() => navigate(`/processes/${record.id}`)}
            style={{
              whiteSpace: "nowrap",
              overflow: "hidden",
              textOverflow: "ellipsis",
            }}
            title={name || `Process #${record.id}`}
          >
            {name || `Process #${record.id}`}
          </a>
          <Text type="secondary" style={{ fontSize: 12 }}>
            ID: {record.uid || record.id}
          </Text>
        </div>
      ),
    },
    {
      title: "状态",
      dataIndex: "status",
      key: "status",
      width: 130,
      render: (status) => {
        if (status === "Success") {
          return <Tag color="success">成功</Tag>;
        }
        if (status === "Running") {
          return <Tag color="processing">运行中</Tag>;
        }
        return <Tag color="error">失败</Tag>;
      },
    },
    {
      title: "开始时间",
      dataIndex: "startAt",
      key: "startAt",
      width: 190,
      render: (value) => (
        <Text type="secondary" style={{ whiteSpace: "nowrap" }}>
          {formatDate(value)}
        </Text>
      ),
    },
    {
      title: "耗时",
      key: "duration",
      width: 120,
      render: (_, record) => (
        <Text type="secondary" style={{ whiteSpace: "nowrap" }}>
          {formatDuration(record.startAt, record.endAt)}
        </Text>
      ),
    },
    {
      title: "操作",
      key: "action",
      width: 100,
      // align: "right",
      render: (_, record) => (
        <a onClick={() => navigate(`/processes/${record.id}`)}>详情</a>
      ),
    },
  ];

  const kpis = [
    {
      title: "流程总数",
      value: stats.definitions,
      trend: "+12%",
      trendColor: "#059669",
      icon: <AppstoreOutlined />,
      iconBg: "#ccfbf1",
      iconColor: "#0d9488",
    },
    {
      title: "运行中",
      value: stats.running,
      trend: "+5.2%",
      trendColor: "#059669",
      icon: <SyncOutlined />,
      iconBg: "#dbeafe",
      iconColor: "#2563eb",
    },
    {
      title: "成功",
      value: stats.success,
      trend: "+8.1%",
      trendColor: "#059669",
      icon: <CheckCircleOutlined />,
      iconBg: "#dcfce7",
      iconColor: "#059669",
    },
    {
      title: "失败",
      value: stats.failed,
      trend: "-2.4%",
      trendColor: "#dc2626",
      icon: <ExclamationCircleOutlined />,
      iconBg: "#fee2e2",
      iconColor: "#dc2626",
    },
  ];

  const donutBackground = `conic-gradient(#0d9488 0 ${statusRatio.success}%, #3b82f6 ${statusRatio.success}% ${statusRatio.success + statusRatio.running}%, #ef4444 ${statusRatio.success + statusRatio.running}% 100%)`;

  return (
    <div
      style={{ padding: 8, display: "flex", flexDirection: "column", gap: 24 }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <Space size={8}>
          <Text style={{ color: "#94a3b8" }}>控制台</Text>
          <Text style={{ color: "#cbd5e1" }}>{">"}</Text>
          <Text style={{ color: "#0f172a", fontWeight: 600 }}>仪表盘概览</Text>
        </Space>
        <Text style={{ color: "#334155", fontWeight: 500 }}>
          {new Date().toLocaleDateString("zh-CN", {
            year: "numeric",
            month: "2-digit",
            day: "2-digit",
            weekday: "long",
          })}
        </Text>
      </div>

      <Row gutter={[16, 16]}>
        {kpis.map((item) => (
          <Col xs={24} md={12} xl={6} key={item.title}>
            <Card
              loading={loading}
              style={{ borderRadius: 12, borderColor: "#f1f5f9" }}
              styles={{ body: { padding: 20 } }}
            >
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  marginBottom: 14,
                }}
              >
                <div
                  style={{
                    width: 34,
                    height: 34,
                    borderRadius: 8,
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    background: item.iconBg,
                    color: item.iconColor,
                    fontSize: 18,
                  }}
                >
                  {item.icon}
                </div>
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    color: item.trendColor,
                    fontSize: 12,
                    fontWeight: 700,
                  }}
                >
                  <RiseOutlined style={{ marginRight: 2 }} />
                  {item.trend}
                </div>
              </div>
              <Text style={{ color: "#64748b", fontSize: 14 }}>
                {item.title}
              </Text>
              <div
                style={{
                  fontSize: 30,
                  lineHeight: "36px",
                  fontWeight: 800,
                  color: "#0f172a",
                  marginTop: 4,
                }}
              >
                {item.value.toLocaleString()}
              </div>
            </Card>
          </Col>
        ))}
      </Row>

      <Row gutter={[16, 16]}>
        <Col xs={24} xl={16}>
          <Card
            style={{ borderRadius: 12, borderColor: "#f1f5f9", height: "100%" }}
            styles={{ body: { padding: 20, height: "100%" } }}
          >
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                marginBottom: 22,
              }}
            >
              <Title level={5} style={{ margin: 0 }}>
                {trendTitle}
              </Title>
              <Space size={6}>
                <button
                  onClick={() => setTrendMode("week")}
                  style={{
                    border: 0,
                    borderRadius: 6,
                    padding: "4px 10px",
                    fontSize: 12,
                    fontWeight: 600,
                    background:
                      trendMode === "week" ? "#f1f5f9" : "transparent",
                    color: trendMode === "week" ? "#475569" : "#94a3b8",
                    cursor: "pointer",
                  }}
                >
                  周
                </button>
                <button
                  onClick={() => setTrendMode("month")}
                  style={{
                    border: 0,
                    borderRadius: 6,
                    padding: "4px 10px",
                    fontSize: 12,
                    fontWeight: 600,
                    background:
                      trendMode === "month" ? "#f1f5f9" : "transparent",
                    color: trendMode === "month" ? "#475569" : "#94a3b8",
                    cursor: "pointer",
                  }}
                >
                  月
                </button>
              </Space>
            </div>
            <div
              style={{
                height: 300,
                display: "flex",
                alignItems: "flex-end",
                gap: 14,
              }}
            >
              {trendData.map((point, index) => {
                const isLast = index === trendData.length - 1;
                return (
                  <div
                    key={point.day}
                    style={{
                      flex: 1,
                      height: "100%",
                      display: "flex",
                      alignItems: "flex-end",
                    }}
                  >
                    <div
                      style={{
                        width: "100%",
                        position: "relative",
                        height: `${point.value}%`,
                        background: isLast
                          ? "#0d9488"
                          : "rgba(13, 148, 136, 0.18)",
                        borderRadius: "4px 4px 0 0",
                      }}
                    >
                      {!isLast ? (
                        <div
                          style={{
                            position: "absolute",
                            inset: 0,
                            background: "#0d9488",
                            opacity: 0.35 + index * 0.08,
                            borderRadius: "4px 4px 0 0",
                          }}
                        />
                      ) : null}
                      {isLast ? (
                        <div
                          style={{
                            position: "absolute",
                            top: -28,
                            left: "50%",
                            transform: "translateX(-50%)",
                            background: "#0f172a",
                            color: "#fff",
                            fontSize: 10,
                            padding: "2px 8px",
                            borderRadius: 6,
                            whiteSpace: "nowrap",
                          }}
                        >
                          {trendMode === "week" ? "今日: 142" : "本月: 482"}
                        </div>
                      ) : null}
                    </div>
                  </div>
                );
              })}
            </div>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                marginTop: 8,
                color: "#94a3b8",
                fontSize: 11,
                fontWeight: 700,
                letterSpacing: "0.08em",
              }}
            >
              {trendData.map((item, index) => {
                if (
                  trendMode === "month" &&
                  index % 5 !== 0 &&
                  index !== trendData.length - 1
                ) {
                  return <span key={item.day}>&nbsp;</span>;
                }
                return (
                  <span key={item.day}>
                    {trendMode === "month" ? `${item.day}d` : item.day}
                  </span>
                );
              })}
            </div>
          </Card>
        </Col>

        <Col xs={24} xl={8}>
          <Card
            style={{ borderRadius: 12, borderColor: "#f1f5f9", height: "100%" }}
            styles={{ body: { padding: 20, height: "100%" } }}
          >
            <Title level={5} style={{ margin: 0, marginBottom: 20 }}>
              状态分布
            </Title>
            <div
              style={{
                height: 300,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                position: "relative",
              }}
            >
              <div
                style={{
                  width: 192,
                  height: 192,
                  borderRadius: "50%",
                  background: donutBackground,
                  padding: 14,
                }}
              >
                <div
                  style={{
                    width: "100%",
                    height: "100%",
                    borderRadius: "50%",
                    background: "#fff",
                  }}
                />
              </div>
              <div style={{ position: "absolute", textAlign: "center" }}>
                <div
                  style={{
                    fontSize: 32,
                    lineHeight: "36px",
                    fontWeight: 800,
                    color: "#0f172a",
                  }}
                >
                  {stats.totalProcesses.toLocaleString()}
                </div>
                <div
                  style={{
                    color: "#94a3b8",
                    fontSize: 11,
                    fontWeight: 700,
                    letterSpacing: "0.08em",
                  }}
                >
                  TOTAL
                </div>
              </div>
            </div>

            <Space orientation="vertical" size={10} style={{ width: "100%" }}>
              <div style={{ display: "flex", justifyContent: "space-between" }}>
                <Space size={8}>
                  <span
                    style={{
                      width: 10,
                      height: 10,
                      borderRadius: 10,
                      background: "#0d9488",
                    }}
                  />
                  <Text style={{ color: "#475569" }}>成功</Text>
                </Space>
                <Text strong>{statusRatio.success}%</Text>
              </div>
              <div style={{ display: "flex", justifyContent: "space-between" }}>
                <Space size={8}>
                  <span
                    style={{
                      width: 10,
                      height: 10,
                      borderRadius: 10,
                      background: "#3b82f6",
                    }}
                  />
                  <Text style={{ color: "#475569" }}>运行中</Text>
                </Space>
                <Text strong>{statusRatio.running}%</Text>
              </div>
              <div style={{ display: "flex", justifyContent: "space-between" }}>
                <Space size={8}>
                  <span
                    style={{
                      width: 10,
                      height: 10,
                      borderRadius: 10,
                      background: "#ef4444",
                    }}
                  />
                  <Text style={{ color: "#475569" }}>失败</Text>
                </Space>
                <Text strong>{statusRatio.failed}%</Text>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>

      <Card
        className="gflow-glass-card"
        style={{ borderRadius: 12, borderColor: "#f1f5f9" }}
        styles={{ body: { padding: 0, overflow: "hidden" } }}
      >
        <div
          style={{
            padding: "18px 20px",
            borderBottom: "1px solid #f1f5f9",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <Title level={5} style={{ margin: 0 }}>
            最近执行
          </Title>
          <a
            style={{
              color: "#0d9488",
              fontWeight: 600,
              display: "inline-flex",
              alignItems: "center",
              gap: 4,
            }}
          >
            查看全部 <ArrowRightOutlined style={{ fontSize: 10 }} />
          </a>
        </div>
        <Table
          rowKey="id"
          columns={recentColumns}
          dataSource={recentProcesses}
          loading={loading}
          pagination={false}
          size="middle"
          locale={{ emptyText: "暂无执行记录" }}
        />
      </Card>
    </div>
  );
};

export default Dashboard;
