<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { currentUser, auth } from "$lib/stores/auth";
  import { config } from "$lib/stores/config";
  import { setupApi } from "$lib/api/setup";
  import { toast } from "$lib/stores/toast";
  import {
    Settings,
    ShieldAlert,
    Store,
    Calculator,
    KeyRound,
    ToggleLeft,
    Save,
    RefreshCw,
    CheckCircle,
    Building2,
    Lock,
    ShieldCheck,
  } from "lucide-svelte";

  let isLoading = true;
  let isSaving = false;

  // Form State
  let outletName = "";
  let outletTagline = "";
  let outletAddress = "";
  let outletPhone = "";
  let receiptFooter = "";

  let taxPercentage = 0;
  let servicePercentage = 0;
  let isTaxInclusive = false;

  let enableTables = true;
  let enableKDS = true;
  let enableWLD = true;
  let enableBOM = true;
  let enableMidtrans = true;

  let midtransServerKey = "";
  let midtransClientKey = "";
  let midtransIsProduction = false;

  onMount(async () => {
    if ($currentUser?.role !== "superadmin") {
      isLoading = false;
      return;
    }
    await loadSettings();
  });

  async function loadSettings() {
    isLoading = true;
    try {
      const res = await setupApi.getSettings();
      if (res.data) {
        const outlet = res.data.outlet || {};
        const flags = res.data.feature_flags || {};

        outletName = outlet.name || "";
        outletTagline = outlet.tagline || "";
        outletAddress = outlet.address || "";
        outletPhone = outlet.phone || "";
        receiptFooter = outlet.receipt_footer || "";

        taxPercentage = outlet.tax_percentage || 0;
        servicePercentage = outlet.service_percentage || 0;
        isTaxInclusive = !!outlet.is_tax_inclusive;

        enableTables = flags.enable_tables ?? true;
        enableKDS = flags.enable_kds ?? true;
        enableWLD = flags.enable_wld ?? true;
        enableBOM = flags.enable_bom ?? true;
        enableMidtrans = flags.enable_midtrans ?? true;

        const sk = res.data.midtrans_server_key || "";
        const ck = res.data.midtrans_client_key || "";
        midtransServerKey = sk.includes("placeholder") ? "" : sk;
        midtransClientKey = ck.includes("placeholder") ? "" : ck;
        midtransIsProduction = !!res.data.midtrans_is_production;
      }
    } catch (err: any) {
      toast.error(err.message || "Gagal memuat pengaturan sistem");
    } finally {
      isLoading = false;
    }
  }

  async function handleSaveClick() {
    if (!outletName.trim()) {
      toast.warning("Nama outlet tidak boleh kosong!");
      return;
    }

    try {
      // Prosedur Otorisasi Supervisor PIN sebelum menyimpan kredensial
      const pin = await auth.requestSupervisorPin(
        "Otorisasi Supervisor Simpan Pengaturan",
      );
      await executeSaveSettings(pin);
    } catch {
      // User cancelled PIN modal
    }
  }

  async function executeSaveSettings(pin: string) {
    isSaving = true;
    try {
      const res = await setupApi.updateSettings({
        supervisor_pin: pin,
        outlet_name: outletName,
        outlet_tagline: outletTagline,
        outlet_address: outletAddress,
        outlet_phone: outletPhone,
        receipt_footer: receiptFooter,
        tax_percentage: taxPercentage,
        service_percentage: servicePercentage,
        is_tax_inclusive: isTaxInclusive,
        enable_tables: enableTables,
        enable_kds: enableKDS,
        enable_wld: enableWLD,
        enable_bom: enableBOM,
        enable_midtrans: enableMidtrans,
        midtrans_server_key: midtransServerKey,
        midtrans_client_key: midtransClientKey,
        midtrans_is_production: midtransIsProduction,
      });

      if (res.success) {
        toast.success("Pengaturan sistem & kredensial berhasil disimpan!");
        // Refresh global stores across app
        await config.checkStatus();
        await loadSettings();
      } else {
        toast.error(res.message || "Gagal menyimpan perubahan pengaturan.");
      }
    } catch (err: any) {
      toast.error(err.message || "Otorisasi PIN gagal atau terjadi kesalahan.");
    } finally {
      isSaving = false;
    }
  }
