import axios, { type AxiosError, type AxiosResponse } from 'axios'

interface ApiEnvelope<T> {
  code?: number
  message?: string
  data?: T
}

const configuredBase = String(import.meta.env.VITE_API_BASE_URL || '/api/v1').trim().replace(/\/+$/, '')

export const api = axios.create({
  baseURL: configuredBase || '/api/v1',
  withCredentials: true,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token')
  if (token && token !== 'smirel-preview-token' && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  if (config.method === 'get') {
    config.params = { ...(config.params || {}), timezone: Intl.DateTimeFormat().resolvedOptions().timeZone }
  }
  return config
})

api.interceptors.response.use((response: AxiosResponse) => {
  const payload = response.data as ApiEnvelope<unknown>
  if (payload && typeof payload === 'object' && typeof payload.code === 'number') {
    if (payload.code !== 0) {
      return Promise.reject(new Error(payload.message || '请求失败'))
    }
    response.data = payload.data
  }
  return response
})

export function getErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<ApiEnvelope<unknown> | Record<string, unknown>>
    const body = axiosError.response?.data
    if (body && typeof body === 'object' && 'message' in body && typeof body.message === 'string') return body.message
    if (!axiosError.response) return '无法连接到服务，请稍后重试'
  }
  return error instanceof Error ? error.message : '请求失败，请稍后重试'
}

export const previewMode = import.meta.env.VITE_UI_PREVIEW === 'true'
