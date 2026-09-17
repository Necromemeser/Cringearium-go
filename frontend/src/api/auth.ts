export type User = {
  id: number
  username: string
  email: string
  role: string
  profile_image_id?: string
}

type LoginResponse = {
  token: string
}

const API_BASE = '/api'

async function parseError(response: Response): Promise<never> {
  const message = (await response.text()).trim()
  throw new Error(message || `Request failed with status ${response.status}`)
}

export async function login(email: string, password: string): Promise<string> {
  const response = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })

  if (!response.ok) {
    return parseError(response)
  }

  const data = (await response.json()) as LoginResponse
  return data.token
}

export async function register(username: string, email: string, password: string): Promise<User> {
  const response = await fetch(`${API_BASE}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password }),
  })

  if (!response.ok) {
    return parseError(response)
  }

  return (await response.json()) as User
}

export async function getMe(token: string): Promise<User> {
  const response = await fetch(`${API_BASE}/auth/me`, {
    headers: { Authorization: `Bearer ${token}` },
  })

  if (!response.ok) {
    return parseError(response)
  }

  return (await response.json()) as User
}
