import { useEffect, useState, useCallback } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Typography, theme, Drawer, Grid } from 'antd';
import {
  DashboardOutlined,
  CarOutlined,
  BarChartOutlined,
  GiftOutlined,
  SettingOutlined,
  LogoutOutlined,
  UserOutlined,
  MenuOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ToolOutlined,
} from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAuthStore, useThemeStore } from '@gastrack/shared';
import type { MenuProps } from 'antd';
import LanguageSwitcher from '../components/LanguageSwitcher';
import NotificationBell from '../components/NotificationBell';

const { Header, Sider, Content } = Layout;
const { useBreakpoint } = Grid;

export default function MainLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const { user, logout } = useAuthStore();
  const resolved = useThemeStore((s) => s.resolved);
  const { token } = theme.useToken();
  const screens = useBreakpoint();

  // 是否为移动端（宽度 < lg = 992px）
  const isMobile = !screens.lg;

  // 同步浏览器标题和 html lang 属性
  useEffect(() => {
    document.title = t('app.title');
    document.documentElement.lang = i18n.language;
  }, [i18n.language, t]);

  // 路由变化时关闭 Drawer
  useEffect(() => {
    if (isMobile) setDrawerOpen(false);
  }, [location.pathname, isMobile]);

  const menuItems: MenuProps['items'] = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: t('nav.dashboard'),
    },
    {
      key: '/vehicles',
      icon: <CarOutlined />,
      label: t('nav.vehicles'),
    },
    {
      key: '/stats',
      icon: <BarChartOutlined />,
      label: t('nav.stats'),
    },
    {
      key: '/invites',
      icon: <GiftOutlined />,
      label: t('nav.invites'),
    },
    {
      key: '/reminders',
      icon: <ToolOutlined />,
      label: t('nav.reminders'),
    },
    {
      key: '/settings',
      icon: <SettingOutlined />,
      label: t('nav.settings'),
    },
  ];

  const handleMenuClick: MenuProps['onClick'] = useCallback(({ key }: { key: string }) => {
    navigate(key);
    if (isMobile) setDrawerOpen(false);
  }, [navigate, isMobile]);

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: user?.nickname || user?.email || 'User',
      disabled: true,
    },
    { type: 'divider' },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: t('nav.settings'),
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('auth.logout'),
      danger: true,
    },
  ];

  const handleUserMenu: MenuProps['onClick'] = ({ key }) => {
    if (key === 'logout') handleLogout();
    if (key === 'settings') navigate('/settings');
  };

  // 获取当前选中的菜单 key
  const selectedKey = '/' + (location.pathname.split('/')[1] || '');

  /** 侧边栏内容（桌面 Sider / 移动 Drawer 共用） */
  const siderContent = (
    <>
      <div
        style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
        }}
      >
        <Typography.Title
          level={4}
          style={{ margin: 0, color: token.colorPrimary, whiteSpace: 'nowrap' }}
        >
          {collapsed && !isMobile ? '⛽' : '⛽ GasTrack'}
        </Typography.Title>
      </div>
      <Menu
        mode="inline"
        selectedKeys={[selectedKey]}
        items={menuItems}
        onClick={handleMenuClick}
        style={{ borderInlineEnd: 'none', marginTop: 8 }}
      />
    </>
  );

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* ── 桌面端: 传统 Sider ── */}
      {!isMobile && (
        <Sider
          collapsible
          collapsed={collapsed}
          onCollapse={setCollapsed}
          breakpoint="lg"
          theme={resolved === 'dark' ? 'dark' : 'light'}
          style={{
            borderRight: `1px solid ${token.colorBorderSecondary}`,
            boxShadow: 'var(--gt-shadow-sider)',
            background: token.colorBgContainer,
          }}
        >
          {siderContent}
        </Sider>
      )}

      {/* ── 移动端: Drawer 抽屉导航 ── */}
      {isMobile && (
        <Drawer
          placement="left"
          open={drawerOpen}
          onClose={() => setDrawerOpen(false)}
          width={256}
          bodyStyle={{ padding: 0, background: token.colorBgContainer }}
          headerStyle={{ display: 'none' }}
        >
          {siderContent}
        </Drawer>
      )}

      <Layout>
        <Header
          style={{
            background: token.colorBgContainer,
            padding: isMobile ? '0 12px' : '0 24px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            boxShadow: 'var(--gt-shadow-header)',
            position: isMobile ? 'sticky' : undefined,
            top: isMobile ? 0 : undefined,
            zIndex: isMobile ? 10 : undefined,
          }}
        >
          <div
            style={{ cursor: 'pointer', fontSize: 18 }}
            onClick={() => (isMobile ? setDrawerOpen(!drawerOpen) : setCollapsed(!collapsed))}
          >
            {isMobile ? <MenuOutlined /> : (collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />)}
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: isMobile ? 8 : 16 }}>
            <LanguageSwitcher style={{ fontSize: 18 }} />
            <NotificationBell />

            <Dropdown
              menu={{ items: userMenuItems, onClick: handleUserMenu }}
              placement="bottomRight"
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
                <Avatar
                  size="small"
                  icon={<UserOutlined />}
                  src={user?.avatar_url || undefined}
                />
                {!isMobile && (
                  <span>{user?.nickname || user?.email || ''}</span>
                )}
              </div>
            </Dropdown>
          </div>
        </Header>

        <Content style={{ margin: 0, overflow: 'auto', paddingBottom: isMobile ? 0 : undefined }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}
