<script lang="ts">
  import { onMount } from "svelte";
  import { shiftsApi } from "$lib/api/shifts";
  import { toast } from "$lib/stores/toast";
  import Modal from "$lib/components/common/Modal.svelte";
  import CurrencyInput from "$lib/components/common/CurrencyInput.svelte";
  import type { Shift } from "$lib/types";
  import {
    Wallet,
    ArrowDownRight,
    ArrowUpRight,
    CheckCircle2,
    Lock,
  } from "lucide-svelte";

  export let isOpen: boolean = false;
  export let onClose: () => void = () => {};

  let activeShift: (Shift & { user?: { name: string } }) | null = null;
  let isLoading = true;
  let openingCash: number = 0;
  let actualCashClose: number = 0;

  // Cash movement form
  let movementType: "in" | "out" = "in";
  let movementAmount: number = 0;
  let movementReason: string = "";

  async function loadShiftStatus() {
    isLoading = true;
    try {
      const res = await shiftsApi.getActiveShift();
      const rawData = res.data as any;
      if (rawData && rawData.shift) {
        activeShift = {
          ...rawData.shift,
          total_cash_sales: rawData.cash_sales_total || 0,
        };
      } else if (rawData && rawData.id) {
        activeShift = rawData;
      } else {
        activeShift = null;
      }
    } catch {
      activeShift = null;
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    if (isOpen) loadShiftStatus();
  });

  $: if (isOpen) {
    loadShiftStatus();
  }

  async function handleOpenShift() {
    if (openingCash < 0) {
      toast.warning("Modal awalan tidak boleh bernilai negatif.");
      return;
    }

    try {
      const res = await shiftsApi.openShift(openingCash);
      const rawData = res.data as any;
      if (rawData && rawData.shift) {
        activeShift = {
          ...rawData.shift,
          total_cash_sales: rawData.cash_sales_total || 0,
        };
      } else if (rawData && rawData.id) {
        activeShift = rawData;
      }
      toast.success("Shift Kasir Berhasil Dibuka!");
      loadShiftStatus();
    } catch (err: any) {
      toast.error(err.message || "Gagal membuka shift kasir");
    }
  }

  async function handleCloseShift() {
    try {
      await shiftsApi.closeShift(actualCashClose);
      toast.success("Shift Kasir Selesai & Dibatasi Rekonsiliasi.");
      activeShift = null;
      onClose();
    } catch (err: any) {
      toast.error(err.message || "Gagal menutup shift");
    }
  }

  async function handleAddMovement() {
    if (movementAmount <= 0 || !movementReason) {
      toast.warning("Masukkan nominal dan alasan petty cash");
      return;
    }
    try {
      await shiftsApi.addCashMovement(
        movementType,
        movementAmount,
        movementReason,
      );
      toast.success(
        `Kas ${movementType === "in" ? "Masuk" : "Keluar"} dicatat`,
      );
      movementAmount = 0;
      movementReason = "";
      loadShiftStatus();
    } catch (err: any) {
      toast.error(err.message || "Gagal mencatat mutasi kas");
    }
  }
</script>

<Modal
  {isOpen}
  title="Manajemen Shift & Cash Drawer Kasir"
  maxWidth="lg"
  {onClose}
