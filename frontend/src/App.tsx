import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import ComposeMerger from './components/ComposeMerger'
import PackageBuilder from './components/PackageBuilder'
import { AuthProvider } from './hooks/useAuth'

function App() {
  return (
    <AuthProvider>
      <Router>
        <Layout>
          <Routes>
            <Route path="/" element={<ComposeMerger />} />
            <Route path="/merge" element={<ComposeMerger />} />
            <Route path="/package" element={<PackageBuilder />} />
          </Routes>
        </Layout>
      </Router>
    </AuthProvider>
  )
}

export default App