import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from '@gastrack/shared';
import MainLayout from './layouts/MainLayout';
import LoginPage from './pages/auth/LoginPage';
import RegisterPage from './pages/auth/RegisterPage';
import DashboardPage from './pages/dashboard/DashboardPage';
import VehicleListPage from './pages/vehicle/VehicleListPage';
import VehicleFormPage from './pages/vehicle/VehicleFormPage';
import RecordListPage from './pages/record/RecordListPage';
import RecordFormPage from './pages/record/RecordFormPage';
import RecordDetailPage from './pages/record/RecordDetailPage';
import StatsPage from './pages/stats/StatsPage';
import SettingsPage from './pages/settings/SettingsPage';
import InviteManagePage from './pages/invite/InviteManagePage';
import ReminderPage from './pages/reminder/ReminderPage';
import GroupPage from './pages/group/GroupPage';
import ExpenseListPage from './pages/expense/ExpenseListPage';
import ExpenseFormPage from './pages/expense/ExpenseFormPage';
import ExpenseDetailPage from './pages/expense/ExpenseDetailPage';
import PrivacyPolicyPage from './pages/legal/PrivacyPolicyPage';
import TermsOfServicePage from './pages/legal/TermsOfServicePage';
import PWAUpdatePrompt from './components/PWAUpdatePrompt';
import InstallPrompt from './components/InstallPrompt';
import { useEffect } from 'react';

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
    </>
  );
}
