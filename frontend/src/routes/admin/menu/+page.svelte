<script lang="ts">
  import { onMount } from 'svelte';
  import { posApi } from '$lib/api/pos';
  import { featureFlags } from '$lib/stores/config';
  import { toast } from '$lib/stores/toast';
  import RecipeMapperModal from '$lib/components/admin/RecipeMapperModal.svelte';
  import Modal from '$lib/components/common/Modal.svelte';
  import ConfirmModal from '$lib/components/common/ConfirmModal.svelte';
  import CurrencyInput from '$lib/components/common/CurrencyInput.svelte';
  import type { Product, Category } from '$lib/types';
  import { Coffee, Plus, Scale, Edit2, Search, FolderPlus, Layers, Trash2 } from 'lucide-svelte';

  let products: Product[] = [];
  let categories: Category[] = [];
  let isLoading = true;
  let searchQuery = '';
  let selectedCategoryFilter: string | number = 'all';

  let selectedProductForBOM: Product | null = null;
  let isBOMModalOpen = false;

  // Product Modal State
  let isProductModalOpen = false;
  let isEditingProduct = false;
  let editingProductId = '';
  let isSavingProduct = false;

  let formProductName = '';
  let formProductSKU = '';
  let formProductCategoryId: string | number = '';
  let formProductPrice = 0;
  let formProductCOGS = 0;
  let formProductIsAvailable = true;

  // Product Delete Modal State
  let isDeleteProductModalOpen = false;
  let productToDelete: Product | null = null;

  // Category Modal State
  let isCategoryModalOpen = false;
  let isEditingCategory = false;
  let editingCategoryId: string | number = '';
  let isSavingCategory = false;
  let formCategoryName = '';
  let formCategorySortOrder = 0;

  // Category Delete Modal State
  let isDeleteCategoryModalOpen = false;
  let categoryToDelete: Category | null = null;
  let isDeletingCategory = false;

  async function loadData() {
    isLoading = true;
    try {
      const [prodRes, catRes] = await Promise.all([
        posApi.getProducts(),
        posApi.getCategories(),
      ]);
      products = prodRes.data || [];
      categories = catRes.data || [];
    } catch (err: any) {
      products = [];
      categories = [];
      toast.error(err.message || 'Gagal memuat data menu & kategori');
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    loadData();
  });

  // Product Handlers
  function openCreateProductModal() {
    isEditingProduct = false;
    editingProductId = '';
    formProductName = '';
    formProductSKU = '';
    formProductCategoryId = categories[0]?.id || '';
    formProductPrice = 0;
    formProductCOGS = 0;
    formProductIsAvailable = true;
    isProductModalOpen = true;
  }

  function openEditProductModal(product: Product) {
    isEditingProduct = true;
    editingProductId = product.id;
    formProductName = product.name;
    formProductSKU = product.sku || '';
    formProductCategoryId = product.category_id;
    formProductPrice = product.price;
    formProductCOGS = product.cogs || 0;
    formProductIsAvailable = product.is_available;
    isProductModalOpen = true;
  }

  function openDeleteProductModal(product: Product) {
    productToDelete = product;
    isDeleteProductModalOpen = true;
  }

  async function handleToggleAvailability(product: Product) {
    try {
      await posApi.saveProduct({
        id: product.id,
        is_available: !product.is_available,
      });
      toast.success(`Status ketersediaan menu "${product.name}" berhasil diubah!`);
      loadData();
    } catch (err: any) {
      toast.error(err.message || 'Gagal mengubah status ketersediaan menu');
    }
  }

  async function handleConfirmDeleteProduct() {
    if (!productToDelete) return;
    try {
      await posApi.deleteProduct(productToDelete.id);
      toast.success(`Menu "${productToDelete.name}" berhasil dihapus secara permanen!`);
      isDeleteProductModalOpen = false;
      productToDelete = null;
      loadData();
    } catch (err: any) {
      toast.error(err.message || 'Gagal menghapus menu');
      isDeleteProductModalOpen = false;
      productToDelete = null;
    }
  }

  async function handleSaveProduct() {
    if (!formProductName || formProductPrice <= 0) {
      toast.warning('Masukkan nama menu dan harga jual yang valid.');
      return;
    }

    isSavingProduct = true;
    try {
      const payload: Partial<Product> = {
        id: isEditingProduct ? editingProductId : undefined,
        name: formProductName,
        sku: formProductSKU || undefined,
        category_id: formProductCategoryId,
        price: formProductPrice,
        cogs: formProductCOGS,
        is_available: formProductIsAvailable,
      };

      await posApi.saveProduct(payload);
      toast.success(`Menu "${formProductName}" berhasil ${isEditingProduct ? 'diperbarui' : 'ditambahkan'}!`);
      isProductModalOpen = false;
      loadData();
    } catch (err: any) {
      toast.error(err.message || 'Gagal menyimpan data menu');
    } finally {
      isSavingProduct = false;
    }
  }

  // Category Handlers
  function openCreateCategoryModal() {
    isEditingCategory = false;
    editingCategoryId = '';
    formCategoryName = '';
    formCategorySortOrder = categories.length + 1;
    isCategoryModalOpen = true;
  }

  function openEditCategoryModal(cat: Category, e?: Event) {
    if (e) e.stopPropagation();
    isEditingCategory = true;
    editingCategoryId = cat.id;
    formCategoryName = cat.name;
    formCategorySortOrder = cat.sort_order || 1;
    isCategoryModalOpen = true;
  }

  function openDeleteCategoryModal(cat: Category, e?: Event) {
    if (e) e.stopPropagation();
    categoryToDelete = cat;
    isDeleteCategoryModalOpen = true;
  }

  async function handleConfirmDeleteCategory() {
    if (!categoryToDelete) return;
    const delId = categoryToDelete.id;
    const catName = categoryToDelete.name;
    isDeletingCategory = true;
    try {
      await posApi.deleteCategory(delId);
      toast.success(`Kategori "${catName}" berhasil dihapus!`);
      isDeleteCategoryModalOpen = false;
      categoryToDelete = null;
      if (selectedCategoryFilter === delId) {
        selectedCategoryFilter = 'all';
      }
      loadData();
    } catch (err: any) {
      toast.error(err.message || 'Gagal menghapus kategori');
    } finally {
      isDeletingCategory = false;
    }
  }

  async function handleSaveCategory() {
    if (!formCategoryName.trim()) {
      toast.warning('Nama kategori wajib diisi.');
      return;
    }

    isSavingCategory = true;
    try {
      if (isEditingCategory) {
        await posApi.updateCategory(editingCategoryId, {
          name: formCategoryName.trim(),
          sort_order: formCategorySortOrder,
        });
        toast.success(`Kategori "${formCategoryName}" berhasil diperbarui!`);
      } else {
        await posApi.createCategory({
          name: formCategoryName.trim(),
          sort_order: formCategorySortOrder,
        });
        toast.success(`Kategori "${formCategoryName}" berhasil dibuat!`);
      }

      isCategoryModalOpen = false;
      loadData();
    } catch (err: any) {
      toast.error(err.message || 'Gagal menyimpan kategori');
    } finally {
      isSavingCategory = false;
    }
  }

  function openRecipeBOM(product: Product) {
    selectedProductForBOM = product;
    isBOMModalOpen = true;
  }

  $: filteredProducts = products.filter((p) => {
    const matchesSearch = p.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCat = selectedCategoryFilter === 'all' || p.category_id === selectedCategoryFilter;
    return matchesSearch && matchesCat;
  });