</script>

<div class="h-full overflow-y-auto bg-slate-100 p-4 sm:p-6 lg:p-8">
  <div class="max-w-5xl mx-auto space-y-6">
    <!-- ACCESS CONTROL GUARD CHECK -->
    {#if $currentUser?.role !== "superadmin"}
      <div
        class="bg-white p-8 rounded-3xl shadow-xl border border-slate-200 text-center space-y-5 my-12"
      >
        <div
          class="inline-flex p-4 bg-red-100 text-brand-danger rounded-full shadow-inner"
        >
          <ShieldAlert class="w-16 h-16" />
        </div>
        <h2 class="text-2xl font-extrabold text-slate-800">
          Akses Ditolak: Khusus Superadmin
        </h2>
        <p class="text-sm text-slate-500 max-w-md mx-auto leading-relaxed">
          Halaman Pengaturan (Setting) Kredensial, Integrasi Midtrans, dan Modul
          Sistem ini terbatas untuk akun ber-role <strong>Superadmin</strong>.
        </p>
        <div class="pt-2">
          <button
            type="button"
            on:click={() => goto("/pos")}
            class="px-6 py-3 bg-brand-primary text-white font-bold rounded-xl shadow-md hover:bg-sky-500 transition min-h-[48px]"
          >
            Kembali ke Kasir POS
          </button>
        </div>
      </div>
    {:else if isLoading}
      <div class="py-20 text-center space-y-3">
        <RefreshCw class="w-10 h-10 animate-spin text-brand-primary mx-auto" />
        <p class="text-sm font-semibold text-slate-500">
          Memuat Pengaturan Sistem...
        </p>
      </div>
    {:else}
      <!-- HEADER BANNER -->
      <div
        class="bg-gradient-to-r from-brand-dark via-slate-900 to-slate-800 p-6 rounded-3xl text-white shadow-xl flex flex-col md:flex-row items-start md:items-center justify-between gap-4 border border-slate-800"
      >
        <div class="space-y-1">
          <div class="flex items-center gap-3">
            <div class="p-2.5 bg-brand-primary/20 text-brand-cyan rounded-2xl">
              <Settings class="w-7 h-7" />
            </div>
            <div>
              <h1 class="text-2xl font-extrabold tracking-wide">
                Pengaturan Sistem & Kredensial
              </h1>
              <p class="text-xs text-slate-300 font-semibold">
                Konfigurasi profil resto, biaya layanan, integrasi Midtrans, dan
                modul aktif.
              </p>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <span
            class="px-3.5 py-1.5 bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 rounded-full text-xs font-bold flex items-center gap-1.5"
          >
            <ShieldCheck class="w-4 h-4" />
            Superadmin Access
          </span>
        </div>
      </div>

      <!-- MAIN FORM SECTIONS -->
      <div class="space-y-6">
        <!-- SECTION 1: PROFIL & IDENTITAS OUTLET -->
        <div
          class="bg-white p-6 rounded-3xl shadow-sm border border-slate-200 space-y-4"
        >
          <div class="flex items-center gap-3 border-b border-slate-100 pb-3">
            <Building2 class="w-6 h-6 text-brand-primary" />
            <div>
              <h3 class="font-bold text-lg text-slate-800">
                Profil & Identitas Resto / Outlet
              </h3>
              <p class="text-xs text-slate-500">
                Informasi identitas resto yang akan tercetak pada struk fisik.
              </p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-1">
            <div class="space-y-1.5">
              <label
                class="block text-xs font-bold text-slate-700 uppercase"
                for="set-name">Nama Resto / Outlet *</label
              >
              <input
                id="set-name"
                type="text"
                bind:value={outletName}
                placeholder="misal: Resto Rasa Nusantara"
                class="w-full px-4 py-3 min-h-[48px] bg-slate-50 border border-slate-200 rounded-xl text-sm font-bold text-slate-800 focus:bg-white focus:border-brand-primary outline-none transition"
              />
            </div>

            <div class="space-y-1.5">
              <label
                class="block text-xs font-bold text-slate-700 uppercase"
                for="set-tagline">Slogan / Tagline Struk</label
              >
              <input
                id="set-tagline"
                type="text"
                bind:value={outletTagline}
                placeholder="misal: Cita Rasa Kuliner Sejati"
                class="w-full px-4 py-3 min-h-[48px] bg-slate-50 border border-slate-200 rounded-xl text-sm font-semibold text-slate-800 focus:bg-white focus:border-brand-primary outline-none transition"
              />
            </div>

            <div class="space-y-1.5">
              <label
                class="block text-xs font-bold text-slate-700 uppercase"
                for="set-phone">Nomor Telepon / WhatsApp</label
              >
              <input
                id="set-phone"
                type="text"
                bind:value={outletPhone}
                placeholder="081234567890"
                class="w-full px-4 py-3 min-h-[48px] bg-slate-50 border border-slate-200 rounded-xl text-sm font-semibold text-slate-800 focus:bg-white focus:border-brand-primary outline-none transition"
              />
            </div>

            <div class="space-y-1.5">
              <label
                class="block text-xs font-bold text-slate-700 uppercase"
                for="set-footer">Pesan Kaki Struk (Receipt Footer)</label
              >
              <input
                id="set-footer"
                type="text"
                bind:value={receiptFooter}
                placeholder="Terima kasih atas kunjungan Anda!"
                class="w-full px-4 py-3 min-h-[48px] bg-slate-50 border border-slate-200 rounded-xl text-sm font-semibold text-slate-800 focus:bg-white focus:border-brand-primary outline-none transition"
              />
            </div>

            <div class="space-y-1.5 md:col-span-2">
              <label
                class="block text-xs font-bold text-slate-700 uppercase"
                for="set-address">Alamat Lengkap Outlet</label
              >
              <textarea
                id="set-address"
                rows="2"
                bind:value={outletAddress}
                placeholder="Jl. Merdeka No. 123, Jakarta"
                class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl text-sm font-semibold text-slate-800 focus:bg-white focus:border-brand-primary outline-none transition"
              />
            </div>
          </div>
        </div>

        <!-- SECTION 2: PAJAK & BIAYA LAYANAN -->
        <div
          class="bg-white p-6 rounded-3xl shadow-sm border border-slate-200 space-y-4"
        >
          <div class="flex items-center gap-3 border-b border-slate-100 pb-3">
            <Calculator class="w-6 h-6 text-brand-primary" />
            <div>
              <h3 class="font-bold text-lg text-slate-800">
                Pajak Resto (PB1) & Biaya Layanan
              </h3>
              <p class="text-xs text-slate-500">
                Kalkulasi otomatis biaya pajak dan service charge transaksi POS.
              </p>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-1">
            <div class="space-y-1.5">
              <label
                class="block text-xs font-bold text-slate-700 uppercase"
                for="set-tax">Pajak Resto PB1 (%)</label
              >
              <input
                id="set-tax"
                type="number"
                step="0.1"
                bind:value={taxPercentage}
                placeholder="10"
                class="w-full px-4 py-3 min-h-[48px] bg-slate-50 border border-slate-200 rounded-xl text-lg font-extrabold text-slate-800 focus:bg-white focus:border-brand-primary outline-none transition"
              />
            </div>

            <div class="space-y-1.5">
              <label
                class="block text-xs font-bold text-slate-700 uppercase"
                for="set-svc">Service Charge (%)</label
              >
              <input
                id="set-svc"
                type="number"
                step="0.1"
                bind:value={servicePercentage}
                placeholder="5"
                class="w-full px-4 py-3 min-h-[48px] bg-slate-50 border border-slate-200 rounded-xl text-lg font-extrabold text-slate-800 focus:bg-white focus:border-brand-primary outline-none transition"
              />
            </div>

            <div class="space-y-1.5 flex flex-col justify-end">
              <label
                class="flex items-center gap-3 p-3 bg-slate-50 border border-slate-200 rounded-xl cursor-pointer min-h-[48px] hover:bg-slate-100 transition"
              >
                <input
                  type="checkbox"
                  bind:checked={isTaxInclusive}
                  class="w-5 h-5 text-brand-primary rounded accent-brand-primary cursor-pointer"
                />
                <span class="text-xs font-bold text-slate-700"
                  >Harga Menu Sudah Termasuk Pajak (Nett)</span
                >
              </label>
            </div>
          </div>
        </div>

        <!-- SECTION 3: KREDENSIAL MIDTRANS API -->
        <div
          class="bg-gradient-to-br from-slate-900 via-slate-900 to-slate-800 p-6 rounded-3xl text-white space-y-4 shadow-xl border border-slate-800"
        >
          <div
            class="flex items-center justify-between border-b border-slate-800 pb-3"
          >
            <div class="flex items-center gap-3">
              <KeyRound class="w-6 h-6 text-brand-cyan" />
              <div>
                <h3 class="font-bold text-lg text-white">
                  Kredensial Integration Midtrans Payment Gateway
                </h3>
                <p class="text-xs text-slate-300">
                  Digunakan untuk generate otomatis QRIS Dinamis pada transaksi
                  kasir.
                </p>
              </div>
            </div>

            <span
              class="text-xs font-semibold px-3 py-1 bg-purple-500/20 text-purple-300 border border-purple-500/30 rounded-full"
            >
              Midtrans Core API
            </span>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-1">
            <div class="space-y-1.5">
              <label
                class="block text-xs font-semibold text-slate-300"
                for="set-mid-server">Midtrans Server Key</label
              >
              <input
                id="set-mid-server"
                type="password"
                bind:value={midtransServerKey}
                placeholder="SB-Mid-server-xxxx..."
                class="w-full px-4 py-3 min-h-[48px] bg-slate-950 border border-slate-700 focus:border-brand-cyan rounded-xl font-mono text-sm text-brand-cyan outline-none transition"
              />
            </div>

            <div class="space-y-1.5">
              <label
                class="block text-xs font-semibold text-slate-300"
                for="set-mid-client">Midtrans Client Key</label
              >
              <input
                id="set-mid-client"
                type="password"
                bind:value={midtransClientKey}
                placeholder="SB-Mid-client-xxxx..."
                class="w-full px-4 py-3 min-h-[48px] bg-slate-950 border border-slate-700 focus:border-brand-cyan rounded-xl font-mono text-sm text-brand-cyan outline-none transition"
              />
            </div>
          </div>

          <div
            class="pt-2 flex items-center justify-between border-t border-slate-800"
          >
            <label class="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                bind:checked={midtransIsProduction}
                class="w-5 h-5 accent-brand-cyan rounded cursor-pointer"
              />
              <span class="text-xs font-bold text-slate-300"
                >Gunakan Mode Production (Centang jika Akun Live Midtrans)</span
              >
            </label>

            <span class="text-xs font-semibold text-slate-400">
              {midtransServerKey
                ? "Kredensial Terisi"
                : "Kredensial Belum Diisi"}
            </span>
          </div>
        </div>

        <!-- SECTION 4: FEATURE FLAGS MODUL AKTIF -->
        <div
          class="bg-white p-6 rounded-3xl shadow-sm border border-slate-200 space-y-4"
        >
          <div class="flex items-center gap-3 border-b border-slate-100 pb-3">
            <ToggleLeft class="w-6 h-6 text-brand-primary" />
            <div>
              <h3 class="font-bold text-lg text-slate-800">
                Aktifkan / Nonaktifkan Modul Sistem (Feature Flags)
              </h3>
              <p class="text-xs text-slate-500">
                Pilih modul bisnis yang ditampilkan di navigasi dan sistem.
              </p>
            </div>
          </div>

          <div
            class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 pt-1"
          >
            <label
              class="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-200 cursor-pointer hover:bg-slate-100 transition"
            >
              <div>
                <h4 class="font-bold text-sm text-slate-800">
                  Modul Meja (Tables)
                </h4>
                <p class="text-[11px] text-slate-500">
                  Tata letak meja Dine-in
                </p>
              </div>
              <input
                type="checkbox"
                bind:checked={enableTables}
                class="w-5 h-5 accent-brand-primary rounded cursor-pointer"
              />
            </label>

            <label
              class="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-200 cursor-pointer hover:bg-slate-100 transition"
            >
              <div>
                <h4 class="font-bold text-sm text-slate-800">
                  Modul KDS (Dapur)
                </h4>
                <p class="text-[11px] text-slate-500">Kitchen Display System</p>
              </div>
              <input
                type="checkbox"
                bind:checked={enableKDS}
                class="w-5 h-5 accent-brand-primary rounded cursor-pointer"
              />
            </label>

            <label
              class="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-200 cursor-pointer hover:bg-slate-100 transition"
            >
              <div>
                <h4 class="font-bold text-sm text-slate-800">
                  Modul WLD (Antrean)
                </h4>
                <p class="text-[11px] text-slate-500">Waiter & Order Monitor</p>
              </div>
              <input
                type="checkbox"
                bind:checked={enableWLD}
                class="w-5 h-5 accent-brand-primary rounded cursor-pointer"
              />
            </label>

            <label
              class="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-200 cursor-pointer hover:bg-slate-100 transition"
            >
              <div>
                <h4 class="font-bold text-sm text-slate-800">
                  Modul BOM Stock
                </h4>
                <p class="text-[11px] text-slate-500">
                  Potong stok bahan resep
                </p>
              </div>
              <input
                type="checkbox"
                bind:checked={enableBOM}
                class="w-5 h-5 accent-brand-primary rounded cursor-pointer"
              />
            </label>

            <label
              class="flex items-center justify-between p-4 bg-slate-50 rounded-2xl border border-slate-200 cursor-pointer hover:bg-slate-100 transition"
            >
              <div>
                <h4 class="font-bold text-sm text-slate-800">
                  Modul Midtrans QRIS
                </h4>
                <p class="text-[11px] text-slate-500">
                  Pembayaran Dynamic QRIS
                </p>
              </div>
              <input
                type="checkbox"
                bind:checked={enableMidtrans}
                class="w-5 h-5 accent-brand-primary rounded cursor-pointer"
              />
            </label>
          </div>
        </div>

        <!-- SAVE BUTTON FOOTER CARD -->
        <div
          class="bg-white p-5 rounded-3xl border border-slate-200 shadow-lg flex items-center justify-between sticky bottom-4 z-20"
        >
          <div class="flex items-center gap-2 text-xs font-bold text-slate-500">
            <Lock class="w-4 h-4 text-amber-500" />
            <span
              >Diperlukan Otorisasi Supervisor PIN 6-Digit untuk menyimpan</span
            >
          </div>

          <button
            type="button"
            on:click={handleSaveClick}
            disabled={isSaving}
            class="px-8 py-3.5 min-h-[48px] bg-brand-primary hover:bg-sky-500 active:scale-95 text-white font-extrabold text-sm rounded-2xl shadow-lg shadow-sky-500/25 flex items-center gap-2 transition disabled:opacity-50"
          >
            {#if isSaving}
              <RefreshCw class="w-5 h-5 animate-spin" />
              Menyimpan...
            {:else}
              <Save class="w-5 h-5" />
              Simpan Perubahan Settings
            {/if}
          </button>
        </div>
      </div>
    {/if}
  </div>
</div>
