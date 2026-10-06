import { useEffect, useState } from 'react'
import type { Chat, ChatMessage } from '../api/chat'
import { createChat, deleteChat, getChatMessages, listChats, streamChatMessage } from '../api/chat'
import { TOKEN_KEY } from '../constants/auth'
import MarkdownContent from '../components/chat/MarkdownContent'

export default function ChatPage() {
  const [chats, setChats] = useState<Chat[]>([])
  const [activeChat, setActiveChat] = useState<number | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [draft, setDraft] = useState('')
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!localStorage.getItem(TOKEN_KEY)) {
      setLoading(false)
      setError('Войдите в аккаунт, чтобы пользоваться Кринжиком.')
      return
    }

    listChats()
      .then(async (items) => {
        setChats(items)
        if (items.length > 0) {
          setActiveChat(items[0].id)
          setMessages(await getChatMessages(items[0].id))
        }
      })
      .catch((e) => setError(e instanceof Error ? e.message : 'Не удалось загрузить чаты'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (activeChat === null) return
    getChatMessages(activeChat)
      .then(setMessages)
      .catch((e) => setError(e instanceof Error ? e.message : 'Не удалось загрузить сообщения'))
  }, [activeChat])

  async function handleCreateChat() {
    try {
      setError('')
      const chat = await createChat()
      setChats((items) => [chat, ...items])
      setActiveChat(chat.id)
      setMessages([])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось создать чат')
    }
  }

  async function handleDeleteChat() {
    if (activeChat === null) return
    try {
      await deleteChat(activeChat)
      const remaining = chats.filter((chat) => chat.id !== activeChat)
      setChats(remaining)
      setActiveChat(remaining[0]?.id ?? null)
      setMessages(remaining.length ? await getChatMessages(remaining[0].id) : [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось удалить чат')
    }
  }

  async function handleSend() {
    const content = draft.trim()
    if (!content || activeChat === null || sending) return

    setError('')
    setDraft('')
    setSending(true)

    const temporaryId = Date.now()
    setMessages((items) => [...items, {
      id: temporaryId,
      content,
      isAiResponse: false,
      timestamp: new Date().toISOString(),
    }, {
      id: temporaryId + 1,
      content: '',
      isAiResponse: true,
      timestamp: new Date().toISOString(),
    }])

    try {
      await streamChatMessage(activeChat, content, (event) => {
        if (event.type === 'chunk') {
          setMessages((items) => {
            const copy = [...items]
            const last = copy[copy.length - 1]
            if (last?.isAiResponse) last.content += event.content
            return copy
          })
        } else if (event.type === 'error') {
          setError(event.message)
        }
      })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Не удалось отправить сообщение')
      setMessages((items) => items.slice(0, -2))
    } finally {
      setSending(false)
    }
  }

  function handleInputKeyDown(event: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key !== 'Enter' || event.shiftKey || event.nativeEvent.isComposing) return

    event.preventDefault()
    void handleSend()
  }

  async function handleSelectChat(id: number) {
    if (sending) return
    setActiveChat(id)
  }

  if (loading) return <main className="page"><div className="container loading-screen">Загружаем Кринжика...</div></main>

  return (
    <main className="page">
      <div className="container chat-page">
        <div className="page-heading">
          <span className="eyebrow">AI-АССИСТЕНТ</span>
          <h1>Кринжик</h1>
          <p>Поможет разобраться с учебными материалами и просто ответит на вопрос.</p>
        </div>

        {error && <div className="form-error">{error}</div>}

        <div className="chat-layout">
          <aside className="chat-sidebar profile-card">
            <button type="button" className="button button-primary chat-new" onClick={handleCreateChat} disabled={sending}>
              + Новый чат
            </button>
            {chats.map((chat) => (
              <button
                type="button"
                key={chat.id}
                className={`chat-item${activeChat === chat.id ? ' active' : ''}`}
                onClick={() => handleSelectChat(chat.id)}
              >
                {chat.chatName}
              </button>
            ))}
            {activeChat !== null && (
              <button type="button" className="chat-delete" onClick={handleDeleteChat} disabled={sending}>
                Удалить чат
              </button>
            )}
          </aside>

          <section className="profile-card chat-panel">
            {activeChat === null ? (
              <div className="empty-state">
                <div className="empty-icon">💬</div>
                <h3>Начните новый чат</h3>
                <button type="button" className="button button-primary" onClick={handleCreateChat}>Создать чат</button>
              </div>
            ) : (
              <>
                <div className="chat-messages">
                  {messages.map((message) => (
                    <div key={message.id} className={`chat-message ${message.isAiResponse ? 'assistant' : 'user'}`}>
                      <div className="chat-message-author">{message.isAiResponse ? 'Кринжик' : 'Вы'}</div>
                      <div className="chat-message-content">
                        {message.isAiResponse ? (
                          message.content ? <MarkdownContent content={message.content} /> : (sending ? '...' : '')
                        ) : (
                          message.content
                        )}
                      </div>
                    </div>
                  ))}
                </div>
                <form className="chat-input" onSubmit={(event) => { event.preventDefault(); void handleSend() }}>
                  <textarea
                    value={draft}
                    onChange={(event) => setDraft(event.target.value)}
                    onKeyDown={handleInputKeyDown}
                    placeholder="Напишите сообщение..."
                    aria-label="Сообщение"
                    maxLength={32768}
                    rows={3}
                    disabled={sending}
                  />
                  <button type="submit" className="button button-primary" disabled={!draft.trim() || sending}>
                    {sending ? 'Кринжик думает...' : 'Отправить'}
                  </button>
                </form>
              </>
            )}
          </section>
        </div>
      </div>
    </main>
  )
}
