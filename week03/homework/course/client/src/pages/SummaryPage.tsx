import { Card, Typography } from 'antd'
import { useEffect, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import rehypeHighlight from 'rehype-highlight'
import remarkGfm from 'remark-gfm'
import { api } from '../api'

export default function SummaryPage() {
  const [content, setContent] = useState('')

  useEffect(() => {
    api.getSummary().then((res) => setContent(res.content))
  }, [])

  return (
    <div className="page-wrap">
      <div className="page-head">
        <Typography.Title level={4}>学习总结</Typography.Title>
      </div>
      <Card className="panel-card">
      <div className="markdown-body">
        <ReactMarkdown
          remarkPlugins={[remarkGfm]}
          rehypePlugins={[rehypeHighlight]}
          components={{
            img: ({ src = '', alt = '' }) => {
              const finalSrc = src.startsWith('http') || src.startsWith('/api/static/')
                ? src
                : `/api/static/${src.replace(/^\/+/, '')}`
              return <img src={finalSrc} alt={alt} style={{ maxWidth: '100%' }} />
            },
          }}
        >
          {content}
        </ReactMarkdown>
      </div>
      </Card>
    </div>
  )
}
