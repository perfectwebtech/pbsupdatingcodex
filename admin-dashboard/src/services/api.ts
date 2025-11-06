import axios, { AxiosInstance, AxiosError } from 'axios';
import toast from 'react-hot-toast';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:80';

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as any;

    // Handle 401 errors
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          const response = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
            refresh_token: refreshToken,
          });

          const { access_token, refresh_token: newRefreshToken } = response.data.data;
          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', newRefreshToken);

          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return api(originalRequest);
        }
      } catch (refreshError) {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // Show error toast
    const message = error.response?.data?.error || error.message || 'An error occurred';
    toast.error(message);

    return Promise.reject(error);
  }
);

// Authentication API
export const authAPI = {
  login: async (username: string, password: string) => {
    const response = await api.post('/api/v1/auth/login', { username, password });
    return response.data;
  },

  register: async (data: { username: string; password: string; email: string }) => {
    const response = await api.post('/api/v1/auth/register', data);
    return response.data;
  },

  logout: async (refreshToken: string) => {
    const response = await api.post('/api/v1/auth/logout', { refresh_token: refreshToken });
    return response.data;
  },

  getCurrentUser: async () => {
    const response = await api.get('/api/v1/auth/me');
    return response.data;
  },

  updateProfile: async (data: { email?: string }) => {
    const response = await api.put('/api/v1/auth/profile', data);
    return response.data;
  },

  changePassword: async (oldPassword: string, newPassword: string) => {
    const response = await api.post('/api/v1/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    });
    return response.data;
  },
};

// Users API (Admin)
export const usersAPI = {
  getAll: async (params?: { page?: number; limit?: number; search?: string }) => {
    const response = await api.get('/api/v1/admin/users', { params });
    return response.data;
  },

  getById: async (id: number) => {
    const response = await api.get(`/api/v1/admin/users/${id}`);
    return response.data;
  },

  create: async (data: any) => {
    const response = await api.post('/api/v1/admin/users', data);
    return response.data;
  },

  update: async (id: number, data: any) => {
    const response = await api.put(`/api/v1/admin/users/${id}`, data);
    return response.data;
  },

  delete: async (id: number) => {
    const response = await api.delete(`/api/v1/admin/users/${id}`);
    return response.data;
  },
};

// Streams API
export const streamsAPI = {
  getAll: async (params?: { type?: string; category_id?: number; search?: string }) => {
    const response = await api.get('/api/v1/streams', { params });
    return response.data;
  },

  getById: async (id: number) => {
    const response = await api.get(`/api/v1/streams/${id}`);
    return response.data;
  },

  create: async (data: any) => {
    const response = await api.post('/api/v1/admin/streams', data);
    return response.data;
  },

  update: async (id: number, data: any) => {
    const response = await api.put(`/api/v1/admin/streams/${id}`, data);
    return response.data;
  },

  delete: async (id: number) => {
    const response = await api.delete(`/api/v1/admin/streams/${id}`);
    return response.data;
  },

  search: async (query: string) => {
    const response = await api.get('/api/v1/streams/search', { params: { q: query } });
    return response.data;
  },
};

// Categories API
export const categoriesAPI = {
  getAll: async () => {
    const response = await api.get('/api/v1/categories');
    return response.data;
  },

  create: async (data: { name: string; icon_url?: string }) => {
    const response = await api.post('/api/v1/admin/categories', data);
    return response.data;
  },

  update: async (id: number, data: any) => {
    const response = await api.put(`/api/v1/admin/categories/${id}`, data);
    return response.data;
  },

  delete: async (id: number) => {
    const response = await api.delete(`/api/v1/admin/categories/${id}`);
    return response.data;
  },
};

// Analytics API
export const analyticsAPI = {
  getDashboardStats: async () => {
    const response = await api.get('/api/v1/admin/analytics/dashboard');
    return response.data;
  },

  getStreamAnalytics: async (streamId: number, period: string = '7d') => {
    const response = await api.get(`/api/v1/admin/analytics/streams/${streamId}`, {
      params: { period },
    });
    return response.data;
  },

  getUserAnalytics: async (period: string = '7d') => {
    const response = await api.get('/api/v1/admin/analytics/users', {
      params: { period },
    });
    return response.data;
  },

  getRevenue: async (period: string = '30d') => {
    const response = await api.get('/api/v1/admin/analytics/revenue', {
      params: { period },
    });
    return response.data;
  },
};

// Sessions API
export const sessionsAPI = {
  getActive: async () => {
    const response = await api.get('/api/v1/admin/sessions/active');
    return response.data;
  },

  terminate: async (sessionId: string) => {
    const response = await api.post(`/api/v1/admin/sessions/${sessionId}/terminate`);
    return response.data;
  },
};

// Packages API
export const packagesAPI = {
  getAll: async () => {
    const response = await api.get('/api/v1/packages');
    return response.data;
  },

  create: async (data: any) => {
    const response = await api.post('/api/v1/admin/packages', data);
    return response.data;
  },

  update: async (id: number, data: any) => {
    const response = await api.put(`/api/v1/admin/packages/${id}`, data);
    return response.data;
  },

  delete: async (id: number) => {
    const response = await api.delete(`/api/v1/admin/packages/${id}`);
    return response.data;
  },
};

// Transcoding API
export const transcodingAPI = {
  getAllJobs: async (params?: { status?: string; limit?: number }) => {
    const response = await api.get('/api/v1/transcode/jobs', { params });
    return response.data;
  },

  getJob: async (jobId: string) => {
    const response = await api.get(`/api/v1/transcode/jobs/${jobId}/status`);
    return response.data;
  },

  createJob: async (data: { input_file: string; preset: string }) => {
    const response = await api.post('/api/v1/transcode', data);
    return response.data;
  },

  cancelJob: async (jobId: string) => {
    const response = await api.post(`/api/v1/transcode/jobs/${jobId}/cancel`);
    return response.data;
  },
};

export default api;
