import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import DashboardLayout from './components/DashboardLayout'
import Dashboard from './pages/Dashboard'
import ComposeMerger from './components/ComposeMerger'
import PackageBuilder from './components/PackageBuilder'
import LoginPage from './components/LoginPage'
import ProtectedRoute from './components/ProtectedRoute'
import { AuthProvider } from './hooks/useAuth'
import { ThemeProvider } from './hooks/useTheme'

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <Router>
          <Routes>
            {/* Public routes */}
            <Route
              path="/login"
              element={
                <ProtectedRoute requireAuth={false}>
                  <LoginPage />
                </ProtectedRoute>
              }
            />

            {/* Protected routes */}
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <Dashboard />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <Dashboard />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/merge"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <ComposeMerger />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/package"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <PackageBuilder />
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/lint"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <div className="p-6 bg-card rounded-lg shadow border border-border">
                      <h2 className="text-2xl font-bold mb-4 text-foreground">Lint Reports</h2>
                      <p className="text-muted-foreground">Lint reports feature coming soon...</p>
                    </div>
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/history"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <div className="p-6 bg-card rounded-lg shadow border border-border">
                      <h2 className="text-2xl font-bold mb-4 text-foreground">Build History</h2>
                      <p className="text-muted-foreground">Build history feature coming soon...</p>
                    </div>
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/cli"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <div className="p-6 bg-card rounded-lg shadow border border-border">
                      <h2 className="text-2xl font-bold mb-4 text-foreground">CLI Tools</h2>
                      <p className="text-muted-foreground">CLI tools documentation coming soon...</p>
                    </div>
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/rbac"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <div className="p-6 bg-card rounded-lg shadow border border-border">
                      <h2 className="text-2xl font-bold mb-4 text-foreground">RBAC Manager</h2>
                      <p className="text-muted-foreground">Role-based access control management coming soon...</p>
                    </div>
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings"
              element={
                <ProtectedRoute>
                  <DashboardLayout>
                    <div className="p-6 bg-card rounded-lg shadow border border-border">
                      <h2 className="text-2xl font-bold mb-4 text-foreground">Settings</h2>
                      <p className="text-muted-foreground">Settings page coming soon...</p>
                    </div>
                  </DashboardLayout>
                </ProtectedRoute>
              }
            />
          </Routes>
        </Router>
      </AuthProvider>
    </ThemeProvider>
  )
}

export default App