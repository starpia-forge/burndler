import { useState, useEffect } from 'react'
import { useAuth } from '../hooks/useAuth'
import { Build } from '../types'
import api from '../services/api'

export default function PackageBuilder() {
  const { isDeveloper } = useAuth()
  const [packageName, setPackageName] = useState('')
  const [composeContent, setComposeContent] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [currentBuild, setCurrentBuild] = useState<Build | null>(null)
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null)

  useEffect(() => {
    return () => {
      if (pollingInterval) {
        clearInterval(pollingInterval)
      }
    }
  }, [pollingInterval])

  const pollBuildStatus = async (buildId: string) => {
    try {
      const status = await api.getBuildStatus(buildId)
      setCurrentBuild(status)

      if (status.status === 'completed' || status.status === 'failed') {
        if (pollingInterval) {
          clearInterval(pollingInterval)
          setPollingInterval(null)
        }
      }
    } catch (err) {
      console.error('Failed to poll build status:', err)
    }
  }

  const handleCreatePackage = async () => {
    if (!isDeveloper) {
      setError('Only Developers can create packages')
      return
    }

    setIsLoading(true)
    setError(null)
    setCurrentBuild(null)

    try {
      const result = await api.createPackage({
        name: packageName,
        compose: composeContent,
      })

      // Start polling for build status
      const interval = setInterval(() => {
        pollBuildStatus(result.build_id)
      }, 2000)
      setPollingInterval(interval)

      // Initial status fetch
      pollBuildStatus(result.build_id)
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create package')
    } finally {
      setIsLoading(false)
    }
  }

  if (!isDeveloper) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6">
        <h3 className="text-lg font-medium text-yellow-800 mb-2">
          Developer Access Required
        </h3>
        <p className="text-yellow-700">
          Package creation is restricted to users with Developer role.
          Engineers have read-only access to the system.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Package Builder</h2>
        <p className="mt-1 text-sm text-gray-600">
          Create offline installer packages with Docker images and resources
        </p>
      </div>

      {/* Package Configuration */}
      <div className="bg-white p-6 rounded-lg shadow space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Package Name
          </label>
          <input
            type="text"
            value={packageName}
            onChange={(e) => setPackageName(e.target.value)}
            placeholder="my-application"
            className="w-full px-3 py-2 border rounded-md"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Merged Compose File
          </label>
          <textarea
            value={composeContent}
            onChange={(e) => setComposeContent(e.target.value)}
            placeholder="Paste your merged docker-compose.yaml content here..."
            rows={15}
            className="w-full px-3 py-2 border rounded-md font-mono text-sm"
          />
        </div>

        <div className="flex justify-end">
          <button
            onClick={handleCreatePackage}
            disabled={isLoading || !packageName || !composeContent}
            className="px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Creating...' : 'Create Package'}
          </button>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      {/* Build Status */}
      {currentBuild && (
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Build Status</h3>

          <div className="space-y-3">
            <div className="flex justify-between">
              <span className="text-gray-600">Build ID:</span>
              <span className="font-mono text-sm">{currentBuild.id}</span>
            </div>

            <div className="flex justify-between">
              <span className="text-gray-600">Status:</span>
              <span
                className={`px-2 py-1 text-xs rounded-full ${
                  currentBuild.status === 'completed'
                    ? 'bg-green-100 text-green-800'
                    : currentBuild.status === 'failed'
                    ? 'bg-red-100 text-red-800'
                    : currentBuild.status === 'building'
                    ? 'bg-blue-100 text-blue-800'
                    : 'bg-gray-100 text-gray-800'
                }`}
              >
                {currentBuild.status}
              </span>
            </div>

            {currentBuild.progress > 0 && (
              <div>
                <div className="flex justify-between text-sm text-gray-600 mb-1">
                  <span>Progress</span>
                  <span>{currentBuild.progress}%</span>
                </div>
                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className="bg-primary-600 h-2 rounded-full transition-all duration-300"
                    style={{ width: `${currentBuild.progress}%` }}
                  />
                </div>
              </div>
            )}

            {currentBuild.error && (
              <div className="bg-red-50 border border-red-200 p-3 rounded">
                <p className="text-sm text-red-700">{currentBuild.error}</p>
              </div>
            )}

            {currentBuild.downloadUrl && (
              <div className="bg-green-50 border border-green-200 p-3 rounded">
                <p className="text-sm text-green-700 mb-2">
                  Package created successfully!
                </p>
                <a
                  href={currentBuild.downloadUrl}
                  download
                  className="inline-flex items-center px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700"
                >
                  Download Package
                </a>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}