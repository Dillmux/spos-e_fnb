<script lang="ts">
  import { onMount } from "svelte";
  import { posApi } from "$lib/api/pos";
  import { cart } from "$lib/stores/cart";
  import { userRole } from "$lib/stores/auth";
  import { toast } from "$lib/stores/toast";
  import Modal from "$lib/components/common/Modal.svelte";
  import ConfirmModal from "$lib/components/common/ConfirmModal.svelte";
  import type { Table } from "$lib/types";
  import {
    LayoutGrid,
    Users,
    RefreshCw,
    Plus,
    Pencil,
    Trash2,
    CheckCircle2,
  } from "lucide-svelte";

  export let selectedTableId: string | null = null;
  export let onSelectTable: (table: Table) => void = () => {};

  let tables: Table[] = [];
  let isLoading = true;

  // Add Table Modal State
  let isAddModalOpen = false;
  let isSavingTable = false;
  let formTableNumber = "";
  let formCapacity = 4;

  // Edit Table Modal State
  let isEditModalOpen = false;
  let isSavingEdit = false;
  let editingTableId: string | number | null = null;
  let editTableNumber = "";
  let editCapacity = 4;
  let editStatus: "available" | "occupied" | "reserved" = "available";

  // Delete Table Modal State
  let isDeleteModalOpen = false;
  let isDeleting = false;
  let tableToDelete: Table | null = null;

  async function loadTables() {
    isLoading = true;
    try {
      const res = await posApi.getTables();
      tables = res.data || [];
    } catch (err: any) {
      tables = [];
      toast.error(err.message || 'Gagal memuat data meja');
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    loadTables();
    const handleRefreshTables = () => loadTables();
    window.addEventListener("refresh-tables", handleRefreshTables);
    return () => {
      window.removeEventListener("refresh-tables", handleRefreshTables);
    };
  });

  function handleTableClick(table: Table) {
    const tableName = table.table_number || table.name || `Meja ${table.id}`;
    cart.setTable(table.id.toString(), tableName);
    onSelectTable(table);
  }

  async function handleQuickReleaseTable(table: Table, e: Event) {
    e.stopPropagation();
    const tableName = table.table_number || table.name || `Meja ${table.id}`;
    try {
      await posApi.updateTable(table.id, { status: "available" });
      toast.success(`Meja "${tableName}" berhasil dikosongkan (Tersedia).`);
      loadTables();
    } catch (err: any) {
      toast.error(err.message || 'Gagal mengubah status meja');
    }
  }

  function openAddModal() {
    formTableNumber = "";
    formCapacity = 0;
    isAddModalOpen = true;
  }

  async function handleSaveTable() {
    if (!formTableNumber.trim()) {
      toast.warning("Nomor meja wajib diisi.");
      return;
    }

    isSavingTable = true;
    try {
      await posApi.createTable({
        table_number: formTableNumber.trim(),
        capacity: formCapacity,
      });

      toast.success(`Meja "${formTableNumber.trim()}" berhasil ditambahkan!`);
      formTableNumber = "";
      formCapacity = 0;
      isAddModalOpen = false;
      loadTables();
    } catch (err: any) {
      toast.error(err.message || 'Gagal menambahkan meja baru');
    } finally {
      isSavingTable = false;
    }
  }

  function openEditModal(table: Table, e: Event) {
    e.stopPropagation();
    editingTableId = table.id;
    editTableNumber = table.table_number || table.name || "";
    editCapacity = table.capacity || 2;
    editStatus = table.status || "available";
    isEditModalOpen = true;
  }

  async function handleSaveEditTable() {
    if (!editTableNumber.trim() || !editingTableId) {
      toast.warning("Nomor/nama meja wajib diisi.");
      return;
    }

    if (editCapacity <= 0 || !editCapacity) {
      toast.warning("Jumlah / kapasitas kursi minimal 1.");
      return;
    }

    isSavingEdit = true;
    try {
      await posApi.updateTable(editingTableId, {
        table_number: editTableNumber.trim(),
        capacity: editCapacity,
        status: editStatus,
      });
      toast.success(`Meja "${editTableNumber}" berhasil diperbarui!`);
      isEditModalOpen = false;
      loadTables();
    } catch (err: any) {
      toast.error(err.message || 'Gagal memperbarui meja');
    } finally {
      isSavingEdit = false;
    }
  }

  function openDeleteModal(table: Table, e: Event) {
    e.stopPropagation();
    tableToDelete = table;
    isDeleteModalOpen = true;
  }

  async function handleConfirmDeleteTable() {
    if (!tableToDelete) return;
    const tableName =
      tableToDelete.table_number ||
      tableToDelete.name ||
      `Meja ${tableToDelete.id}`;
    isDeleting = true;
    try {
      await posApi.deleteTable(tableToDelete.id);
      toast.success(`Meja "${tableName}" berhasil dihapus!`);
      isDeleteModalOpen = false;
      tableToDelete = null;
      loadTables();
    } catch (err: any) {
      toast.error(err.message || 'Gagal menghapus meja');
      isDeleteModalOpen = false;
      tableToDelete = null;
    } finally {
      isDeleting = false;
    }
  }
