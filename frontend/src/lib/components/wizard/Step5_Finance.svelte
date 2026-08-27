<script lang="ts">
  import { Calculator, CheckCircle, RefreshCw, KeyRound } from "lucide-svelte";
  import { setupApi } from "$lib/api/setup";
  import { toast } from "$lib/stores/toast";
  import type { SetupPayload } from "$lib/types";

  export let formData: SetupPayload;

  let isTestingMidtrans = false;
  let midtransStatus = "";

  async function testMidtransCredentials() {
    if (!formData.midtrans_server_key || !formData.midtrans_client_key) {
      toast.warning(
        "Masukkan Server Key & Client Key Midtrans terlebih dahulu",
      );
      return;
    }

    isTestingMidtrans = true;
    midtransStatus = "";

    try {
      const res = await setupApi.testMidtrans(
        formData.midtrans_server_key,
        formData.midtrans_client_key,
      );
      if (res.data?.valid || (res.data as any)?.status === "valid") {
        midtransStatus = "Kredensial Midtrans Valid!";
        toast.success(midtransStatus);
      } else {
        midtransStatus =
          "Kredensial Midtrans tidak valid atau akun Sandbox terblokir.";
        toast.error(midtransStatus);
      }
    } catch (err: any) {
      midtransStatus = err.message || "Error pengujian Midtrans";
    } finally {
      isTestingMidtrans = false;
    }
  }
</script>

<div class="space-y-6">
  <div>
    <h2 class="text-2xl font-bold text-slate-800 flex items-center gap-3">
      <Calculator class="w-7 h-7 text-brand-primary" />
      Langkah 5: Pajak, Service Fee & Payment Gateways
    </h2>
    <p class="text-slate-500 text-sm mt-1">
      Konfigurasi persentase pajak resto (PB1), biaya layanan, dan kunci
      integrasi Midtrans.
    </p>
  </div>

  <div
    class="bg-slate-50 p-5 rounded-2xl border border-slate-200 grid grid-cols-1 md:grid-cols-3 gap-4"
  >
    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="tax-pct"
        >Pajak Resto / PB1 (%)</label
      >
      <input
        id="tax-pct"
        type="number"
        step="0.1"
        bind:value={formData.tax_percentage}
        placeholder="0"
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-bold text-slate-800 text-lg shadow-sm outline-none transition"
      />
    </div>

    <div class="space-y-2">
      <label class="block text-sm font-bold text-slate-700" for="service-pct"
        >Service Charge (%)</label
      >
      <input
        id="service-pct"
        type="number"
        step="0.1"
        bind:value={formData.service_percentage}
        placeholder="0"
        class="w-full px-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-bold text-slate-800 text-lg shadow-sm outline-none transition"
      />
    </div>

    <div class="space-y-2 flex flex-col justify-end">
      <label
        class="flex items-center gap-3 p-3 bg-white border border-slate-300 rounded-xl cursor-pointer min-h-[48px]"
      >
        <input
          type="checkbox"
          bind:checked={formData.is_tax_inclusive}
          class="w-5 h-5 text-brand-primary rounded accent-brand-primary cursor-pointer"
        />
        <span class="text-xs font-bold text-slate-700"
          >Harga Menu Sudah Termasuk Pajak (Nett)</span
        >
      </label>
    </div>
  </div>

  {#if formData.active_modules.enable_midtrans}
    <div
      class="bg-gradient-to-br from-slate-900 to-slate-800 p-6 rounded-2xl text-white space-y-4 shadow-xl"
    >
      <div class="flex items-center gap-3">
        <KeyRound class="w-6 h-6 text-brand-cyan" />
        <h4 class="font-bold text-lg text-white">
          Integrasi Midtrans Snap / Dynamic QRIS
        </h4>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="space-y-1">
          <label
            class="block text-xs font-semibold text-slate-300"
            for="mid-server">Midtrans Server Key</label
          >
          <input
            id="mid-server"
            type="password"
            bind:value={formData.midtrans_server_key}
            placeholder="SB-Mid-server-xxxx..."
            class="w-full px-4 py-3 min-h-[48px] bg-slate-950/80 border border-slate-700 focus:border-brand-cyan rounded-xl font-mono text-sm text-brand-cyan shadow-sm outline-none transition"
          />
        </div>

        <div class="space-y-1">
          <label
            class="block text-xs font-semibold text-slate-300"
            for="mid-client">Midtrans Client Key</label
          >
          <input
            id="mid-client"
            type="password"
            bind:value={formData.midtrans_client_key}
            placeholder="SB-Mid-client-xxxx..."
            class="w-full px-4 py-3 min-h-[48px] bg-slate-950/80 border border-slate-700 focus:border-brand-cyan rounded-xl font-mono text-sm text-brand-cyan shadow-sm outline-none transition"
          />
        </div>
      </div>

      <div
        class="pt-3 border-t border-slate-800 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3"
      >
        <label class="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            bind:checked={formData.midtrans_is_production}
            class="w-5 h-5 accent-brand-cyan rounded cursor-pointer"
          />
          <span class="text-xs font-bold text-slate-300"
            >Gunakan Mode Production (Centang jika Akun Live Midtrans)</span
          >
        </label>

        <div class="flex items-center gap-3">
          <button
            type="button"
            on:click={testMidtransCredentials}
            disabled={isTestingMidtrans}
            class="px-5 py-2.5 min-h-[48px] font-bold text-slate-900 bg-brand-cyan hover:bg-cyan-300 rounded-xl flex items-center gap-2 transition"
          >
            {#if isTestingMidtrans}
              <RefreshCw class="w-4 h-4 animate-spin" />
              Pengujian...
            {:else}
              <CheckCircle class="w-4 h-4" />
              Uji Midtrans API Key
            {/if}
          </button>

          {#if midtransStatus}
            <span class="text-xs font-semibold text-brand-cyan"
              >{midtransStatus}</span
            >
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
