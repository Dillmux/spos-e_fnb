<script lang="ts">
  import { UserCheck, ShieldCheck, Lock } from "lucide-svelte";
  import type { SetupPayload } from "$lib/types";

  export let formData: SetupPayload;

  function handlePinInput(e: Event) {
    const input = e.target as HTMLInputElement;
    formData.admin_pin = input.value.replace(/\D/g, "").slice(0, 6);
  }
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-slate-800 flex items-center gap-3">
      <UserCheck class="w-7 h-7 text-brand-primary" />
      Langkah 6: Akun Super Admin & PIN Otorisasi
    </h2>
    <p class="text-slate-500 text-sm mt-1">
      Buat kredensial login admin utama serta PIN supervisor 6-digit untuk
      otorisasi void & split bill.
    </p>
  </div>

  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="admin-name"
        >Nama Lengkap Admin *</label
      >
      <input
        id="admin-name"
        type="text"
        bind:value={formData.admin_name}
        placeholder="Manager Utama / Owner"
        required
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-bold text-slate-800 shadow-sm outline-none transition"
      />
    </div>

    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="admin-user"
        >Username Login *</label
      >
      <input
        id="admin-user"
        type="text"
        bind:value={formData.admin_username}
        placeholder="username"
        required
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-bold text-slate-800 shadow-sm outline-none transition"
      />
    </div>

    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="admin-pass"
        >Password Login *</label
      >
      <div class="relative">
        <input
          id="admin-pass"
          type="password"
          bind:value={formData.admin_password}
          placeholder="••••••••"
          required
          class="w-full pl-4 pr-10 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-bold text-slate-800 shadow-sm outline-none transition"
        />
        <Lock class="w-5 h-5 absolute right-3 top-3.5 text-slate-400" />
      </div>
    </div>

    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="admin-pin">
        PIN Supervisor (6-Digit Angka) *
      </label>
      <div class="relative">
        <input
          id="admin-pin"
          type="password"
          inputmode="numeric"
          value={formData.admin_pin}
          on:input={handlePinInput}
          placeholder="cth. 123456"
          maxlength="6"
          required
          class="w-full pl-4 pr-10 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-mono text-lg font-bold text-slate-800 tracking-widest shadow-sm outline-none transition"
        />
        <ShieldCheck
          class="w-5 h-5 absolute right-3 top-3.5 text-brand-primary"
        />
      </div>
      <p class="text-xs text-slate-500">
        PIN ini akan diminta secara otomatis saat kasir melakukan tindakan
        terbatas (Void Item, Split Bill).
      </p>
    </div>
  </div>
</div>
