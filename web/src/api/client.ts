import axios from 'axios';
import type { APIResponse, TokenResponse } from '../types';

const API_BASE = '/api/v1';

const api = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
});

// Request interceptor: attach JWT token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: handle 401 + auto refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      const refreshToken = localStorage.getItem('refresh_token');
      if (refreshToken) {
        try {
          const res = await axios.post(`${API_BASE}/auth/refresh`, {
            refresh_token: refreshToken,
          });
          const { access_token, refresh_token } = res.data.data;
          localStorage.setItem('access_token', access_token);
          localStorage.setItem('refresh_token', refresh_token);
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return api(originalRequest);
        } catch {
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  login: (email: string, password: string) =>
    api.post<APIResponse<TokenResponse>>('/auth/login', { email, password }),
  register: (data: { email: string; password: string; name: string; role: string; branch_id: string }) =>
    api.post<APIResponse<TokenResponse>>('/auth/register', data),
  me: () => api.get<APIResponse<User>>('/auth/me'),
};

// Menu API
export const menuAPI = {
  list: (params?: { branch_id?: string; category?: string; search?: string; page?: number; limit?: number }) =>
    api.get<APIResponse<MenuItem[]>>('/menu', { params }),
  get: (id: string) => api.get<APIResponse<MenuItem>>(`/menu/${id}`),
  create: (data: Partial<MenuItem>) => api.post<APIResponse<MenuItem>>('/menu', data),
  update: (id: string, data: Partial<MenuItem>) => api.put<APIResponse<MenuItem>>(`/menu/${id}`, data),
  delete: (id: string) => api.delete(`/menu/${id}`),
  stats: (branchId?: string) => api.get<APIResponse<CategoryStats[]>>('/menu/stats', { params: { branch_id: branchId } }),
};

// Order API
export const orderAPI = {
  list: (params?: { branch_id?: string; status?: string; page?: number; limit?: number }) =>
    api.get<APIResponse<Order[]>>('/orders', { params }),
  get: (id: string) => api.get<APIResponse<Order>>(`/orders/${id}`),
  create: (data: { branch_id: string; customer_name: string; order_type: string; items: { menu_item_id: string; quantity: number; unit_price: number }[]; discount?: number; notes?: string }) =>
    api.post<APIResponse<Order>>('/orders', data),
  updateStatus: (id: string, status: string) =>
    api.patch<APIResponse<Order>>(`/orders/${id}/status`, { status }),
  cancel: (id: string) => api.post<APIResponse<Order>>(`/orders/${id}/cancel`),
};

// Inventory API
export const inventoryAPI = {
  list: (params?: { branch_id?: string; page?: number; limit?: number }) =>
    api.get<APIResponse<InventoryItem[]>>('/inventory', { params }),
  get: (id: string) => api.get<APIResponse<InventoryItem>>(`/inventory/${id}`),
  create: (data: Partial<InventoryItem>) => api.post<APIResponse<InventoryItem>>('/inventory', data),
  update: (id: string, data: Partial<InventoryItem>) => api.patch<APIResponse<InventoryItem>>(`/inventory/${id}`, data),
  alerts: (branchId?: string) => api.get<APIResponse<InventoryItem[]>>('/inventory/alerts', { params: { branch_id: branchId } }),
  reserve: (itemId: string, quantity: number) => api.post<APIResponse<InventoryItem>>('/inventory/reserve', { item_id: itemId, quantity }),
};

// Payment API
export const paymentAPI = {
  create: (data: { order_id: string; amount: number; payment_method: string }) =>
    api.post<APIResponse<unknown>>('/payments/create', data),
  getStatus: (orderId: string) => api.get<APIResponse<unknown>>(`/payments/${orderId}`),
};

// Types re-exported for convenience
import type { User, MenuItem, CategoryStats, Order, InventoryItem } from '../types';

export default api;