>
  {#if isLoading}
    <div class="py-12 text-center text-slate-400 font-semibold">
      Memuat Status Shift...
    </div>
  {:else if !activeShift}
    <!-- OPEN SHIFT FORM -->
    <div class="space-y-5 py-2">
      <div
        class="flex items-center gap-3 p-4 bg-blue-50 text-blue-900 rounded-2xl border border-blue-200"
      >
        <Wallet class="w-8 h-8 text-brand-primary" />
        <div>
          <h4 class="font-bold text-base">Buka Shift Kasir Baru</h4>
          <p class="text-xs text-blue-700">
            Input saldo kas awal di laci kasir (modal awal kembalian).
          </p>
        </div>
      </div>

      <div class="space-y-2">
        <label
          class="block text-xs font-bold text-slate-600 uppercase"
          for="modal-open-cash">Modal Awalan (Opening Cash)</label
        >
        <CurrencyInput bind:value={openingCash} placeholder="200.000" />
      </div>

      <button
        type="button"
        on:click={handleOpenShift}
        class="w-full py-4 min-h-[48px] bg-brand-primary hover:bg-sky-500 text-white font-extrabold text-lg rounded-2xl shadow-lg transition"
      >
        BUKA SHIFT KASIR SEKARANG
      </button>
    </div>
  {:else}
    <!-- ACTIVE SHIFT RECONCILIATION -->
    <div class="space-y-6">
      <div class="p-4 bg-slate-900 text-white rounded-2xl space-y-3">
        <div
          class="flex items-center justify-between border-b border-slate-700 pb-2"
        >
          <div class="flex items-center gap-2">
            <CheckCircle2 class="w-5 h-5 text-brand-mint-active" />
            <span class="font-bold text-sm"
              >Shift Aktif: {activeShift?.user?.name ||
                activeShift?.user_name ||
                "Kasir"}</span
            >
          </div>
          <span class="text-xs text-slate-400"
            >Mulai: {activeShift?.start_time
              ? new Date(activeShift.start_time).toLocaleTimeString("id-ID")
              : "-"}</span
          >
        </div>

        <div class="grid grid-cols-2 gap-4 text-xs">
          <div>
            <span class="text-slate-400 block">Modal Awal:</span>
            <strong class="text-base text-white"
              >Rp {(activeShift?.opening_cash || 0).toLocaleString(
                "id-ID",
              )}</strong
            >
          </div>
          <div>
            <span class="text-slate-400 block">Penjualan Tunai:</span>
            <strong class="text-base text-brand-cyan"
              >Rp {(activeShift?.total_cash_sales || 0).toLocaleString(
                "id-ID",
              )}</strong
            >
          </div>
        </div>
      </div>

      <!-- Petty Cash Entry -->
      <div
        class="bg-slate-50 p-4 rounded-2xl border border-slate-200 space-y-3"
      >
        <h5 class="font-bold text-slate-800 text-sm flex items-center gap-2">
          <Wallet class="w-4 h-4 text-brand-primary" />
          Catat Petty Cash (Uang Keluar / Masuk Laci)
        </h5>

        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            on:click={() => (movementType = "in")}
            class="py-2 text-xs font-bold rounded-xl border flex items-center justify-center gap-1 transition
            {movementType === 'in'
              ? 'bg-emerald-600 text-white border-emerald-600'
              : 'bg-white text-slate-700 border-slate-200'}"
          >
            <ArrowDownRight class="w-4 h-4" /> Uang Masuk
          </button>
          <button
            type="button"
            on:click={() => (movementType = "out")}
            class="py-2 text-xs font-bold rounded-xl border flex items-center justify-center gap-1 transition
            {movementType === 'out'
              ? 'bg-brand-danger text-white border-brand-danger'
              : 'bg-white text-slate-700 border-slate-200'}"
          >
            <ArrowUpRight class="w-4 h-4" /> Uang Keluar
          </button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
          <CurrencyInput bind:value={movementAmount} placeholder="Nominal Rp" />
          <input
            type="text"
            bind:value={movementReason}
            placeholder="Alasan (misal: Beli Es Batu)"
            class="px-3 py-2 bg-white border border-slate-300 rounded-xl text-xs font-semibold outline-none focus:border-brand-primary"
          />
        </div>

        <button
          type="button"
          on:click={handleAddMovement}
          class="w-full py-2.5 bg-slate-800 text-white font-bold text-xs rounded-xl hover:bg-slate-700 transition"
        >
          Simpan Mutasi Kas
        </button>
      </div>

      <!-- Close Shift Form -->
      <div class="pt-2 border-t border-slate-200 space-y-3">
        <h5 class="font-bold text-slate-800 text-sm flex items-center gap-2">
          <Lock class="w-4 h-4 text-brand-danger" />
          Tutup Shift & Hitung Kas Fisik
        </h5>

        <CurrencyInput
          bind:value={actualCashClose}
          placeholder="Hitung Fisik Uang di Laci Kasir"
        />

        <button
          type="button"
          on:click={handleCloseShift}
          class="w-full py-3.5 bg-brand-danger hover:bg-red-600 text-white font-extrabold text-sm rounded-xl shadow-md transition"
        >
          TUTUP SHIFT & HITUNG SELISIH KAS
        </button>
      </div>
    </div>
  {/if}
</Modal>
