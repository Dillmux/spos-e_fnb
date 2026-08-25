<script lang="ts">
  import { inventoryApi } from "$lib/api/inventory";
  import { featureFlags } from "$lib/stores/config";
  import { toast } from "$lib/stores/toast";
  import Modal from "$lib/components/common/Modal.svelte";
  import DisabledFeatureState from "$lib/components/common/DisabledFeatureState.svelte";
  import type { Product, Ingredient, ProductRecipe } from "$lib/types";
  import {
    Plus,
    Trash2,
    Save,
    Scale,
    DollarSign,
    Calculator,
  } from "lucide-svelte";

  export let isOpen: boolean = false;
  export let product: Product | null = null;
  export let onClose: () => void = () => {};

  let availableIngredients: Ingredient[] = [];
  let recipes: ProductRecipe[] = [];
  let isLoading = false;
  let isSaving = false;

  async function loadData() {
    if (!product) return;
    isLoading = true;
    try {
      const [ingRes, recRes] = await Promise.all([
        inventoryApi.getIngredients(),
        inventoryApi.getRecipe(product.id),
      ]);
      availableIngredients = ingRes.data || [];

      const fetchedRecipes = recRes.data || [];
      if (fetchedRecipes.length > 0) {
        recipes = fetchedRecipes.map((r: any) => {
          const ingId = r.ingredient_id || r.IngredientID || r.Ingredient?.id || "";
          const matchedIng = availableIngredients.find((i) => i.id === ingId);
          return {
            ingredient_id: ingId,
            ingredient_name: matchedIng?.name || r.Ingredient?.name || r.ingredient_name || "",
            quantity: r.quantity_required ?? r.quantity ?? 1,
            quantity_required: r.quantity_required ?? r.quantity ?? 1,
            unit: matchedIng?.unit || r.Ingredient?.unit || r.unit || "gram",
          };
        });
      } else {
        recipes = product.Recipes || product.recipe || [];
      }
    } catch (err: any) {
      availableIngredients = [];
      recipes = [];
      toast.error(err.message || 'Gagal memuat data resep & bahan baku');
    } finally {
      isLoading = false;
    }
  }

  $: if (isOpen && product) {
    loadData();
  }

  function addRecipeLine() {
    if (availableIngredients.length === 0) return;
    const firstIng = availableIngredients[0];
    recipes = [
      ...recipes,
      {
        ingredient_id: firstIng.id,
        ingredient_name: firstIng.name,
        quantity: 1,
        quantity_required: 1,
        unit: firstIng.unit,
      },
    ];
  }

  function removeRecipeLine(index: number) {
    recipes = recipes.filter((_, i) => i !== index);
  }

  function handleIngredientSelect(index: number, ingId: string) {
    const ing = availableIngredients.find((i) => i.id === ingId);
    if (ing) {
      recipes[index].ingredient_id = ing.id;
      recipes[index].ingredient_name = ing.name;
      recipes[index].unit = ing.unit;
    }
  }

  // Live estimated COGS calculation
  $: estimatedCOGS = recipes.reduce((total, r) => {
    const ing = availableIngredients.find((i) => i.id === r.ingredient_id);
    const cost = ing?.cost_per_unit ?? ing?.cogs_per_unit ?? 0;
    const qty = r.quantity_required ?? r.quantity ?? 0;
    return total + cost * qty;
  }, 0);

  $: estimatedMargin = (product?.price || 0) - estimatedCOGS;

  async function handleSave() {
    if (!product) return;
    isSaving = true;
    try {
      const payload = recipes.map((r) => ({
        ingredient_id: r.ingredient_id,
        quantity_required: Number(r.quantity_required ?? r.quantity ?? 0),
      }));

      await inventoryApi.saveRecipe(product.id, payload as any);
      toast.success(
        `Resep BOM untuk "${product.name}" berhasil disimpan! Total HPP: Rp ${Math.round(estimatedCOGS).toLocaleString("id-ID")}`,
      );
      onClose();
    } catch (err: any) {
      toast.error(err.message || "Gagal menyimpan resep BOM");
    } finally {
      isSaving = false;
    }
  }
</script>

<Modal
  {isOpen}
  title={`Setting Resep Formula (BOM) - ${product?.name || ""}`}
  maxWidth="xl"
  {onClose}
