<script lang="ts">
  import { Layers, Check } from 'lucide-svelte';
  import type { SetupPayload } from '$lib/types';

  export let formData: SetupPayload;

  function toggleMode(mode: 'dine_in' | 'take_away' | 'delivery') {
    if (formData.service_modes.includes(mode)) {
      if (formData.service_modes.length > 1) {
        formData.service_modes = formData.service_modes.filter((m) => m !== mode);
      }
    } else {
      formData.service_modes = [...formData.service_modes, mode];
    }
  }
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-slate-800 flex items-center gap-3">
      <Layers class="w-7 h-7 text-brand-primary" />
      Langkah 3: Mode Layanan Resto
    </h2>
    <p class="text-slate-500 text-sm mt-1">Pilih metode pelayanan yang disediakan di outlet Anda (dapat memilih lebih dari satu).</p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <!-- Dine In -->
    <button
      type="button"
      on:click={() => toggleMode('dine_in')}
      class="p-6 rounded-2xl border-2 flex flex-col justify-between h-44 text-left transition
      {formData.service_modes.includes('dine_in') ? 'border-brand-primary bg-brand-primary/5 ring-2 ring-brand-cyan/50' : 'border-slate-200 bg-white hover:border-slate-300'}"
    >
      <div class="flex items-center justify-between">
        <span class="text-3xl">🍽️</span>
        {#if formData.service_modes.includes('dine_in')}
          <div class="w-7 h-7 rounded-full bg-brand-primary text-white flex items-center justify-center">
            <Check class="w-4 h-4" />
          </div>
        {/if}
      </div>
      <div>
        <h4 class="font-bold text-lg text-slate-800">Dine-In (Makan di Tempat)</h4>
        <p class="text-xs text-slate-500 mt-1">Mendukung visual denah meja, pemesanan per nomor meja, dan cetak bil.</p>
      </div>
    </button>

    <!-- Take Away -->
    <button
      type="button"
      on:click={() => toggleMode('take_away')}
      class="p-6 rounded-2xl border-2 flex flex-col justify-between h-44 text-left transition
      {formData.service_modes.includes('take_away') ? 'border-brand-primary bg-brand-primary/5 ring-2 ring-brand-cyan/50' : 'border-slate-200 bg-white hover:border-slate-300'}"
    >
      <div class="flex items-center justify-between">
        <span class="text-3xl">🛍️</span>
        {#if formData.service_modes.includes('take_away')}
          <div class="w-7 h-7 rounded-full bg-brand-primary text-white flex items-center justify-center">
            <Check class="w-4 h-4" />
          </div>
        {/if}
      </div>
      <div>
        <h4 class="font-bold text-lg text-slate-800">Take-Away (Bawa Pulang)</h4>
        <p class="text-xs text-slate-500 mt-1">Pesanan langsung bayar di kasir tanpa pemetaan nomor meja.</p>
      </div>
    </button>

    <!-- Delivery -->
    <button
      type="button"
      on:click={() => toggleMode('delivery')}
      class="p-6 rounded-2xl border-2 flex flex-col justify-between h-44 text-left transition
      {formData.service_modes.includes('delivery') ? 'border-brand-primary bg-brand-primary/5 ring-2 ring-brand-cyan/50' : 'border-slate-200 bg-white hover:border-slate-300'}"
    >
      <div class="flex items-center justify-between">
        <span class="text-3xl">🛵</span>
        {#if formData.service_modes.includes('delivery')}
          <div class="w-7 h-7 rounded-full bg-brand-primary text-white flex items-center justify-center">
            <Check class="w-4 h-4" />
          </div>
        {/if}
      </div>
      <div>
        <h4 class="font-bold text-lg text-slate-800">Pesan Antar / Online</h4>
        <p class="text-xs text-slate-500 mt-1">Pencatatan kurir & pengiriman via driver online.</p>
      </div>
    </button>
  </div>
</div>
