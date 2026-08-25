<script lang="ts">
  import { onMount } from "svelte";
  import { authApi } from "$lib/api/auth";
  import { toast } from "$lib/stores/toast";
  import Modal from "$lib/components/common/Modal.svelte";
  import ConfirmModal from "$lib/components/common/ConfirmModal.svelte";
  import type { User } from "$lib/types";
  import {
    Users,
    UserPlus,
    Search,
    Edit2,
    ShieldCheck,
    Key,
    Lock,
    UserX,
    CheckCircle,
    ShieldAlert,
    Trash2,
  } from "lucide-svelte";

  interface StaffUser extends User {
    has_pin?: boolean;
  }

  let users: StaffUser[] = [];
  let isLoading = true;
  let searchQuery = "";
  let selectedRole = "all";

  // Create/Edit Modal State
  let isModalOpen = false;
  let isEditing = false;
  let editingUserId = "";
  let isSaving = false;

  // Toggle Confirm Modal State
  let isToggleModalOpen = false;
  let userToToggle: StaffUser | null = null;
  let isToggling = false;

  // Delete Confirm Modal State
  let isDeleteModalOpen = false;
  let userToDelete: StaffUser | null = null;
  let isDeleting = false;

  // Form Fields
  let formName = "";
  let formUsername = "";
  let formEmail = "";
  let formRole: "superadmin" | "manager" | "cashier" | "kitchen" = "cashier";
  let formPassword = "";
  let formPin = "";
  let formIsActive = true;

  async function loadUsers() {
    isLoading = true;
    try {
      const res = await authApi.getUsers();
      users = res.data || [];
    } catch (err: any) {
      users = [];
      toast.error(err.message || "Gagal memuat data staf & pengguna");
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    loadUsers();
  });

  function openCreateModal() {
    isEditing = false;
    editingUserId = "";
    formName = "";
    formUsername = "";
    formEmail = "";
    formRole = "cashier";
    formPassword = "";
    formPin = "";
    formIsActive = true;
    isModalOpen = true;
  }

  function openEditModal(user: StaffUser) {
    isEditing = true;
    editingUserId = user.id;
    formName = user.name;
    formUsername = user.username;
    formEmail = user.email || "";
    formRole = user.role;
    formPassword = "";
    formPin = "";
    formIsActive = user.is_active;
    isModalOpen = true;
  }

  function openToggleModal(user: StaffUser) {
    userToToggle = user;
    isToggleModalOpen = true;
  }

  async function handleConfirmToggleStatus() {
    if (!userToToggle) return;
    isToggling = true;
    try {
      await authApi.toggleUserStatus(userToToggle.id, !userToToggle.is_active);
      toast.success(`Status akun ${userToToggle.name} berhasil diubah.`);
      isToggleModalOpen = false;
      loadUsers();
    } catch (err: any) {
      toast.error(err.message || "Gagal mengubah status akun");
    } finally {
      isToggling = false;
    }
  }

  function openDeleteModal(user: StaffUser) {
    userToDelete = user;
    isDeleteModalOpen = true;
  }

  async function handleConfirmDeleteUser() {
    if (!userToDelete) return;
    const userName = userToDelete.name;
    isDeleting = true;
    try {
      await authApi.deleteUser(userToDelete.id);
      toast.success(
        `Akun pengguna "${userName}" berhasil dihapus secara permanen.`,
      );
      isDeleteModalOpen = false;
      userToDelete = null;
      loadUsers();
    } catch (err: any) {
      toast.error(err.message || "Gagal menghapus pengguna");
    } finally {
      isDeleting = false;
    }
  }

  async function handleSaveUser() {
    if (!formName || !formUsername) {
      toast.warning("Nama lengkap dan username wajib diisi.");
      return;
    }

    if (!isEditing && !formPassword) {
      toast.warning("Password wajib diisi untuk pengguna baru.");
      return;
    }

    if (
      (formRole === "superadmin" || formRole === "manager") &&
      formPin &&
      formPin.length !== 6
    ) {
      toast.warning("PIN Supervisor harus 6-digit angka.");
      return;
    }

    isSaving = true;
    try {
      const payload = {
        name: formName,
        username: formUsername,
        email: formEmail || undefined,
        role: formRole,
        password: formPassword || undefined,
        auth_pin: formPin || undefined,
        is_active: formIsActive,
      };

      if (isEditing) {
        await authApi.updateUser(editingUserId, payload);
        toast.success("Akun pengguna berhasil diperbarui!");
      } else {
        await authApi.createUser(payload);
        toast.success("Pengguna baru berhasil ditambahkan!");
      }

      isModalOpen = false;
      loadUsers();
    } catch (err: any) {
      toast.error(err.message || "Gagal menyimpan data pengguna");
    } finally {
      isSaving = false;
    }
  }

  function handlePinInput(e: Event) {
    const input = e.target as HTMLInputElement;
    formPin = input.value.replace(/\D/g, "").slice(0, 6);
  }

  $: filteredUsers = users.filter((u) => {
    const matchesSearch =
      u.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      u.username.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesRole = selectedRole === "all" || u.role === selectedRole;
    return matchesSearch && matchesRole;
  });

  function getRoleBadge(role: string) {
    switch (role) {
      case "superadmin":
        return {
          label: "Super Admin",
          bg: "bg-purple-100 text-purple-800 border-purple-200",
        };
      case "manager":
        return {
          label: "Manager / Supervisor",
          bg: "bg-blue-100 text-blue-800 border-blue-200",
        };
      case "kitchen":
        return {
          label: "Staf Dapur / KDS",
          bg: "bg-amber-100 text-amber-800 border-amber-200",
        };
      default:
        return {
          label: "Kasir",
          bg: "bg-emerald-100 text-emerald-800 border-emerald-200",
        };
    }
  }
