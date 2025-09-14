export interface User {
  id: string
  email: string
  name: string
  role: 'Developer' | 'Engineer'
}

export interface Module {
  name: string
  compose: string
  variables?: Record<string, string>
}

export interface MergeRequest {
  modules: Module[]
  projectVariables?: Record<string, string>
}

export interface MergeResult {
  mergedCompose: string
  mappings: Record<string, string>
  warnings: string[]
}

export interface LintRequest {
  compose: string
  strictMode?: boolean
}

export interface LintResult {
  valid: boolean
  errors: LintIssue[]
  warnings: LintIssue[]
}

export interface LintIssue {
  rule: string
  message: string
  line?: number
}

export interface PackageRequest {
  name: string
  compose: string
  resources?: Resource[]
}

export interface Resource {
  module: string
  version: string
  files: string[]
}

export interface Build {
  id: string
  name: string
  status: 'queued' | 'building' | 'completed' | 'failed'
  progress: number
  downloadUrl?: string
  error?: string
  createdAt: string
  completedAt?: string
}