import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Space, Breadcrumb, Badge, Button, Dropdown } from 'antd';
import type { ReactNode } from 'react';
import {
  DashboardOutlined,
  FileTextOutlined,
  BranchesOutlined,
  UserOutlined,
  AuditOutlined,
  AppstoreOutlined,
  CloudServerOutlined,
  GlobalOutlined,
  BellOutlined,
  RightOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  SettingOutlined,
  SunOutlined,
  MoonOutlined,
  DesktopOutlined,
} from '@ant-design/icons';
import { useUserStore, useAppStore } from '../stores';
import type { Theme } from '../stores/app';

const { Header, Sider, Content, Footer } = Layout;

const menuItems = [
  { path: '/dashboard', name: '仪表盘', icon: <DashboardOutlined /> },
  { path: '/definitions', name: '流程定义', icon: <FileTextOutlined /> },
  { path: '/processes', name: '流程实例', icon: <BranchesOutlined /> },
  { path: '/runners', name: '组件管理', icon: <CloudServerOutlined /> },
  { path: '/endpoints', name: '系统配置', icon: <GlobalOutlined /> },
  { path: '/users', name: '用户管理', icon: <UserOutlined /> },
  { path: '/audit-logs', name: '审计日志', icon: <AuditOutlined /> },
];

const routeNameMap: Record<string, string> = {
  '/dashboard': '仪表盘',
  '/definitions': '流程定义',
  '/processes': '流程实例',
  '/runners': '组件管理',
  '/endpoints': '系统配置',
  '/users': '用户管理',
  '/audit-logs': '审计日志',
};

const MainLayout = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const { user, logout } = useUserStore();
  const { sidebarCollapsed, toggleSidebar, theme: appTheme, setTheme } = useAppStore();

  const activeMenuKey = menuItems.find((item) => location.pathname.startsWith(item.path))?.path || '/dashboard';

  const getInitials = (name: string) => {
    return name.slice(0, 2).toUpperCase();
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人信息',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
      onClick: handleLogout,
    },
  ];

  const themeOptions: { value: Theme; label: string; icon: ReactNode }[] = [
    { value: 'light', label: '浅色', icon: <SunOutlined /> },
    { value: 'dark', label: '深色', icon: <MoonOutlined /> },
    { value: 'system', label: '跟随系统', icon: <DesktopOutlined /> },
  ];

  const getCurrentThemeIcon = () => {
    if (appTheme === 'dark') {
      return <MoonOutlined />;
    }
    if (appTheme === 'light') {
      return <SunOutlined />;
    }
    return <DesktopOutlined />;
  };

  const sideMenuItems = [
    {
      type: 'group' as const,
      label: '流程管理',
      children: menuItems
        .filter((item) => ['/dashboard', '/definitions', '/processes', '/runners'].includes(item.path))
        .map((item) => ({
          key: item.path,
          icon: item.icon,
          label: item.name,
          onClick: () => navigate(item.path),
        })),
    },
    {
      type: 'group' as const,
      label: '系统管理',
      children: menuItems
        .filter((item) => ['/endpoints', '/users', '/audit-logs'].includes(item.path))
        .map((item) => ({
          key: item.path,
          icon: item.icon,
          label: item.name,
          onClick: () => navigate(item.path),
        })),
    },
  ];

  return (
    <Layout className="gflow-main-layout">
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        className="gflow-main-sider"
        theme="dark"
        width={240}
        collapsedWidth={88}
      >
        <div className={`gflow-main-logo ${sidebarCollapsed ? 'is-collapsed' : ''}`}>
          <AppstoreOutlined className="gflow-main-logo-icon" />
          {!sidebarCollapsed && 'GFlow Admin'}
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[activeMenuKey]}
          items={sideMenuItems}
        />
      </Sider>
      <Layout className="gflow-main-frame" style={{ marginLeft: sidebarCollapsed ? 88 : 240 }}>
        <Header className="gflow-main-header">
          <Space size={8}>
            <Button
              type="text"
              className="gflow-main-collapse-btn"
              icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={toggleSidebar}
            />
            <Breadcrumb
              separator={<RightOutlined style={{ fontSize: 10, color: '#cbd5e1' }} />}
              items={[
                { title: '控制台' },
                { title: routeNameMap[location.pathname] || '详情' },
              ]}
            />
          </Space>
          <Space size={18}>
            <Dropdown
              menu={{
                items: themeOptions.map((option) => ({
                  key: option.value,
                  icon: option.icon,
                  label: option.label,
                })),
                onClick: ({ key }) => setTheme(key as Theme),
              }}
              trigger={['click']}
            >
              <Button type="text" className="gflow-main-header-theme-btn" icon={getCurrentThemeIcon()}>
                {themeOptions.find((item) => item.value === appTheme)?.label || '主题'}
              </Button>
            </Dropdown>
            <Badge dot>
              <BellOutlined className="gflow-main-header-bell" />
            </Badge>
            <div className="gflow-main-header-divider" />
            <span className="gflow-main-header-date">
              {new Date().toLocaleDateString('zh-CN', {
                year: 'numeric',
                month: '2-digit',
                day: '2-digit',
                weekday: 'long',
              })}
            </span>
            <Dropdown
              menu={{
                items: userMenuItems,
              }}
              trigger={['click']}
            >
              <div className="gflow-main-header-user">
                <Avatar className="gflow-main-user-avatar">{getInitials(user?.username || 'U')}</Avatar>
                <span>{user?.username || 'Admin User'}</span>
              </div>
            </Dropdown>
          </Space>
        </Header>
        <Content className="gflow-main-content">
          <div className="gflow-main-content-inner">
            <Outlet />
          </div>
        </Content>
        <Footer className="gflow-main-footer">
          <span>© 2024 GFlow Automation Systems. 保留所有权利。</span>
          <Space size={20}>
            <a>文档</a>
            <a>支持</a>
            <a>API 参考</a>
          </Space>
        </Footer>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
