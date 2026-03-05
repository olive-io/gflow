import axios from 'axios';
import type { AxiosInstance } from 'axios';
import NProgress from 'nprogress';
import 'nprogress/nprogress.css';
import { useUserStore } from '../stores/user';

NProgress.configure({ showSpinner: false, speed: 500 });

const BASE_URL = '/v1';

let requestsCounter = 0;

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.client.interceptors.request.use((config) => {
      if (requestsCounter === 0) {
        NProgress.start();
      }
      requestsCounter++;

      const token = useUserStore.getState().token;
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    this.client.interceptors.response.use(
      (response) => {
        requestsCounter--;
        if (requestsCounter === 0) {
          NProgress.done();
        }
        return response;
      },
      (error) => {
        requestsCounter--;
        if (requestsCounter === 0) {
          NProgress.done();
        }

        if (error.response?.status === 401) {
          useUserStore.getState().logout();
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  async get<T>(path: string, params?: Record<string, unknown>): Promise<T> {
    const response = await this.client.get<T>(path, { params });
    return response.data;
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    const response = await this.client.post<T>(path, body);
    return response.data;
  }

  async put<T>(path: string, body?: unknown): Promise<T> {
    const response = await this.client.put<T>(path, body);
    return response.data;
  }

  async patch<T>(path: string, body?: unknown): Promise<T> {
    const response = await this.client.patch<T>(path, body);
    return response.data;
  }

  async delete<T>(path: string, params?: Record<string, unknown>): Promise<T> {
    const response = await this.client.delete<T>(path, { params });
    return response.data;
  }
}

export const apiClient = new ApiClient();
