import { useEffect, useMemo } from 'react';
import { ConfigProvider, App as AntApp, theme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import { RouterProvider } from 'react-router-dom';
import { router } from './router';
import { useAppStore } from './stores';

const darkThemeToken = {
  colorPrimary: '#1d4ed8',
  colorSuccess: '#22c55e',
  colorWarning: '#f59e0b',
  colorError: '#ef4444',
  colorInfo: '#3b82f6',
  colorTextBase: '#f8fafc',
  colorBgBase: '#0a1120',
  colorPrimaryBg: '#111b34',
  colorPrimaryBgHover: '#162447',
  colorPrimaryBorder: '#22396f',
  colorPrimaryBorderHover: '#2b4f95',
  colorPrimaryHover: '#2563eb',
  colorPrimaryActive: '#1e40af',
  colorPrimaryText: '#7bb4ff',
  colorPrimaryTextHover: '#b8d5ff',
  colorPrimaryTextActive: '#deebff',
  colorSuccessBg: '#052e16',
  colorSuccessBgHover: '#064e3b',
  colorSuccessBorder: '#047857',
  colorSuccessBorderHover: '#059669',
  colorSuccessHover: '#16a34a',
  colorSuccessActive: '#15803d',
  colorSuccessText: '#4ade80',
  colorSuccessTextHover: '#86efac',
  colorSuccessTextActive: '#bbf7d0',
  colorWarningBg: '#451a03',
  colorWarningBgHover: '#713f12',
  colorWarningBorder: '#b45309',
  colorWarningBorderHover: '#d97706',
  colorWarningHover: '#f97316',
  colorWarningActive: '#ea580c',
  colorWarningText: '#fbbf24',
  colorWarningTextHover: '#fcd34d',
  colorWarningTextActive: '#fde68a',
  colorErrorBg: '#450a0a',
  colorErrorBgHover: '#7f1d1d',
  colorErrorBorder: '#b91c1c',
  colorErrorBorderHover: '#dc2626',
  colorErrorHover: '#f87171',
  colorErrorActive: '#ef4444',
  colorErrorText: '#f87171',
  colorErrorTextHover: '#fca5a5',
  colorErrorTextActive: '#fecaca',
  colorInfoBg: '#0f172a',
  colorInfoBgHover: '#1e293b',
  colorInfoBorder: '#334155',
  colorInfoBorderHover: '#475569',
  colorInfoHover: '#2563eb',
  colorInfoActive: '#1d4ed8',
  colorInfoText: '#60a5fa',
  colorInfoTextHover: '#93c5fd',
  colorInfoTextActive: '#bfdbfe',
  colorText: '#d8e2f2',
  colorTextSecondary: '#9bb0cf',
  colorTextTertiary: '#7083a0',
  colorTextQuaternary: '#5b6c87',
  colorTextDisabled: '#5b6c87',
  colorBgContainer: '#101a30',
  colorBgElevated: '#162447',
  colorBgLayout: '#0a1120',
  colorBgSpotlight: 'rgba(15, 23, 42, 0.8)',
  colorBgMask: 'rgba(2, 6, 23, 0.6)',
  colorBorder: '#23365e',
  colorBorderSecondary: '#314977',
  borderRadius: 10,
  borderRadiusXS: 2,
  borderRadiusSM: 4,
  borderRadiusLG: 14,
  padding: 16,
  paddingSM: 12,
  paddingLG: 20,
  margin: 16,
  marginSM: 12,
  marginLG: 20,
  boxShadow: '0 6px 24px rgba(7, 20, 45, 0.35)',
  boxShadowSecondary: '0 12px 36px rgba(5, 15, 33, 0.4)',
};

const lightThemeToken = {
  colorPrimary: '#1d4ed8',
  colorSuccess: '#22c55e',
  colorWarning: '#f59e0b',
  colorError: '#ef4444',
  colorInfo: '#3b82f6',
  colorTextBase: '#0f172a',
  colorBgBase: '#ffffff',
  colorPrimaryBg: '#ebf3ff',
  colorPrimaryBgHover: '#dbeafe',
  colorPrimaryBorder: '#8db8ff',
  colorPrimaryBorderHover: '#5f93f0',
  colorPrimaryHover: '#2563eb',
  colorPrimaryActive: '#1d4ed8',
  colorPrimaryText: '#245fcd',
  colorPrimaryTextHover: '#1d4ed8',
  colorPrimaryTextActive: '#1e40af',
  colorSuccessBg: '#f0fdf4',
  colorSuccessBgHover: '#dcfce7',
  colorSuccessBorder: '#86efac',
  colorSuccessBorderHover: '#4ade80',
  colorSuccessHover: '#16a34a',
  colorSuccessActive: '#15803d',
  colorSuccessText: '#22c55e',
  colorSuccessTextHover: '#16a34a',
  colorSuccessTextActive: '#15803d',
  colorWarningBg: '#fffbeb',
  colorWarningBgHover: '#fef3c7',
  colorWarningBorder: '#fcd34d',
  colorWarningBorderHover: '#fbbf24',
  colorWarningHover: '#d97706',
  colorWarningActive: '#b45309',
  colorWarningText: '#f59e0b',
  colorWarningTextHover: '#d97706',
  colorWarningTextActive: '#b45309',
  colorErrorBg: '#fef2f2',
  colorErrorBgHover: '#fee2e2',
  colorErrorBorder: '#fca5a5',
  colorErrorBorderHover: '#f87171',
  colorErrorHover: '#dc2626',
  colorErrorActive: '#b91c1c',
  colorErrorText: '#ef4444',
  colorErrorTextHover: '#dc2626',
  colorErrorTextActive: '#b91c1c',
  colorInfoBg: '#eff6ff',
  colorInfoBgHover: '#dbeafe',
  colorInfoBorder: '#93c5fd',
  colorInfoBorderHover: '#60a5fa',
  colorInfoHover: '#2563eb',
  colorInfoActive: '#1d4ed8',
  colorInfoText: '#3b82f6',
  colorInfoTextHover: '#2563eb',
  colorInfoTextActive: '#1d4ed8',
  colorText: '#0f172a',
  colorTextSecondary: '#4b5f7f',
  colorTextTertiary: '#62718a',
  colorTextQuaternary: '#8a98b0',
  colorTextDisabled: '#98a6bb',
  colorBgContainer: '#ffffff',
  colorBgElevated: '#ffffff',
  colorBgLayout: '#f2f6fc',
  colorBgSpotlight: 'rgba(15, 23, 42, 0.9)',
  colorBgMask: 'rgba(15, 23, 42, 0.6)',
  colorBorder: '#d7dfed',
  colorBorderSecondary: '#eaf0fa',
  borderRadius: 10,
  borderRadiusXS: 2,
  borderRadiusSM: 4,
  borderRadiusLG: 14,
  padding: 16,
  paddingSM: 12,
  paddingLG: 20,
  margin: 16,
  marginSM: 12,
  marginLG: 20,
  boxShadow: '0 8px 30px rgba(26, 70, 145, 0.08)',
  boxShadowSecondary: '0 14px 42px rgba(22, 52, 102, 0.12)',
};

const App = () => {
  const { theme: appTheme } = useAppStore();

  const isDark = useMemo(() => {
    if (appTheme === 'system') {
      return window.matchMedia('(prefers-color-scheme: dark)').matches;
    }
    return appTheme === 'dark';
  }, [appTheme]);

  useEffect(() => {
    const applyTheme = (themeValue: 'light' | 'dark' | 'system') => {
      const root = document.documentElement;
      root.classList.remove('light', 'dark');

      if (themeValue === 'system') {
        const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        root.classList.add(systemDark ? 'dark' : 'light');
      } else {
        root.classList.add(themeValue);
      }
    };

    applyTheme(appTheme);

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = () => {
      if (appTheme === 'system') {
        applyTheme('system');
      }
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [appTheme]);

  const themeConfig = useMemo(() => ({
    algorithm: isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: isDark ? darkThemeToken : lightThemeToken,
    components: {
      Button: {
        borderRadius: 10,
        controlHeight: 36,
        fontWeight: 600,
      },
      Card: {
        borderRadius: 16,
      },
      Table: {
        borderRadius: 14,
      },
      Modal: {
        borderRadius: 16,
      },
      Input: {
        borderRadius: 10,
      },
      Select: {
        borderRadius: 10,
      },
      Tag: {
        borderRadius: 9999,
      },
      Layout: {
        headerBg: 'transparent',
        bodyBg: 'transparent',
        siderBg: 'transparent',
      },
      Menu: {
        darkItemBg: 'transparent',
        darkSubMenuItemBg: 'transparent',
        darkItemSelectedBg: 'rgba(255, 255, 255, 0.16)',
        darkItemHoverBg: 'rgba(255, 255, 255, 0.12)',
      },
    },
  }), [isDark]);

  return (
    <ConfigProvider
      locale={zhCN}
      theme={themeConfig}
    >
      <AntApp>
        <RouterProvider router={router} />
      </AntApp>
    </ConfigProvider>
  );
};

export default App;
