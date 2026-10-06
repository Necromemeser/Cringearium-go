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
    <main className="chat-page">
      <aside className="chat-sidebar">
        <div className="chat-sidebar-header">
          <div className="chat-brand">Кринжик</div>
          <button type="button" className="chat-new" onClick={handleCreateChat} disabled={sending}>
            <span>＋</span>
            Новый чат
          </button>
        </div>

        <div className="chat-history">
          {chats.length === 0 ? (
            <div className="chat-history-empty">Здесь появятся ваши чаты</div>
          ) : (
            chats.map((chat) => (
              <button
                type="button"
                key={chat.id}
                className={`chat-item${activeChat === chat.id ? ' active' : ''}`}
                onClick={() => handleSelectChat(chat.id)}
              >
                <span className="chat-item-icon">◌</span>
                <span className="chat-item-name">{chat.chatName}</span>
              </button>
            ))
          )}
        </div>

        {activeChat !== null && (
          <button type="button" className="chat-delete" onClick={handleDeleteChat} disabled={sending}>
            Удалить текущий чат
          </button>
        )}
      </aside>

      <section className="chat-panel">
        <header className="chat-header">
          <div>
            <div className="chat-header-title">Кринжик</div>
            <div className="chat-header-subtitle">Персональный ИИ-помощник</div>
          </div>
        </header>

        {error && <div className="chat-error">{error}</div>}

        {activeChat === null ? (
          <div className="chat-empty">
            <div className="chat-empty-icon">✦</div>
            <h1>Чем могу помочь?</h1>
            <p>Разберём учебный материал, решим задачу или просто обсудим вопрос.</p>
            <button type="button" className="button button-primary" onClick={handleCreateChat}>Начать новый чат</button>
          </div>
        ) : (
          <>
            <div ref={messagesRef} className="chat-messages" onScroll={handleMessagesScroll}>
              <div className="chat-messages-inner">
                {messages.length === 0 && (
                  <div className="chat-welcome">
                    <div className="chat-welcome-icon">✦</div>
                    <h1>Чем могу помочь?</h1>
                    <p>Задайте вопрос Кринжику. Он поможет разобраться с учебными материалами.</p>
                  </div>
                )}

                {messages.map((message) => (
                  <div key={message.id} className={`chat-message ${message.isAiResponse ? 'assistant' : 'user'}`}>
                    {message.isAiResponse && <div className="chat-avatar">К</div>}
                    <div className="chat-message-body">
                      <div className="chat-message-author">{message.isAiResponse ? 'Кринжик' : 'Вы'}</div>
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
            </div>

            <form className="chat-input-wrap" onSubmit={(event) => { event.preventDefault(); void handleSend() }}>
              <div className="chat-input">
                <textarea
                  value={draft}
                  onChange={(event) => setDraft(event.target.value)}
                  onKeyDown={handleInputKeyDown}
                  placeholder="Сообщение для Кринжика..."
                  aria-label="Сообщение"
                  maxLength={32768}
                  rows={1}
                  disabled={sending}
                />
                <button type="submit" className="chat-send" disabled={!draft.trim() || sending} aria-label="Отправить сообщение">
                  ↑
                </button>
              </div>
              <div className="chat-input-hint">Enter — отправить · Shift + Enter — новая строка</div>
            </form>
          </>
        )}
      </section>
    </main>
  )
}
