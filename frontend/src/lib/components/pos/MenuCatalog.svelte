<script lang="ts">
  import { onMount } from "svelte";
  import { posApi } from "$lib/api/pos";
  import { inventoryApi } from "$lib/api/inventory";
  import { cart } from "$lib/stores/cart";
  import type { Category, Product, Ingredient } from "$lib/types";
  import {
    Search,
    Plus,
    Coffee,
    Utensils,
    AlertCircle,
    ShieldAlert,
  } from "lucide-svelte";
  import { toast } from "$lib/stores/toast";
  import StockAlertModal from "./StockAlertModal.svelte";
  import { validateBOMStockAddition } from "$lib/utils/bomStock";

  let categories: Category[] = [];
  let products: Product[] = [];
  let ingredients: Ingredient[] = [];
  let selectedCategory: string | number = "all";
  let searchQuery: string = "";
  let searchInput: HTMLInputElement;
  let isLoading = true;

  let isStockAlertOpen = false;
  let hasAutoAlerted = false;

  interface StockIngredientAlert {
    id: string;
    name: string;
    current_stock: number;
    min_stock_alert: number;
    unit: string;
    affected_products: string[];
  }

  let lowStockList: StockIngredientAlert[] = [];
  let depletedList: StockIngredientAlert[] = [];

  function checkProductStockStatus(product: Product) {
    if (!product.is_available) {
      return {
        isOutOfStock: true,
        isLowStock: false,
        depletedReason: "Menu Nonaktif",
        remainingPortions: 0,
      };
    }

    const validation = validateBOMStockAddition(product, 1, $cart.items);
    const recipes =
      product.recipes ||
      (product as any).Recipes ||
      (product as any).recipe ||
      [];

    if (recipes.length === 0) {
      return {
        isOutOfStock: false,
        isLowStock: false,
        depletedReason: "",
        remainingPortions: 999,
      };
    }

    let isLow = false;
    for (const r of recipes) {
      const ing: Ingredient | undefined = r.Ingredient || (r as any).ingredient;
      if (!ing) continue;

      const currentStock = ing.current_stock ?? ing.stock ?? 0;
      const alertThreshold = ing.min_stock_alert ?? ing.min_stock ?? 0;

      if (alertThreshold > 0 && currentStock <= alertThreshold) {
        isLow = true;
        break;
      }
    }

    if (!validation.allowed) {
      return {
        isOutOfStock: true,
        isLowStock: isLow,
        depletedReason:
          validation.reason ||
          "Stok bahan baku di keranjang telah mencapai batas",
        remainingPortions: validation.maxOrderablePortions ?? 0,
      };
    }

    return {
      isOutOfStock: false,
      isLowStock: isLow,
      depletedReason: "",
      remainingPortions: validation.maxOrderablePortions ?? 999,
    };
  }

  function evaluateAllInventoryAlerts() {
    const lowStockMap = new Map<string, StockIngredientAlert>();
    const depletedMap = new Map<string, StockIngredientAlert>();

    // 1. Map raw ingredients from Inventory API
    for (const ing of ingredients) {
      const curStock = ing.current_stock ?? ing.stock ?? 0;
      const minAlert = ing.min_stock_alert ?? ing.min_stock ?? 0;
      const unit = ing.unit || "unit";

      if (curStock <= 0) {
        depletedMap.set(ing.id, {
          id: ing.id,
          name: ing.name,
          current_stock: curStock,
          min_stock_alert: minAlert,
          unit,
          affected_products: [],
        });
      } else if (minAlert > 0 && curStock <= minAlert) {
        lowStockMap.set(ing.id, {
          id: ing.id,
          name: ing.name,
          current_stock: curStock,
          min_stock_alert: minAlert,
          unit,
          affected_products: [],
        });
      }
    }

    // 2. Associate affected product names
    for (const prod of products) {
      const recipes =
        prod.recipes || (prod as any).Recipes || (prod as any).recipe || [];
      for (const r of recipes) {
        const ing: Ingredient | undefined =
          r.Ingredient || (r as any).ingredient;
        const ingId = r.ingredient_id || ing?.id;
        if (!ingId) continue;

        const curStock = ing?.current_stock ?? ing?.stock ?? 0;
        const minAlert = ing?.min_stock_alert ?? ing?.min_stock ?? 0;
        const ingName = ing?.name || r.ingredient_name || "Bahan Baku";
        const unit = ing?.unit || r.unit || "unit";

        if (curStock <= 0) {
          if (!depletedMap.has(ingId)) {
            depletedMap.set(ingId, {
              id: ingId,
              name: ingName,
              current_stock: curStock,
              min_stock_alert: minAlert,
              unit,
              affected_products: [prod.name],
            });
          } else {
            const existing = depletedMap.get(ingId)!;
            if (!existing.affected_products.includes(prod.name)) {
              existing.affected_products.push(prod.name);
            }
          }
        } else if (minAlert > 0 && curStock <= minAlert) {
          if (!lowStockMap.has(ingId)) {
            lowStockMap.set(ingId, {
              id: ingId,
              name: ingName,
              current_stock: curStock,
              min_stock_alert: minAlert,
              unit,
              affected_products: [prod.name],
            });
          } else {
            const existing = lowStockMap.get(ingId)!;
            if (!existing.affected_products.includes(prod.name)) {
              existing.affected_products.push(prod.name);
            }
          }
        }
      }
    }

    lowStockList = Array.from(lowStockMap.values());
    depletedList = Array.from(depletedMap.values());

    if (
      (lowStockList.length > 0 || depletedList.length > 0) &&
      !hasAutoAlerted
    ) {
      isStockAlertOpen = true;
      hasAutoAlerted = true;
    }
  }

  function getCategoryName(product: Product): string {
    if (product.category_name) return product.category_name;
    if (product.Category?.name) return product.Category.name;
    if ((product as any).category?.name) return (product as any).category.name;
    if (product.category_id) {
      const match = categories.find(
        (c) => String(c.id) === String(product.category_id),
      );
      if (match) return match.name;
    }
    return "Resto Item";
  }

  async function loadCatalog() {
    isLoading = true;
    try {
      const [catRes, prodRes, ingRes] = await Promise.all([
        posApi.getCategories(),
        posApi.getProducts(),
        inventoryApi.getIngredients(),
      ]);
      categories = catRes.data || [];
      const rawProducts: Product[] = prodRes.data || [];
      products = rawProducts.map((p) => ({
        ...p,
        category_name: getCategoryName(p),
      }));
      ingredients = ingRes.data || [];
      evaluateAllInventoryAlerts();
    } catch (err: any) {
      categories = [];
      products = [];
      ingredients = [];
      toast.error(err.message || "Gagal memuat katalog menu");
    } finally {
      isLoading = false;
    }
  }

  function handleAddToCart(product: Product) {
    const validation = validateBOMStockAddition(product, 1, $cart.items);
    if (!validation.allowed) {
      toast.warning(
        validation.reason || `Menu "${product.name}" tidak dapat dipesan.`,
      );
      return;
    }
    cart.addItem(product);
  }

  onMount(() => {
    loadCatalog();

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "F2") {
        e.preventDefault();
        searchInput?.focus();
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  });

  $: filteredProducts = products.filter((p) => {
    const matchCategory =
      selectedCategory === "all" || p.category_id === selectedCategory;
    const matchSearch = p.name
      .toLowerCase()
      .includes(searchQuery.toLowerCase());
    return matchCategory && matchSearch;
  });
