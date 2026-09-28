export interface SessionResponse {
  authenticated: boolean
  username?: string
  version?: string
}

export interface DiagramMeta {
  id: string
  title: string
  createdAt: string
  updatedAt: string
  sortOrder: number
}

export interface Diagram extends DiagramMeta {
  code: string
}

export interface ListDiagramsResponse {
  items: DiagramMeta[]
}

export interface NextNameResponse {
  name: string
}

export interface ApiError {
  code: string
  message: string
}

export interface UpdateCheckResponse {
  current_version: string
  latest_version: string
  has_update: boolean
  release_title: string
  release_notes: string
  published_at: string
  release_url: string
  asset_url?: string
  asset_name?: string
  can_auto_update: boolean
  update_message?: string
  manual_command?: string
  os: string
  arch: string
}

export interface UpdateApplyResponse {
  success: boolean
  message: string
}
