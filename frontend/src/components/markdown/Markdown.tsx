import type { ReactNode } from 'react'

function renderInlineMarkdown(text: string): ReactNode[] {
  const pattern = /(\*\*([^*]+)\*\*|__([^_]+)__|\x60([^\x60]+)\x60|\[([^\]]+)\]\(([^)]+)\)|(\*|_)([^*_]+)\7)/g
  const result: ReactNode[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null
  let key = 0
  while ((match = pattern.exec(text)) !== null) {
    if (match.index > lastIndex) result.push(text.slice(lastIndex, match.index))
    if (match[2] || match[3]) result.push(<strong key={key++}>{match[2] || match[3]}</strong>)
    else if (match[4]) result.push(<code key={key++}>{match[4]}</code>)
    else if (match[5] && match[6]) result.push(<a key={key++} href={match[6]} target="_blank" rel="noreferrer">{match[5]}</a>)
    else if (match[8]) result.push(<em key={key++}>{match[8]}</em>)
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < text.length) result.push(text.slice(lastIndex))
  return result
}

export default function Markdown({ content }: { content: string }) {
  const lines = content.replace(/\r\n?/g, '\n').split('\n')
  const blocks: ReactNode[] = []
  let paragraph: string[] = []
  let list: string[] = []
  let orderedList = false
  let codeBlock: string[] | null = null
  let codeLanguage = ''
  const flushParagraph = () => {
    if (!paragraph.length) return
    blocks.push(<p key={`p-${blocks.length}`}>{renderInlineMarkdown(paragraph.join(' '))}</p>)
    paragraph = []
  }
  const flushList = () => {
    if (!list.length) return
    const items = list.map((item, index) => <li key={index}>{renderInlineMarkdown(item)}</li>)
    blocks.push(orderedList ? <ol key={`ol-${blocks.length}`}>{items}</ol> : <ul key={`ul-${blocks.length}`}>{items}</ul>)
    list = []
  }
  lines.forEach((line) => {
    const trimmed = line.trim()
    if (codeBlock) {
      if (trimmed.startsWith('```')) {
        blocks.push(
          <pre key={`code-${blocks.length}`}>
            {codeLanguage && <div className="markdown-code-language">{codeLanguage}</div>}
            <code>{codeBlock.join('\n')}</code>
          </pre>,
        )
        codeBlock = null
        codeLanguage = ''
      } else {
        codeBlock.push(line)
      }
      return
    }
    if (trimmed.startsWith('```')) {
      flushParagraph()
      flushList()
      codeBlock = []
      codeLanguage = trimmed.slice(3).trim()
      return
    }
    if (!trimmed) { flushParagraph(); flushList(); return }
    const image = /^!\\[([^\\]]*)\\]\\(([^)]+)\\)$/.exec(trimmed)
    if (image) {
      flushParagraph()
      flushList()
      blocks.push(
        <figure className="markdown-image" key={`img-${blocks.length}`}>
          <img src={image[2]} alt={image[1]} loading="lazy" />
          {image[1] && <figcaption>{image[1]}</figcaption>}
        </figure>,
      )
      return
    }
    const heading = /^(#{1,6})\s+(.+)$/.exec(trimmed)
    if (heading) {
      flushParagraph(); flushList()
      const level = Math.min(heading[1].length, 4)
      const text = renderInlineMarkdown(heading[2])
      if (level === 1) blocks.push(<h3 key={`h-${blocks.length}`}>{text}</h3>)
      else if (level === 2) blocks.push(<h4 key={`h-${blocks.length}`}>{text}</h4>)
      else blocks.push(<h5 key={`h-${blocks.length}`}>{text}</h5>)
      return
    }
    const unordered = /^[-*+]\s+(.+)$/.exec(trimmed)
    const ordered = /^\d+[.)]\s+(.+)$/.exec(trimmed)
    if (unordered || ordered) {
      flushParagraph()
      if (list.length && orderedList !== Boolean(ordered)) flushList()
      orderedList = Boolean(ordered)
      list.push((unordered || ordered)![1])
      return
    }
    if (trimmed.startsWith('> ')) { flushParagraph(); flushList(); blocks.push(<blockquote key={`q-${blocks.length}`}>{renderInlineMarkdown(trimmed.slice(2))}</blockquote>); return }
    if (/^---+$/.test(trimmed)) { flushParagraph(); flushList(); blocks.push(<hr key={`hr-${blocks.length}`} />); return }
    paragraph.push(trimmed)
  })
  if (codeBlock) {
    blocks.push(
      <pre key={`code-${blocks.length}`}>
        {codeLanguage && <div className="markdown-code-language">{codeLanguage}</div>}
        <code>{codeBlock.join('\n')}</code>
      </pre>,
    )
  }
  flushParagraph(); flushList()
  return <div className="markdown-content">{blocks.length ? blocks : <p>Материал этой страницы пока не добавлен.</p>}</div>
}
