import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import DashboardLayout from './components/DashboardLayout';
import Dashboard from './pages/Dashboard';
import PackageBuilder from './components/PackageBuilder';
import LoginPage from './components/LoginPage';
import SetupWizard from './pages/SetupWizard';
import ContainersPage from './pages/ContainersPage';
import ContainerDetailPage from './pages/ContainerDetailPage';
import ServicesPage from './pages/ServicesPage';
import ServiceDetailPage from './pages/ServiceDetailPage';
import ProtectedRoute from './components/ProtectedRoute';
import SetupGuard from './components/SetupGuard';
import PlaceholderPage from './components/PlaceholderPage';
import { AuthProvider } from './hooks/useAuth';
import { SetupProvider } from './hooks/useSetup';
import { BackendConnectionProvider } from './hooks/useBackendConnection';
import { ThemeProvider } from './hooks/useTheme';

function App() {
  return (
    <ThemeProvider>
      <BackendConnectionProvider>
        <SetupProvider>
          <AuthProvider>
            <Router>
              <Routes>
                {/* Setup route - always accessible during setup */}
                <Route path="/setup" element={<SetupWizard />} />
                {/* Public routes - guarded by setup status */}
                <Route
                  path="/login"
                  element={
                    <SetupGuard>
                      <ProtectedRoute requireAuth={false}>
                        <LoginPage />
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />

                {/* Protected routes - guarded by setup status */}
                <Route
                  path="/"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <Dashboard />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/dashboard"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <Dashboard />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/package"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <PackageBuilder />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/containers"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <ContainersPage />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/containers/:id"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <ContainerDetailPage />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/services"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <ServicesPage />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/services/:id"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <ServiceDetailPage />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/lint"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <PlaceholderPage
                            titleKey="navigation.lintReports"
                            descriptionKey="pages.lintReportsComingSoon"
                          />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/history"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <PlaceholderPage
                            titleKey="navigation.buildHistory"
                            descriptionKey="pages.buildHistoryComingSoon"
                          />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/cli"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <PlaceholderPage
                            titleKey="navigation.cliTools"
                            descriptionKey="pages.cliToolsComingSoon"
                          />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/rbac"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <PlaceholderPage
                            titleKey="navigation.rbacManager"
                            descriptionKey="pages.rbacComingSoon"
                          />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
                <Route
                  path="/settings"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <PlaceholderPage
                            titleKey="navigation.settings"
                            descriptionKey="pages.settingsComingSoon"
                          />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />

                {/* Fallback route - redirect to login if not authenticated */}
                <Route
                  path="*"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <Dashboard />
                        </DashboardLayout>
                      </ProtectedRoute>
                    </SetupGuard>
                  }
                />
              </Routes>
            </Router>
          </AuthProvider>
        </SetupProvider>
      </BackendConnectionProvider>
    </ThemeProvider>
  );
}

export default App;
