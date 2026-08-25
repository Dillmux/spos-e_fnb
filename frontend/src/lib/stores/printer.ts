import { writable, derived } from 'svelte/store';
import type { Order, Outlet, Payment } from '$lib/types';
import { toast } from './toast';

export interface PrintableReceiptPayload {
  outlet: Outlet;
  order: Order;
  payment?: Payment;
  paperWidth: 58 | 80;
}

export const printableReceiptData = writable<PrintableReceiptPayload | null>(null);

interface PrinterState {
  paperWidth: 58 | 80;
}

const store = writable<PrinterState>({
  paperWidth: 58,
});

export const printer = {
  subscribe: store.subscribe,

  setPaperWidth: (width: 58 | 80) => {
    store.update((s) => ({ ...s, paperWidth: width }));
  },

  connect: async () => {
    toast.info('Printer bawaan browser aktif (siap mencetak via window.print)');
    return true;
  },

  disconnect: () => {
    toast.info('Printer bawaan browser aktif');
  },

  printReceipt: (outlet: Outlet, order: Order, payment?: Payment) => {
    let paperWidth: 58 | 80 = 58;
    store.subscribe((s) => {
      paperWidth = s.paperWidth;
    })();

    printableReceiptData.set({
      outlet,
      order,
      payment,
      paperWidth,
    });

    // Wait for Svelte DOM update then trigger native browser print dialog
    setTimeout(() => {
      if (typeof window !== 'undefined') {
        window.print();
        toast.success('Membuka dialog pencetakan browser...');
      }
    }, 150);
  },
};

export const isPrinterConnected = derived(store, () => true);
export const connectedPrinterName = derived(store, () => 'Browser Native Printer');
