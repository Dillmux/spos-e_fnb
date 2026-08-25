<script lang="ts">
  import { goto } from "$app/navigation";
  import { setupApi } from "$lib/api/setup";
  import { config } from "$lib/stores/config";
  import { auth } from "$lib/stores/auth";
  import { toast } from "$lib/stores/toast";
  import Step1_DB from "$lib/components/wizard/Step1_DB.svelte";
  import Step2_Outlet from "$lib/components/wizard/Step2_Outlet.svelte";
  import Step3_Type from "$lib/components/wizard/Step3_Type.svelte";
  import Step4_Modules from "$lib/components/wizard/Step4_Modules.svelte";
  import Step5_Finance from "$lib/components/wizard/Step5_Finance.svelte";
  import Step6_Admin from "$lib/components/wizard/Step6_Admin.svelte";
  import Step7_Finish from "$lib/components/wizard/Step7_Finish.svelte";
  import type { SetupPayload } from "$lib/types";
  import { Store, ChevronLeft, ChevronRight } from "lucide-svelte";

  let currentStep = 1;
  let isSubmitting = false;

  let formData: SetupPayload = {
    db_driver: "sqlite",
    db_connection_string: "",
    outlet_name: "",
    outlet_tagline: "",
    outlet_address: "",
    outlet_phone: "",
    business_type: "resto",
    service_modes: ["dine_in", "take_away"],
    active_modules: {
      enable_tables: true,
      enable_kds: true,
      enable_wld: true,
      enable_bom: true,
      enable_midtrans: true,
    },
    tax_percentage: 0,
    service_percentage: 0,
    is_tax_inclusive: false,
    midtrans_server_key: "",
    midtrans_client_key: "",
    admin_name: "",
    admin_username: "",
    admin_password: "",
    admin_pin: "",
  };

  const steps = [
    "Basis Data",
    "Profil Outlet",
    "Mode Layanan",
    "Modul Fitur",
    "Pajak & Gateways",
    "Akun Super Admin",
    "Selesai",
  ];

  function nextStep() {
    if (
      currentStep === 2 &&
      (!formData.outlet_name ||
        !formData.outlet_phone ||
        !formData.outlet_address)
    ) {
      toast.warning(
        "Lengkapi nama, telepon, dan alamat resto terlebih dahulu.",
      );
      return;
    }
    if (
      currentStep === 6 &&
      (!formData.admin_username ||
        !formData.admin_password ||
        formData.admin_pin.length !== 6)
    ) {
      toast.warning("Lengkapi kredensial admin dan 6-digit PIN supervisor.");
      return;
    }
    if (currentStep < 7) {
      currentStep += 1;
    }
  }

  function prevStep() {
    if (currentStep > 1) {
      currentStep -= 1;
    }
  }

  async function handleInitializeSystem() {
    isSubmitting = true;
    try {
      const res = await setupApi.initialize(formData);
      if (res.data?.initialized) {
        if (res.data.token) {
          auth.setToken(res.data.token);
          auth.setUser({
            id: res.data.user_id || "superadmin-id",
            name: formData.admin_name,
            username: res.data.username || formData.admin_username,
            role: (res.data.role as any) || "superadmin",
            is_active: true,
          });
        }
        config.setConfig({
          initialized: true,
          featureFlags: formData.active_modules,
        });
        toast.success("Inisialisasi Smart POS Berhasil!");
        goto("/pos");
      } else {
        toast.error("Inisialisasi gagal. Periksa kembali log backend.");
      }
    } catch (err: any) {
      toast.error(err.message || "Gagal mengeksekusi inisialisasi wizard");
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div
  class="min-h-full bg-brand-surface py-8 px-4 sm:px-6 lg:px-8 flex flex-col items-center justify-center"
>
  <div
    class="w-full max-w-4xl bg-white rounded-3xl shadow-2xl overflow-hidden border border-slate-100 flex flex-col"
  >
    <!-- Header Banner -->
    <div
      class="bg-brand-dark p-6 text-white flex items-center justify-between border-b border-slate-800"
    >
      <div class="flex items-center gap-3">
        <div class="p-3 bg-brand-primary rounded-2xl text-white shadow-md">
          <Store class="w-7 h-7" />
        </div>
        <div>
          <h1 class="font-black text-2xl tracking-wide">
            Smart POS Inisialisasi Wizard
          </h1>
          <p class="text-xs text-brand-cyan font-semibold">
            Konfigurasi awal resto & basis data modular
          </p>
        </div>
      </div>
      <span
        class="text-xs font-extrabold px-3 py-1.5 bg-slate-800 text-brand-cyan rounded-full border border-slate-700"
      >
        Langkah {currentStep} dari 7
      </span>
    </div>

    <!-- Stepper Progress Indicator -->
    <div
      class="bg-slate-50 px-6 py-4 border-b border-slate-200 overflow-x-auto"
    >
      <div class="flex items-center justify-between min-w-[600px]">
        {#each steps as stepLabel, idx}
          {@const stepNum = idx + 1}
          <div class="flex items-center gap-2">
            <div
              class="w-8 h-8 rounded-full font-extrabold text-xs flex items-center justify-center transition-all shadow-sm
              {currentStep === stepNum
                ? 'bg-brand-primary text-white ring-4 ring-brand-cyan/40 scale-110'
                : currentStep > stepNum
                  ? 'bg-emerald-500 text-white'
                  : 'bg-slate-200 text-slate-500'}"
            >
              {currentStep > stepNum ? "✓" : stepNum}
            </div>
            <span
              class="text-xs font-bold {currentStep === stepNum
                ? 'text-slate-900 font-extrabold'
                : 'text-slate-500'}"
            >
              {stepLabel}
            </span>
          </div>
          {#if idx < steps.length - 1}
            <div class="h-0.5 flex-1 bg-slate-200 mx-2"></div>
          {/if}
        {/each}
      </div>
    </div>

    <!-- Step Body Component Render -->
    <div class="p-8 flex-1 overflow-y-auto min-h-[420px]">
      {#if currentStep === 1}
        <Step1_DB bind:formData />
      {:else if currentStep === 2}
        <Step2_Outlet bind:formData />
      {:else if currentStep === 3}
        <Step3_Type bind:formData />
      {:else if currentStep === 4}
        <Step4_Modules bind:formData />
      {:else if currentStep === 5}
        <Step5_Finance bind:formData />
      {:else if currentStep === 6}
        <Step6_Admin bind:formData />
      {:else if currentStep === 7}
        <Step7_Finish
          bind:formData
          onInitialize={handleInitializeSystem}
          {isSubmitting}
        />
      {/if}
    </div>

    <!-- Wizard Navigation Footer -->
    {#if currentStep < 7}
      <div
        class="p-6 bg-slate-50 border-t border-slate-200 flex items-center justify-between"
      >
        <button
          type="button"
          on:click={prevStep}
          disabled={currentStep === 1}
          class="px-5 py-3 font-bold text-slate-700 bg-white border border-slate-300 hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed rounded-xl flex items-center gap-2 transition"
        >
          <ChevronLeft class="w-5 h-5" />
          Sebelumnya
        </button>

        <button
          type="button"
          on:click={nextStep}
          class="px-8 py-3 font-extrabold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md flex items-center gap-2 transition"
        >
          Lanjut
          <ChevronRight class="w-5 h-5" />
        </button>
      </div>
    {/if}
  </div>
</div>
