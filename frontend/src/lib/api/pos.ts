import { request } from './client';
import type { Category, Product, Table, Order, OrderItem, Payment } from '$lib/types';

export const posApi = {
  getCategories: async () => {
    return request<Category[]>('/pos/categories');
  },

  createCategory: async (categoryData: { name: string; sort_order?: number }) => {
    return request<Category>('/pos/categories', {
      method: 'POST',
      body: JSON.stringify(categoryData),
    });
  },

  updateCategory: async (id: number | string, categoryData: { name?: string; sort_order?: number }) => {
    return request<Category>(`/pos/categories/${id}`, {
      method: 'PUT',
      body: JSON.stringify(categoryData),
    });
  },

  deleteCategory: async (id: number | string) => {
    return request<void>(`/pos/categories/${id}`, {
      method: 'DELETE',
    });
  },

  getProducts: async (categoryId?: string) => {
    const query = categoryId ? `?category_id=${categoryId}` : '';
    return request<Product[]>(`/pos/products${query}`);
  },

  saveProduct: async (productData: Partial<Product>) => {
    const method = productData.id ? 'PUT' : 'POST';
    const url = productData.id ? `/pos/products/${productData.id}` : '/pos/products';
    return request<Product>(url, {
      method,
      body: JSON.stringify(productData),
    });
  },

  deleteProduct: async (id: string) => {
    return request<{ id: string }>(`/pos/products/${id}`, {
      method: 'DELETE',
    });
  },

  getTables: async () => {
    return request<Table[]>('/pos/tables');
  },

  createTable: async (tableData: { table_number: string; capacity: number }) => {
    return request<Table>('/pos/tables', {
      method: 'POST',
      body: JSON.stringify(tableData),
    });
  },

  updateTable: async (id: number | string, tableData: { table_number?: string; capacity?: number; status?: string }) => {
    return request<Table>(`/pos/tables/${id}`, {
      method: 'PUT',
      body: JSON.stringify(tableData),
    });
  },

  deleteTable: async (id: number | string) => {
    return request<void>(`/pos/tables/${id}`, {
      method: 'DELETE',
    });
  },

  createOrder: async (orderData: {
    table_id?: number | string | null;
    service_type?: string;
    service_mode?: string;
    customer_name?: string;
    items: { product_id: string; quantity: number; unit_price?: number; item_notes?: string; notes?: string }[];
    discount_amount?: number;
  }) => {
    const tableIdNum = orderData.table_id ? Number(orderData.table_id) : undefined;
    const formattedPayload = {
      table_id: tableIdNum && !isNaN(tableIdNum) ? tableIdNum : null,
      service_type: orderData.service_type || orderData.service_mode || 'dine_in',
      customer_name: orderData.customer_name || '',
      discount_amount: orderData.discount_amount || 0,
      items: orderData.items.map((item) => ({
        product_id: item.product_id,
        quantity: item.quantity,
        unit_price: item.unit_price || 0,
        item_notes: item.item_notes || item.notes || '',
      })),
    };
    return request<Order>('/pos/orders', {
      method: 'POST',
      body: JSON.stringify(formattedPayload),
    });
  },

  getOrders: async (
    statusOrFilters?: string | { status?: string; shiftId?: string; date?: string; startDate?: string; endDate?: string },
    shiftId?: string,
    date?: string,
    startDate?: string,
    endDate?: string
  ) => {
    const params = new URLSearchParams();
    if (typeof statusOrFilters === 'object' && statusOrFilters !== null) {
      if (statusOrFilters.status) params.append('status', statusOrFilters.status);
      if (statusOrFilters.shiftId) params.append('shift_id', statusOrFilters.shiftId);
      if (statusOrFilters.date) params.append('date', statusOrFilters.date);
      if (statusOrFilters.startDate) params.append('start_date', statusOrFilters.startDate);
      if (statusOrFilters.endDate) params.append('end_date', statusOrFilters.endDate);
    } else {
      if (statusOrFilters) params.append('status', statusOrFilters);
      if (shiftId) params.append('shift_id', shiftId);
      if (date) params.append('date', date);
      if (startDate) params.append('start_date', startDate);
      if (endDate) params.append('end_date', endDate);
    }
    const query = params.toString() ? `?${params.toString()}` : '';
    return request<Order[]>(`/pos/orders${query}`);
  },

  getOrder: async (id: string) => {
    return request<Order>(`/pos/orders/${id}`);
  },

  splitBill: async (orderId: string, payload: { items: { order_item_id: string; quantity: number }[] } | { items_bill_b: { order_item_id: string; quantity: number }[]; pin?: string }) => {
    const items = 'items' in payload ? payload.items : payload.items_bill_b;
    return request<{ parent_order: Order; child_order: Order }>(`/pos/orders/${orderId}/split`, {
      method: 'POST',
      body: JSON.stringify({ items }),
    });
  },

  mergeBill: async (targetOrderId: string, sourceOrderIds: string[]) => {
    return request<Order>('/pos/orders/merge', {
      method: 'POST',
      body: JSON.stringify({ target_order_id: targetOrderId, source_order_ids: sourceOrderIds }),
    });
  },

  voidOrder: async (orderId: string, pin: string, reason: string) => {
    return request<{ voided: boolean; order_id: string }>(`/pos/orders/${orderId}/void`, {
      method: 'POST',
      body: JSON.stringify({ pin, reason }),
    });
  },

  checkoutCash: async (orderId: string, amountPaid: number) => {
    return request<Payment>('/payment/cash', {
      method: 'POST',
      body: JSON.stringify({ order_id: orderId, amount_paid: amountPaid }),
    });
  },

  checkoutQRIS: async (orderId: string) => {
    return request<{
      payment_id: string;
      order_id: string;
      midtrans_transaction_id: string;
      gross_amount: number;
      qris_url: string;
      qr_string: string;
      status: string;
    }>('/payment/midtrans/charge', {
      method: 'POST',
      body: JSON.stringify({ order_id: orderId, payment_type: 'qris_dynamic' }),
    });
  },

  getKdsOrders: async () => {
    return request<Order[]>('/pos/orders');
  },

  updateKdsItemStatus: async (orderItemId: string, status: string) => {
    return request<OrderItem>(`/pos/orders/items/${orderItemId}/status`, {
      method: 'PUT',
      body: JSON.stringify({ kitchen_status: status }),
    });
  },

  completeOrder: async (orderId: string) => {
    return request<{ order_id: string }>(`/pos/orders/${orderId}/complete`, {
      method: 'PUT',
    });
  },

  getPaymentStatus: async (orderId: string) => {
    return request<Payment>(`/payment/status/${orderId}`);
  },
};

