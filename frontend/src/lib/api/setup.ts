import { request } from './client';
import type { SetupStatus, SetupPayload } from '$lib/types';

export const setupApi = {
  getStatus: async () => {
    return request<SetupStatus>('/setup/status');
  },

  testDatabase: async (driver: string, connectionString?: string) => {
    return request<{ connected: boolean; latency_ms?: number }>('/setup/test-db', {
      method: 'POST',
      body: JSON.stringify({ driver, connection_string: connectionString }),
    });
  },

  testMidtrans: async (serverKey: string, clientKey: string) => {
    return request<{ valid: boolean }>('/setup/test-midtrans', {
      method: 'POST',
      body: JSON.stringify({ server_key: serverKey, client_key: clientKey }),
    });
  },

  initialize: async (payload: SetupPayload) => {
    const formattedBody = {
      outlet: {
        name: payload.outlet_name,
        tagline: payload.outlet_tagline,
        address: payload.outlet_address,
        phone: payload.outlet_phone,
        business_type: payload.business_type,
        service_modes: payload.service_modes,
        tax_percentage: payload.tax_percentage,
        service_percentage: payload.service_percentage,
        is_tax_inclusive: payload.is_tax_inclusive,
      },
      user: {
        name: payload.admin_name,
        username: payload.admin_username,
        email: `${payload.admin_username}@spos.local`,
        password: payload.admin_password,
        auth_pin: payload.admin_pin,
      },
      ...payload,
    };
    return request<{ initialized: boolean; token?: string; user_id?: string; username?: string; role?: string }>('/setup/initialize', {
      method: 'POST',
      body: JSON.stringify(formattedBody),
    });
  },
};
