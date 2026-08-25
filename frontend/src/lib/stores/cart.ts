import { writable, derived } from 'svelte/store';
import type { CartItem, Product } from '$lib/types';

interface CartState {
  items: CartItem[];
  selectedTableId: string | null;
  selectedTableName: string | null;
  customerName: string;
  serviceMode: 'dine_in' | 'take_away' | 'delivery';
  manualDiscount: number; // Global order discount
  taxPercentage: number;
  servicePercentage: number;
  isTaxInclusive: boolean;
}

const initialState: CartState = {
  items: [],
  selectedTableId: null,
  selectedTableName: null,
  customerName: '',
  serviceMode: 'dine_in',
  manualDiscount: 0,
  taxPercentage: 0,
  servicePercentage: 0,
  isTaxInclusive: false,
};

const store = writable<CartState>(initialState);

export const cart = {
  subscribe: store.subscribe,

  addItem: (product: Product, notes = '') => {
    store.update((s) => {
      const existingIndex = s.items.findIndex(
        (i) => i.product.id === product.id && i.notes === notes
      );
      if (existingIndex >= 0) {
        const updatedItems = [...s.items];
        updatedItems[existingIndex].quantity += 1;
        return { ...s, items: updatedItems };
      }

      const newItem: CartItem = {
        id: `${product.id}_${Date.now()}_${Math.random().toString(36).substr(2, 4)}`,
        product,
        quantity: 1,
        notes,
        discount_amount: 0,
        split_group: 'A',
      };
      return { ...s, items: [...s.items, newItem] };
    });
  },

  removeItem: (itemId: string) => {
    store.update((s) => ({
      ...s,
      items: s.items.filter((i) => i.id !== itemId),
    }));
  },

  updateQuantity: (itemId: string, delta: number) => {
    store.update((s) => {
      const updated = s.items
        .map((item) => {
          if (item.id === itemId) {
            const newQty = item.quantity + delta;
            return newQty > 0 ? { ...item, quantity: newQty } : null;
          }
          return item;
        })
        .filter(Boolean) as CartItem[];
      return { ...s, items: updated };
    });
  },

  setItemQuantity: (itemId: string, qty: number) => {
    if (qty <= 0) {
      cart.removeItem(itemId);
      return;
    }
    store.update((s) => ({
      ...s,
      items: s.items.map((i) => (i.id === itemId ? { ...i, quantity: qty } : i)),
    }));
  },

  setItemNotes: (itemId: string, notes: string) => {
    store.update((s) => ({
      ...s,
      items: s.items.map((i) => (i.id === itemId ? { ...i, notes } : i)),
    }));
  },

  setItemDiscount: (itemId: string, discountAmount: number) => {
    store.update((s) => ({
      ...s,
      items: s.items.map((i) => (i.id === itemId ? { ...i, discount_amount: Math.max(0, discountAmount) } : i)),
    }));
  },

  setTable: (tableId: string | null, tableName: string | null = null) => {
    store.update((s) => ({ ...s, selectedTableId: tableId, selectedTableName: tableName }));
  },

  setCustomerName: (name: string) => {
    store.update((s) => ({ ...s, customerName: name }));
  },

  setServiceMode: (mode: 'dine_in' | 'take_away' | 'delivery') => {
    store.update((s) => ({ ...s, serviceMode: mode }));
  },

  setManualDiscount: (amount: number) => {
    store.update((s) => ({ ...s, manualDiscount: Math.max(0, amount) }));
  },

  setTaxAndService: (taxPct: number, servicePct: number, inclusive: boolean) => {
    store.update((s) => ({
      ...s,
      taxPercentage: taxPct,
      servicePercentage: servicePct,
      isTaxInclusive: inclusive,
    }));
  },

  assignSplitGroup: (itemId: string, group: 'A' | 'B') => {
    store.update((s) => ({
      ...s,
      items: s.items.map((i) => (i.id === itemId ? { ...i, split_group: group } : i)),
    }));
  },

  clear: () => {
    store.update((s) => ({
      ...s,
      items: [],
      selectedTableId: null,
      selectedTableName: null,
      customerName: '',
      manualDiscount: 0,
    }));
  },
};

// Derived financial totals
export const cartTotals = derived(store, ($s) => {
  const itemSubtotal = $s.items.reduce((sum, item) => {
    const itemTotal = item.product.price * item.quantity - item.discount_amount;
    return sum + Math.max(0, itemTotal);
  }, 0);

  const subtotal = Math.max(0, itemSubtotal - $s.manualDiscount);

  let serviceAmount = 0;
  let taxAmount = 0;

  if ($s.isTaxInclusive) {
    // Tax is included inside price
    taxAmount = Math.round(subtotal * ($s.taxPercentage / (100 + $s.taxPercentage)));
    serviceAmount = Math.round(subtotal * ($s.servicePercentage / 100));
  } else {
    // Tax & Service added on top
    serviceAmount = Math.round(subtotal * ($s.servicePercentage / 100));
    const taxableAmount = subtotal + serviceAmount;
    taxAmount = Math.round(taxableAmount * ($s.taxPercentage / 100));
  }

  const grandTotal = Math.max(0, subtotal + ($s.isTaxInclusive ? 0 : serviceAmount + taxAmount));

  return {
    itemSubtotal,
    subtotal,
    manualDiscount: $s.manualDiscount,
    serviceAmount,
    taxAmount,
    grandTotal,
    itemCount: $s.items.reduce((acc, item) => acc + item.quantity, 0),
  };
});
