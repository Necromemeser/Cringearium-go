import type { ElementType, ReactNode } from 'react'

type Block =
  | { type: 'code'; content: string; language?: string }
  | { type: 'heading'; level: number; content: string }
  | { type: 'list'; ordered: boolean; items: string[] }
  | { type: 'quote'; content: string }
  | { type: 'paragraph'; content: string }

function parseBlocks(markdown: string): Block[] {
  const lines = markdown.replace(/\r\n/g, '\n').split('\n')
  const blocks: Block[] = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    if (!line.trim()) {
      i++
      continue
    }

    const fence = line.match(/^\s*\`\`\`(.*)$/)
    if (fence) {
      const language = fence[1].trim() || undefined
      const code: string[] = []
      i++

      while (i < lines.length && !/^\s*\`\`\`\s*$/.test(lines[i])) {
        code.push(lines[i])
        i++
      }

      if (i < lines.length) i++
      blocks.push({ type: 'code', content: code.join('\n'), language })
      continue
    }

    const heading = line.match(/^\s*(#{1,6})\s+(.+?)\s*#*\s*$/)
    if (heading) {
      blocks.push({ type: 'heading', level: heading[1].length, content: heading[2] })
      i++
      continue
    }

    if (/^\s*>/.test(line)) {
      const quote: string[] = []
      while (i < lines.length && /^\s*>/.test(lines[i])) {
        quote.push(lines[i].replace(/^\s*>\s?/, ''))
        i++
      }
      blocks.push({ type: 'quote', content: quote.join('\n') })
      continue
    }

    const listMatch = line.match(/^\s*([-*+]|\d+[.)])\s+(.+)$/)
    if (listMatch) {
      const ordered = /^\d/.test(listMatch[1])
      const items: string[] = []

      while (i < lines.length) {
        const item = lines[i].match(/^\s*([-*+]|\d+[.)])\s+(.+)$/)
        if (!item || /^\d/.test(item[1]) !== ordered) break
        items.push(item[2])
        i++
      }

      blocks.push({ type: 'list', ordered, items })
      continue
    }

    const paragraph: string[] = [line]
    i++

    while (i < lines.length && lines[i].trim()) {
      if (
        /^\s*\`\`\`/.test(lines[i]) ||
        /^\s*#{1,6}\s+/.test(lines[i]) ||
        /^\s*>/.test(lines[i]) ||
        /^\s*([-*+]|\d+[.)])\s+/.test(lines[i])
      ) {
        break
      }

      paragraph.push(lines[i])
      i++
    }

    blocks.push({ type: 'paragraph', content: paragraph.join('\n') })
  }

  return blocks
}

function renderInline(text: string): ReactNode[] {
  const tokenPattern = /(\`[^\`]+\`|\*\*[^*]+\*\*|__[^_]+__|~~[^~]+~~|\*[^*]+\*|_[^_]+_|\[[^\]]+\]\([^\)]+\))/g
  const parts = text.split(tokenPattern)

  return parts.map((part, index) => {
    if (!part) return null

    if (part.startsWith('\`') && part.endsWith('\`')) {
      return <code key={index}>{part.slice(1, -1)}</code>
    }

    if ((part.startsWith('**') && part.endsWith('**')) || (part.startsWith('__') && part.endsWith('__'))) {
      return <strong key={index}>{part.slice(2, -2)}</strong>
    }

    if (part.startsWith('~~') && part.endsWith('~~')) {
      return <del key={index}>{part.slice(2, -2)}</del>
    }

    if ((part.startsWith('*') && part.endsWith('*')) || (part.startsWith('_') && part.endsWith('_'))) {
      return <em key={index}>{part.slice(1, -1)}</em>
    }

    const link = part.match(/^\[([^\]]+)\]\(([^\)]+)\)$/)
    if (link) {
      return (
        <a key={index} href={link[2]} target="_blank" rel="noreferrer">
          {link[1]}
        </a>
      )
    }

    return part
  })
}

export default function MarkdownContent({ content }: { content: string }) {
  const blocks = parseBlocks(content)

  return (
    <div className="markdown-content">
      {blocks.map((block, index) => {
        switch (block.type) {
          case 'code':
            return (
              <pre key={index}>
                {block.language && <div className="markdown-code-language">{block.language}</div>}
                <code>{block.content}</code>
              </pre>
            )
          case 'heading': {
            const Tag = `h${block.level}` as ElementType
            return <Tag key={index}>{renderInline(block.content)}</Tag>
          }
          case 'list': {
            const Tag = block.ordered ? 'ol' : 'ul'
            return (
              <Tag key={index}>
                {block.items.map((item, itemIndex) => (
                  <li key={itemIndex}>{renderInline(item)}</li>
                ))}
              </Tag>
            )
          }
          case 'quote':
            return <blockquote key={index}>{renderInline(block.content)}</blockquote>
          case 'paragraph':
            return <p key={index}>{renderInline(block.content)}</p>
        }
      })}
    </div>
  )
}
