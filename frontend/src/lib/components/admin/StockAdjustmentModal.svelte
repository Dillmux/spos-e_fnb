<script lang="ts">
  import Modal from "$lib/components/common/Modal.svelte";
  import type { Ingredient } from "$lib/types";
  import {
    PackageCheck,
    ArrowUpRight,
    ArrowDownRight,
    RefreshCw,
    Save,
  } from "lucide-svelte";

  export let isOpen: boolean = false;
  export let ingredient: Ingredient | null = null;
  export let onClose: () => void = () => {};
  export let onSave: (payload: {
    change_type: string;
    quantity_change: number;
    notes: string;
  }) => Promise<void> = async () => {};

  let changeType: "restock" | "waste" | "adjustment" = "restock";
  let quantityChange: number = 100;
  let notes: string = "";
  let isSaving: boolean = false;

  $: if (isOpen && ingredient) {
    changeType = "restock";
    quantityChange = 100;
    notes = "";
  }

  async function handleConfirm() {
    if (quantityChange <= 0 && changeType !== "adjustment") {
      return;
    }
    isSaving = true;
    try {
      const finalQty =
        changeType === "waste"
          ? -Math.abs(quantityChange)
          : Math.abs(quantityChange);
      await onSave({
        change_type: changeType,
        quantity_change: finalQty,
        notes: notes.trim() || `Stok ${changeType} via Backoffice`,
      });
      onClose();
    } finally {
      isSaving = false;
    }
  }
</script>

<Modal
  {isOpen}
  title={`Penyesuaian Stok (Stock Opname) - ${ingredient?.name || ""}`}
  maxWidth="md"
  {onClose}
>
  <div class="space-y-4 py-2">
    <!-- Ingredient Summary Header -->
    {#if ingredient}
      <div
        class="p-3.5 bg-slate-50 border border-slate-200 rounded-xl flex items-center justify-between"
      >
        <div>
          <h4 class="font-bold text-slate-800 text-sm">{ingredient.name}</h4>
          <p class="text-xs text-slate-500 font-semibold">
            Satuan: {ingredient.unit}
          </p>
        </div>
        <div class="text-right">
          <span class="text-xs text-slate-400 font-bold block uppercase"
            >Stok Saat Ini</span
          >
          <span class="text-base font-extrabold text-brand-primary">
            {ingredient.current_stock ?? ingredient.stock ?? 0}
            {ingredient.unit}
          </span>
        </div>
      </div>
    {/if}

    <!-- Change Type Radio Group -->
    <div class="space-y-1">
      <span class="block text-xs font-bold text-slate-700 uppercase"
        >Tipe Penyesuaian Mutasi *</span
      >
      <div class="grid grid-cols-3 gap-2">
        <button
          type="button"
          on:click={() => (changeType = "restock")}
          class="touch-target p-3 rounded-xl font-bold text-xs border flex items-center justify-center gap-1.5 transition
          {changeType === 'restock'
            ? 'bg-emerald-500 text-white border-emerald-600 shadow-sm'
            : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'}"
        >
          <ArrowUpRight class="w-4 h-4" /> Restock (+Stok)
        </button>

        <button
          type="button"
          on:click={() => (changeType = "waste")}
          class="touch-target p-3 rounded-xl font-bold text-xs border flex items-center justify-center gap-1.5 transition
          {changeType === 'waste'
            ? 'bg-red-600 text-white border-red-700 shadow-sm'
            : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'}"
        >
          <ArrowDownRight class="w-4 h-4" /> Waste (-Rusak)
        </button>

        <button
          type="button"
          on:click={() => (changeType = "adjustment")}
          class="touch-target p-3 rounded-xl font-bold text-xs border flex items-center justify-center gap-1.5 transition
          {changeType === 'adjustment'
            ? 'bg-brand-primary text-white border-sky-600 shadow-sm'
            : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'}"
        >
          <RefreshCw class="w-4 h-4" /> Korreksi
        </button>
      </div>
    </div>

    <!-- Quantity Change Input -->
    <div class="space-y-1">
      <label
        class="block text-xs font-bold text-slate-700 uppercase"
        for="qty-change"
      >
        Jumlah Kuantitas ({ingredient?.unit || "unit"}) *
      </label>
      <input
        id="qty-change"
        type="number"
        step="0.01"
        bind:value={quantityChange}
        placeholder="100"
        required
        class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
      />
    </div>

    <!-- Notes Input -->
    <div class="space-y-1">
      <label
        class="block text-xs font-bold text-slate-700 uppercase"
        for="adj-notes">Catatan Audit Mutasi</label
      >
      <input
        id="adj-notes"
        type="text"
        bind:value={notes}
        placeholder="misal: Penerimaan pasokan supplier A / bahan terbuang kadaluarsa"
        class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
      />
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
      disabled={isSaving || quantityChange <= 0}
      on:click={handleConfirm}
      class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md transition flex items-center gap-2"
    >
      <Save class="w-4 h-4" />
      {isSaving ? "Simpan..." : "Simpan Penyesuaian"}
    </button>
  </svelte:fragment>
</Modal>
