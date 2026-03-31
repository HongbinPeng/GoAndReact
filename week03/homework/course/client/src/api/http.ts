import axios from 'axios'
import { message } from 'antd'
import { clearToken, getToken } from '../utils/auth'
import type { ApiResponse } from '../types'

const http = axios.create({
  baseURL: '/api',
  timeout: 10000,
})

http.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    if (status === 401) {
      clearToken()
      if (location.pathname !== '/login') {
        message.error('登录已过期，请重新登录')
        location.href = '/login'
      }
    } else {
      message.error(error.response?.data?.msg || error.message || '网络错误')
    }
    return Promise.reject(error)
  }
)

async function unwrap<T>(promise: Promise<{ data: ApiResponse<T> }>) {
  const res = await promise
  if (res.data.code !== 0) {
    message.error(res.data.msg || '请求失败')
    throw new Error(res.data.msg || 'Request failed')
  }
  return res.data.data
}

const request = {
  get: <T>(url: string, config?: object) => unwrap<T>(http.get(url, config)),
  post: <T>(url: string, data?: object, config?: object) => unwrap<T>(http.post(url, data, config)),
  put: <T>(url: string, data?: object, config?: object) => unwrap<T>(http.put(url, data, config)),
  patch: <T>(url: string, data?: object, config?: object) => unwrap<T>(http.patch(url, data, config)),
  delete: <T>(url: string, config?: object) => unwrap<T>(http.delete(url, config)),
}

export default request
