import { request } from './client';
import type { DailyReport } from '$lib/types';

export const reportsApi = {
  getDailySales: async (date?: string) => {
    const query = date ? `?date=${date}` : '';
    return request<{
      date: string;
      total_orders: number;
      total_revenue: number;
      total_tax: number;
      total_service: number;
      payment_breakdowns: { payment_method: string; total_amount: number; count: number }[];
    }>(`/report/daily-sales${query}`);
  },

  getDailyReport: async (date?: string) => {
    const query = date ? `?date=${date}` : '';
    return request<DailyReport>(`/report/daily-sales${query}`);
  },

  getGrossProfit: async (startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    const query = params.toString() ? `?${params.toString()}` : '';
    return request<{
      start_date: string;
      end_date: string;
      net_revenue: number;
      total_cogs: number;
      gross_profit: number;
      profit_margin: number;
    }>(`/report/gross-profit${query}`);
  },

  getTopSelling: async (date?: string, startDate?: string, endDate?: string) => {
    const params = new URLSearchParams();
    if (date) params.append('date', date);
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    const query = params.toString() ? `?${params.toString()}` : '';
    return request<{
      product_id: string;
      product_name: string;
      total_sold: number;
      total_revenue: number;
    }[]>(`/report/top-selling${query}`);
  },
};


