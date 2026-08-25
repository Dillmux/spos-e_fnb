import { writable } from 'svelte/store';

export interface ToastMessage {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  text: string;
  duration?: number;
}

function createToastStore() {
  const { subscribe, update } = writable<ToastMessage[]>([]);

  function show(text: string, type: ToastMessage['type'] = 'info', duration = 3500) {
    if (!text) return;
    const id = Math.random().toString(36).substring(2, 9);
    let added = false;

    update((toasts) => {
      // Deduplicate: Don't show identical toast if message and type match an active toast
      if (toasts.some((t) => t.text === text && t.type === type)) {
        return toasts;
      }
      added = true;
      const next = [...toasts, { id, type, text, duration }];
      // Limit to at most 4 active toasts to prevent screen clutter
      return next.length > 4 ? next.slice(next.length - 4) : next;
    });

    if (added) {
      setTimeout(() => {
        dismiss(id);
      }, duration);
    }
  }

  function dismiss(id: string) {
    update((toasts) => toasts.filter((t) => t.id !== id));
  }

  function clear() {
    update(() => []);
  }

  return {
    subscribe,
    show,
    success: (text: string, duration?: number) => show(text, 'success', duration),
    error: (text: string, duration?: number) => show(text, 'error', duration),
    warning: (text: string, duration?: number) => show(text, 'warning', duration),
    info: (text: string, duration?: number) => show(text, 'info', duration),
    dismiss,
    clear,
  };
}

export const toast = createToastStore();