>
  {#if !$featureFlags.enable_bom}
    <DisabledFeatureState
      title="Modul Otomatisasi Stok BOM Nonaktif"
      description="Fitur kalkulasi HPP dan pengurangan stok bahan baku otomatis belum diaktifkan di Setup Wizard / Pengaturan Modul."
      icon={Scale}
      actionText="Tutup Dialog"
      actionHref=""
    />
  {:else if isLoading}
    <div class="py-8 text-center text-slate-400 font-semibold">
      Memuat Data Resep & Bahan Baku...
    </div>
  {:else}
    <div class="space-y-4">
      <!-- Info & Live COGS Banner -->
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div
          class="p-3 bg-emerald-50 text-emerald-900 rounded-xl text-xs font-semibold flex items-center gap-2 border border-emerald-200"
        >
          <Scale class="w-5 h-5 text-emerald-600 flex-shrink-0" />
          <span
            >Tentukan takaran bahan baku (gramasi/ml) yang akan terpotong
            otomatis saat transaksi lunas.</span
          >
        </div>

        <div
          class="p-3 bg-slate-900 text-white rounded-xl text-xs flex items-center justify-between shadow-sm"
        >
          <div>
            <span class="text-[10px] uppercase font-bold text-slate-400 block"
              >Kalkulasi HPP (COGS) Live</span
            >
            <span class="text-base font-extrabold text-brand-cyan">
              Rp {Math.round(estimatedCOGS).toLocaleString("id-ID")} / porsi
            </span>
          </div>
          <div class="text-right">
            <span class="text-[10px] uppercase font-bold text-slate-400 block"
              >Estimasi Laba Kotor</span
            >
            <span
              class="text-xs font-extrabold {estimatedMargin >= 0
                ? 'text-emerald-400'
                : 'text-red-400'}"
            >
              Rp {Math.round(estimatedMargin).toLocaleString("id-ID")}
            </span>
          </div>
        </div>
      </div>

      <!-- Recipe Rows -->
      {#if recipes.length === 0}
        <div
          class="p-8 text-center bg-slate-50 rounded-2xl border border-dashed border-slate-300 text-slate-500 space-y-2"
        >
          <Calculator class="w-8 h-8 mx-auto text-slate-400" />
          <p class="font-bold text-slate-700">
            Belum ada takaran bahan baku untuk menu ini
          </p>
          <p class="text-xs text-slate-400">
            Klik tombol di bawah untuk menambahkan komposisi bahan baku.
          </p>
        </div>
      {:else}
        <div class="space-y-2.5 max-h-80 overflow-y-auto pr-1">
          {#each recipes as recipe, index}
            {@const currentIng = availableIngredients.find(
              (i) => i.id === recipe.ingredient_id,
            )}
            {@const unitCost =
              currentIng?.cost_per_unit ?? currentIng?.cogs_per_unit ?? 0}
            {@const activeUnit = currentIng?.unit || recipe.unit || "unit"}
            {@const subtotalCost =
              unitCost * (recipe.quantity_required ?? recipe.quantity ?? 0)}

            <div
              class="p-3.5 bg-slate-50 rounded-xl border border-slate-200 grid grid-cols-12 gap-3 items-center"
            >
              <div class="col-span-5">
                <span
                  class="block text-[10px] font-bold text-slate-400 uppercase mb-0.5"
                  >Bahan Baku</span
                >
                <select
                  value={recipe.ingredient_id}
                  on:change={(e) =>
                    handleIngredientSelect(index, e.currentTarget.value)}
                  class="w-full px-3 py-2 bg-white border border-slate-300 rounded-lg text-xs font-bold text-slate-800 outline-none focus:border-brand-primary"
                >
                  {#each availableIngredients as ing}
                    <option value={ing.id}>
                      {ing.name} (Rp {(ing.cost_per_unit ?? ing.cogs_per_unit ?? 0).toLocaleString("id-ID")}/{ing.unit})
                    </option>
                  {/each}
                </select>
              </div>

              <div class="col-span-4">
                <span
                  class="block text-[10px] font-bold text-slate-400 uppercase mb-0.5"
                  >Takaran Per Porsi</span
                >
                <div class="flex items-center gap-1.5">
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    bind:value={recipe.quantity_required}
                    placeholder="Takaran"
                    class="w-full px-3 py-2 bg-white border border-slate-300 rounded-lg text-xs font-bold text-slate-800 outline-none focus:border-brand-primary"
                  />
                  <span
                    class="bg-slate-200 text-slate-800 text-xs px-2.5 py-2 rounded-lg font-extrabold uppercase whitespace-nowrap"
                  >
                    {activeUnit}
                  </span>
                </div>
              </div>

              <div class="col-span-2 text-right">
                <span
                  class="block text-[10px] font-bold text-slate-400 uppercase"
                  >Subtotal HPP</span
                >
                <span class="text-xs font-extrabold text-slate-900 block"
                  >Rp {Math.round(subtotalCost).toLocaleString("id-ID")}</span
                >
                <span class="text-[10px] text-slate-400 font-semibold"
                  >@ Rp {unitCost.toLocaleString("id-ID")}/{activeUnit}</span
                >
              </div>

              <div class="col-span-1 text-right">
                <button
                  type="button"
                  on:click={() => removeRecipeLine(index)}
                  class="p-2 text-slate-400 hover:text-brand-danger transition rounded-lg hover:bg-slate-200"
                  title="Hapus Komposisi Bahan"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}

      <button
        type="button"
        on:click={addRecipeLine}
        class="w-full py-3 bg-white hover:bg-slate-50 border-2 border-dashed border-slate-300 text-slate-700 font-bold text-xs rounded-xl flex items-center justify-center gap-2 transition"
      >
        <Plus class="w-4 h-4 text-brand-primary" />
        Tambah Komposisi Bahan Baku
      </button>
    </div>
  {/if}

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={onClose}
      class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
    >
      Batal
    </button>
    {#if $featureFlags.enable_bom}
      <button
        type="button"
        disabled={isSaving}
        on:click={handleSave}
        class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md flex items-center gap-2 transition"
      >
        <Save class="w-4 h-4" />
        {isSaving ? "Menyimpan..." : "Simpan Formula BOM"}
      </button>
    {/if}
  </svelte:fragment>
</Modal>
