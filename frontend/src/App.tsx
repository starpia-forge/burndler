import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import DashboardLayout from './components/DashboardLayout'
import Dashboard from './pages/Dashboard'
import ComposeMerger from './components/ComposeMerger'
import PackageBuilder from './components/PackageBuilder'
import { AuthProvider } from './hooks/useAuth'

function App() {
  return (
    <AuthProvider>
      <Router>
        <DashboardLayout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/merge" element={<ComposeMerger />} />
            <Route path="/package" element={<PackageBuilder />} />
            <Route path="/lint" element={<div className="p-6 bg-white rounded-lg shadow"><h2 className="text-2xl font-bold mb-4">Lint Reports</h2><p className="text-gray-600">Lint reports feature coming soon...</p></div>} />
            <Route path="/history" element={<div className="p-6 bg-white rounded-lg shadow"><h2 className="text-2xl font-bold mb-4">Build History</h2><p className="text-gray-600">Build history feature coming soon...</p></div>} />
            <Route path="/cli" element={<div className="p-6 bg-white rounded-lg shadow"><h2 className="text-2xl font-bold mb-4">CLI Tools</h2><p className="text-gray-600">CLI tools documentation coming soon...</p></div>} />
            <Route path="/rbac" element={<div className="p-6 bg-white rounded-lg shadow"><h2 className="text-2xl font-bold mb-4">RBAC Manager</h2><p className="text-gray-600">Role-based access control management coming soon...</p></div>} />
            <Route path="/settings" element={<div className="p-6 bg-white rounded-lg shadow"><h2 className="text-2xl font-bold mb-4">Settings</h2><p className="text-gray-600">Settings page coming soon...</p></div>} />
          </Routes>
        </DashboardLayout>
      </Router>
    </AuthProvider>
  )
}

export default App