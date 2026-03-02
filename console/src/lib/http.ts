class HttpClient {
  private baseURL: string

  constructor() {
    this.baseURL = import.meta.env.VITE_API_BASE_URL || ''
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }
    const token = localStorage.getItem('token')
    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }
    return headers
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      if (response.status === 401) {
        localStorage.removeItem('token')
        localStorage.removeItem('user')
        window.location.href = '/login'
        throw new Error('Unauthorized')
      }
      const error = await response.text()
      throw new Error(error || `HTTP Error: ${response.status}`)
    }
    const text = await response.text()
    if (!text) {
      return {} as T
    }
    return JSON.parse(text) as T
  }

  async get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
    const fullURL = new URL(url, window.location.origin)
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          fullURL.searchParams.append(key, String(value))
        }
      })
    }
    const response = await fetch(fullURL.toString(), {
      method: 'GET',
      headers: this.getHeaders(),
    })
    return this.handleResponse<T>(response)
  }

  async post<T>(url: string, data?: unknown): Promise<T> {
    const response = await fetch(url, {
      method: 'POST',
      headers: this.getHeaders(),
      body: data ? JSON.stringify(data) : undefined,
    })
    return this.handleResponse<T>(response)
  }

  async put<T>(url: string, data?: unknown): Promise<T> {
    const response = await fetch(url, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: data ? JSON.stringify(data) : undefined,
    })
    return this.handleResponse<T>(response)
  }

  async patch<T>(url: string, data?: unknown): Promise<T> {
    const response = await fetch(url, {
      method: 'PATCH',
      headers: this.getHeaders(),
      body: data ? JSON.stringify(data) : undefined,
    })
    return this.handleResponse<T>(response)
  }

  async delete<T>(url: string): Promise<T> {
    const response = await fetch(url, {
      method: 'DELETE',
      headers: this.getHeaders(),
    })
    return this.handleResponse<T>(response)
  }
}

export const http = new HttpClient()
