export interface SetupStatus {
  is_completed: boolean
  requires_setup: boolean
  admin_exists: boolean
  setup_token?: string
}

export interface SetupConfig {
  company_name: string
  system_settings: Record<string, string>
}

export interface AdminCreateRequest {
  email: string
  password: string
  name: string
}

export interface AdminCreateResponse {
  id: number
  email: string
  name: string
  role: string
  active: boolean
  created_at: string
  updated_at: string
}

export interface SetupCompleteRequest {
  company_name: string
  system_settings: Record<string, string>
}

export interface SetupError {
  error: string
  message: string
}