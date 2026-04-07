import { SendOutlined, SyncOutlined, ClearOutlined } from '@ant-design/icons'
import { Alert, Avatar, Button, Card, Collapse, Empty, Input, Space, Typography, Tag } from 'antd'
import { useEffect, useRef, useState } from 'react'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { api } from '../api'
import lingxiAvatar from '../assets/灵犀.svg'
import userAvatar from '../assets/user.png'

type ChatRole = 'user' | 'assistant'

type ChatMessage = {
  id: string
  role: ChatRole
  content: string
  trace?: TraceItem[]
  isStreaming?: boolean
}

type TraceItem = {
  step: string
  detail: string
  sql: string | null
  at: string
}

export default function AiChatPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputValue, setInputValue] = useState('')
  const [sending, setSending] = useState(false)
  const [errorText, setErrorText] = useState('')
  const [lastQuestion, setLastQuestion] = useState('')
  const [sessionId, setSessionId] = useState('')
  const [showTrace, setShowTrace] = useState(false)
  const messageListRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    if (messageListRef.current) {
      messageListRef.current.scrollTop = messageListRef.current.scrollHeight
    }
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages, sending, errorText])

  useEffect(() => {
    document.body.classList.add('ai-page-lock')
    return () => {
      document.body.classList.remove('ai-page-lock')
    }
  }, [])

  const sendQuestion = async (rawQuestion: string, appendUser: boolean) => {
    const question = rawQuestion.trim()
    if (!question || sending) return

    setErrorText('')
    setLastQuestion(question)

    if (appendUser) {
      setMessages((prev) => [
        ...prev,
        { id: `user-${Date.now()}`, role: 'user', content: question },
      ])
      setInputValue('')
    }

    // 先添加一个空的 AI 消息，标记为流式中
    const aiMsgId = `assistant-${Date.now()}`
    setMessages((prev) => [
      ...prev,
      { id: aiMsgId, role: 'assistant', content: '', isStreaming: true },
    ])
    setSending(true)

    try {
      await api.chatWithAiStream(
        { question, sessionId: sessionId || undefined },
        {
          onStart: (data) => {
            setSessionId(data.sessionId || sessionId)
          },
          onChunk: (data) => {
            // 实时更新消息内容
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === aiMsgId
                  ? { ...msg, content: data.fullContent, isStreaming: true }
                  : msg
              )
            )
          },
          onDone: (data) => {
            // 流式完成，更新 trace
            setMessages((prev) =>
              prev.map((msg) =>
                msg.id === aiMsgId
                  ? { ...msg, isStreaming: false, trace: data.trace as TraceItem[] }
                  : msg
              )
            )
            setShowTrace(true)
            setSending(false)
          },
          onError: (error) => {
            setErrorText(error.message)
            // 移除空的 AI 消息
            setMessages((prev) => prev.filter((msg) => msg.id !== aiMsgId))
            setSending(false)
          },
        }
      )
    } catch (error) {
      setErrorText(error instanceof Error ? error.message : '发送失败，请稍后重试')
      setMessages((prev) => prev.filter((msg) => msg.id !== aiMsgId))
      setSending(false)
    }
  }

  const clearMessages = () => {
    setMessages([])
    setSessionId('')
    setShowTrace(false)
    setErrorText('')
  }

  return (
    <div className="ai-chat-page">
      <Card
        className="ai-chat-card"
        title={
          <Space>
            <span className="ai-chat-title-icon">🤖</span>
            <span>AI 智能助手</span>
            <Tag color="blue">流式</Tag>
          </Space>
        }
        extra={
          <Space>
            <Button
              size="small"
              icon={<ClearOutlined />}
              onClick={clearMessages}
              disabled={messages.length === 0}
            >
              清空对话
            </Button>
            <Button
              size="small"
              icon={<SyncOutlined spin={sending} />}
              onClick={() => setShowTrace(!showTrace)}
              disabled={messages.length === 0}
            >
              {showTrace ? '隐藏' : '查看'} 执行链路
            </Button>
          </Space>
        }
      >
        <div className="ai-chat-wrap">
          <div className="ai-chat-messages" ref={messageListRef}>
            {messages.length === 0 ? (
              <div className="ai-chat-empty">
                <Empty 
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description={
                    <div className="ai-chat-empty-text">
                      <p>开始与 AI 助手对话吧！</p>
                      <p className="ai-chat-hint">
                        试试问我："有多少门课程？" 或 "学生总数是多少？"
                      </p>
                    </div>
                  }
                />
              </div>
            ) : (
              messages.map((message) => (
                <div
                  key={message.id}
                  className={`ai-chat-item ${message.role === 'user' ? 'ai-chat-item-user' : 'ai-chat-item-assistant'}`}
                >
                  <Avatar
                    className="ai-chat-avatar"
                    src={message.role === 'user' ? userAvatar : lingxiAvatar}
                    style={{ marginBottom: 8, backgroundColor: 'transparent' }}
                    shape="circle"
                    size={40}
                  >
                    {null}
                  </Avatar>
                  <div className="ai-chat-content">
                    <div className="ai-chat-header">
                      <Typography.Text className="ai-chat-role">
                        {message.role === 'user' ? '我' : 'AI 助手'}
                      </Typography.Text>
                      {message.isStreaming && (
                        <Tag color="processing" className="streaming-tag">
                          <span className="streaming-dot" /> 生成中
                        </Tag>
                      )}
                    </div>
                    {message.role === 'assistant' ? (
                      <div className="ai-chat-bubble markdown-body">
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                          {message.content || (message.isStreaming ? '正在思考...' : '')}
                        </ReactMarkdown>
                        {message.isStreaming && <span className="streaming-indicator">▊</span>}
                      </div>
                    ) : (
                      <div className="ai-chat-bubble user-bubble">
                        <Typography.Text>{message.content}</Typography.Text>
                      </div>
                    )}
                    {showTrace && message.trace && message.trace.length > 0 && (
                      <div className="ai-chat-trace">
                        <div className="trace-header">
                          <SyncOutlined spin={false} />
                          <Typography.Text type="secondary" style={{ fontSize: 12, marginLeft: 8 }}>
                            执行链路 ({message.trace.length} 步)
                          </Typography.Text>
                        </div>
                        <Collapse
                          size="small"
                          bordered={false}
                          items={message.trace.map((item, idx) => ({
                            key: idx,
                            label: (
                              <Space>
                                <Tag color={idx === 0 ? 'blue' : idx === 1 ? 'purple' : 'default'}>
                                  {idx + 1}
                                </Tag>
                                <span>{item.step}</span>
                                {item.sql && <Tag color="orange">SQL</Tag>}
                              </Space>
                            ),
                            children: (
                              <div className="trace-content">
                                <div className="trace-detail">{item.detail}</div>
                                {item.sql && (
                                  <pre className="trace-sql">
                                    <code>{item.sql}</code>
                                  </pre>
                                )}
                                <div className="trace-time">
                                  {new Date(item.at).toLocaleTimeString()}
                                </div>
                              </div>
                            ),
                          }))}
                        />
                      </div>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>

          <div className="ai-chat-input-area">
            {errorText ? (
              <Alert
                type="error"
                showIcon
                message={errorText}
                action={
                  <Button
                    size="small"
                    icon={<SyncOutlined />}
                    onClick={() => sendQuestion(lastQuestion, false)}
                    disabled={!lastQuestion || sending}
                  >
                    重试
                  </Button>
                }
                style={{ marginBottom: 12 }}
              />
            ) : null}
            <div className="input-wrapper">
              <Input.TextArea
                className="chat-input"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                autoSize={{ minRows: 2, maxRows: 6 }}
                placeholder="请输入问题，按 Enter 发送，Shift + Enter 换行"
                onPressEnter={(e) => {
                  if (e.shiftKey) return
                  e.preventDefault()
                  void sendQuestion(inputValue, true)
                }}
                disabled={sending}
              />
              <Button
                className="send-button"
                type="primary"
                size="large"
                icon={<SendOutlined />}
                loading={sending}
                onClick={() => sendQuestion(inputValue, true)}
                disabled={!inputValue.trim() || sending}
              >
                发送
              </Button>
            </div>
            <div className="input-tips">
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                支持查询课程、学生、学习记录等相关信息
              </Typography.Text>
            </div>
          </div>
        </div>
      </Card>
    </div>
  )
}
