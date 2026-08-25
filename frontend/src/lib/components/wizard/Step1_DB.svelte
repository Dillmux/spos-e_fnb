<script lang="ts">
  import { setupApi } from '$lib/api/setup';
  import { toast } from '$lib/stores/toast';
  import { Database, CheckCircle, RefreshCw } from 'lucide-svelte';
  import type { SetupPayload } from '$lib/types';

  export let formData: SetupPayload;

  let isTesting = false;
  let testSuccess = false;
  let testMessage = '';

  async function testConnection() {
    isTesting = true;
    testSuccess = false;
    testMessage = '';
    try {
      const res = await setupApi.testDatabase(formData.db_driver, formData.db_connection_string);
      if (res.data?.connected || (res.data as any)?.status === 'connected') {
        testSuccess = true;
        testMessage = `Koneksi Database Berhasil! (${res.data?.latency_ms || 12}ms)`;
        toast.success(testMessage);
      } else {
        testMessage = 'Gagal terhubung ke database. Periksa parameter string koneksi.';
        toast.error(testMessage);
      }
    } catch (err: any) {
      testMessage = err.message || 'Error saat menguji database';
    } finally {
      isTesting = false;
    }
  }
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-slate-800 flex items-center gap-3">
      <Database class="w-7 h-7 text-brand-primary" />
      Langkah 1: Konfigurasi Basis Data
    </h2>
    <p class="text-slate-500 text-sm mt-1">Pilih mesin database yang akan digunakan untuk menyimpan data POS & transaksi.</p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <button
      type="button"
      on:click={() => { formData.db_driver = 'sqlite'; testSuccess = false; }}
      class="p-5 rounded-2xl border-2 text-left transition flex flex-col justify-between h-36
      {formData.db_driver === 'sqlite' ? 'border-brand-primary bg-brand-primary/5 ring-2 ring-brand-cyan/50' : 'border-slate-200 bg-white hover:border-slate-300'}"
    >
      <div>
        <h4 class="font-bold text-lg text-slate-800">SQLite 3 (Local File)</h4>
        <p class="text-xs text-slate-500 mt-1">Sangat cepat, cocok untuk offline single-terminal resto.</p>
      </div>
      <span class="text-xs font-semibold px-2.5 py-1 bg-emerald-100 text-emerald-800 rounded-full w-max">Rekomendasi Local</span>
    </button>

    <button
      type="button"
      on:click={() => { formData.db_driver = 'postgres'; testSuccess = false; }}
      class="p-5 rounded-2xl border-2 text-left transition flex flex-col justify-between h-36
      {formData.db_driver === 'postgres' ? 'border-brand-primary bg-brand-primary/5 ring-2 ring-brand-cyan/50' : 'border-slate-200 bg-white hover:border-slate-300'}"
    >
      <div>
        <h4 class="font-bold text-lg text-slate-800">PostgreSQL</h4>
        <p class="text-xs text-slate-500 mt-1">Skala besar, multi-outlet cloud sync.</p>
      </div>
      <span class="text-xs font-semibold px-2.5 py-1 bg-blue-100 text-blue-800 rounded-full w-max">Cloud Multi-Branch</span>
    </button>

    <button
      type="button"
      on:click={() => { formData.db_driver = 'mysql'; testSuccess = false; }}
      class="p-5 rounded-2xl border-2 text-left transition flex flex-col justify-between h-36
      {formData.db_driver === 'mysql' ? 'border-brand-primary bg-brand-primary/5 ring-2 ring-brand-cyan/50' : 'border-slate-200 bg-white hover:border-slate-300'}"
    >
      <div>
        <h4 class="font-bold text-lg text-slate-800">MySQL / MariaDB</h4>
        <p class="text-xs text-slate-500 mt-1">Kompatibel dengan VPS & Shared Hosting.</p>
      </div>
      <span class="text-xs font-semibold px-2.5 py-1 bg-amber-100 text-amber-800 rounded-full w-max">Enterprise Standard</span>
    </button>
  </div>

  {#if formData.db_driver !== 'sqlite'}
    <div class="bg-slate-50 p-5 rounded-2xl border border-slate-200 space-y-3">
      <label class="block text-sm font-bold text-slate-700" for="db-conn">
        Database Connection String (DSN)
      </label>
      <input
        id="db-conn"
        type="text"
        bind:value={formData.db_connection_string}
        placeholder="user:password@tcp(localhost:5432)/spos_db?sslmode=disable"
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-mono text-sm shadow-sm outline-none transition"
      />
    </div>
  {/if}

  <div class="pt-2 flex items-center justify-between">
    <button
      type="button"
      on:click={testConnection}
      disabled={isTesting}
      class="px-6 py-3 min-h-[48px] font-bold text-brand-dark bg-brand-cyan hover:bg-cyan-300 active:scale-95 rounded-xl flex items-center gap-2 shadow-sm transition"
    >
      {#if isTesting}
        <RefreshCw class="w-5 h-5 animate-spin" />
        Menguji Koneksi...
      {:else}
        <CheckCircle class="w-5 h-5" />
        Uji Koneksi Database
      {/if}
    </button>

    {#if testMessage}
      <div class="text-sm font-semibold {testSuccess ? 'text-emerald-600' : 'text-brand-danger'}">
        {testMessage}
      </div>
    {/if}
  </div>
</div>
