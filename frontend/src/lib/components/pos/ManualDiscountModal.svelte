<script lang="ts">
  import Modal from "$lib/components/common/Modal.svelte";
  import CurrencyInput from "$lib/components/common/CurrencyInput.svelte";
  import { Percent, CheckCircle } from "lucide-svelte";

  export let isOpen: boolean = false;
  export let subtotal: number = 0;
  export let currentDiscount: number = 0;
  export let onClose: () => void = () => {};
  export let onApplyDiscount: (amount: number) => void = () => {};

  let discountAmount: number = 0;

  $: if (isOpen) {
    discountAmount = currentDiscount;
  }

  function applyPresetPercent(percent: number) {
    discountAmount = Math.round((subtotal * percent) / 100);
  }

  function handleSave() {
    onApplyDiscount(discountAmount);
    onClose();
  }
</script>

<Modal {isOpen} title="Diskon Manual Transaksi POS" maxWidth="md" {onClose}>
  <div class="space-y-4 py-2">
    <div
      class="p-3.5 bg-blue-50 border border-blue-200 text-blue-900 rounded-xl text-xs font-semibold flex items-center justify-between"
    >
      <span>Subtotal Keranjang Saat Ini:</span>
      <span class="font-bold text-sm text-brand-primary"
        >Rp {subtotal.toLocaleString("id-ID")}</span
      >
    </div>

    <!-- Preset Percentage Buttons -->
    <div class="space-y-1">
      <span class="block text-xs font-bold text-slate-700 uppercase"
        >Preset Persentase Diskon</span
      >
      <div class="grid grid-cols-4 gap-2">
        {#each [5, 10, 15, 20] as pct}
          <button
            type="button"
            on:click={() => applyPresetPercent(pct)}
            class="touch-target py-2.5 bg-slate-100 hover:bg-brand-primary hover:text-white rounded-xl font-bold text-xs border border-slate-200 transition text-center"
          >
            {pct}%
          </button>
        {/each}
      </div>
    </div>

    <!-- Manual Currency Input -->
    <div class="space-y-1">
      <label
        class="block text-xs font-bold text-slate-700 uppercase"
        for="disc-input">Nominal Potongan Diskon (Rp) *</label
      >
      <CurrencyInput bind:value={discountAmount} placeholder="Rp 0" />
    </div>
  </div>

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={onClose}
      class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
    >
      Batal
    </button>

    <button
      type="button"
      on:click={handleSave}
      class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md flex items-center gap-2 transition"
    >
      <CheckCircle class="w-4 h-4" /> Terapkan Diskon
    </button>
  </svelte:fragment>
</Modal>