</script>

<div
  class="h-full bg-brand-surface p-6 flex flex-col space-y-6 overflow-hidden"
>
  <!-- Header Banner -->
  <div class="flex items-center justify-between">
    <div>
      <h2
        class="text-2xl font-extrabold text-slate-900 flex items-center gap-3"
      >
        <Users class="w-7 h-7 text-brand-primary" />
        Manajemen Pengguna & Hak Akses Staf
      </h2>
      <p class="text-xs text-slate-500 font-semibold mt-1">
        Kelola akun login kasir, staf dapur, supervisor, serta 6-digit PIN
        otorisasi transaksi.
      </p>
    </div>

    <button
      type="button"
      on:click={openCreateModal}
      class="px-5 py-2.5 bg-brand-primary hover:bg-sky-500 text-white text-xs font-bold rounded-xl shadow-md flex items-center gap-2 transition"
    >
      <UserPlus class="w-4 h-4" />
      Tambah Pengguna Baru
    </button>
  </div>

  <!-- Filter & Search Bar -->
  <div
    class="flex flex-col sm:flex-row items-center justify-between gap-3 bg-white p-4 rounded-2xl border border-slate-200 shadow-sm"
  >
    <!-- Role Tabs Filter -->
    <div class="flex items-center gap-1.5 overflow-x-auto w-full sm:w-auto">
      {#each [{ id: "all", label: "Semua Staf" }, { id: "superadmin", label: "Superadmin" }, { id: "manager", label: "Manager" }, { id: "cashier", label: "Kasir" }, { id: "kitchen", label: "Dapur" }] as tab}
        <button
          type="button"
          on:click={() => (selectedRole = tab.id)}
          class="px-3.5 py-1.5 rounded-xl text-xs font-bold transition whitespace-nowrap
          {selectedRole === tab.id
            ? 'bg-slate-900 text-white shadow-sm'
            : 'bg-slate-100 text-slate-600 hover:bg-slate-200'}"
        >
          {tab.label}
        </button>
      {/each}
    </div>

    <!-- Search Bar -->
    <div class="relative w-full sm:w-64">
      <Search class="w-4 h-4 absolute left-3 top-2.5 text-slate-400" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Cari nama atau username..."
        class="w-full pl-9 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold outline-none focus:border-brand-primary focus:bg-white transition"
      />
    </div>
  </div>

  <!-- Users Table Container -->
  <div
    class="flex-1 bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex flex-col"
  >
    {#if isLoading}
      <div
        class="flex-1 flex items-center justify-center text-slate-400 font-semibold"
      >
        Memuat Daftar Pengguna...
      </div>
    {:else if filteredUsers.length === 0}
      <div
        class="flex-1 flex flex-col items-center justify-center text-slate-400 p-8 text-center space-y-2"
      >
        <Users class="w-12 h-12 text-slate-300" />
        <p class="font-bold text-slate-600">Tidak ada pengguna ditemukan</p>
        <p class="text-xs text-slate-400">
          Coba ubah kata kunci pencarian atau filter peranan staf.
        </p>
      </div>
    {:else}
      <div class="flex-1 overflow-y-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr
              class="bg-slate-50 border-b border-slate-200 text-xs font-bold text-slate-500 uppercase tracking-wider"
            >
              <th class="py-3.5 px-6">Nama Pengguna</th>
              <th class="py-3.5 px-6">Username Login</th>
              <th class="py-3.5 px-6">Peran & Akses</th>
              <th class="py-3.5 px-6">PIN Supervisor</th>
              <th class="py-3.5 px-6">Status Akun</th>
              <th class="py-3.5 px-6 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody
            class="divide-y divide-slate-100 text-sm font-semibold text-slate-800"
          >
            {#each filteredUsers as user}
              {@const roleBadge = getRoleBadge(user.role)}
              <tr class="hover:bg-slate-50/80 transition">
                <td class="py-4 px-6">
                  <div class="flex items-center gap-3">
                    <div
                      class="w-9 h-9 rounded-full bg-slate-900 text-white font-black text-xs flex items-center justify-center shadow-sm"
                    >
                      {user.name ? user.name[0].toUpperCase() : "U"}
                    </div>
                    <div>
                      <p class="font-bold text-slate-900">{user.name}</p>
                      <p class="text-xs text-slate-400 font-normal">
                        {user.email || "Tanpa email"}
                      </p>
                    </div>
                  </div>
                </td>

                <td
                  class="py-4 px-6 font-mono text-xs font-bold text-slate-700"
                >
                  @{user.username}
                </td>

                <td class="py-4 px-6">
                  <span
                    class="text-xs font-bold px-3 py-1 rounded-full border {roleBadge.bg}"
                  >
                    {roleBadge.label}
                  </span>
                </td>

                <td class="py-4 px-6">
                  {#if user.role === "superadmin" || user.role === "manager"}
                    {#if user.has_pin}
                      <span
                        class="inline-flex items-center gap-1.5 text-xs font-bold text-emerald-700 bg-emerald-50 px-2.5 py-1 rounded-lg border border-emerald-200"
                      >
                        <ShieldCheck class="w-3.5 h-3.5" /> PIN Aktif
                      </span>
                    {:else}
                      <span
                        class="inline-flex items-center gap-1.5 text-xs font-bold text-amber-700 bg-amber-50 px-2.5 py-1 rounded-lg border border-amber-200"
                      >
                        <ShieldAlert class="w-3.5 h-3.5" /> Belum Set PIN
                      </span>
                    {/if}
                  {:else}
                    <span class="text-xs text-slate-400 font-normal">-</span>
                  {/if}
                </td>

                <td class="py-4 px-6">
                  <button
                    type="button"
                    on:click={() => openToggleModal(user)}
                    class="text-xs font-bold px-3 py-1 rounded-full transition flex items-center gap-1.5
                    {user.is_active
                      ? 'bg-emerald-100 text-emerald-800 hover:bg-emerald-200'
                      : 'bg-red-100 text-red-800 hover:bg-red-200'}"
                  >
                    {#if user.is_active}
                      <CheckCircle class="w-3.5 h-3.5" /> Aktif
                    {:else}
                      <UserX class="w-3.5 h-3.5" /> Nonaktif
                    {/if}
                  </button>
                </td>

                <td class="py-4 px-6 text-right space-x-1">
                  <button
                    type="button"
                    on:click={() => openEditModal(user)}
                    class="p-2 text-slate-400 hover:text-brand-primary transition rounded-lg hover:bg-slate-100"
                    title="Edit Pengguna"
                  >
                    <Edit2 class="w-4 h-4" />
                  </button>
                  <button
                    type="button"
                    on:click={() => openDeleteModal(user)}
                    class="p-2 text-slate-400 hover:text-red-600 transition rounded-lg hover:bg-red-50"
                    title="Hapus Pengguna"
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

  <!-- Create/Edit User Modal -->
  <Modal
    isOpen={isModalOpen}
    title={isEditing ? "Edit Akun Pengguna" : "Tambah Pengguna Baru"}
    maxWidth="md"
    onClose={() => (isModalOpen = false)}
  >
    <div class="space-y-4 py-2">
      <!-- Full Name & Email -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="form-name">Nama Lengkap *</label
          >
          <input
            id="form-name"
            type="text"
            bind:value={formName}
            placeholder="misal: Budi Santoso"
            required
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary focus:bg-white rounded-xl text-sm font-bold outline-none transition"
          />
        </div>

        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="form-email">Email Pengguna</label
          >
          <input
            id="form-email"
            type="email"
            bind:value={formEmail}
            placeholder="budi@spos.local"
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary focus:bg-white rounded-xl text-sm font-bold outline-none transition"
          />
        </div>
      </div>

      <!-- Username & Role -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="form-username">Username Login *</label
          >
          <input
            id="form-username"
            type="text"
            bind:value={formUsername}
            placeholder="kasir1"
            required
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary focus:bg-white rounded-xl text-sm font-bold outline-none transition"
          />
        </div>

        <div class="space-y-1">
          <label
            class="block text-xs font-bold text-slate-700 uppercase"
            for="form-role">Peran / Hak Akses *</label
          >
          <select
            id="form-role"
            bind:value={formRole}
            class="w-full px-4 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary focus:bg-white rounded-xl text-sm font-bold outline-none transition"
          >
            <option value="cashier">Kasir</option>
            <option value="manager">Manager / Supervisor</option>
            <option value="kitchen">Staf Dapur / KDS</option>
            <option value="superadmin">Super Admin</option>
          </select>
        </div>
      </div>

      <!-- Password -->
      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-700 uppercase"
          for="form-pass"
        >
          {isEditing
            ? "Password Baru (Kosongkan jika tidak diubah)"
            : "Password Login *"}
        </label>
        <div class="relative">
          <input
            id="form-pass"
            type="password"
            bind:value={formPassword}
            placeholder="••••••••"
            class="w-full pl-4 pr-10 py-2.5 bg-slate-50 border border-slate-300 focus:border-brand-primary focus:bg-white rounded-xl text-sm font-bold outline-none transition"
          />
          <Lock class="w-4 h-4 absolute right-3 top-3 text-slate-400" />
        </div>
      </div>

      <!-- Status Keaktifan Akun -->
      <div class="pt-1">
        <label class="flex items-center gap-3 p-3 bg-slate-50 border border-slate-200 rounded-xl cursor-pointer">
          <input
            type="checkbox"
            bind:checked={formIsActive}
            class="w-5 h-5 text-brand-primary rounded accent-brand-primary cursor-pointer"
          />
          <span class="text-xs font-bold text-slate-700">Status Akun Aktif (Dapat Digunakan Login)</span>
        </label>
      </div>

      <!-- Supervisor 6-digit PIN (Only shown for superadmin or manager) -->
      {#if formRole === "superadmin" || formRole === "manager"}
        <div
          class="space-y-1 bg-amber-50 p-4 rounded-xl border border-amber-200"
        >
          <label
            class="block text-xs font-bold text-amber-900 uppercase"
            for="form-pin"
          >
            PIN Supervisor (6-Digit Angka Otorisasi)
          </label>
          <div class="relative">
            <input
              id="form-pin"
              type="password"
              inputmode="numeric"
              value={formPin}
              on:input={handlePinInput}
              placeholder="123456"
              maxlength="6"
              class="w-full pl-4 pr-10 py-2.5 bg-white border border-amber-300 focus:border-amber-500 rounded-xl font-mono text-base font-bold text-amber-900 tracking-widest outline-none transition"
            />
            <Key class="w-4 h-4 absolute right-3 top-3.5 text-amber-600" />
          </div>
          <p class="text-[11px] text-amber-700 font-semibold">
            PIN ini akan diminta secara otomatis saat kasir melakukan transaksi
            khusus (Void item, Split Bill).
          </p>
        </div>
      {/if}
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
        on:click={handleSaveUser}
        class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md transition"
      >
        {isSaving
          ? "Menyimpan..."
          : isEditing
            ? "Simpan Perubahan"
            : "Buat Pengguna"}
      </button>
    </svelte:fragment>
  </Modal>

  <ConfirmModal
    isOpen={isToggleModalOpen}
    title="Ubah Status Akun Pengguna"
    message={`Apakah Anda yakin ingin ${userToToggle?.is_active ? "menonaktifkan" : "mengaktifkan"} akun staf "${userToToggle?.name || ""}"?`}
    confirmText={userToToggle?.is_active
      ? "Ya, Nonaktifkan Akun"
      : "Ya, Aktifkan Akun"}
    cancelText="Batal"
    isDanger={userToToggle?.is_active ?? true}
    isLoading={isToggling}
    onConfirm={handleConfirmToggleStatus}
    onClose={() => (isToggleModalOpen = false)}
  />

  <ConfirmModal
    isOpen={isDeleteModalOpen}
    title="Hapus Akun Pengguna"
    message={`Apakah Anda yakin ingin menghapus akun staf "${userToDelete?.name || ""}" secara permanen? Tindakan ini tidak dapat dibatalkan.`}
    confirmText="Ya, Hapus Pengguna"
    cancelText="Batal"
    isDanger={true}
    isLoading={isDeleting}
    onConfirm={handleConfirmDeleteUser}
    onClose={() => {
      isDeleteModalOpen = false;
      userToDelete = null;
    }}
  />
</div>
