<script lang="ts">
  import { onDestroy } from "svelte";
  import { cart, cartTotals } from "$lib/stores/cart";
  import { printer } from "$lib/stores/printer";
  import { config, featureFlags } from "$lib/stores/config";
  import { posApi } from "$lib/api/pos";
  import { shiftsApi } from "$lib/api/shifts";
  import { toast } from "$lib/stores/toast";
  import Modal from "$lib/components/common/Modal.svelte";
  import ReceiptModal from "$lib/components/pos/ReceiptModal.svelte";
  import Numpad from "$lib/components/common/Numpad.svelte";
  import CurrencyInput from "$lib/components/common/CurrencyInput.svelte";
  import { getCartBOMSummaryList } from "$lib/utils/bomStock";
  import {
    Banknote,
    QrCode,
    Printer,
    CheckCircle2,
    Clock,
    RefreshCw,
    Utensils,
  } from "lucide-svelte";

  export let isOpen: boolean = false;
  export let existingOrder: any = null;
  export let onClose: () => void = () => {};

  let activeMethod: "cash" | "qris" = "cash";
  let cashTendered: number = 0;
  let isProcessing = false;
  let isPaidSuccess = false;
  let isReceiptModalOpen = false;
  let lastPaidOrder: any = null;
  let lastPaymentResult: any = null;

  // Snapshots for Success Screen display after cart.clear()
  let savedTotalDue = 0;
  let savedCashTendered = 0;
  let savedChangeAmount = 0;

  // QRIS State
  let qrisUrl: string = "";
  let qrisTimerSeconds = 900; // 15 minutes countdown
  let timerInterval: any = null;
  let pollInterval: any = null;

  $: totalDue = existingOrder
    ? existingOrder.total_amount || existingOrder.total || 0
    : $cartTotals.grandTotal;
  $: changeAmount = Math.max(0, cashTendered - totalDue);
  $: bomSummaryList = getCartBOMSummaryList($cart.items);

  // Quick Cash Preset Amounts
  $: presetPas = totalDue;
  $: preset50k =
    totalDue <= 50000 ? 50000 : Math.ceil(totalDue / 50000) * 50000;
  $: preset100k =
    totalDue <= 100000 ? 100000 : Math.ceil(totalDue / 100000) * 100000;

  function selectPreset(amount: number) {
    cashTendered = amount;
  }

  async function handleCashCheckout() {
    if (cashTendered < totalDue) {
      toast.warning("Jumlah uang tunai kurang dari total tagihan!");
      return;
    }

    isProcessing = true;
    try {
      // Verify active shift status before cash checkout
      try {
        const shiftRes = await shiftsApi.getActiveShift();
        const rawShift = shiftRes.data as any;
        const activeShift =
          rawShift?.shift || (rawShift && rawShift.id ? rawShift : null);

        if (!activeShift || activeShift.status !== "open") {
          toast.warning(
            "Shift register belum dibuka! Silakan buka shift kasir terlebih dahulu sebelum melunasi transaksi.",
          );
          window.dispatchEvent(new CustomEvent("open-shift-modal"));
          isProcessing = false;
          return;
        }
      } catch (err: any) {
        toast.warning(
          "Gagal verifikasi status shift kasir. Pastikan shift kasir sudah dibuka.",
        );
        window.dispatchEvent(new CustomEvent("open-shift-modal"));
        isProcessing = false;
        return;
      }

      // Capture financial snapshots before cart clearing
      savedTotalDue = totalDue;
      savedCashTendered = cashTendered;
      savedChangeAmount = changeAmount;

      let targetOrder = existingOrder;
      if (!targetOrder) {
        const tableIdParsed = $cart.selectedTableId
          ? typeof $cart.selectedTableId === "string"
            ? parseInt($cart.selectedTableId, 10)
            : $cart.selectedTableId
          : undefined;

        const orderRes = await posApi.createOrder({
          table_id: isNaN(tableIdParsed as any)
            ? $cart.selectedTableId || undefined
            : tableIdParsed,
          service_type: $cart.serviceMode,
          customer_name: $cart.customerName || undefined,
          items: $cart.items.map((i) => ({
            product_id: i.product.id,
            quantity: i.quantity,
            unit_price: i.product.price,
            item_notes: i.notes,
            discount_amount: i.discount_amount,
          })),
          discount_amount: $cartTotals.manualDiscount,
        });

        if (!orderRes.data) {
          throw new Error(
            orderRes.error?.message || "Gagal membuat transaksi order",
          );
        }

        window.dispatchEvent(new CustomEvent("refresh-tables"));
        targetOrder = orderRes.data;
      }

      const payRes = await posApi.checkoutCash(targetOrder.id, cashTendered);
      if (!payRes.data) {
        throw new Error(
          payRes.error?.message || "Gagal memproses pelunasan cash",
        );
      }

      lastPaidOrder = targetOrder;
      lastPaymentResult = payRes.data;
      isPaidSuccess = true;

      toast.success("Pembayaran Tunai Berhasil!");
      window.dispatchEvent(new CustomEvent("refresh-tables"));
      if (!existingOrder) {
        cart.clear();
      }
    } catch (err: any) {
      toast.error(err.message || "Gagal memproses pembayaran cash");
    } finally {
      isProcessing = false;
    }
  }

  async function startQRISFlow() {
    isProcessing = true;
    try {
      // Verify active shift status before QRIS checkout
      try {
        const shiftRes = await shiftsApi.getActiveShift();
        const rawShift = shiftRes.data as any;
        const activeShift =
          rawShift?.shift || (rawShift && rawShift.id ? rawShift : null);

        if (!activeShift || activeShift.status !== "open") {
          toast.warning(
            "Shift register belum dibuka! Silakan buka shift kasir terlebih dahulu sebelum melunasi transaksi.",
          );
          window.dispatchEvent(new CustomEvent("open-shift-modal"));
          isProcessing = false;
          return;
        }
      } catch (err: any) {
        toast.warning(
          "Gagal verifikasi status shift kasir. Pastikan shift kasir sudah dibuka.",
        );
        window.dispatchEvent(new CustomEvent("open-shift-modal"));
        isProcessing = false;
        return;
      }

      savedTotalDue = totalDue;
      savedCashTendered = 0;
      savedChangeAmount = 0;

      let targetOrder = existingOrder;
      if (!targetOrder) {
        const tableIdParsed = $cart.selectedTableId
          ? typeof $cart.selectedTableId === "string"
            ? parseInt($cart.selectedTableId, 10)
            : $cart.selectedTableId
          : undefined;

        const orderRes = await posApi.createOrder({
          table_id: isNaN(tableIdParsed as any)
            ? $cart.selectedTableId || undefined
            : tableIdParsed,
          service_type: $cart.serviceMode,
          customer_name: $cart.customerName || undefined,
          items: $cart.items.map((i) => ({
            product_id: i.product.id,
            quantity: i.quantity,
            unit_price: i.product.price,
            item_notes: i.notes,
            discount_amount: i.discount_amount,
          })),
          discount_amount: $cartTotals.manualDiscount,
        });

        if (!orderRes.data) {
          throw new Error(
            orderRes.error?.message || "Gagal membuat transaksi order",
          );
        }

        window.dispatchEvent(new CustomEvent("refresh-tables"));
        targetOrder = orderRes.data;
      }

      lastPaidOrder = targetOrder;

      try {
        const qrisRes = await posApi.checkoutQRIS(targetOrder.id);
        if (qrisRes.data?.qris_url) {
          qrisUrl = qrisRes.data.qris_url;
        } else {
          const qrData = qrisRes.data?.qr_string || targetOrder.id;
          qrisUrl = `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(qrData)}`;
        }
      } catch {
        qrisUrl = `https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=${encodeURIComponent(targetOrder.id)}`;
      }

      startTimers(targetOrder.id);
    } catch (err: any) {
      toast.error(err.message || "Gagal membuat QRIS payment");
    } finally {
      isProcessing = false;
    }
  }

  function startTimers(orderId: string) {
    qrisTimerSeconds = 900;

    clearInterval(timerInterval);
    timerInterval = setInterval(() => {
      if (qrisTimerSeconds > 0) {
        qrisTimerSeconds -= 1;
      } else {
        clearInterval(timerInterval);
        clearInterval(pollInterval);
        toast.error("Waktu QRIS telah kadaluarsa");
      }
    }, 1000);

    clearInterval(pollInterval);
    pollInterval = setInterval(async () => {
      try {
        const statusRes = await posApi.getPaymentStatus(orderId);
        if (statusRes.data?.status === "settlement") {
          clearInterval(timerInterval);
          clearInterval(pollInterval);
          lastPaymentResult = statusRes.data;
          isPaidSuccess = true;
          toast.success("Pembayaran QRIS Terverifikasi Settlement!");
          cart.clear();
        }
      } catch {
        // Ignored during polling
      }
    }, 3000);
  }

  function handleCloseModal() {
    clearInterval(timerInterval);
    clearInterval(pollInterval);
    isPaidSuccess = false;
    qrisUrl = "";
    cashTendered = 0;
    savedTotalDue = 0;
    savedCashTendered = 0;
    savedChangeAmount = 0;
    lastPaidOrder = null;
    lastPaymentResult = null;
    window.dispatchEvent(new CustomEvent("refresh-tables"));
    onClose();
  }

  onDestroy(() => {
    clearInterval(timerInterval);
    clearInterval(pollInterval);
  });

  $: formattedTimer = `${Math.floor(qrisTimerSeconds / 60)}:${(qrisTimerSeconds % 60).toString().padStart(2, "0")}`;
