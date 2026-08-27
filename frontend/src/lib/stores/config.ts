import { writable, derived } from 'svelte/store';
import type { FeatureFlags, Outlet } from '$lib/types';
import { setupApi } from '$lib/api/setup';

interface ConfigState {
  initialized: boolean;
  checkingStatus: boolean;
  outlet: Outlet | null;
  featureFlags: FeatureFlags;
  midtransConfigured: boolean;
}

const defaultConfigState: ConfigState = {
  initialized: false,
  checkingStatus: true,
  outlet: null,
  featureFlags: {
    enable_tables: true,
    enable_kds: true,
    enable_wld: true,
    enable_bom: true,
    enable_midtrans: true,
  },
  midtransConfigured: false,
};

const store = writable<ConfigState>(defaultConfigState);

export const config = {
  subscribe: store.subscribe,

  checkStatus: async () => {
    store.update((s) => ({ ...s, checkingStatus: true }));
    try {
      const res = await setupApi.getStatus();
      if (res.data) {
        store.update((s) => ({
          ...s,
          initialized: res.data?.initialized ?? false,
          outlet: res.data?.outlet || s.outlet,
          featureFlags: res.data?.feature_flags
            ? { ...s.featureFlags, ...res.data.feature_flags }
            : s.featureFlags,
          midtransConfigured: res.data?.midtrans_configured ?? false,
          checkingStatus: false,
        }));
      }
    } catch {
      store.update((s) => ({ ...s, initialized: false, checkingStatus: false }));
    }
  },

  setConfig: (data: { initialized: boolean; outlet?: Outlet; featureFlags?: FeatureFlags; midtransConfigured?: boolean }) => {
    store.update((s) => ({
      ...s,
      initialized: data.initialized,
      outlet: data.outlet || s.outlet,
      featureFlags: data.featureFlags || s.featureFlags,
      midtransConfigured: data.midtransConfigured ?? s.midtransConfigured,
      checkingStatus: false,
    }));
  },

  setFeatureFlags: (flags: Partial<FeatureFlags>) => {
    store.update((s) => ({
      ...s,
      featureFlags: { ...s.featureFlags, ...flags },
    }));
  },
};

export const isInitialized = derived(store, ($s) => $s.initialized);
export const featureFlags = derived(store, ($s) => $s.featureFlags);
export const outletProfile = derived(store, ($s) => $s.outlet);
export const midtransConfigured = derived(store, ($s) => $s.midtransConfigured);
