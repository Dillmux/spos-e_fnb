<script lang="ts">
  import { onMount } from 'svelte';
  import { featureFlags } from '$lib/stores/config';
  import { cart, cartTotals } from '$lib/stores/cart';
  import { shiftsApi } from '$lib/api/shifts';
  import { toast } from '$lib/stores/toast';
  import TableGrid from '$lib/components/pos/TableGrid.svelte';
  import MenuCatalog from '$lib/components/pos/MenuCatalog.svelte';
  import CartPanel from '$lib/components/pos/CartPanel.svelte';
  import SplitBillModal from '$lib/components/pos/SplitBillModal.svelte';
  import PaymentModal from '$lib/components/pos/PaymentModal.svelte';
  import type { Table } from '$lib/types';
  import { ShoppingBag, X, ChevronUp, CreditCard } from 'lucide-svelte';

  let isSplitBillOpen = false;
  let isPaymentOpen = false;
  let isMobileCartDrawerOpen = false;
  let selectedTable: Table | null = null;

  function handleSelectTable(table: Table) {
    selectedTable = table;
  }

  async function handleOpenPayment() {
    if ($cart.items.length === 0) {
      toast.warning('Keranjang pesanan masih kosong! Pilih menu makanan / minuman terlebih dahulu.');
      return;
    }

    try {
      const res = await shiftsApi.getActiveShift();
      const raw = res.data as any;
      const activeShift = raw?.shift || (raw && raw.id ? raw : null);
      
      if (!activeShift || activeShift.status !== 'open') {
        toast.warning('Shift register belum dibuka! Silakan buka shift kasir terlebih dahulu sebelum menerima pesanan / transaksi.');
        window.dispatchEvent(new CustomEvent('open-shift-modal'));
        return;
      }
    } catch {
      // Backend offline fallback - allow opening payment modal with warning
    }

    isMobileCartDrawerOpen = false;
    isPaymentOpen = true;
  }

  onMount(() => {
    function handleGlobalKeys(e: KeyboardEvent) {
      if (e.key === 'F4') {
        e.preventDefault();
        handleOpenPayment();
      } else if (e.key === 'Escape') {
        isSplitBillOpen = false;
        isPaymentOpen = false;
        isMobileCartDrawerOpen = false;
      }
    }
    window.addEventListener('keydown', handleGlobalKeys);
    return () => window.removeEventListener('keydown', handleGlobalKeys);
  });
</script>

