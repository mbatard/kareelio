import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { LanguageProvider } from './contexts/LanguageContext';
import { Navbar } from './components/Navbar';
import { ProtectedRoute, AdminRoute } from './components/ProtectedRoute';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { VerifyEmailPage } from './pages/VerifyEmailPage';
import { DashboardPage } from './pages/DashboardPage';
import { ProfilePage } from './pages/ProfilePage';
import { AboutPage } from './pages/AboutPage';
import { AdminPage } from './pages/AdminPage';
import { AdminUsersPage } from './pages/AdminUsersPage';
import { AdminUserEditPage } from './pages/AdminUserEditPage';
import { AdminAuditPage } from './pages/AdminAuditPage';
import { ApplicationsPage } from './pages/ApplicationsPage';
import { JobApplicationForm } from './pages/JobApplicationForm';

function AppRoutes() {
  const { user, loading, isAdmin } = useAuth();

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen bg-gray-50 dark:bg-gray-900">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors">
      <Navbar />
      <Routes>
        <Route path="/login" element={user ? <Navigate to={isAdmin ? '/admin' : '/'} /> : <LoginPage />} />
        <Route path="/register" element={user ? <Navigate to={isAdmin ? '/admin' : '/'} /> : <RegisterPage />} />
        <Route path="/verify-email" element={<VerifyEmailPage />} />
        {!isAdmin && (
          <>
            <Route path="/" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
            <Route path="/applications" element={<ProtectedRoute><ApplicationsPage /></ProtectedRoute>} />
            <Route path="/applications/new" element={<ProtectedRoute><JobApplicationForm /></ProtectedRoute>} />
            <Route path="/applications/:id/edit" element={<ProtectedRoute><JobApplicationForm /></ProtectedRoute>} />
          </>
        )}
        {isAdmin && (
          <Route path="/" element={<Navigate to="/admin" />} />
        )}
        <Route path="/profile" element={<ProtectedRoute><ProfilePage /></ProtectedRoute>} />
        <Route path="/about" element={<ProtectedRoute><AboutPage /></ProtectedRoute>} />
        <Route path="/admin" element={<AdminRoute><AdminPage /></AdminRoute>} />
        <Route path="/admin/audit" element={<AdminRoute><AdminAuditPage /></AdminRoute>} />
        <Route path="/admin/users" element={<AdminRoute><AdminUsersPage /></AdminRoute>} />
        <Route path="/admin/users/:id/edit" element={<AdminRoute><AdminUserEditPage /></AdminRoute>} />
        <Route path="*" element={<Navigate to={isAdmin ? '/admin' : '/'} />} />
      </Routes>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <LanguageProvider>
          <AuthProvider>
            <AppRoutes />
          </AuthProvider>
        </LanguageProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}

export default App;
