import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from '@gastrack/shared';
import MainLayout from './layouts/MainLayout';
import PWAUpdatePrompt from './components/PWAUpdatePrompt';
import InstallPrompt from './components/InstallPrompt';
import { Suspense, lazy, useEffect } from 'react';
import PageFallback from './components/PageFallback';

// 页面按路由拆分打包，首屏只加载当前页面
const LoginPage = lazy(() => import('./pages/auth/LoginPage'));
const RegisterPage = lazy(() => import('./pages/auth/RegisterPage'));
const DashboardPage = lazy(() => import('./pages/dashboard/DashboardPage'));
const VehicleListPage = lazy(() => import('./pages/vehicle/VehicleListPage'));
const VehicleFormPage = lazy(() => import('./pages/vehicle/VehicleFormPage'));
const RecordListPage = lazy(() => import('./pages/record/RecordListPage'));
const RecordFormPage = lazy(() => import('./pages/record/RecordFormPage'));
const RecordDetailPage = lazy(() => import('./pages/record/RecordDetailPage'));
const StatsPage = lazy(() => import('./pages/stats/StatsPage'));
const SettingsPage = lazy(() => import('./pages/settings/SettingsPage'));
const InviteManagePage = lazy(() => import('./pages/invite/InviteManagePage'));
const ReminderPage = lazy(() => import('./pages/reminder/ReminderPage'));
const GroupPage = lazy(() => import('./pages/group/GroupPage'));
const ExpenseListPage = lazy(() => import('./pages/expense/ExpenseListPage'));
const ExpenseFormPage = lazy(() => import('./pages/expense/ExpenseFormPage'));
const ExpenseDetailPage = lazy(() => import('./pages/expense/ExpenseDetailPage'));
const PrivacyPolicyPage = lazy(() => import('./pages/legal/PrivacyPolicyPage'));
const TermsOfServicePage = lazy(() => import('./pages/legal/TermsOfServicePage'));

/** 需要认证的路由守卫 */
function RequireAuth({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return <>{children}</>;
}

/** 已登录时重定向到首页 */
function GuestOnly({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  if (isAuthenticated) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  const { isAuthenticated, fetchProfile } = useAuthStore();

  useEffect(() => {
    if (isAuthenticated) {
      fetchProfile();
    }
  }, [isAuthenticated, fetchProfile]);

  return (
    <>
      <PWAUpdatePrompt />
      <InstallPrompt />
      <Suspense fallback={<PageFallback />}>
      <Routes>
      {/* 公开路由 */}
      <Route path="/login" element={<GuestOnly><LoginPage /></GuestOnly>} />
      <Route path="/register" element={<GuestOnly><RegisterPage /></GuestOnly>} />
      <Route path="/privacy" element={<PrivacyPolicyPage />} />
      <Route path="/terms" element={<TermsOfServicePage />} />

      {/* 需要认证的路由 */}
      <Route
        path="/"
        element={
          <RequireAuth>
            <MainLayout />
          </RequireAuth>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="vehicles" element={<VehicleListPage />} />
        <Route path="vehicles/new" element={<VehicleFormPage />} />
        <Route path="vehicles/:id/edit" element={<VehicleFormPage />} />
        <Route path="vehicles/:vehicleId/records" element={<RecordListPage />} />
        <Route path="vehicles/:vehicleId/records/new" element={<RecordFormPage />} />
        <Route path="vehicles/:vehicleId/records/:recordId" element={<RecordDetailPage />} />
        <Route path="vehicles/:vehicleId/records/:recordId/edit" element={<RecordFormPage />} />
        <Route path="vehicles/:vehicleId/expenses" element={<ExpenseListPage />} />
        <Route path="vehicles/:vehicleId/expenses/new" element={<ExpenseFormPage />} />
        <Route path="vehicles/:vehicleId/expenses/:expenseId" element={<ExpenseDetailPage />} />
        <Route path="vehicles/:vehicleId/expenses/:expenseId/edit" element={<ExpenseFormPage />} />
        <Route path="records/new" element={<RecordFormPage />} />
        <Route path="stats" element={<StatsPage />} />
        <Route path="invites" element={<InviteManagePage />} />
        <Route path="reminders" element={<ReminderPage />} />
        <Route path="groups" element={<GroupPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>

      {/* 404 重定向 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
      </Suspense>
    </>
  );
}
