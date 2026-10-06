import { useEffect, useRef, useState } from 'react'
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
  const messagesRef = useRef<HTMLDivElement>(null)
  const shouldAutoScrollRef = useRef(true)

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

  function handleMessagesScroll() {
    const container = messagesRef.current
    if (!container) return

    const distanceFromBottom = container.scrollHeight - container.scrollTop - container.clientHeight
    shouldAutoScrollRef.current = distanceFromBottom < 80
  }

  useEffect(() => {
    const container = messagesRef.current
    if (!container || !shouldAutoScrollRef.current) return

    container.scrollTop = container.scrollHeight
  }, [messages])

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

  async function handleDeleteChat(chatId: number) {
    if (sending) return

    try {
      setError('')
      await deleteChat(chatId)

      const remaining = chats.filter((chat) => chat.id !== chatId)
      setChats(remaining)

      if (activeChat !== chatId) return

      const nextChat = remaining[0]
      setActiveChat(nextChat?.id ?? null)
      setMessages(nextChat ? await getChatMessages(nextChat.id) : [])
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

  function handleSelectChat(id: number) {
    if (sending) return
    setActiveChat(id)
  }

  if (loading) return <main className="page"><div className="container loading-screen">Загружаем Кринжика...</div></main>

  return (
    <main className="page">
      <div className="container chat-page">
        {error && (
          <div className={error.startsWith("I'm sowwy") ? 'chat-quota-notice' : 'form-error'} role="alert">
            {error.startsWith("I'm sowwy") ? (
              <>
                <div className="chat-quota-icon" aria-hidden="true">🥺</div>
                <div>
                  <strong>Кринжик на сегодня всё</strong>
                  <p>{error}</p>
                </div>
              </>
            ) : (
              error
            )}
          </div>
        )}

        <div className="chat-layout">
          <aside className="chat-sidebar profile-card">
            <button type="button" className="button button-primary chat-new" onClick={handleCreateChat} disabled={sending}>
              + Новый чат
            </button>
            {chats.map((chat) => (
              <div key={chat.id} className={`chat-item-wrapper${activeChat === chat.id ? ' active' : ''}`}>
                <button
                  type="button"
                  className="chat-item"
                  onClick={() => handleSelectChat(chat.id)}
                >
                  <span className="chat-item-name">{chat.chatName}</span>
                </button>
                <button
                  type="button"
                  className="chat-delete"
                  onClick={() => void handleDeleteChat(chat.id)}
                  disabled={sending}
                  aria-label={`Удалить чат «${chat.chatName}»`}
                  title="Удалить чат"
                >
                  <svg viewBox="0 0 24 24" aria-hidden="true">
                    <path d="M9 3h6l1 2h4v2H4V5h4l1-2Zm-3 6h12l-.8 11.2a2 2 0 0 1-2 1.8H8.8a2 2 0 0 1-2-1.8L6 9Zm4 2v7h2v-7h-2Zm4 0v7h2v-7h-2Z" />
                  </svg>
                </button>
              </div>
            ))}
          </aside>

          <section className="profile-card chat-panel">
            <header className="chat-header">
              <div className="chat-avatar">К</div>
              <div>
                <strong>Кринжик</strong>
                <span>Персональный ИИ-помощник</span>
              </div>
            </header>

            {activeChat === null ? (
              <div className="empty-state">
                <div className="empty-icon">💬</div>
                <h3>Начните новый чат</h3>
                <button type="button" className="button button-primary" onClick={handleCreateChat}>Создать чат</button>
              </div>
            ) : (
              <>
                <div ref={messagesRef} className="chat-messages" onScroll={handleMessagesScroll}>
                  {messages.map((message) => (
                    <div key={message.id} className={`chat-message ${message.isAiResponse ? 'assistant' : 'user'}`}>
                      {message.isAiResponse && <div className="chat-message-avatar">К</div>}
                      <div className="chat-message-body">
                        {message.isAiResponse && <div className="chat-message-author">Кринжик</div>}
                        <div className="chat-message-content">
                          {message.isAiResponse ? (
                            message.content ? <MarkdownContent content={message.content} /> : (sending ? '...' : '')
                          ) : (
                            message.content
                          )}
                        </div>
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
                    rows={1}
                    disabled={sending}
                  />
                  <button type="submit" className="chat-send" disabled={!draft.trim() || sending} aria-label="Отправить сообщение">
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                      <path d="M4.5 19.5 21 12 4.5 4.5 4 10l11 2-11 2 .5 5.5Z" />
                    </svg>
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
