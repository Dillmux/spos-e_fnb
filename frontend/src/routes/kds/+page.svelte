<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { posApi } from '$lib/api/pos';
  import { featureFlags } from '$lib/stores/config';
  import { toast } from '$lib/stores/toast';
  import KitchenTicketCard from '$lib/components/kds/KitchenTicketCard.svelte';
  import StatusFilter from '$lib/components/kds/StatusFilter.svelte';
  import DisabledFeatureState from '$lib/components/common/DisabledFeatureState.svelte';
  import type { Order } from '$lib/types';
  import { Monitor, RefreshCw, ChefHat } from 'lucide-svelte';

  let orders: Order[] = [];
  let filterStatus: string = 'all';
  let isLoading = true;
  let refreshInterval: any;

  async function loadKdsOrders() {
    try {
      const res = await posApi.getKdsOrders();
      const rawOrders = res.data || [];
      // Normalize items and order properties
      orders = rawOrders.map((o: any) => ({
        ...o,
        service_type: o.service_type || o.service_mode || 'dine_in',
        service_mode: o.service_mode || o.service_type || 'dine_in',
        items: (o.items || o.OrderItems || []).map((i: any) => ({
          ...i,
          product_name: i.product_name || i.product?.name || i.Product?.name || 'Produk',
          kitchen_status: i.kitchen_status || 'queued',
        })),
      }));
    } catch (err: any) {
      orders = [];
      toast.error(err.message || 'Gagal memuat daftar antrean dapur');
    } finally {
      isLoading = false;
    }
  }

  let eventSource: EventSource | null = null;

  onMount(() => {
    loadKdsOrders();
    refreshInterval = setInterval(loadKdsOrders, 10000);

    if (typeof window !== 'undefined' && 'EventSource' in window) {
      try {
        eventSource = new EventSource('/api/v1/events/stream');
        eventSource.onmessage = () => {
          loadKdsOrders();
        };
        eventSource.addEventListener('order_created', () => loadKdsOrders());
        eventSource.addEventListener('order_updated', () => loadKdsOrders());
        eventSource.addEventListener('kitchen_status_updated', () => loadKdsOrders());
      } catch {}
    }
  });

  onDestroy(() => {
    clearInterval(refreshInterval);
    if (eventSource) {
      eventSource.close();
    }
  });

  async function handleUpdateStatus(orderItemId: string, status: string) {
    try {
      await posApi.updateKdsItemStatus(orderItemId, status);
      toast.success('Status dapur diperbarui');
      loadKdsOrders();
    } catch (err: any) {
      toast.error(err.message || 'Gagal memperbarui status item dapur');
    }
  }

  $: filteredOrders = orders.filter((o) => {
    if (o.status === 'cancelled' || o.status === 'void' || o.status === 'completed') {
      return false;
    }
    const items = o.items || o.OrderItems || [];
    if (items.length === 0) return false;

    // Pesanan otomatis disembunyikan/hilang dari KDS jika semua item sudah tersajikan (served)
    const allItemsServed = items.every((i) => i.kitchen_status === 'served');
    if (allItemsServed) return false;

    if (filterStatus === 'all') return true;

    return items.some((i) => {
      const st = i.kitchen_status || 'queued';
      if (filterStatus === 'pending') return st === 'pending' || st === 'queued';
      if (filterStatus === 'preparing') return st === 'preparing' || st === 'cooking';
      return st === filterStatus;
    });
  });
</script>

{#if !$featureFlags.enable_kds}
  <DisabledFeatureState
    title="Modul Kitchen Display System (KDS) Nonaktif"
    description="Fitur antrean koki dapur (KDS) belum diaktifkan di Setup Wizard / Konfigurasi Modul. Silakan aktifkan modul KDS di menu pengaturan."
    icon={ChefHat}
    actionText="Kembali ke Kasir POS"
    actionHref="/pos"
  />
{:else}
  <div class="h-full bg-brand-dark p-3 sm:p-6 flex flex-col space-y-4 sm:space-y-6 overflow-hidden">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <div class="flex items-center gap-3 text-white">
        <div class="p-2.5 sm:p-3 bg-brand-cyan/20 text-brand-cyan rounded-2xl">
          <ChefHat class="w-6 h-6 sm:w-7 sm:h-7" />
        </div>
        <div>
          <h2 class="text-xl sm:text-2xl font-black tracking-wide text-white">Kitchen Display System (KDS)</h2>
          <p class="text-[11px] sm:text-xs text-brand-cyan font-semibold">Live Antrean Koki Dapur & Status Tiket Masak</p>
        </div>
      </div>

      <div class="flex items-center gap-2 sm:gap-3 w-full sm:w-auto justify-between sm:justify-end">
        <StatusFilter
          selectedStatus={filterStatus}
          onChangeStatus={(st) => (filterStatus = st)}
        />

        <button
          type="button"
          on:click={loadKdsOrders}
          class="p-2.5 bg-slate-800 hover:bg-slate-700 text-brand-cyan rounded-2xl border border-slate-700 transition min-h-[44px] min-w-[44px] flex items-center justify-center flex-shrink-0"
          title="Manual Refresh"
        >
          <RefreshCw class="w-5 h-5 {isLoading ? 'animate-spin' : ''}" />
        </button>
      </div>
    </div>

    <!-- Kanban Grid -->
    {#if filteredOrders.length === 0}
      <div class="flex-1 flex flex-col items-center justify-center text-slate-500 py-16 border-2 border-dashed border-slate-800 rounded-3xl">
        <ChefHat class="w-16 h-16 mb-2 text-slate-700" />
        <h4 class="font-bold text-slate-400 text-lg">Tidak Ada Tiket Pesanan Aktif</h4>
        <p class="text-xs text-slate-600 mt-1">Semua pesanan sudah selesai dimasak atau belum ada transaksi baru.</p>
      </div>
    {:else}
      <div class="flex-1 overflow-x-auto grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 pb-2">
        {#each filteredOrders as order (order.id)}
          <KitchenTicketCard
            {order}
            onUpdateItemStatus={handleUpdateStatus}
          />
        {/each}
      </div>
    {/if}
  </div>
{/if}