</script>

<div class="h-full bg-brand-surface p-6 flex flex-col space-y-6 overflow-hidden">
  <!-- Top Bar -->
  <div class="flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
    <div>
      <h2 class="text-2xl font-extrabold text-slate-900 flex items-center gap-3">
        <Coffee class="w-7 h-7 text-brand-primary" />
        Manajemen Menu Katalog & Formula BOM
      </h2>
      <p class="text-xs text-slate-500 font-semibold mt-1">
        Kelola kategori, harga jual menu, dan takaran resep bahan baku per porsi.
      </p>
    </div>

    <!-- Action Buttons -->
    <div class="flex items-center gap-2.5">
      <button
        type="button"
        on:click={openCreateCategoryModal}
        class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white text-xs font-bold rounded-xl shadow-sm flex items-center gap-2 transition"
      >
        <FolderPlus class="w-4 h-4 text-brand-cyan" />
        Tambah Kategori
      </button>

      <button
        type="button"
        on:click={openCreateProductModal}
        class="px-4 py-2 bg-brand-primary hover:bg-sky-500 text-white text-xs font-bold rounded-xl shadow-sm flex items-center gap-2 transition"
      >
        <Plus class="w-4 h-4" />
        Tambah Menu Baru
      </button>
    </div>
  </div>

  <!-- Category Filter & Search Bar -->
  <div class="flex flex-col sm:flex-row items-center justify-between gap-3 bg-white p-4 rounded-2xl border border-slate-200 shadow-sm">
    <!-- Category Tabs Filter -->
    <div class="flex items-center gap-1.5 overflow-x-auto w-full sm:w-auto no-scrollbar">
      <button
        type="button"
        on:click={() => (selectedCategoryFilter = 'all')}
        class="px-4 py-2.5 rounded-xl text-xs font-bold transition whitespace-nowrap min-h-[44px] flex items-center
        {selectedCategoryFilter === 'all' ? 'bg-slate-900 text-white shadow-sm font-extrabold' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'}"
      >
        <Layers class="w-4 h-4 inline mr-1" /> Semua Menu
      </button>
      {#each categories as cat}
        <div class="flex items-center gap-1 bg-slate-100 rounded-xl px-1">
          <button
            type="button"
            on:click={() => (selectedCategoryFilter = cat.id)}
            class="px-3 py-2.5 rounded-xl text-xs font-bold transition whitespace-nowrap min-h-[44px] flex items-center gap-1.5
            {selectedCategoryFilter === cat.id ? 'bg-slate-900 text-white shadow-sm font-extrabold' : 'text-slate-700 hover:bg-slate-200'}"
          >
            {cat.name}
          </button>
          <button
            type="button"
            on:click={(e) => openEditCategoryModal(cat, e)}
            class="p-1.5 text-slate-400 hover:text-brand-primary transition rounded-lg hover:bg-slate-200"
            title="Edit Kategori"
          >
            <Edit2 class="w-3.5 h-3.5" />
          </button>
          <button
            type="button"
            on:click={(e) => openDeleteCategoryModal(cat, e)}
            class="p-1.5 text-slate-400 hover:text-red-600 transition rounded-lg hover:bg-red-50"
            title="Hapus Kategori"
          >
            <Trash2 class="w-3.5 h-3.5" />
          </button>
        </div>
      {/each}
    </div>

    <!-- Search Input -->
    <div class="relative w-full sm:w-64">
      <Search class="w-4 h-4 absolute left-3 top-3 text-slate-400" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Cari nama menu..."
        class="w-full pl-9 pr-4 py-2.5 min-h-[44px] bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold outline-none focus:border-brand-primary focus:bg-white transition"
      />
    </div>
  </div>

  <!-- Product List Table -->
  <div class="flex-1 bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex flex-col">
    {#if isLoading}
      <div class="flex-1 flex items-center justify-center text-slate-400 font-semibold">Memuat Data Menu...</div>
    {:else if filteredProducts.length === 0}
      <div class="flex-1 flex flex-col items-center justify-center text-slate-400 p-8 text-center space-y-2">
        <Coffee class="w-12 h-12 text-slate-300" />
        <p class="font-bold text-slate-600">Tidak ada menu ditemukan</p>
        <p class="text-xs text-slate-400">Coba ubah kata kunci atau pilih kategori menu lainnya.</p>
      </div>
    {:else}
      <div class="flex-1 overflow-x-auto overflow-y-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-xs font-bold text-slate-500 uppercase tracking-wider">
              <th class="py-3.5 px-6">Nama Menu / SKU</th>
              <th class="py-3.5 px-6">Kategori</th>
              <th class="py-3.5 px-6">Harga Jual / HPP</th>
              <th class="py-3.5 px-6">Status Ketersediaan</th>
              <th class="py-3.5 px-6 text-right">Aksi Formulasi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-sm font-semibold text-slate-800">
            {#each filteredProducts as product}
              <tr class="hover:bg-slate-50 transition">
                <td class="py-4 px-6">
                  <p class="font-bold text-slate-900">{product.name}</p>
                  {#if product.sku}
                    <p class="text-[11px] font-mono text-slate-400">SKU: {product.sku}</p>
                  {/if}
                </td>
                <td class="py-4 px-6 text-xs text-slate-500 font-bold">{product.category_name || 'Resto Item'}</td>
                <td class="py-4 px-6">
                  <p class="text-brand-primary font-extrabold">Rp {product.price.toLocaleString('id-ID')}</p>
                  {#if product.cogs && product.cogs > 0}
                    <p class="text-[11px] text-slate-400 font-semibold">HPP: Rp {product.cogs.toLocaleString('id-ID')}</p>
                  {/if}
                </td>
                <td class="py-4 px-6">
                  <button
                    type="button"
                    on:click={() => handleToggleAvailability(product)}
                    class="text-xs font-bold px-2.5 py-1 rounded-full transition cursor-pointer hover:opacity-80 {product.is_available ? 'bg-emerald-100 text-emerald-800' : 'bg-red-100 text-red-800'}"
                    title="Klik untuk ubah ketersediaan (Tersedia/Habis)"
                  >
                    {product.is_available ? 'Tersedia' : 'Habis'}
                  </button>
                </td>
                <td class="py-4 px-6 text-right space-x-2">
                  {#if $featureFlags.enable_bom}
                    <button
                      type="button"
                      on:click={() => openRecipeBOM(product)}
                      class="px-3 py-1.5 bg-emerald-50 hover:bg-emerald-100 text-emerald-700 text-xs font-bold rounded-xl border border-emerald-200 inline-flex items-center gap-1 transition"
                    >
                      <Scale class="w-3.5 h-3.5" />
                      Setting BOM Resep
                    </button>
                  {/if}

                  <button
                    type="button"
                    on:click={() => openEditProductModal(product)}
                    class="p-1.5 text-slate-400 hover:text-slate-700 transition"
                    title="Edit Menu"
                  >
                    <Edit2 class="w-4 h-4" />
                  </button>

                  <button
                    type="button"
                    on:click={() => openDeleteProductModal(product)}
                    class="p-1.5 text-slate-400 hover:text-red-600 transition"
                    title="Hapus Menu Permanen"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>

  <!-- Create/Edit Product Modal -->
  <Modal
    isOpen={isProductModalOpen}
    title={isEditingProduct ? 'Edit Menu Resto' : 'Tambah Menu Baru'}
    maxWidth="md"
    onClose={() => (isProductModalOpen = false)}
  >
    <div class="space-y-4 py-2">
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
        <div class="md:col-span-2 space-y-1">
          <label class="block text-xs font-bold text-slate-700 uppercase" for="prod-name">Nama Menu *</label>
          <input
            id="prod-name"
            type="text"
            bind:value={formProductName}
            placeholder="misal: Kopi Susu Gula Aren"
            required
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
          />
        </div>

        <div class="space-y-1">
          <label class="block text-xs font-bold text-slate-700 uppercase" for="prod-sku">Kode SKU / Barcode</label>
          <input
            id="prod-sku"
            type="text"
            bind:value={formProductSKU}
            placeholder="KPS-001"
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold font-mono text-slate-800 outline-none transition"
          />
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="space-y-1">
          <label class="block text-xs font-bold text-slate-700 uppercase" for="prod-cat">Kategori Menu *</label>
          <select
            id="prod-cat"
            bind:value={formProductCategoryId}
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
          >
            {#each categories as cat}
              <option value={cat.id}>{cat.name}</option>
            {/each}
          </select>
        </div>

        <div class="space-y-1">
          <label class="block text-xs font-bold text-slate-700 uppercase" for="prod-price">Harga Jual (Rp) *</label>
          <CurrencyInput bind:value={formProductPrice} placeholder="Nominal Rp" />
        </div>
      </div>

      <div class="space-y-1">
        <label class="block text-xs font-bold text-slate-700 uppercase" for="prod-cogs">Modal HPP (COGS)</label>
        <CurrencyInput bind:value={formProductCOGS} placeholder="Estimasi Modal HPP" />
      </div>

      <div class="pt-2">
        <label class="flex items-center gap-3 p-3 bg-slate-50 border border-slate-200 rounded-xl cursor-pointer">
          <input
            type="checkbox"
            bind:checked={formProductIsAvailable}
            class="w-5 h-5 text-brand-primary rounded accent-brand-primary cursor-pointer"
          />
          <span class="text-xs font-bold text-slate-700">Status Menu Tersedia (Siap Dijual)</span>
        </label>
      </div>
    </div>

    <svelte:fragment slot="footer">
      <button
        type="button"
        on:click={() => (isProductModalOpen = false)}
        class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
      >
        Batal
      </button>

      <button
        type="button"
        disabled={isSavingProduct}
        on:click={handleSaveProduct}
        class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md transition"
      >
        {isSavingProduct ? 'Menyimpan...' : isEditingProduct ? 'Simpan Perubahan' : 'Tambah Menu'}
      </button>
    </svelte:fragment>
  </Modal>

  <!-- Create/Edit Category Modal -->
  <Modal
    isOpen={isCategoryModalOpen}
    title={isEditingCategory ? 'Edit Kategori Menu' : 'Tambah Kategori Menu Baru'}
    maxWidth="sm"
    onClose={() => (isCategoryModalOpen = false)}
  >
    <div class="space-y-4 py-2">
      <div class="space-y-1">
        <label class="block text-xs font-bold text-slate-700 uppercase" for="cat-name">Nama Kategori *</label>
        <input
          id="cat-name"
          type="text"
          bind:value={formCategoryName}
          placeholder="misal: Snack & Dessert"
          required
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        />
      </div>

      <div class="space-y-1">
        <label class="block text-xs font-bold text-slate-700 uppercase" for="cat-order">Urutan Tampilan</label>
        <input
          id="cat-order"
          type="number"
          bind:value={formCategorySortOrder}
          placeholder="1"
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        />
      </div>
    </div>

    <svelte:fragment slot="footer">
      <button
        type="button"
        on:click={() => (isCategoryModalOpen = false)}
        class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
      >
        Batal
      </button>

      <button
        type="button"
        disabled={isSavingCategory}
        on:click={handleSaveCategory}
        class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md transition"
      >
        {isSavingCategory ? 'Menyimpan...' : isEditingCategory ? 'Simpan Perubahan' : 'Buat Kategori'}
      </button>
    </svelte:fragment>
  </Modal>

  <ConfirmModal
    isOpen={isDeleteProductModalOpen}
    title="Hapus Menu Permanen"
    message={`Apakah Anda yakin ingin menghapus menu "${productToDelete?.name || ''}" secara permanen? Operasi ini tidak dapat dibatalkan.`}
    confirmText="Ya, Hapus Permanen"
    cancelText="Batal"
    isDanger={true}
    onConfirm={handleConfirmDeleteProduct}
    onClose={() => (isDeleteProductModalOpen = false)}
  />

  <ConfirmModal
    isOpen={isDeleteCategoryModalOpen}
    title="Hapus Kategori Menu"
    message={`Apakah Anda yakin ingin menghapus kategori "${categoryToDelete?.name || ''}"? Tindakan ini tidak dapat dibatalkan.`}
    confirmText="Ya, Hapus Kategori"
    cancelText="Batal"
    isDanger={true}
    isLoading={isDeletingCategory}
    onConfirm={handleConfirmDeleteCategory}
    onClose={() => {
      isDeleteCategoryModalOpen = false;
      categoryToDelete = null;
    }}
  />

  <RecipeMapperModal
    isOpen={isBOMModalOpen}
    product={selectedProductForBOM}
    onClose={() => (isBOMModalOpen = false)}
  />
</div>

