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

  getSettings: async () => {
    return request<{
      outlet: any;
      feature_flags: Record<string, boolean>;
      midtrans_server_key: string;
      midtrans_client_key: string;
      midtrans_is_production: boolean;
    }>('/setup/settings');
  },

  updateSettings: async (payload: {
    supervisor_pin: string;
    outlet_name?: string;
    outlet_tagline?: string;
    outlet_address?: string;
    outlet_phone?: string;
    receipt_footer?: string;
    tax_percentage?: number;
    service_percentage?: number;
    is_tax_inclusive?: boolean;
    enable_tables?: boolean;
    enable_kds?: boolean;
    enable_wld?: boolean;
    enable_bom?: boolean;
    enable_midtrans?: boolean;
    midtrans_server_key?: string;
    midtrans_client_key?: string;
    midtrans_is_production?: boolean;
  }) => {
    return request<{ outlet_id: string; updated: boolean }>('/setup/settings', {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
  },
};
