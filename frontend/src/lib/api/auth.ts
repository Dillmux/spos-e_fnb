import { request } from './client';
import type { User } from '$lib/types';

export const authApi = {
  login: async (username: string, password_hash: string) => {
    return request<{ token: string; user: User }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password: password_hash }),
    });
  },

  verifyPin: async (pin: string) => {
    return request<{ valid?: boolean; verified_by?: string; role?: string; supervisor_name?: string }>('/auth/verify-pin', {
      method: 'POST',
      body: JSON.stringify({ pin }),
    });
  },

  getMe: async () => {
    return request<User>('/auth/me');
  },

  refreshToken: async () => {
    return request<{ token: string }>('/auth/refresh', {
      method: 'POST',
    });
  },

  getUsers: async () => {
    return request<User[]>('/auth/users');
  },

  createUser: async (payload: Partial<User> & { password?: string; auth_pin?: string }) => {
    return request<User>('/auth/users', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  updateUser: async (id: string, payload: Partial<User> & { password?: string; auth_pin?: string }) => {
    return request<User>(`/auth/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
  },

  toggleUserStatus: async (id: string, is_active: boolean) => {
    return request<User>(`/auth/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ is_active }),
    });
  },

  deleteUser: async (id: string) => {
    return request<{ id: string }>(`/auth/users/${id}`, {
      method: 'DELETE',
    });
  },
};
