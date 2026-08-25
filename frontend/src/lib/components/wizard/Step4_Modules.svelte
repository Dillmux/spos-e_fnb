<script lang="ts">
  import { ToggleLeft, ToggleRight, LayoutGrid, MonitorCheck, PackageCheck, QrCode, Tv } from 'lucide-svelte';
  import type { SetupPayload } from '$lib/types';

  export let formData: SetupPayload;
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-slate-800 flex items-center gap-3">
      <ToggleRight class="w-7 h-7 text-brand-primary" />
      Langkah 4: Modul Fitur Akselerasi
    </h2>
    <p class="text-slate-500 text-sm mt-1">Aktifkan modul yang dibutuhkan. Antarmuka POS akan menyesuaikan secara dinamis.</p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <!-- Enable Tables -->
    <div
      class="p-5 rounded-2xl border-2 flex items-center justify-between transition
      {formData.active_modules.enable_tables ? 'border-brand-primary bg-brand-primary/5' : 'border-slate-200 bg-white'}"
    >
      <div class="flex items-center gap-4">
        <div class="p-3 bg-blue-100 text-blue-700 rounded-xl">
          <LayoutGrid class="w-6 h-6" />
        </div>
        <div>
          <h4 class="font-bold text-base text-slate-800">Visual Denah Meja</h4>
          <p class="text-xs text-slate-500">Status ketersediaan meja terisi/kosong real-time.</p>
        </div>
      </div>
      <button
        type="button"
        on:click={() => (formData.active_modules.enable_tables = !formData.active_modules.enable_tables)}
        class="touch-target p-2 text-brand-primary hover:opacity-80 transition"
      >
        {#if formData.active_modules.enable_tables}
          <ToggleRight class="w-10 h-10 text-brand-primary" />
        {:else}
          <ToggleLeft class="w-10 h-10 text-slate-400" />
        {/if}
      </button>
    </div>

    <!-- Enable KDS -->
    <div
      class="p-5 rounded-2xl border-2 flex items-center justify-between transition
      {formData.active_modules.enable_kds ? 'border-brand-primary bg-brand-primary/5' : 'border-slate-200 bg-white'}"
    >
      <div class="flex items-center gap-4">
        <div class="p-3 bg-amber-100 text-amber-700 rounded-xl">
          <MonitorCheck class="w-6 h-6" />
        </div>
        <div>
          <h4 class="font-bold text-base text-slate-800">Kitchen Display System (KDS)</h4>
          <p class="text-xs text-slate-500">Layar antrean tiket koki dapur tanpa kertas.</p>
        </div>
      </div>
      <button
        type="button"
        on:click={() => (formData.active_modules.enable_kds = !formData.active_modules.enable_kds)}
        class="touch-target p-2 text-brand-primary hover:opacity-80 transition"
      >
        {#if formData.active_modules.enable_kds}
          <ToggleRight class="w-10 h-10 text-brand-primary" />
        {:else}
          <ToggleLeft class="w-10 h-10 text-slate-400" />
        {/if}
      </button>
    </div>

    <!-- Enable WLD (Waiting Line Display) -->
    <div
      class="p-5 rounded-2xl border-2 flex items-center justify-between transition
      {formData.active_modules.enable_wld ? 'border-brand-primary bg-brand-primary/5' : 'border-slate-200 bg-white'}"
    >
      <div class="flex items-center gap-4">
        <div class="p-3 bg-sky-100 text-sky-700 rounded-xl">
          <Tv class="w-6 h-6" />
        </div>
        <div>
          <h4 class="font-bold text-base text-slate-800">Waiting Line Display (WLD)</h4>
          <p class="text-xs text-slate-500">Layar TV panggilan nomor antrean pesanan pelanggan.</p>
        </div>
      </div>
      <button
        type="button"
        on:click={() => (formData.active_modules.enable_wld = !formData.active_modules.enable_wld)}
        class="touch-target p-2 text-brand-primary hover:opacity-80 transition"
      >
        {#if formData.active_modules.enable_wld}
          <ToggleRight class="w-10 h-10 text-brand-primary" />
        {:else}
          <ToggleLeft class="w-10 h-10 text-slate-400" />
        {/if}
      </button>
    </div>

    <!-- Enable BOM -->
    <div
      class="p-5 rounded-2xl border-2 flex items-center justify-between transition
      {formData.active_modules.enable_bom ? 'border-brand-primary bg-brand-primary/5' : 'border-slate-200 bg-white'}"
    >
      <div class="flex items-center gap-4">
        <div class="p-3 bg-emerald-100 text-emerald-700 rounded-xl">
          <PackageCheck class="w-6 h-6" />
        </div>
        <div>
          <h4 class="font-bold text-base text-slate-800">BOM & Otomatisasi Stok Bahan</h4>
          <p class="text-xs text-slate-500">Pemotongan gramasi stok bahan baku per porsi menu.</p>
        </div>
      </div>
      <button
        type="button"
        on:click={() => (formData.active_modules.enable_bom = !formData.active_modules.enable_bom)}
        class="touch-target p-2 text-brand-primary hover:opacity-80 transition"
      >
        {#if formData.active_modules.enable_bom}
          <ToggleRight class="w-10 h-10 text-brand-primary" />
        {:else}
          <ToggleLeft class="w-10 h-10 text-slate-400" />
        {/if}
      </button>
    </div>

    <!-- Enable Midtrans -->
    <div
      class="p-5 rounded-2xl border-2 flex items-center justify-between transition
      {formData.active_modules.enable_midtrans ? 'border-brand-primary bg-brand-primary/5' : 'border-slate-200 bg-white'}"
    >
      <div class="flex items-center gap-4">
        <div class="p-3 bg-purple-100 text-purple-700 rounded-xl">
          <QrCode class="w-6 h-6" />
        </div>
        <div>
          <h4 class="font-bold text-base text-slate-800">Midtrans Dynamic QRIS</h4>
          <p class="text-xs text-slate-500">Pembayaran QRIS otomatis langsung settlement.</p>
        </div>
      </div>
      <button
        type="button"
        on:click={() => (formData.active_modules.enable_midtrans = !formData.active_modules.enable_midtrans)}
        class="touch-target p-2 text-brand-primary hover:opacity-80 transition"
      >
        {#if formData.active_modules.enable_midtrans}
          <ToggleRight class="w-10 h-10 text-brand-primary" />
        {:else}
          <ToggleLeft class="w-10 h-10 text-slate-400" />
        {/if}
      </button>
    </div>
  </div>
</div>

