<script lang="ts">
  import Modal from '$lib/components/common/Modal.svelte';
  import { goto } from '$app/navigation';
  import { AlertTriangle, PackageX, AlertCircle, ArrowRight, ShieldAlert } from 'lucide-svelte';

  export let isOpen: boolean = false;
  export let depletedIngredients: Array<{
    id: string;
    name: string;
    current_stock: number;
    unit: string;
    affected_products: string[];
  }> = [];

  export let lowStockIngredients: Array<{
    id: string;
    name: string;
    current_stock: number;
    min_stock_alert: number;
    unit: string;
    affected_products: string[];
  }> = [];

  export let onClose: () => void = () => {};

  function handleGoToInventory() {
    onClose();
    goto('/admin/inventory');
  }
</script>

<Modal {isOpen} title="Peringatan Stok Bahan Baku (BOM Stock Control)" maxWidth="2xl" {onClose}>
  <div class="space-y-5">
    <!-- Header Warning Banner -->
    <div class="p-4 bg-amber-50 border border-amber-200 rounded-2xl flex items-start gap-3">
      <div class="p-2 bg-amber-100 text-amber-700 rounded-xl mt-0.5">
        <ShieldAlert class="w-6 h-6" />
      </div>
      <div>
        <h4 class="font-extrabold text-amber-900 text-base">Peringatan Inventaris Bahan Baku</h4>
        <p class="text-xs text-amber-700 font-semibold mt-0.5">
          Sistem mengidentifikasi bahan baku yang telah mencapai batas minimum alert atau habis. Menu terkait tidak dapat dipesan atau memerlukan pengisian ulang stok.
        </p>
      </div>
    </div>

    <!-- Depleted Ingredients (Red Alert Section) -->
    {#if depletedIngredients.length > 0}
      <div class="space-y-2">
        <div class="flex items-center gap-2 text-xs font-extrabold text-brand-danger uppercase tracking-wider">
          <PackageX class="w-4 h-4" />
          <span>Bahan Baku Habis ({depletedIngredients.length} Bahan)</span>
        </div>

        <div class="space-y-2 max-h-44 overflow-y-auto pr-1">
          {#each depletedIngredients as ing}
            <div class="p-3 bg-red-50/70 border border-red-200 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <div class="flex items-center gap-2">
                  <span class="font-extrabold text-red-900 text-sm">{ing.name}</span>
                  <span class="text-[10px] font-extrabold px-2 py-0.5 bg-brand-danger text-white rounded-full uppercase">
                    Habis (0 {ing.unit})
                  </span>
                </div>
                {#if ing.affected_products.length > 0}
                  <p class="text-xs text-red-700 mt-1 font-semibold">
                    Menu Tidak Dapat Dipesan: <span class="font-bold text-red-900">{ing.affected_products.join(', ')}</span>
                  </p>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- Low Stock Ingredients (Amber Warning Section) -->
    {#if lowStockIngredients.length > 0}
      <div class="space-y-2">
        <div class="flex items-center gap-2 text-xs font-extrabold text-amber-700 uppercase tracking-wider">
          <AlertCircle class="w-4 h-4" />
          <span>Bahan Baku Mencapai Batas Alert ({lowStockIngredients.length} Bahan)</span>
        </div>

        <div class="space-y-2 max-h-48 overflow-y-auto pr-1">
          {#each lowStockIngredients as ing}
            <div class="p-3 bg-amber-50/70 border border-amber-200 rounded-xl flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <div class="flex items-center gap-2">
                  <span class="font-extrabold text-amber-950 text-sm">{ing.name}</span>
                  <span class="text-xs font-bold text-amber-800 bg-amber-200/70 px-2 py-0.5 rounded-md">
                    Sisa: <strong class="text-amber-950 font-extrabold">{ing.current_stock} {ing.unit}</strong> (Batas Alert: {ing.min_stock_alert} {ing.unit})
                  </span>
                </div>
                {#if ing.affected_products.length > 0}
                  <p class="text-xs text-amber-800 mt-1 font-semibold">
                    Menu Terkait: <span class="font-bold text-amber-950">{ing.affected_products.join(', ')}</span>
                  </p>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}
  </div>

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={onClose}
      class="px-5 py-3 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
    >
      Mengerti & Lanjutkan Kasir
    </button>

    <button
      type="button"
      on:click={handleGoToInventory}
      class="px-6 py-3 font-extrabold text-white bg-amber-500 hover:bg-amber-600 rounded-xl shadow-md flex items-center gap-2 transition"
    >
      <span>Kelola Stok Bahan (Inventory)</span>
      <ArrowRight class="w-4 h-4" />
    </button>
  </svelte:fragment>
</Modal>