</script>

<Modal
  {isOpen}
  title="Kasir Checkout & Pembayaran"
  maxWidth="2xl"
  onClose={handleCloseModal}
>
  {#if isPaidSuccess}
    <!-- SUCCESS SCREEN -->
    <div class="py-8 text-center space-y-4 select-none">
      <div
        class="inline-flex p-4 bg-emerald-100 text-emerald-600 rounded-full animate-bounce"
      >
        <CheckCircle2 class="w-16 h-16" />
      </div>
      <h3 class="text-3xl font-extrabold text-slate-800">
        Transaksi Selesai & Lunas!
      </h3>
      <p class="text-slate-500 font-semibold text-sm">
        No. Struk: <strong class="text-slate-800"
          >{lastPaidOrder?.invoice_number ||
            lastPaidOrder?.order_number ||
            `#${lastPaidOrder?.id}`}</strong
        >
      </p>

      <div
        class="bg-slate-50 max-w-md mx-auto p-4 rounded-2xl border border-slate-200 space-y-2 text-left text-sm"
      >
        <div class="flex justify-between text-slate-600">
          <span>Total Tagihan:</span>
          <span class="font-bold text-slate-900">
            Rp {(lastPaidOrder?.total_amount || savedTotalDue).toLocaleString(
              "id-ID",
            )}
          </span>
        </div>
        {#if activeMethod === "cash"}
          <div class="flex justify-between text-slate-600">
            <span>Uang Tunai:</span>
            <span class="font-bold text-slate-900">
              Rp {(
                lastPaymentResult?.cash_tendered || savedCashTendered
              ).toLocaleString("id-ID")}
            </span>
          </div>
          <div
            class="flex justify-between text-emerald-600 font-bold text-base pt-1 border-t border-slate-200"
          >
            <span>Kembalian:</span>
            <span>
              Rp {(
                lastPaymentResult?.change_amount ?? savedChangeAmount
              ).toLocaleString("id-ID")}
            </span>
          </div>
        {/if}
      </div>

      <div class="pt-4 flex items-center justify-center gap-3">
        <button
          type="button"
          on:click={() => (isReceiptModalOpen = true)}
          class="px-6 py-3 min-h-[48px] bg-brand-primary text-white font-bold rounded-xl hover:bg-sky-500 shadow-md flex items-center gap-2 transition"
        >
          <Printer class="w-5 h-5" />
          Lihat & Cetak Struk
        </button>

        <button
          type="button"
          on:click={handleCloseModal}
          class="px-6 py-3 min-h-[48px] bg-slate-200 text-slate-800 font-bold rounded-xl hover:bg-slate-300 transition"
        >
          Transaksi Baru
        </button>
      </div>
    </div>
  {:else}
    <!-- PAYMENT WORKSPACE -->
    <div class="space-y-5">
      <!-- Method Selector Tabs -->
      <div
        class="grid grid-cols-2 gap-3 p-1 bg-slate-100 rounded-2xl border border-slate-200"
      >
        <button
          type="button"
          on:click={() => (activeMethod = "cash")}
          class="touch-target py-3 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition
          {activeMethod === 'cash'
            ? 'bg-white text-brand-primary shadow-sm'
            : 'text-slate-600 hover:text-slate-900'}"
        >
          <Banknote class="w-5 h-5" />
          Tunai / Cash Engine
        </button>

        <button
          type="button"
          on:click={() => {
            activeMethod = "qris";
            startQRISFlow();
          }}
          class="touch-target py-3 rounded-xl font-bold text-sm flex items-center justify-center gap-2 transition
          {activeMethod === 'qris'
            ? 'bg-white text-purple-600 shadow-sm'
            : 'text-slate-600 hover:text-slate-900'}"
        >
          <QrCode class="w-5 h-5" />
          Midtrans Dynamic QRIS
        </button>
      </div>

      <!-- Total Due Banner -->
      <div
        class="bg-brand-dark p-5 rounded-2xl text-white flex items-center justify-between shadow-inner"
      >
        <div>
          <span
            class="text-xs font-semibold text-slate-400 uppercase tracking-wider block"
            >Total Pembayaran</span
          >
          <h2 class="display-currency text-brand-cyan">
            Rp {totalDue.toLocaleString("id-ID")}
          </h2>
        </div>

        {#if activeMethod === "cash"}
          <div class="text-right">
            <span class="text-xs font-semibold text-slate-400 block"
              >Kembalian Kasir</span
            >
            <span
              class="text-2xl font-bold {changeAmount >= 0
                ? 'text-brand-mint-light'
                : 'text-brand-danger'}"
            >
              Rp {changeAmount.toLocaleString("id-ID")}
            </span>
          </div>
        {/if}
      </div>

      <!-- CASH TAB CONTENT -->
      {#if activeMethod === "cash"}
        <div class="space-y-4">
          <!-- Quick Cash Presets -->
          <div class="space-y-2">
            <span class="block text-xs font-bold text-slate-500 uppercase"
              >Preset Uang Pas / Pecahan</span
            >
            <div class="grid grid-cols-3 gap-2">
              <button
                type="button"
                on:click={() => selectPreset(presetPas)}
                class="touch-target p-3 bg-slate-100 hover:bg-brand-primary hover:text-white rounded-xl font-bold text-xs border border-slate-200 transition text-center"
              >
                Uang Pas (Rp {presetPas.toLocaleString("id-ID")})
              </button>
              <button
                type="button"
                on:click={() => selectPreset(preset50k)}
                class="touch-target p-3 bg-slate-100 hover:bg-brand-primary hover:text-white rounded-xl font-bold text-xs border border-slate-200 transition text-center"
              >
                Rp {preset50k.toLocaleString("id-ID")}
              </button>
              <button
                type="button"
                on:click={() => selectPreset(preset100k)}
                class="touch-target p-3 bg-slate-100 hover:bg-brand-primary hover:text-white rounded-xl font-bold text-xs border border-slate-200 transition text-center"
              >
                Rp {preset100k.toLocaleString("id-ID")}
              </button>
            </div>
          </div>

          <!-- Direct Input + Touch Numpad -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="space-y-2">
              <label
                class="block text-xs font-bold text-slate-500 uppercase"
                for="cash-input"
              >
                Nominal Diterima
              </label>
              <CurrencyInput bind:value={cashTendered} placeholder="0" />
            </div>

            <Numpad
              value={cashTendered.toString()}
              onChange={(val) => (cashTendered = parseInt(val, 10) || 0)}
            />
          </div>
        </div>

        <!-- QRIS TAB CONTENT -->
      {:else if !$featureFlags.enable_midtrans}
        <div
          class="p-6 bg-slate-50 rounded-2xl border border-slate-200 flex flex-col items-center justify-center text-center space-y-3"
        >
          <QrCode class="w-12 h-12 text-slate-300" />
          <h4 class="font-bold text-slate-700">Modul Midtrans QRIS Nonaktif</h4>
          <p class="text-xs text-slate-500 max-w-sm">
            Fitur pembayaran Midtrans QRIS belum diaktifkan di Setup Wizard /
            Konfigurasi Modul. Gunakan metode pembayaran Tunai (Cash Engine)
            atau aktifkan modul Midtrans di pengaturan.
          </p>
        </div>
      {:else}
        <div
          class="p-6 bg-slate-50 rounded-2xl border border-slate-200 flex flex-col items-center justify-center text-center space-y-4"
        >
          <div
            class="flex items-center gap-2 text-xs font-bold text-amber-700 bg-amber-100 px-3 py-1.5 rounded-full"
          >
            <Clock class="w-4 h-4 animate-spin" />
            Batas Waktu Pembayaran: {formattedTimer}
          </div>

          {#if qrisUrl}
            <div
              class="p-4 bg-white rounded-2xl shadow-md border border-slate-200"
            >
              <img
                src={qrisUrl}
                alt="Midtrans Dynamic QRIS"
                class="w-52 h-52 object-contain mx-auto"
              />
            </div>
          {:else}
            <div
              class="w-52 h-52 bg-slate-200 rounded-2xl flex items-center justify-center text-slate-400"
            >
              <RefreshCw class="w-8 h-8 animate-spin" />
            </div>
          {/if}

          <p class="text-xs text-slate-500 font-semibold max-w-xs">
            Tunjukkan kode QR di atas kepada pelanggan. Sistem memantau status
            secara otomatis.
          </p>
        </div>
      {/if}
    </div>
  {/if}

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={handleCloseModal}
      class="px-5 py-3 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
    >
      Batal
    </button>

    {#if !isPaidSuccess && activeMethod === "cash"}
      <button
        type="button"
        disabled={cashTendered < totalDue || isProcessing}
        on:click={handleCashCheckout}
        class="px-8 py-3 font-extrabold text-white bg-brand-primary hover:bg-sky-500 active:scale-95 disabled:opacity-50 rounded-xl shadow-lg shadow-sky-500/20 transition flex items-center gap-2"
      >
        {isProcessing ? "Memproses Transaksi..." : "PROSES PELUNASAN CASH"}
      </button>
    {/if}
  </svelte:fragment>
</Modal>

<ReceiptModal
  isOpen={isReceiptModalOpen}
  order={lastPaidOrder}
  payment={lastPaymentResult}
  onClose={() => (isReceiptModalOpen = false)}
/>
