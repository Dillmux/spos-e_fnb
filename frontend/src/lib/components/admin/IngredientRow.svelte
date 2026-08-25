<script lang="ts">
  import type { Ingredient } from '$lib/types';
  import { Package, AlertTriangle, RefreshCw, Edit2, Trash2 } from 'lucide-svelte';

  export let ingredient: Ingredient;
  export let onAdjustStock: (ing: Ingredient) => void = () => {};
  export let onEditDetail: (ing: Ingredient) => void = () => {};
  export let onDelete: (ing: Ingredient) => void = () => {};

  $: currentStock = ingredient.current_stock ?? ingredient.stock ?? 0;
  $: minStockAlert = ingredient.min_stock_alert ?? ingredient.min_stock ?? 0;
  $: costPerUnit = ingredient.cost_per_unit ?? ingredient.cogs_per_unit ?? 0;

  $: isLowStock = currentStock <= minStockAlert;
</script>

<div class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:border-slate-300 transition">
  <div class="flex items-center gap-3">
    <div class="p-3 bg-slate-100 rounded-xl text-slate-700">
      <Package class="w-5 h-5 text-brand-primary" />
    </div>
    <div>
      <h4 class="font-bold text-slate-800 text-base">{ingredient.name}</h4>
      <p class="text-xs text-slate-500 font-semibold">
        Satuan: <span class="font-extrabold text-slate-700 bg-slate-100 px-2 py-0.5 rounded-md">{ingredient.unit}</span>
        | HPP/COGS: Rp {costPerUnit.toLocaleString('id-ID')} / {ingredient.unit}
      </p>
    </div>
  </div>

  <div class="flex items-center justify-between sm:justify-end gap-3">
    <!-- Stock Badge -->
    <div class="text-right">
      <span class="text-lg font-extrabold text-slate-900">
        {currentStock} <span class="text-xs font-semibold text-slate-500">{ingredient.unit}</span>
      </span>

      {#if isLowStock}
        <span class="flex items-center justify-end gap-1 text-[11px] font-bold text-brand-danger bg-red-50 px-2 py-0.5 rounded-full mt-0.5">
          <AlertTriangle class="w-3 h-3" /> Stok Kritis (&le; {minStockAlert})
        </span>
      {/if}
    </div>

    <!-- Action Buttons -->
    <div class="flex items-center gap-1">
      <button
        type="button"
        on:click={() => onAdjustStock(ingredient)}
        class="p-2.5 text-slate-600 hover:text-brand-primary hover:bg-slate-100 rounded-xl transition min-h-[44px] min-w-[44px] flex items-center justify-center"
        title="Stock Opname / Penyesuaian Stok"
      >
        <RefreshCw class="w-4 h-4" />
      </button>

      <button
        type="button"
        on:click={() => onEditDetail(ingredient)}
        class="p-2.5 text-slate-600 hover:text-slate-800 hover:bg-slate-100 rounded-xl transition min-h-[44px] min-w-[44px] flex items-center justify-center"
        title="Edit Detail & Satuan Ukur"
      >
        <Edit2 class="w-4 h-4" />
      </button>

      <button
        type="button"
        on:click={() => onDelete(ingredient)}
        class="p-2.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-xl transition min-h-[44px] min-w-[44px] flex items-center justify-center"
        title="Hapus Bahan Baku"
      >
        <Trash2 class="w-4 h-4" />
      </button>
    </div>
  </div>
</div>


