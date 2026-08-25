<script lang="ts">
  import { cart, cartTotals } from "$lib/stores/cart";
  import { activePinRequest, auth as authStore } from "$lib/stores/auth";
  import { toast } from "$lib/stores/toast";
  import {
    validateBOMStockAddition,
    getCartBOMSummaryList,
  } from "$lib/utils/bomStock";
  import type { CartItem } from "$lib/types";
  import {
    ShoppingBag,
    Minus,
    Plus,
    Trash2,
    Split,
    Percent,
    ArrowRight,
    Edit3,
    User as UserIcon,
    Utensils,
  } from "lucide-svelte";
  import ConfirmModal from "$lib/components/common/ConfirmModal.svelte";
  import ManualDiscountModal from "./ManualDiscountModal.svelte";

  import { outletProfile } from "$lib/stores/config";
  import { currentUser } from "$lib/stores/auth";

  export let onOpenSplitBill: () => void = () => {};
  export let onOpenPayment: () => void = () => {};

  let editingNoteId: string | null = null;
  let tempNote: string = "";
  let isDiscountModalOpen = false;
  let isClearCartConfirmOpen = false;

  function getAvailableServiceModes(outlet: any): ('dine_in' | 'take_away' | 'delivery')[] {
    if (!outlet || !outlet.service_modes) {
      return ['dine_in', 'take_away', 'delivery'];
    }
    let modes: string[] = [];
    if (Array.isArray(outlet.service_modes)) {
      modes = outlet.service_modes;
    } else if (typeof outlet.service_modes === 'string') {
      try {
        modes = JSON.parse(outlet.service_modes);
      } catch {
        modes = ['dine_in', 'take_away', 'delivery'];
      }
    }
    const validModes: ('dine_in' | 'take_away' | 'delivery')[] = [];
    for (const m of modes) {
      if (m === 'dine_in' || m === 'take_away' || m === 'delivery') {
        validModes.push(m);
      }
    }
    return validModes.length > 0 ? validModes : ['dine_in', 'take_away', 'delivery'];
  }

  $: activeOutlet = $outletProfile || ($currentUser as any)?.outlet;
  $: availableServiceModes = getAvailableServiceModes(activeOutlet);

  $: {
    if (activeOutlet) {
      cart.setTaxAndService(
        activeOutlet.tax_percentage ?? 0,
        activeOutlet.service_percentage ?? 0,
        activeOutlet.is_tax_inclusive ?? false
      );
      if (availableServiceModes.length > 0 && !availableServiceModes.includes($cart.serviceMode)) {
        cart.setServiceMode(availableServiceModes[0]);
      }
    }
  }

  $: bomSummaryList = getCartBOMSummaryList($cart.items);

  function handleIncrementQuantity(item: CartItem) {
    const validation = validateBOMStockAddition(item.product, 1, $cart.items);
    if (!validation.allowed) {
      toast.warning(
        validation.reason ||
          `Batas stok bahan baku tercapai untuk menu "${item.product.name}".`,
      );
      return;
    }
    cart.updateQuantity(item.id, 1);
  }

  function openNoteEditor(itemId: string, currentNote: string) {
    editingNoteId = itemId;
    tempNote = currentNote;
  }

  function saveNote(itemId: string) {
    cart.setItemNotes(itemId, tempNote);
    editingNoteId = null;
  }

  function triggerClearCartModal() {
    if ($cart.items.length === 0) return;
    isClearCartConfirmOpen = true;
  }

  async function handleConfirmClearCart() {
    try {
      await authStore.requestSupervisorPin("Otorisasi Void / Clear Cart");
      cart.clear();
      toast.success("Keranjang berhasil dibersihkan");
      isClearCartConfirmOpen = false;
    } catch {
      // User cancelled pin
      isClearCartConfirmOpen = false;
    }
  }

  async function triggerDiscountModal() {
    try {
      await authStore.requestSupervisorPin("Otorisasi Diskon Manual Transaksi");
      isDiscountModalOpen = true;
    } catch {
      // Cancelled pin
    }
  }

  function handleApplyDiscount(amount: number) {
    cart.setManualDiscount(amount);
    toast.success(
      `Diskon manual Rp ${amount.toLocaleString("id-ID")} diterapkan`,
    );
  }

  async function handleTriggerSplitBill() {
    if ($cart.items.length === 0) {
      toast.warning("Keranjang masih kosong");
      return;
    }
    try {
      await authStore.requestSupervisorPin("Otorisasi Split Bill");
      onOpenSplitBill();
    } catch {
      // Cancelled
    }
  }
</script>

<div
  class="bg-white rounded-2xl border border-slate-200 shadow-lg flex flex-col h-full overflow-hidden"
