import type {
  Diagram,
  DiagramMeta,
  ListDiagramsResponse,
  NextNameResponse,
  SessionResponse,
} from './types'

export const AUTH_EXPIRED_EVENT = 'online-mermaid:auth-expired'

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {})
  if (options.body && typeof options.body === 'string' && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(path, {
    ...options,
    headers,
    credentials: 'same-origin',
  })

  if (response.status === 401 && path !== '/api/auth/login' && path !== '/api/auth/session') {
    window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT))
  }

  const contentType = response.headers.get('Content-Type') || ''
  const isJson = contentType.includes('application/json')
  const payload = isJson ? await response.json().catch(() => null) : null

  if (!response.ok) {
    const message = payload?.message || `请求失败 (${response.status})`
    throw new Error(message)
  }

  return payload as T
}

export async function login(username: string, password: string): Promise<SessionResponse> {
  return request<SessionResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
}

export async function logout(): Promise<SessionResponse> {
  return request<SessionResponse>('/api/auth/logout', {
    method: 'POST',
  })
}

export async function getSession(): Promise<SessionResponse> {
  return request<SessionResponse>('/api/auth/session', {
    method: 'GET',
  })
}

export async function fetchDiagrams(): Promise<DiagramMeta[]> {
  const data = await request<ListDiagramsResponse>('/api/diagrams', { method: 'GET' })
  return data.items || []
}

export async function fetchNextName(): Promise<string> {
  const data = await request<NextNameResponse>('/api/diagrams/next-name', { method: 'GET' })
  return data.name || 'graph1'
}

export async function fetchDiagram(id: string): Promise<Diagram> {
  return request<Diagram>(`/api/diagrams/${encodeURIComponent(id)}`, { method: 'GET' })
}

export async function createDiagram(title?: string, code?: string): Promise<Diagram> {
  return request<Diagram>('/api/diagrams', {
    method: 'POST',
    body: JSON.stringify({ title, code }),
  })
}

export async function updateDiagram(id: string, title: string, code: string): Promise<Diagram> {
  return request<Diagram>(`/api/diagrams/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify({ title, code }),
  })
}

export async function deleteDiagram(id: string): Promise<{ deleted: string }> {
  return request<{ deleted: string }>(`/api/diagrams/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export async function duplicateDiagram(id: string): Promise<Diagram> {
  return request<Diagram>(`/api/diagrams/${encodeURIComponent(id)}/duplicate`, {
    method: 'POST',
  })
}