<div class="h-full p-2 sm:p-4 flex flex-col lg:grid lg:grid-cols-12 gap-3 sm:gap-4 overflow-hidden relative">
  <!-- Left Workspace: Table Map & Menu Catalog (Cols 1-8 Desktop, 100% Mobile/Tablet) -->
  <div class="lg:col-span-8 flex flex-col space-y-3 sm:space-y-4 h-full overflow-hidden flex-1">
    <!-- Visual Table Grid (Conditional rendering based on active feature flag) -->
    {#if $featureFlags.enable_tables}
      <TableGrid
        selectedTableId={$cart.selectedTableId}
        onSelectTable={handleSelectTable}
      />
    {/if}

    <!-- Menu Catalog Grid -->
    <div class="flex-1 overflow-hidden">
      <MenuCatalog />
    </div>
  </div>

  <!-- Desktop Right Workspace: Sticky Bill Cart Panel (Cols 9-12, Hidden on Mobile/Tablet) -->
  <div class="hidden lg:block lg:col-span-4 h-full overflow-hidden">
    <CartPanel
      onOpenSplitBill={() => (isSplitBillOpen = true)}
      onOpenPayment={handleOpenPayment}
    />
  </div>

  <!-- Mobile/Tablet Floating Bottom Cart Bar (Visible < lg when items in cart or to toggle cart) -->
  <div class="lg:hidden block sticky bottom-0 left-0 right-0 z-30 pt-2 bg-gradient-to-t from-slate-900/40 to-transparent">
    <div class="bg-brand-dark text-white p-3 rounded-2xl shadow-2xl border border-slate-700 flex items-center justify-between gap-3">
      <button
        type="button"
        on:click={() => (isMobileCartDrawerOpen = true)}
        class="flex items-center gap-3 flex-1 text-left min-h-[48px]"
      >
        <div class="relative">
          <div class="p-2.5 bg-brand-primary text-white rounded-xl shadow-md">
            <ShoppingBag class="w-5 h-5" />
          </div>
          {#if $cartTotals.itemCount > 0}
            <span class="absolute -top-1.5 -right-1.5 bg-amber-400 text-slate-950 font-extrabold text-[10px] w-5 h-5 rounded-full flex items-center justify-center border-2 border-brand-dark">
              {$cartTotals.itemCount}
            </span>
          {/if}
        </div>
        <div>
          <p class="text-[11px] text-slate-400 font-bold uppercase tracking-wider leading-none">
            Total Ringkasan (F4)
          </p>
          <p class="font-extrabold text-base text-brand-cyan mt-0.5">
            Rp {$cartTotals.grandTotal.toLocaleString('id-ID')}
          </p>
        </div>
      </button>

      <div class="flex items-center gap-2">
        <button
          type="button"
          on:click={() => (isMobileCartDrawerOpen = true)}
          class="px-3.5 py-3 bg-slate-800 hover:bg-slate-700 text-white text-xs font-bold rounded-xl flex items-center gap-1.5 min-h-[48px] border border-slate-700"
        >
          <ChevronUp class="w-4 h-4 text-brand-cyan" />
          Cart
        </button>

        <button
          type="button"
          on:click={handleOpenPayment}
          disabled={$cart.items.length === 0}
          class="px-4 py-3 bg-brand-primary hover:bg-sky-500 disabled:opacity-50 text-white text-xs font-extrabold rounded-xl shadow-lg flex items-center gap-1.5 min-h-[48px]"
        >
          <CreditCard class="w-4 h-4" />
          Bayar
        </button>
      </div>
    </div>
  </div>

  <!-- Mobile Cart Bottom Sheet / Slide-up Drawer -->
  {#if isMobileCartDrawerOpen}
    <button
      type="button"
      aria-label="Tutup Keranjang Mobile"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-40 lg:hidden transition-opacity cursor-default w-full h-full border-none p-0"
      on:click={() => (isMobileCartDrawerOpen = false)}
    />
    <div
      class="fixed bottom-0 left-0 right-0 top-14 bg-brand-surface z-50 rounded-t-3xl shadow-2xl flex flex-col overflow-hidden lg:hidden animate-in slide-in-from-bottom duration-250 border-t-2 border-brand-primary"
    >
      <div class="bg-brand-dark text-white px-4 py-3 flex items-center justify-between border-b border-slate-800">
        <div class="flex items-center gap-2">
          <ShoppingBag class="w-5 h-5 text-brand-cyan" />
          <h3 class="font-extrabold text-base">Keranjang Pesanan Kasir</h3>
        </div>
        <button
          type="button"
          on:click={() => (isMobileCartDrawerOpen = false)}
          class="p-2 text-slate-400 hover:text-white rounded-xl transition min-h-[48px] min-w-[48px] flex items-center justify-center"
        >
          <X class="w-6 h-6" />
        </button>
      </div>

      <div class="flex-1 overflow-hidden p-3">
        <CartPanel
          onOpenSplitBill={() => {
            isMobileCartDrawerOpen = false;
            isSplitBillOpen = true;
          }}
          onOpenPayment={handleOpenPayment}
        />
      </div>
    </div>
  {/if}

  <!-- Modals -->
  <SplitBillModal
    isOpen={isSplitBillOpen}
    onClose={() => (isSplitBillOpen = false)}
  />

  <PaymentModal
    isOpen={isPaymentOpen}
    onClose={() => (isPaymentOpen = false)}
  />
</div>

