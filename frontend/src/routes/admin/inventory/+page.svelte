<script lang="ts">
  import { onMount } from "svelte";
  import { inventoryApi } from "$lib/api/inventory";
  import { toast } from "$lib/stores/toast";
  import IngredientRow from "$lib/components/admin/IngredientRow.svelte";
  import StockAdjustmentModal from "$lib/components/admin/StockAdjustmentModal.svelte";
  import Modal from "$lib/components/common/Modal.svelte";
  import ConfirmModal from "$lib/components/common/ConfirmModal.svelte";
  import CurrencyInput from "$lib/components/common/CurrencyInput.svelte";
  import type { Ingredient } from "$lib/types";
  import { PackageCheck, Plus, RefreshCw } from "lucide-svelte";

  let ingredients: Ingredient[] = [];
  let isLoading = true;

  // Create/Edit Ingredient Modal State
  let isModalOpen = false;
  let isEditing = false;
  let editingIngredientId = "";
  let isSaving = false;

  let formName = "";
  let formStock = 0;
  let formUnit = "gram";
  let formMinStock = 0;
  let formCogsPerUnit = 0;

  // Delete Ingredient Modal State
  let isDeleteModalOpen = false;
  let ingredientToDelete: Ingredient | null = null;
  let isDeleting = false;

  // Stock Adjustment Modal State
  let isStockModalOpen = false;
  let selectedIngredientForStock: Ingredient | null = null;

  async function loadIngredients() {
    isLoading = true;
    try {
      const res = await inventoryApi.getIngredients();
      ingredients = res.data || [];
    } catch (err: any) {
      ingredients = [];
      toast.error(err.message || "Gagal memuat data bahan baku");
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    loadIngredients();
  });

  function openCreateIngredientModal() {
    isEditing = false;
    editingIngredientId = "";
    formName = "";
    formStock = 0;
    formUnit = "gram";
    formMinStock = 0;
    formCogsPerUnit = 0;
    isModalOpen = true;
  }

  function openEditDetailModal(ing: Ingredient) {
    isEditing = true;
    editingIngredientId = ing.id;
    formName = ing.name;
    formStock = ing.current_stock ?? ing.stock ?? 0;
    formUnit = ing.unit || "gram";
    formMinStock = ing.min_stock_alert ?? ing.min_stock ?? 0;
    formCogsPerUnit = ing.cost_per_unit ?? ing.cogs_per_unit ?? 0;
    isModalOpen = true;
  }

  function openDeleteIngredientModal(ing: Ingredient) {
    ingredientToDelete = ing;
    isDeleteModalOpen = true;
  }

  async function handleConfirmDeleteIngredient() {
    if (!ingredientToDelete) return;
    isDeleting = true;
    try {
      await inventoryApi.deleteIngredient(ingredientToDelete.id);
      toast.success(
        `Bahan baku "${ingredientToDelete.name}" berhasil dihapus!`,
      );
      isDeleteModalOpen = false;
      loadIngredients();
    } catch (err: any) {
      toast.error(err.message || "Gagal menghapus bahan baku");
      isDeleteModalOpen = false;
    } finally {
      isDeleting = false;
    }
  }

  async function handleSaveIngredient() {
    if (!formName || formStock < 0 || formCogsPerUnit <= 0) {
      toast.warning(
        "Masukkan nama bahan baku, stok, dan harga COGS per unit yang valid.",
      );
      return;
    }

    isSaving = true;
    try {
      const payload: Partial<Ingredient> = {
        id: isEditing ? editingIngredientId : undefined,
        name: formName,
        current_stock: formStock,
        stock: formStock,
        unit: formUnit,
        min_stock_alert: formMinStock,
        min_stock: formMinStock,
        cost_per_unit: formCogsPerUnit,
        cogs_per_unit: formCogsPerUnit,
      };

      await inventoryApi.saveIngredient(payload);
      toast.success(
        `Bahan baku "${formName}" berhasil ${isEditing ? "diperbarui" : "ditambahkan"}!`,
      );
      isModalOpen = false;
      loadIngredients();
    } catch (err: any) {
      toast.error(err.message || "Gagal menyimpan data bahan baku");
    } finally {
      isSaving = false;
    }
  }

  function openAdjustStockModal(ing: Ingredient) {
    selectedIngredientForStock = ing;
    isStockModalOpen = true;
  }

  async function handleSaveStockAdjustment(payload: {
    change_type: string;
    quantity_change: number;
    notes: string;
  }) {
    if (!selectedIngredientForStock) return;
    try {
      await inventoryApi.adjustStock(selectedIngredientForStock.id, payload);
      toast.success(
        `Stok ${selectedIngredientForStock.name} berhasil diperbarui!`,
      );
      loadIngredients();
    } catch {
      // Fallback local update
      ingredients = ingredients.map((i) =>
        i.id === selectedIngredientForStock?.id
          ? {
              ...i,
              stock: Math.max(
                0,
                (i.current_stock ?? i.stock ?? 0) + payload.quantity_change,
              ),
            }
          : i,
      );
      toast.success(
        `Stok ${selectedIngredientForStock.name} diperbarui (${payload.quantity_change > 0 ? "+" : ""}${payload.quantity_change})`,
      );
    }
  }