>
  <!-- Header: Table & Service Mode Selector -->
  <div class="p-4 bg-slate-900 text-white flex items-center justify-between">
    <div class="flex items-center gap-2">
      <ShoppingBag class="w-5 h-5 text-brand-cyan" />
      <div>
        <h3 class="font-bold text-base leading-tight">Keranjang Pesanan</h3>
        <p class="text-xs text-slate-300">
          {$cart.selectedTableName
            ? `Meja: ${$cart.selectedTableName}`
            : "Tanpa Meja"}
        </p>
      </div>
    </div>

    <!-- Service Mode Selector Pills (Correlated with Setup Step 3) -->
    <div
      class="flex items-center bg-slate-800 p-1 rounded-xl text-xs font-bold gap-0.5"
    >
      {#each availableServiceModes as mode}
        <button
          type="button"
          on:click={() => cart.setServiceMode(mode)}
          class="px-2.5 py-1.5 rounded-lg transition min-h-[38px] flex items-center {$cart.serviceMode === mode
            ? 'bg-brand-primary text-white shadow-sm font-extrabold'
            : 'text-slate-400 hover:text-white'}"
        >
          {#if mode === 'dine_in'}
            Dine-In
          {:else if mode === 'take_away'}
            Takeaway
          {:else if mode === 'delivery'}
            Delivery
          {/if}
        </button>
      {/each}
    </div>
  </div>

  <!-- Customer Name Input Bar -->
  <div
    class="px-4 py-2.5 bg-slate-50 border-b border-slate-200 flex items-center gap-2"
  >
    <UserIcon class="w-4 h-4 text-brand-primary flex-shrink-0" />
    <span class="text-xs font-bold text-slate-600 whitespace-nowrap"
      >Pemesan:</span
    >
    <input
      type="text"
      value={$cart.customerName}
      on:input={(e) => cart.setCustomerName(e.currentTarget.value)}
      placeholder="Nama Pemesan (misal: Budi / Ibu Siska)"
      class="w-full px-3 py-2 min-h-[44px] bg-white border border-slate-300 focus:border-brand-primary rounded-xl text-xs font-bold text-slate-800 outline-none shadow-sm"
    />
  </div>

  <!-- Cart Items List -->
  <div class="flex-1 overflow-y-auto p-4 space-y-3 divide-y divide-slate-100">
    {#if $cart.items.length === 0}
      <div
        class="h-full flex flex-col items-center justify-center text-slate-400 py-12"
      >
        <ShoppingBag class="w-12 h-12 mb-2 text-slate-200" />
        <p class="font-bold text-slate-600">Keranjang Masih Kosong</p>
        <p class="text-xs text-slate-400 mt-1">
          Pilih menu dari katalog untuk menambahkan pesanan.
        </p>
      </div>
    {:else}
      {#each $cart.items as item (item.id)}
        <div class="pt-3 first:pt-0 space-y-2">
          <div class="flex items-start justify-between">
            <div>
              <h5 class="font-bold text-slate-800 text-sm">
                {item.product.name}
              </h5>
              <span class="text-xs text-brand-primary font-semibold">
                Rp {item.product.price.toLocaleString("id-ID")}
              </span>
            </div>
            <button
              type="button"
              on:click={() => cart.removeItem(item.id)}
              class="text-slate-400 hover:text-brand-danger transition p-2 min-h-[40px] min-w-[40px] flex items-center justify-center"
              title="Hapus Menu"
            >
              <Trash2 class="w-4 h-4" />
            </button>
          </div>

          <!-- Notes & Discount indicators -->
          {#if item.notes}
            <p
              class="text-xs italic text-amber-700 bg-amber-50 px-2.5 py-1.5 rounded-md border border-amber-200"
            >
              Catatan: {item.notes}
            </p>
          {/if}

          <!-- Note Editor Input -->
          {#if editingNoteId === item.id}
            <div class="flex items-center gap-2 mt-1">
              <input
                type="text"
                bind:value={tempNote}
                placeholder="Misal: Tanpa gula, es sedikit..."
                class="flex-1 px-3 py-2 min-h-[44px] text-xs bg-slate-50 border border-slate-300 rounded-lg outline-none focus:border-brand-primary"
              />
              <button
                type="button"
                on:click={() => saveNote(item.id)}
                class="px-3.5 py-2 min-h-[44px] bg-brand-primary text-white text-xs font-bold rounded-lg"
              >
                Simpan
              </button>
            </div>
          {/if}

          <!-- Quantity Modifiers -->
          <div class="flex items-center justify-between pt-1">
            <button
              type="button"
              on:click={() => openNoteEditor(item.id, item.notes)}
              class="text-[11px] font-semibold text-slate-600 hover:text-brand-primary flex items-center gap-1 bg-slate-100 hover:bg-slate-200 px-2.5 py-2 rounded-lg transition min-h-[40px]"
            >
              <Edit3 class="w-3.5 h-3.5" />
              {item.notes ? "Edit Catatan" : "+ Catatan"}
            </button>

            <div
              class="flex items-center border border-slate-300 rounded-xl overflow-hidden bg-slate-50 shadow-sm"
            >
              <button
                type="button"
                on:click={() => cart.updateQuantity(item.id, -1)}
                class="w-10 h-10 flex items-center justify-center text-slate-700 hover:bg-slate-200 transition min-h-[40px] min-w-[40px]"
              >
                <Minus class="w-4 h-4" />
              </button>
              <span class="w-8 text-center font-bold text-sm text-slate-800"
                >{item.quantity}</span
              >
              <button
                type="button"
                on:click={() => handleIncrementQuantity(item)}
                class="w-10 h-10 flex items-center justify-center text-slate-700 hover:bg-slate-200 transition min-h-[40px] min-w-[40px]"
              >
                <Plus class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      {/each}
    {/if}
  </div>

  <!-- Financial Summary & Action Bar -->
  <div class="p-4 bg-slate-50 border-t border-slate-200 space-y-3">
    <!-- Quick Utility Bar -->
    <div class="flex items-center justify-between text-xs font-bold gap-2">
      <button
        type="button"
        on:click={triggerDiscountModal}
        class="flex-1 py-2.5 min-h-[44px] bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 rounded-xl flex items-center justify-center gap-1 shadow-sm transition"
      >
        <Percent class="w-3.5 h-3.5 text-brand-primary" />
        Diskon
      </button>

      <button
        type="button"
        on:click={handleTriggerSplitBill}
        class="flex-1 py-2.5 min-h-[44px] bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 rounded-xl flex items-center justify-center gap-1 shadow-sm transition"
      >
        <Split class="w-3.5 h-3.5 text-brand-primary" />
        Split Bill
      </button>

      <button
        type="button"
        on:click={triggerClearCartModal}
        class="px-3.5 py-2.5 min-h-[44px] bg-white hover:bg-red-50 text-brand-danger border border-slate-200 rounded-xl flex items-center justify-center shadow-sm transition"
        title="Clear Cart"
      >
        <Trash2 class="w-4 h-4" />
      </button>
    </div>

    <!-- Calculations -->
    <div class="space-y-1.5 text-xs">
      <div class="flex justify-between text-slate-600">
        <span>Subtotal ({$cartTotals.itemCount} Item):</span>
        <span class="font-semibold"
          >Rp {$cartTotals.subtotal.toLocaleString("id-ID")}</span
        >
      </div>
      {#if $cartTotals.manualDiscount > 0}
        <div class="flex justify-between text-brand-danger font-semibold">
          <span>Diskon Manual:</span>
          <span>-Rp {$cartTotals.manualDiscount.toLocaleString("id-ID")}</span>
        </div>
      {/if}
      {#if $cartTotals.serviceAmount > 0}
        <div class="flex justify-between text-slate-600">
          <span>Service Charge ({$cart.servicePercentage}%):</span>
          <span class="font-semibold"
            >Rp {$cartTotals.serviceAmount.toLocaleString("id-ID")}</span
          >
        </div>
      {/if}
      {#if $cartTotals.taxAmount > 0}
        <div class="flex justify-between text-slate-600">
          <span>Pajak Resto ({$cart.taxPercentage}%):</span>
          <span class="font-semibold"
            >Rp {$cartTotals.taxAmount.toLocaleString("id-ID")}</span
          >
        </div>
      {/if}
    </div>

    <!-- Grand Total Display -->
    <div
      class="pt-2 border-t border-slate-200 flex items-baseline justify-between"
    >
      <span class="text-xs font-bold text-slate-500 uppercase tracking-wider"
        >Total Tagihan</span
      >
      <span class="display-currency text-brand-primary text-2xl sm:text-3xl font-extrabold">
        Rp {$cartTotals.grandTotal.toLocaleString("id-ID")}
      </span>
    </div>

    <!-- Main Unified Process & Payment Button (F4) -->
    <button
      type="button"
      disabled={$cart.items.length === 0}
      on:click={onOpenPayment}
      class="w-full py-4 min-h-[48px] font-extrabold text-white text-base sm:text-lg bg-brand-primary hover:bg-sky-500 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed rounded-2xl shadow-lg shadow-sky-500/25 flex items-center justify-center gap-2.5 transition"
    >
      <span>BAYAR / PROSES PESANAN (F4)</span>
      <ArrowRight class="w-5 h-5" />
    </button>
  </div>
</div>

<ManualDiscountModal
  isOpen={isDiscountModalOpen}
  subtotal={$cartTotals.subtotal}
  currentDiscount={$cartTotals.manualDiscount}
  onClose={() => (isDiscountModalOpen = false)}
  onApplyDiscount={handleApplyDiscount}
/>

<ConfirmModal
  isOpen={isClearCartConfirmOpen}
  title="Konfirmasi Void / Bersihkan Keranjang"
  message="Apakah Anda yakin ingin membatalkan dan mengosongkan seluruh pesanan di keranjang saat ini?"
  confirmText="Ya, Void Pesanan"
  cancelText="Batal"
  isDanger={true}
  onConfirm={handleConfirmClearCart}
  onClose={() => (isClearCartConfirmOpen = false)}
/>
