import { request } from './client';
import type { Shift, CashMovement } from '$lib/types';

export const shiftsApi = {
  getActiveShift: async () => {
    return request<{
      shift: Shift;
      cash_sales_total: number;
      cash_in_total: number;
      cash_out_total: number;
      expected_cash_now: number;
      movements: CashMovement[];
    }>('/shift/active', { suppressToast: true });
  },

  openShift: async (openingCash: number, notes: string = '') => {
    return request<Shift>('/shift/open', {
      method: 'POST',
      body: JSON.stringify({ opening_cash: openingCash, notes }),
    });
  },

  closeShift: async (actualCash: number, notes: string = '') => {
    return request<Shift>('/shift/close', {
      method: 'POST',
      body: JSON.stringify({ actual_cash: actualCash, notes }),
    });
  },

  addCashMovement: async (type: 'in' | 'out' | 'cash_in' | 'cash_out', amount: number, reason: string) => {
    const mType = type === 'in' ? 'cash_in' : type === 'out' ? 'cash_out' : type;
    return request<CashMovement>('/shift/cash-movement', {
      method: 'POST',
      body: JSON.stringify({ type: mType, amount, notes: reason, reason }),
    });
  },
};

