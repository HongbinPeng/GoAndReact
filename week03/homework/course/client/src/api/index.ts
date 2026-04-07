import http from './http'
import { getToken } from '../utils/auth'
import type { Course, DashboardData, PaginationData, Student } from '../types'

export const api = {
  login: (payload: { username: string; password: string }) =>
    http.post<{ token: string; user: { id: number; username: string; name: string } }>('/auth/login', payload),

  getDashboard: () => http.get<DashboardData>('/dashboard'),

  getCourseList: (params: Record<string, unknown>) =>
    http.get<PaginationData<Course>>('/courses', { params }),
  getCourseCategories: () => http.get<string[]>('/courses/categories'),
  createCourse: (payload: Partial<Course>) => http.post<Course>('/courses', payload),
  updateCourse: (id: number, payload: Partial<Course>) => http.put<Course>(`/courses/${id}`, payload),
  deleteCourse: (id: number) => http.delete(`/courses/${id}`),
  toggleCourseStatus: (id: number) => http.patch<Course>(`/courses/${id}/status`),
  getStudentList: (params: Record<string, unknown>) =>
    http.get<PaginationData<Student>>('/students', { params }),
  getClasses: () => http.get<string[]>('/students/classes'),
  createStudent: (payload: Partial<Student>) => http.post<Student>('/students', payload),
  updateStudent: (id: number, payload: Partial<Student>) => http.put<Student>(`/students/${id}`, payload),
  deleteStudent: (id: number) => http.delete(`/students/${id}`),

  getSummary: () => http.get<{ content: string }>('/summary'),

  /**
   * 流式 AI 对话 (SSE)
   * @param payload - { question, sessionId }
   * @param callbacks - { onStart, onChunk, onDone, onError }
   */
  chatWithAiStream: async (
    payload: { question: string; sessionId?: string },
    callbacks: {
      onStart?: (data: { sessionId: string; question: string }) => void
      onChunk?: (data: { increment: string; fullContent: string }) => void
      onDone?: (data: { trace: unknown[]; sqlInfo?: unknown }) => void
      onError?: (error: Error) => void
    }
  ) => {
    const token = getToken()
    const response = await fetch('/api/ai/chat/stream', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: token ? `Bearer ${token}` : '',
      },
      body: JSON.stringify(payload),
    })

    if (!response.ok) {
      const text = await response.text()
      let msg = `请求失败: ${response.status}`
      try {
        const data = JSON.parse(text)
        msg = data?.msg || msg
      } catch {}
      callbacks.onError?.(new Error(msg))
      return
    }

    const reader = response.body?.getReader()
    if (!reader) {
      callbacks.onError?.(new Error('无法获取响应流'))
      return
    }

    const decoder = new TextDecoder('utf-8')
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })

      // 解析 SSE 格式
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        const trimmed = line.trim()
        if (!trimmed) continue

        if (trimmed.startsWith('event:')) {
          // 事件类型行，跳过等待 data 行
          continue
        }

        if (trimmed.startsWith('data:')) {
          const dataStr = trimmed.slice(5).trim()
          try {
            const data = JSON.parse(dataStr)
            // 根据数据字段判断事件类型
            if (data.sessionId && data.question) {
              callbacks.onStart?.(data)
            } else if (data.increment !== undefined) {
              callbacks.onChunk?.(data)
            } else if (data.trace) {
              callbacks.onDone?.(data)
            } else if (data.message) {
              callbacks.onError?.(new Error(data.message))
            }
          } catch {
            // 忽略解析失败
          }
        }
      }
    }
  },
}