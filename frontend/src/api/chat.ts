import { TOKEN_KEY } from '../constants/auth'

export type Chat = {
  id: number
  chatName: string
  createdAt: string
  updatedAt: string
}

export type ChatMessage = {
  id: number
  content: string
  isAiResponse: boolean
  timestamp: string
  userId?: number
}

const API_BASE = '/api'

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem(TOKEN_KEY)
  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers ?? {}),
    },
  })

  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || `Request failed with status ${response.status}`)
  }

  if (response.status === 204) return undefined as T
  return (await response.json()) as T
}

export function listChats(): Promise<Chat[]> {
  return request<Chat[]>('/chats')
}

export function createChat(name?: string): Promise<Chat> {
  return request<Chat>('/chats', {
    method: 'POST',
    body: JSON.stringify(name ? { name } : {}),
  })
}

export function getChatMessages(chatId: number): Promise<ChatMessage[]> {
  return request<ChatMessage[]>(`/chats/${chatId}/messages`)
}

export async function deleteChat(chatId: number): Promise<void> {
  await request<void>(`/chats/${chatId}`, { method: 'DELETE' })
}

export type StreamEvent =
  | { type: 'chunk'; content: string }
  | { type: 'done' }
  | { type: 'error'; message: string }

export async function streamChatMessage(
  chatId: number,
  content: string,
  onEvent: (event: StreamEvent) => void,
): Promise<void> {
  const token = localStorage.getItem(TOKEN_KEY)
  const response = await fetch(`${API_BASE}/chats/${chatId}/messages`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ content }),
  })

  if (!response.ok) {
    const message = (await response.text()).trim()
    throw new Error(message || `Request failed with status ${response.status}`)
  }

  if (!response.body) throw new Error('Streaming is not supported by this browser')

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const events = buffer.split('\n\n')
    buffer = events.pop() ?? ''

    for (const raw of events) {
      const line = raw.split('\n').find((item) => item.startsWith('data:'))
      if (!line) continue
      const data = JSON.parse(line.slice(5).trim()) as StreamEvent
      onEvent(data)
      if (data.type === 'done' || data.type === 'error') return
    }
  }
}