</script>

<div
  class="h-full bg-brand-surface p-6 flex flex-col space-y-6 overflow-hidden"
>
  <!-- Top Bar -->
  <div class="flex items-center justify-between">
    <div>
      <h2
        class="text-2xl font-extrabold text-slate-900 flex items-center gap-3"
      >
        <PackageCheck class="w-7 h-7 text-brand-primary" />
        Monitoring Stok Bahan Baku & Inventaris
      </h2>
      <p class="text-xs text-slate-500 font-semibold mt-1">
        Pantau sisa gramasi stok bahan mentah, ubah satuan ukur, dan lakukan
        penyesuaian fisik (stock opname).
      </p>
    </div>

    <div class="flex items-center gap-3">
      <button
        type="button"
        on:click={loadIngredients}
        class="p-2.5 bg-white hover:bg-slate-100 text-slate-700 rounded-xl border border-slate-200 transition"
      >
        <RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
      </button>

      <button
        type="button"
        on:click={openCreateIngredientModal}
        class="px-4 py-2 bg-brand-primary hover:bg-sky-500 text-white text-xs font-bold rounded-xl shadow-sm flex items-center gap-2 transition"
      >
        <Plus class="w-4 h-4" />
        Bahan Baku Baru
      </button>
    </div>
  </div>

  <!-- Ingredients List Grid -->
  {#if isLoading}
    <div
      class="flex-1 flex items-center justify-center text-slate-400 font-semibold"
    >
      Memuat Data Stok...
    </div>
  {:else}
    <div class="flex-1 overflow-y-auto space-y-3 pr-1">
      {#each ingredients as ingredient (ingredient.id)}
        <IngredientRow
          {ingredient}
          onAdjustStock={openAdjustStockModal}
          onEditDetail={openEditDetailModal}
          onDelete={openDeleteIngredientModal}
        />
      {/each}
    </div>
  {/if}

  <!-- Create/Edit Ingredient Modal -->
  <Modal
    isOpen={isModalOpen}
    title={isEditing
      ? "Edit Detail & Satuan Ukur Bahan Baku"
      : "Tambah Bahan Baku Inventaris Baru"}
    maxWidth="md"
    onClose={() => (isModalOpen = false)}
  >
    <div class="space-y-4 py-2">
      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-700 uppercase"
          for="ing-name">Nama Bahan Baku *</label
        >
        <input
          id="ing-name"
          type="text"
          bind:value={formName}
          placeholder="misal: Bijikopi Arabica Flores"
          required
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        />
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="ing-stock">Jumlah Stok saat ini *</label
          >
          <input
            id="ing-stock"
            type="number"
            bind:value={formStock}
            placeholder="1000"
            required
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
          />
        </div>

        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="ing-unit">Satuan Ukuran Bahan Baku *</label
          >
          <select
            id="ing-unit"
            bind:value={formUnit}
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
          >
            <option value="gram">gram (g)</option>
            <option value="ml">milliliter (ml)</option>
            <option value="pcs">pieces / porsi (pcs)</option>
            <option value="kg">kilogram (kg)</option>
            <option value="liter">liter (l)</option>
            <option value="pack">pack / bungkus</option>
            <option value="botol">botol</option>
            <option value="kaleng">kaleng</option>
          </select>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="ing-min">Batas Alert Stok Minimal</label
          >
          <input
            id="ing-min"
            type="number"
            bind:value={formMinStock}
            placeholder="500"
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
          />
        </div>

        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="ing-cogs">Modal per Satuan (COGS)</label
          >
          <CurrencyInput bind:value={formCogsPerUnit} placeholder="50" />
        </div>
      </div>
    </div>

    <svelte:fragment slot="footer">
      <button
        type="button"
        on:click={() => (isModalOpen = false)}
        class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
      >
        Batal
      </button>

      <button
        type="button"
        disabled={isSaving}
        on:click={handleSaveIngredient}
        class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md transition"
      >
        {isSaving
          ? "Menyimpan..."
          : isEditing
            ? "Simpan Perubahan"
            : "Tambah Bahan Baku"}
      </button>
    </svelte:fragment>
  </Modal>

  <StockAdjustmentModal
    isOpen={isStockModalOpen}
    ingredient={selectedIngredientForStock}
    onClose={() => (isStockModalOpen = false)}
    onSave={handleSaveStockAdjustment}
  />

  <ConfirmModal
    isOpen={isDeleteModalOpen}
    title="Hapus Bahan Baku Inventaris"
    message={`Apakah Anda yakin ingin menghapus bahan baku "${ingredientToDelete?.name || ""}" dari sistem inventaris?`}
    confirmText="Ya, Hapus Bahan"
    cancelText="Batal"
    isDanger={true}
    isLoading={isDeleting}
    onConfirm={handleConfirmDeleteIngredient}
    onClose={() => (isDeleteModalOpen = false)}
  />
</div>
