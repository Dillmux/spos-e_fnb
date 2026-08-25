import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import type { User } from '$lib/types';
import { authApi } from '$lib/api/auth';

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
}

const initialToken = browser ? localStorage.getItem('spos_jwt') : null;
const initialUserStr = browser ? localStorage.getItem('spos_user') : null;
let initialUser: User | null = null;
if (initialUserStr) {
  try {
    initialUser = JSON.parse(initialUserStr);
  } catch {}
}

const state = writable<AuthState>({
  user: initialUser,
  token: initialToken,
  isLoading: true,
});

// Global Pin Modal Request State
export interface PinRequest {
  title: string;
  onSuccess: (pin: string) => void;
  onCancel: () => void;
}

export const activePinRequest = writable<PinRequest | null>(null);

export const auth = {
  subscribe: state.subscribe,

  setToken: (token: string | null) => {
    if (browser) {
      if (token) {
        localStorage.setItem('spos_jwt', token);
      } else {
        localStorage.removeItem('spos_jwt');
      }
    }
    state.update((s) => ({ ...s, token }));
  },

  setUser: (user: User | null) => {
    if (browser) {
      if (user) {
        localStorage.setItem('spos_user', JSON.stringify(user));
      } else {
        localStorage.removeItem('spos_user');
      }
    }
    state.update((s) => ({ ...s, user, isLoading: false }));
  },

  initSession: async () => {
    if (!browser) return;
    const token = localStorage.getItem('spos_jwt');
    if (!token) {
      auth.logout();
      state.update((s) => ({ ...s, isLoading: false }));
      return;
    }

    try {
      const res = await authApi.getMe();
      const userData = (res.data as any)?.user || res.data;
      if (userData && userData.id) {
        auth.setUser(userData);
      } else {
        auth.logout();
      }
    } catch (err: any) {
      if (err?.status === 401 || err?.status === 403) {
        auth.logout();
      }
    } finally {
      state.update((s) => ({ ...s, isLoading: false }));
    }
  },

  login: async (username: string, password: string) => {
    state.update((s) => ({ ...s, isLoading: true }));
    try {
      const res = await authApi.login(username, password);
      if (res.data?.token && res.data?.user) {
        auth.setToken(res.data.token);
        auth.setUser(res.data.user);
        return res.data;
      }
    } finally {
      state.update((s) => ({ ...s, isLoading: false }));
    }
  },

  logout: () => {
    auth.setToken(null);
    auth.setUser(null);
  },

  refreshToken: async () => {
    try {
      const res = await authApi.refreshToken();
      if (res.data?.token) {
        auth.setToken(res.data.token);
        return res.data.token;
      }
    } catch {}
    return null;
  },

  // Trigger global supervisor PIN modal
  requestSupervisorPin: (title: string): Promise<string> => {
    return new Promise((resolve, reject) => {
      activePinRequest.set({
        title,
        onSuccess: (pin: string) => {
          activePinRequest.set(null);
          resolve(pin);
        },
        onCancel: () => {
          activePinRequest.set(null);
          reject(new Error('PIN authorization cancelled'));
        },
      });
    });
  },
};

export const isAuthenticated = derived(state, ($s) => !!$s.token && !!$s.user);
export const currentUser = derived(state, ($s) => $s.user);
export const userRole = derived(state, ($s) => $s.user?.role || null);
