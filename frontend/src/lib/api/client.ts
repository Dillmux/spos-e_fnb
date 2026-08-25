import type { ApiResponse } from '$lib/types';
import { toast } from '$lib/stores/toast';
import { browser } from '$app/environment';

const BASE_URL = (import.meta.env.PUBLIC_API_BASE_URL as string) || '/api/v1';

export class ApiError extends Error {
  status: number;
  data: any;
  constructor(message: string, status: number, data?: any) {
    super(message);
    this.status = status;
    this.data = data;
  }
}

export interface RequestOptions extends RequestInit {
  suppressToast?: boolean;
}

export async function request<T = any>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<ApiResponse<T>> {
  const { suppressToast, ...fetchOptions } = options;
  const token = browser ? localStorage.getItem('spos_jwt') : null;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(fetchOptions.headers as Record<string, string> || {}),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const url = endpoint.startsWith('http') ? endpoint : `${BASE_URL}${endpoint}`;

  try {
    const res = await fetch(url, {
      ...fetchOptions,
      headers,
    });

    const contentType = res.headers.get('content-type');
    let data: ApiResponse<T>;

    if (contentType && contentType.includes('application/json')) {
      data = await res.json();
    } else {
      const text = await res.text();
      data = {
        success: res.ok,
        message: text || res.statusText,
      };
    }

    if (!res.ok || !data.success) {
      const errMsg = data.message || `Request failed with status ${res.status}`;
      if (!suppressToast) {
        toast.error(errMsg);
      }
      throw new ApiError(errMsg, res.status, data);
    }

    return data;
  } catch (err: any) {
    if (err instanceof ApiError) {
      throw err;
    }
    const msg = err.message || 'Network communication error';
    if (!suppressToast) {
      toast.error(msg);
    }
    throw err;
  }
}
