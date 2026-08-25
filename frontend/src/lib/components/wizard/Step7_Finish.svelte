<script lang="ts">
  import { CheckCircle2, Rocket, Store, Database, ShieldCheck, ToggleRight } from 'lucide-svelte';
  import type { SetupPayload } from '$lib/types';

  export let formData: SetupPayload;
  export let onInitialize: () => void;
  export let isSubmitting: boolean = false;
</script>

<div class="space-y-6">
  <div class="text-center max-w-xl mx-auto space-y-2">
    <div class="inline-flex p-4 bg-emerald-100 text-emerald-600 rounded-full mb-2">
      <CheckCircle2 class="w-12 h-12" />
    </div>
    <h2 class="text-3xl font-bold text-slate-800">Ringkasan Inisialisasi System</h2>
    <p class="text-slate-500 text-sm">Semua parameter siap diinstalasi. Periksa kembali konfigurasi Anda sebelum mengunci inisialisasi.</p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl mx-auto">
    <!-- Card Database & Outlet -->
    <div class="p-5 bg-slate-50 rounded-2xl border border-slate-200 space-y-3">
      <div class="flex items-center gap-2 font-bold text-slate-800 border-b border-slate-200 pb-2">
        <Store class="w-5 h-5 text-brand-primary" />
        Profil Business & DB
      </div>
      <div class="space-y-1 text-sm">
        <p><span class="text-slate-500">Nama Resto:</span> <strong class="text-slate-800">{formData.outlet_name}</strong></p>
        <p><span class="text-slate-500">Tipe Usaha:</span> <strong class="text-slate-800">{formData.business_type.toUpperCase()}</strong></p>
        <p><span class="text-slate-500">Basis Data:</span> <strong class="text-slate-800">{formData.db_driver.toUpperCase()}</strong></p>
        <p><span class="text-slate-500">Telepon:</span> <strong class="text-slate-800">{formData.outlet_phone}</strong></p>
      </div>
    </div>

    <!-- Card Admin & Financial -->
    <div class="p-5 bg-slate-50 rounded-2xl border border-slate-200 space-y-3">
      <div class="flex items-center gap-2 font-bold text-slate-800 border-b border-slate-200 pb-2">
        <ShieldCheck class="w-5 h-5 text-brand-primary" />
        Admin & Kualifikasi Keuangan
      </div>
      <div class="space-y-1 text-sm">
        <p><span class="text-slate-500">Superadmin:</span> <strong class="text-slate-800">{formData.admin_username}</strong></p>
        <p><span class="text-slate-500">Supervisor PIN:</span> <strong class="text-slate-800">•••••• (6-digit)</strong></p>
        <p><span class="text-slate-500">Pajak Resto:</span> <strong class="text-slate-800">{formData.tax_percentage}% ({formData.is_tax_inclusive ? 'Nett' : 'Excl'})</strong></p>
        <p><span class="text-slate-500">Service Fee:</span> <strong class="text-slate-800">{formData.service_percentage}%</strong></p>
      </div>
    </div>

    <!-- Active Modules -->
    <div class="md:col-span-2 p-5 bg-slate-50 rounded-2xl border border-slate-200 space-y-3">
      <div class="flex items-center gap-2 font-bold text-slate-800 border-b border-slate-200 pb-2">
        <ToggleRight class="w-5 h-5 text-brand-primary" />
        Modul Aktif
      </div>
      <div class="flex flex-wrap gap-2">
        {#if formData.active_modules.enable_tables}
          <span class="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-xs font-bold">✓ Denah Meja</span>
        {/if}
        {#if formData.active_modules.enable_kds}
          <span class="px-3 py-1 bg-amber-100 text-amber-800 rounded-full text-xs font-bold">✓ Kitchen Display (KDS)</span>
        {/if}
        {#if formData.active_modules.enable_wld}
          <span class="px-3 py-1 bg-sky-100 text-sky-800 rounded-full text-xs font-bold">✓ Waiting Line (WLD)</span>
        {/if}
        {#if formData.active_modules.enable_bom}
          <span class="px-3 py-1 bg-emerald-100 text-emerald-800 rounded-full text-xs font-bold">✓ Otomatisasi Stok BOM</span>
        {/if}
        {#if formData.active_modules.enable_midtrans}
          <span class="px-3 py-1 bg-purple-100 text-purple-800 rounded-full text-xs font-bold">✓ Midtrans QRIS</span>
        {/if}
      </div>
    </div>
  </div>

  <div class="pt-4 text-center">
    <button
      type="button"
      on:click={onInitialize}
      disabled={isSubmitting}
      class="px-10 py-4 min-h-[48px] font-extrabold text-white text-lg bg-brand-primary hover:bg-sky-500 active:scale-95 disabled:opacity-50 rounded-2xl shadow-xl shadow-sky-500/20 inline-flex items-center gap-3 transition"
    >
      <Rocket class="w-6 h-6 animate-pulse" />
      {isSubmitting ? 'Menginisialisasi Sistem...' : 'Jalankan Inisialisasi Smart POS'}
    </button>
  </div>
</div>
