import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import DashboardLayout from './components/DashboardLayout';
import Dashboard from './pages/Dashboard';
import ComposeMerger from './components/ComposeMerger';
import PackageBuilder from './components/PackageBuilder';
import LoginPage from './components/LoginPage';
import SetupWizard from './pages/SetupWizard';
import ProtectedRoute from './components/ProtectedRoute';
import SetupGuard from './components/SetupGuard';
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
                  path="/merge"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <ComposeMerger />
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
                  path="/lint"
                  element={
                    <SetupGuard>
                      <ProtectedRoute>
                        <DashboardLayout>
                          <div className="p-6 bg-card rounded-lg shadow border border-border">
                            <h2 className="text-2xl font-bold mb-4 text-foreground">
                              Lint Reports
                            </h2>
                            <p className="text-muted-foreground">
                              Lint reports feature coming soon...
                            </p>
                          </div>
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
                          <div className="p-6 bg-card rounded-lg shadow border border-border">
                            <h2 className="text-2xl font-bold mb-4 text-foreground">
                              Build History
                            </h2>
                            <p className="text-muted-foreground">
                              Build history feature coming soon...
                            </p>
                          </div>
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
                          <div className="p-6 bg-card rounded-lg shadow border border-border">
                            <h2 className="text-2xl font-bold mb-4 text-foreground">CLI Tools</h2>
                            <p className="text-muted-foreground">
                              CLI tools documentation coming soon...
                            </p>
                          </div>
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
                          <div className="p-6 bg-card rounded-lg shadow border border-border">
                            <h2 className="text-2xl font-bold mb-4 text-foreground">
                              RBAC Manager
                            </h2>
                            <p className="text-muted-foreground">
                              Role-based access control management coming soon...
                            </p>
                          </div>
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
                          <div className="p-6 bg-card rounded-lg shadow border border-border">
                            <h2 className="text-2xl font-bold mb-4 text-foreground">Settings</h2>
                            <p className="text-muted-foreground">Settings page coming soon...</p>
                          </div>
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
