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
