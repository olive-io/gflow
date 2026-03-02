import { useUserStore } from '@/stores'

const BASE_URL = '/v1'

export interface ApiResponse<T> {
  data?: T
  error?: string
  message?: string
}

export interface ApiError {
  code: number
  message: string
  details?: string
}

class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string = BASE_URL) {
    this.baseUrl = baseUrl
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }

    const userStore = useUserStore()
    if (userStore.token) {
      headers['Authorization'] = `Bearer ${userStore.token}`
    }

    return headers
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      if (response.status === 401) {
        const userStore = useUserStore()
        userStore.logout()
        window.location.href = '/login'
        throw new Error('Unauthorized')
      }

      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `HTTP Error: ${response.status}`)
    }

    if (response.status === 204) {
      return {} as T
    }

    return response.json()
  }

  async get<T>(path: string, params?: Record<string, string | number | boolean>): Promise<T> {
    let url = `${this.baseUrl}${path}`
    if (params) {
      const searchParams = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value))
        }
      })
      const queryString = searchParams.toString()
      if (queryString) {
        url += `?${queryString}`
      }
    }

    const response = await fetch(url, {
      method: 'GET',
      headers: this.getHeaders(),
    })

    return this.handleResponse<T>(response)
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    })

    return this.handleResponse<T>(response)
  }

  async put<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    })

    return this.handleResponse<T>(response)
  }

  async patch<T>(path: string, body?: unknown): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'PATCH',
      headers: this.getHeaders(),
      body: body ? JSON.stringify(body) : undefined,
    })

    return this.handleResponse<T>(response)
  }

  async delete<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      method: 'DELETE',
      headers: this.getHeaders(),
    })

    return this.handleResponse<T>(response)
  }
}

export const apiClient = new ApiClient()
