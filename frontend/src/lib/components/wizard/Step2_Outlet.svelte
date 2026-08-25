<script lang="ts">
  import { Store, Coffee, Utensils, UtensilsCrossed } from "lucide-svelte";
  import type { SetupPayload } from "$lib/types";

  export let formData: SetupPayload;
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-slate-800 flex items-center gap-3">
      <Store class="w-7 h-7 text-brand-primary" />
      Langkah 2: Profil Outlet & Resto
    </h2>
    <p class="text-slate-500 text-sm mt-1">
      Identitas usaha ini akan dicetak pada kepala struk transaksi fisik dan
      laporan.
    </p>
  </div>

  <!-- Business Type Selector -->
  <div>
    <label class="block text-sm font-bold text-slate-700 mb-3" for="biz-type"
      >Tipe Usaha Kuliner</label
    >
    <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
      <button
        type="button"
        on:click={() => (formData.business_type = "cafe")}
        class="p-4 rounded-xl border-2 flex items-center gap-3 transition text-left
        {formData.business_type === 'cafe'
          ? 'border-brand-primary bg-brand-primary/5 font-bold text-brand-primary'
          : 'border-slate-200 bg-white text-slate-700'}"
      >
        <Coffee class="w-6 h-6 flex-shrink-0" />
        <div>
          <p class="font-bold text-base">Kafe & Coffee Shop</p>
          <p class="text-xs text-slate-500 font-normal">
            Minuman, pastry, quick service
          </p>
        </div>
      </button>

      <button
        type="button"
        on:click={() => (formData.business_type = "resto")}
        class="p-4 rounded-xl border-2 flex items-center gap-3 transition text-left
        {formData.business_type === 'resto'
          ? 'border-brand-primary bg-brand-primary/5 font-bold text-brand-primary'
          : 'border-slate-200 bg-white text-slate-700'}"
      >
        <Utensils class="w-6 h-6 flex-shrink-0" />
        <div>
          <p class="font-bold text-base">Restoran Full-Service</p>
          <p class="text-xs text-slate-500 font-normal">
            Dine-in, meja, kitchen workflow
          </p>
        </div>
      </button>

      <button
        type="button"
        on:click={() => (formData.business_type = "cloud_kitchen")}
        class="p-4 rounded-xl border-2 flex items-center gap-3 transition text-left
        {formData.business_type === 'cloud_kitchen'
          ? 'border-brand-primary bg-brand-primary/5 font-bold text-brand-primary'
          : 'border-slate-200 bg-white text-slate-700'}"
      >
        <UtensilsCrossed class="w-6 h-6 flex-shrink-0" />
        <div>
          <p class="font-bold text-base">Cloud Kitchen / Takeaway</p>
          <p class="text-xs text-slate-500 font-normal">
            Pengiriman & pesan antar online
          </p>
        </div>
      </button>
    </div>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="outlet-name"
        >Nama Outlet / Resto *</label
      >
      <input
        id="outlet-name"
        type="text"
        bind:value={formData.outlet_name}
        placeholder="Nama Outlet"
        required
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-bold text-slate-800 shadow-sm outline-none transition"
      />
    </div>

    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="outlet-tagline"
        >Slogan / Tagline Struk</label
      >
      <input
        id="outlet-tagline"
        type="text"
        bind:value={formData.outlet_tagline}
        placeholder="Jargon / Tagline"
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-semibold text-slate-800 shadow-sm outline-none transition"
      />
    </div>

    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="outlet-phone"
        >Nomor Telepon / WA *</label
      >
      <input
        id="outlet-phone"
        type="text"
        bind:value={formData.outlet_phone}
        placeholder="08**-****-****"
        required
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-semibold text-slate-800 shadow-sm outline-none transition"
      />
    </div>

    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="outlet-address"
        >Alamat Fisik Resto *</label
      >
      <input
        id="outlet-address"
        type="text"
        bind:value={formData.outlet_address}
        placeholder="Nama jalan, blok, nomor bangunan"
        required
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-semibold text-slate-800 shadow-sm outline-none transition"
      />
    </div>
  </div>
</div>