</script>

<div
  class="bg-white p-4 rounded-2xl border border-slate-200 shadow-sm space-y-4"
>
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-2">
      <LayoutGrid class="w-5 h-5 text-brand-primary" />
      <h3 class="font-bold text-slate-800">Daftar Meja</h3>
    </div>
    <div class="flex items-center gap-2">
      {#if $userRole === "superadmin" || $userRole === "manager"}
        <button
          type="button"
          on:click={openAddModal}
          class="px-3 py-1.5 bg-brand-primary hover:bg-sky-500 text-white text-xs font-bold rounded-xl flex items-center gap-1 transition"
        >
          <Plus class="w-3.5 h-3.5" /> Tambah Meja
        </button>
      {/if}
      <button
        type="button"
        on:click={loadTables}
        class="p-2 hover:bg-slate-100 rounded-lg text-slate-500 transition"
        title="Refresh Meja"
      >
        <RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
      </button>
    </div>
  </div>

  {#if isLoading}
    <div
      class="h-36 flex items-center justify-center text-slate-400 text-sm font-semibold"
    >
      Memuat Denah Meja...
    </div>
  {:else}
    <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-3">
      {#each tables as table}
        {@const tableName =
          table.table_number || table.name || `Meja ${table.id}`}
        <button
          type="button"
          on:click={() => handleTableClick(table)}
          class="touch-target p-3.5 rounded-2xl border-2 flex flex-col justify-between h-28 text-left transition-all duration-200 active:scale-95 shadow-sm relative group
          {table.id.toString() === selectedTableId?.toString()
            ? 'ring-4 ring-brand-cyan/60 font-extrabold'
            : ''}
          {table.status === 'available'
            ? 'bg-[#D8FFC5] border-emerald-300 text-emerald-950 hover:bg-[#C4F7CA]'
            : ''}
          {table.status === 'occupied'
            ? 'bg-brand-primary border-sky-400 text-white hover:bg-sky-500'
            : ''}
          {table.status === 'reserved'
            ? 'bg-amber-100 border-amber-300 text-amber-900'
            : ''}"
        >
          <div class="flex items-center justify-between w-full">
            <span class="font-bold text-base">{tableName}</span>
            <div class="flex items-center gap-1 text-xs opacity-90">
              <Users class="w-3.5 h-3.5" />
              {table.capacity}
            </div>
          </div>
          <div class="flex items-center justify-between w-full mt-2">
            <span
              class="text-[10px] uppercase tracking-wider font-extrabold px-2 py-0.5 rounded-full
              {table.status === 'available'
                ? 'bg-emerald-200 text-emerald-900'
                : ''}
              {table.status === 'occupied' ? 'bg-sky-700 text-white' : ''}
              {table.status === 'reserved'
                ? 'bg-amber-200 text-amber-900'
                : ''}"
            >
              {table.status === "available"
                ? "Tersedia"
                : table.status === "occupied"
                  ? "Terisi"
                  : "Reservasi"}
            </span>

            {#if table.status === "occupied"}
              <button
                type="button"
                on:click={(e) => handleQuickReleaseTable(table, e)}
                class="px-2 py-0.5 bg-emerald-500 hover:bg-emerald-600 text-white rounded-md text-[10px] font-bold flex items-center gap-1 shadow-sm transition"
                title="Kosongkan Meja (Set Tersedia)"
              >
                <CheckCircle2 class="w-3 h-3" /> Kosongkan
              </button>
            {/if}

            {#if $userRole === "superadmin" || $userRole === "manager"}
              <div
                class="flex items-center gap-1 opacity-90 group-hover:opacity-100 transition"
              >
                <button
                  type="button"
                  on:click={(e) => openEditModal(table, e)}
                  class="p-1 hover:bg-black/10 rounded-md transition text-slate-700 hover:text-black"
                  title="Edit Meja"
                >
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button
                  type="button"
                  on:click={(e) => openDeleteModal(table, e)}
                  class="p-1 hover:bg-red-500/20 rounded-md transition text-slate-700 hover:text-red-900"
                  title="Hapus Meja"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            {/if}
          </div>
        </button>
      {/each}
    </div>
  {/if}

  <!-- Add Table Modal -->
  <Modal
    isOpen={isAddModalOpen}
    title="Tambah Meja Resto Baru"
    maxWidth="sm"
    onClose={() => (isAddModalOpen = false)}
  >
    <div class="space-y-4 py-2">
      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-700 uppercase"
          for="table-num">Nomor / Nama Meja *</label
        >
        <input
          id="table-num"
          type="text"
          bind:value={formTableNumber}
          placeholder="misal: Meja 05 atau VIP 1"
          required
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        />
      </div>

      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-700 uppercase"
          for="table-cap">Kapasitas Kursi *</label
        >
        <input
          id="table-cap"
          type="number"
          bind:value={formCapacity}
          min="1"
          placeholder="4"
          required
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        />
      </div>
    </div>

    <svelte:fragment slot="footer">
      <button
        type="button"
        on:click={() => (isAddModalOpen = false)}
        class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
      >
        Batal
      </button>

      <button
        type="button"
        disabled={isSavingTable}
        on:click={handleSaveTable}
        class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md transition"
      >
        {isSavingTable ? "Menyimpan..." : "Tambah Meja"}
      </button>
    </svelte:fragment>
  </Modal>

  <!-- Edit Table Modal -->
  <Modal
    isOpen={isEditModalOpen}
    title="Edit Data Meja Resto"
    maxWidth="sm"
    onClose={() => (isEditModalOpen = false)}
  >
    <div class="space-y-4 py-2">
      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-700 uppercase"
          for="edit-table-num">Nomor / Nama Meja *</label
        >
        <input
          id="edit-table-num"
          type="text"
          bind:value={editTableNumber}
          placeholder="misal: Meja 05 atau VIP 1"
          required
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        />
      </div>

      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-700 uppercase"
          for="edit-table-cap">Kapasitas Kursi *</label
        >
        <input
          id="edit-table-cap"
          type="number"
          bind:value={editCapacity}
          min="1"
          placeholder="4"
          required
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        />
      </div>

      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-700 uppercase"
          for="edit-table-status">Status Meja</label
        >
        <select
          id="edit-table-status"
          bind:value={editStatus}
          class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-sm font-bold text-slate-800 outline-none transition"
        >
          <option value="available">Tersedia</option>
          <option value="occupied">Terisi</option>
          <option value="reserved">Reservasi</option>
        </select>
      </div>
    </div>

    <svelte:fragment slot="footer">
      <button
        type="button"
        on:click={() => (isEditModalOpen = false)}
        class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
      >
        Batal
      </button>

      <button
        type="button"
        disabled={isSavingEdit}
        on:click={handleSaveEditTable}
        class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md transition"
      >
        {isSavingEdit ? "Menyimpan..." : "Simpan Perubahan"}
      </button>
    </svelte:fragment>
  </Modal>

  <!-- Delete Confirm Modal -->
  <ConfirmModal
    isOpen={isDeleteModalOpen}
    title="Hapus Meja Resto"
    message={`Apakah Anda yakin ingin menghapus "${tableToDelete?.table_number || tableToDelete?.name || "meja ini"}"? Tindakan ini tidak dapat dibatalkan.`}
    confirmText="Ya, Hapus Meja"
    isDanger={true}
    isLoading={isDeleting}
    onConfirm={handleConfirmDeleteTable}
    onClose={() => {
      isDeleteModalOpen = false;
      tableToDelete = null;
    }}
  />
</div>
