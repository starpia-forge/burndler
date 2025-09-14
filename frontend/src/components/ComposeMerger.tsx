import { useState } from 'react'
import { Module, MergeResult, LintResult } from '../types'
import api from '../services/api'

export default function ComposeMerger() {
  const [modules, setModules] = useState<Module[]>([
    { name: 'module1', compose: '' },
  ])
  const [projectVariables, setProjectVariables] = useState<Record<string, string>>({})
  const [mergeResult, setMergeResult] = useState<MergeResult | null>(null)
  const [lintResult, setLintResult] = useState<LintResult | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const addModule = () => {
    setModules([...modules, { name: `module${modules.length + 1}`, compose: '' }])
  }

  const removeModule = (index: number) => {
    setModules(modules.filter((_, i) => i !== index))
  }

  const updateModule = (index: number, field: keyof Module, value: string) => {
    const updated = [...modules]
    updated[index] = { ...updated[index], [field]: value }
    setModules(updated)
  }

  const handleMerge = async () => {
    setIsLoading(true)
    setError(null)
    setMergeResult(null)
    setLintResult(null)

    try {
      // Merge compose files
      const result = await api.mergeCompose({
        modules,
        projectVariables,
      })
      setMergeResult(result)

      // Lint the merged result
      if (result.mergedCompose) {
        const lintRes = await api.lintCompose({
          compose: result.mergedCompose,
          strictMode: true,
        })
        setLintResult(lintRes)
      }
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to merge compose files')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Compose Merger</h2>
        <p className="mt-1 text-sm text-gray-600">
          Merge multiple Docker Compose files with namespace prefixing and variable substitution
        </p>
      </div>

      {/* Modules */}
      <div className="space-y-4">
        <div className="flex justify-between items-center">
          <h3 className="text-lg font-medium text-gray-900">Modules</h3>
          <button
            onClick={addModule}
            className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
          >
            Add Module
          </button>
        </div>

        {modules.map((module, index) => (
          <div key={index} className="bg-white p-4 rounded-lg shadow space-y-3">
            <div className="flex justify-between items-center">
              <input
                type="text"
                value={module.name}
                onChange={(e) => updateModule(index, 'name', e.target.value)}
                placeholder="Module name (namespace)"
                className="px-3 py-2 border rounded-md flex-1 mr-3"
              />
              {modules.length > 1 && (
                <button
                  onClick={() => removeModule(index)}
                  className="px-3 py-2 text-red-600 hover:bg-red-50 rounded-md"
                >
                  Remove
                </button>
              )}
            </div>
            <textarea
              value={module.compose}
              onChange={(e) => updateModule(index, 'compose', e.target.value)}
              placeholder="Paste your docker-compose.yaml content here..."
              rows={10}
              className="w-full px-3 py-2 border rounded-md font-mono text-sm"
            />
          </div>
        ))}
      </div>

      {/* Project Variables */}
      <div className="bg-white p-4 rounded-lg shadow">
        <h3 className="text-lg font-medium text-gray-900 mb-3">
          Project Variables (Optional)
        </h3>
        <textarea
          value={Object.entries(projectVariables)
            .map(([k, v]) => `${k}=${v}`)
            .join('\n')}
          onChange={(e) => {
            const vars: Record<string, string> = {}
            e.target.value.split('\n').forEach((line) => {
              const [key, value] = line.split('=')
              if (key && value) {
                vars[key.trim()] = value.trim()
              }
            })
            setProjectVariables(vars)
          }}
          placeholder="KEY=value (one per line)"
          rows={5}
          className="w-full px-3 py-2 border rounded-md font-mono text-sm"
        />
      </div>

      {/* Actions */}
      <div className="flex justify-end">
        <button
          onClick={handleMerge}
          disabled={isLoading || modules.every((m) => !m.compose)}
          className="px-6 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? 'Merging...' : 'Merge & Lint'}
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      {/* Merge Result */}
      {mergeResult && (
        <div className="bg-white p-4 rounded-lg shadow space-y-4">
          <h3 className="text-lg font-medium text-gray-900">Merge Result</h3>

          {/* Warnings */}
          {mergeResult.warnings.length > 0 && (
            <div className="bg-yellow-50 border border-yellow-200 p-3 rounded">
              <h4 className="font-medium text-yellow-800 mb-2">Warnings</h4>
              <ul className="list-disc list-inside text-sm text-yellow-700">
                {mergeResult.warnings.map((warning, i) => (
                  <li key={i}>{warning}</li>
                ))}
              </ul>
            </div>
          )}

          {/* Mappings */}
          {Object.keys(mergeResult.mappings).length > 0 && (
            <div>
              <h4 className="font-medium text-gray-700 mb-2">Name Mappings</h4>
              <div className="bg-gray-50 p-3 rounded font-mono text-sm">
                {Object.entries(mergeResult.mappings).map(([old, new_]) => (
                  <div key={old}>
                    {old} → {new_}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Merged Compose */}
          <div>
            <h4 className="font-medium text-gray-700 mb-2">Merged Compose</h4>
            <textarea
              value={mergeResult.mergedCompose}
              readOnly
              rows={15}
              className="w-full px-3 py-2 border rounded-md font-mono text-sm bg-gray-50"
            />
          </div>
        </div>
      )}

      {/* Lint Result */}
      {lintResult && (
        <div className="bg-white p-4 rounded-lg shadow space-y-4">
          <h3 className="text-lg font-medium text-gray-900">
            Lint Result: {lintResult.valid ? '✅ Valid' : '❌ Invalid'}
          </h3>

          {/* Errors */}
          {lintResult.errors.length > 0 && (
            <div className="bg-red-50 border border-red-200 p-3 rounded">
              <h4 className="font-medium text-red-800 mb-2">Errors</h4>
              <ul className="space-y-1">
                {lintResult.errors.map((error, i) => (
                  <li key={i} className="text-sm text-red-700">
                    <span className="font-medium">[{error.rule}]</span> {error.message}
                    {error.line && <span className="text-xs"> (line {error.line})</span>}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Warnings */}
          {lintResult.warnings.length > 0 && (
            <div className="bg-yellow-50 border border-yellow-200 p-3 rounded">
              <h4 className="font-medium text-yellow-800 mb-2">Warnings</h4>
              <ul className="space-y-1">
                {lintResult.warnings.map((warning, i) => (
                  <li key={i} className="text-sm text-yellow-700">
                    <span className="font-medium">[{warning.rule}]</span> {warning.message}
                    {warning.line && <span className="text-xs"> (line {warning.line})</span>}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  )
}