</script>

<div class="flex flex-col h-full space-y-4">
  <!-- Search Bar + Keyboard Shortcut F2 Badge -->
  <div class="relative">
    <Search class="w-5 h-5 absolute left-4 top-3.5 text-slate-400" />
    <input
      bind:this={searchInput}
      type="text"
      bind:value={searchQuery}
      placeholder="Cari Menu / Ketik F2..."
      class="w-full pl-12 pr-16 py-3 min-h-[48px] bg-white border border-slate-200 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-2xl font-semibold text-slate-800 text-base shadow-sm outline-none transition"
    />
    <span
      class="absolute right-3 top-3 text-xs font-bold px-2 py-1 bg-slate-100 text-slate-500 rounded-md border border-slate-200"
    >
      F2
    </span>
  </div>

  <!-- Persistent Stock Alert Banner (Always Visible If Any Low Stock or Depleted Material Exists) -->
  {#if lowStockList.length > 0 || depletedList.length > 0}
    <button
      type="button"
      on:click={() => (isStockAlertOpen = true)}
      class="w-full py-3 px-4 bg-amber-50 border-2 border-amber-300 text-amber-950 hover:bg-amber-100 rounded-2xl text-xs font-extrabold flex items-center justify-between shadow-sm transition"
    >
      <span class="flex items-center gap-2">
        <ShieldAlert class="w-5 h-5 text-amber-600 animate-pulse" />
        <span>
          Peringatan Inventaris Bahan Baku:
          {#if depletedList.length > 0}
            <strong
              class="text-brand-danger font-extrabold bg-red-100 px-2 py-0.5 rounded-md"
              >{depletedList.length} Bahan Habis</strong
            >
          {/if}
          {#if depletedList.length > 0 && lowStockList.length > 0}
            |
          {/if}
          {#if lowStockList.length > 0}
            <strong
              class="text-amber-800 font-extrabold bg-amber-200/80 px-2 py-0.5 rounded-md"
              >{lowStockList.length} Mencapai Batas Alert</strong
            >
          {/if}
        </span>
      </span>
      <span
        class="underline text-amber-900 font-extrabold flex items-center gap-1"
      >
        Lihat Alert &rarr;
      </span>
    </button>
  {/if}

  <!-- Category Pills -->
  <div class="flex items-center gap-2 overflow-x-auto pb-1 no-scrollbar">
    <button
      type="button"
      on:click={() => (selectedCategory = "all")}
      class="touch-target px-5 py-2.5 rounded-xl font-bold text-sm whitespace-nowrap transition shadow-sm
      {selectedCategory === 'all'
        ? 'bg-brand-primary text-white'
        : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'}"
    >
      Semua Menu
    </button>
    {#each categories as cat}
      <button
        type="button"
        on:click={() => (selectedCategory = cat.id)}
        class="touch-target px-5 py-2.5 rounded-xl font-bold text-sm whitespace-nowrap transition shadow-sm
        {selectedCategory === cat.id
          ? 'bg-brand-primary text-white'
          : 'bg-white text-slate-600 hover:bg-slate-100 border border-slate-200'}"
      >
        {cat.name}
      </button>
    {/each}
  </div>

  <!-- Product Catalog Grid -->
  {#if isLoading}
    <div
      class="flex-1 flex items-center justify-center text-slate-400 font-semibold"
    >
      Memuat Katalog Menu...
    </div>
  {:else if filteredProducts.length === 0}
    <div
      class="flex-1 flex flex-col items-center justify-center text-slate-400 p-8"
    >
      <AlertCircle class="w-12 h-12 mb-2 text-slate-300" />
      <p class="font-bold text-slate-600">Menu tidak ditemukan</p>
      <p class="text-xs text-slate-400 mt-1">
        Coba gunakan kata kunci pencarian atau kategori lain.
      </p>
    </div>
  {:else}
    <div
      class="flex-1 overflow-y-auto grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-3 xl:grid-cols-4 gap-3.5 pr-1"
    >
      {#each filteredProducts as product}
        {@const stockInfo = checkProductStockStatus(product)}
        <button
          type="button"
          disabled={stockInfo.isOutOfStock || !product.is_available}
          on:click={() => handleAddToCart(product)}
          class="touch-target p-4 bg-white hover:bg-slate-50 active:bg-sky-50 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed rounded-2xl border border-slate-200 shadow-sm hover:shadow-md flex flex-col justify-between text-left transition-all duration-150 h-36 relative group overflow-hidden"
        >
          <div class="text-center">
            <h4
              class="font-bold text-slate-800 text-base leading-snug line-clamp-2"
            >
              {product.name}
            </h4>
            <span class="text-xs font-semibold text-slate-400 block mt-1">
              {getCategoryName(product)}
            </span>
          </div>

          <div
            class="flex items-center justify-between mt-2 pt-2 border-t border-slate-100 space-x-2"
          >
            <span class="font-extrabold text-brand-primary text-base">
              Rp {product.price.toLocaleString("id-ID")}
            </span>

            <div
              class="w-9 h-9 rounded-xl bg-brand-primary/10 group-hover:bg-brand-primary group-hover:text-white text-brand-primary flex items-center justify-center transition"
            >
              <Plus class="w-5 h-5" />
            </div>
          </div>

          {#if stockInfo.isOutOfStock}
            <span
              class="absolute top-2 right-2 text-[10px] font-extrabold px-2 py-0.5 bg-brand-danger text-white rounded-md uppercase shadow-sm"
              title={stockInfo.depletedReason}
            >
              Stok Bahan Habis
            </span>
          {:else if stockInfo.isLowStock}
            <span
              class="absolute top-2 right-2 text-[10px] font-extrabold px-2 py-0.5 bg-amber-500 text-white rounded-md uppercase shadow-sm"
            >
              Stok Minim
            </span>
          {/if}
        </button>
      {/each}
    </div>
  {/if}
</div>

<StockAlertModal
  isOpen={isStockAlertOpen}
  depletedIngredients={depletedList}
  lowStockIngredients={lowStockList}
  onClose={() => (isStockAlertOpen = false)}
/>
