<script lang="ts">
  import { onMount } from "svelte";
  import { posApi } from "$lib/api/pos";
  import { toast } from "$lib/stores/toast";
  import { auth as authStore } from "$lib/stores/auth";
  import Modal from "$lib/components/common/Modal.svelte";
  import ReceiptModal from "$lib/components/pos/ReceiptModal.svelte";
  import PaymentModal from "$lib/components/pos/PaymentModal.svelte";
  import type { Order } from "$lib/types";
  import {
    History,
    Calendar,
    RefreshCw,
    Filter,
    Search,
    Printer,
    Eye,
    Receipt,
    ShoppingBag,
    DollarSign,
    CheckCircle2,
    XCircle,
    Clock,
    User,
    Utensils,
    CalendarDays,
    CreditCard,
    FileText,
    QrCode,
  } from "lucide-svelte";

  function getTodayString() {
    return new Date().toISOString().split("T")[0];
  }

  function getYesterdayString() {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    return d.toISOString().split("T")[0];
  }

  let filterMode: "today" | "yesterday" | "single" | "range" = "today";
  let selectedDate: string = getTodayString();
  let startDate: string = getTodayString();
  let endDate: string = getTodayString();
  let selectedStatus: string = "all";
  let searchQuery: string = "";

  let orders: Order[] = [];
  let isLoading = true;

  // Modals state
  let selectedDetailOrder: Order | null = null;
  let isDetailModalOpen = false;

  let selectedReceiptOrder: Order | null = null;
  let isReceiptModalOpen = false;

  let selectedPendingOrder: Order | null = null;
  let isPaymentModalOpen = false;

  function openPaymentModalForOrder(order: Order) {
    selectedPendingOrder = order;
    isPaymentModalOpen = true;
  }

  let isVoidReasonModalOpen = false;
  let voidTargetOrder: Order | null = null;
  let voidReason = "";
  let isVoiding = false;
  let verifiedSupervisorPin = "";

  function getOrderAgeMinutes(createdAt: string): number {
    const created = new Date(createdAt).getTime();
    const now = new Date().getTime();
    return Math.max(0, (now - created) / (1000 * 60));
  }

  function getVoidEligibility(order: Order): {
    allowed: boolean;
    remainingMinutes: number;
    reason?: string;
  } {
    if (order.status === "cancelled" || order.status === "void") {
      return {
        allowed: false,
        remainingMinutes: 0,
        reason: "Pesanan sudah dibatalkan",
      };
    }
    const ageMin = getOrderAgeMinutes(order.created_at);
    if (ageMin > 30) {
      return {
        allowed: false,
        remainingMinutes: 0,
        reason: `Telah berusia ${Math.floor(ageMin)} menit (Maksimal 30 menit)`,
      };
    }
    const remaining = Math.max(0, Math.ceil(30 - ageMin));
    return { allowed: true, remainingMinutes: remaining };
  }

  async function openVoidModal(order: Order) {
    const eligibility = getVoidEligibility(order);
    if (!eligibility.allowed) {
      toast.warning(eligibility.reason || "Pesanan tidak dapat dibatalkan.");
      return;
    }

    try {
      verifiedSupervisorPin = await authStore.requestSupervisorPin(
        "Otorisasi Pembatalan Pesanan",
      );
      voidTargetOrder = order;
      voidReason = "";
      isVoidReasonModalOpen = true;
    } catch {
      // Pin entry cancelled by user
    }
  }

  async function handleConfirmVoidOrder() {
    if (!voidTargetOrder || !verifiedSupervisorPin) return;
    isVoiding = true;
    try {
      const res = await posApi.voidOrder(
        voidTargetOrder.id,
        verifiedSupervisorPin,
        voidReason || "Dibatalkan dari Riwayat Pemesanan",
      );
      if (res.error) {
        throw new Error(res.error.message || "Gagal membatalkan pesanan");
      }
      toast.success(
        `Pesanan #${voidTargetOrder.invoice_number || voidTargetOrder.id.slice(0, 8)} berhasil dibatalkan!`,
      );
      isVoidReasonModalOpen = false;
      if (isDetailModalOpen) isDetailModalOpen = false;
      await loadOrders();
    } catch (err: any) {
      toast.error(err.message || "Gagal membatalkan pesanan");
    } finally {
      isVoiding = false;
    }
  }

  async function loadOrders() {
    isLoading = true;
    try {
      if (filterMode === "today") {
        selectedDate = getTodayString();
      } else if (filterMode === "yesterday") {
        selectedDate = getYesterdayString();
      }

      let res: any;
      if (filterMode === "range") {
        res = await posApi.getOrders({
          startDate,
          endDate,
          status: selectedStatus === "all" ? undefined : selectedStatus,
        });
      } else {
        res = await posApi.getOrders({
          date: selectedDate,
          status: selectedStatus === "all" ? undefined : selectedStatus,
        });
      }

      orders = (res?.data || (Array.isArray(res) ? res : [])) as Order[];
    } catch (err: any) {
      orders = [];
      toast.error(err.message || "Gagal memuat data histori pesanan");
    } finally {
      isLoading = false;
    }
  }

  function setPreset(mode: "today" | "yesterday") {
    filterMode = mode;
    if (mode === "today") selectedDate = getTodayString();
    if (mode === "yesterday") selectedDate = getYesterdayString();
    loadOrders();
  }

  function openDetailModal(order: Order) {
    selectedDetailOrder = order;
    isDetailModalOpen = true;
  }

  function openReceiptModal(order: Order) {
    selectedReceiptOrder = order;
    isReceiptModalOpen = true;
  }

  $: filteredOrders = orders.filter((o) => {
    if (!searchQuery.trim()) return true;
    const q = searchQuery.toLowerCase();
    const inv = (o.invoice_number || o.order_number || "").toLowerCase();
    const cust = (o.customer_name || "").toLowerCase();
    const tbl = (o.table_name || o.Table?.name || o.table_number || "")
      .toString()
      .toLowerCase();
    const idStr = o.id.toLowerCase();
    return (
      inv.includes(q) ||
      cust.includes(q) ||
      tbl.includes(q) ||
      idStr.includes(q)
    );
  });

  // Metrics
  $: totalOrdersCount = filteredOrders.length;
  $: paidOrders = filteredOrders.filter(
    (o) => o.status === "paid" || o.status === "completed",
  );
  $: totalRevenue = paidOrders.reduce(
    (sum, o) => sum + (o.total_amount || 0),
    0,
  );
  $: avgOrderValue =
    paidOrders.length > 0 ? Math.round(totalRevenue / paidOrders.length) : 0;
  $: voidOrdersCount = filteredOrders.filter(
    (o) => o.status === "cancelled" || o.status === "void",
  ).length;

  function formatPrice(amount: number = 0) {
    return `Rp ${amount.toLocaleString("id-ID")}`;
  }

  function formatDateTime(dateStr: string) {
    if (!dateStr) return "-";
    const d = new Date(dateStr);
    return d.toLocaleString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function getStatusBadge(status: string) {
    switch (status) {
      case "paid":
      case "completed":
        return {
          label: "LUNAS / SELESAI",
          bg: "bg-emerald-100 text-emerald-800 border-emerald-300",
        };
      case "preparing":
      case "queued":
      case "cooking":
        return {
          label: "DIPROSES",
          bg: "bg-amber-100 text-amber-800 border-amber-300",
        };
      case "ready":
        return {
          label: "SIAP",
          bg: "bg-cyan-100 text-cyan-800 border-cyan-300",
        };
      case "open":
      case "pending":
      case "pending_payment":
        return {
          label: "PENDING",
          bg: "bg-blue-100 text-blue-800 border-blue-300",
        };
      case "cancelled":
      case "void":
        return {
          label: "BATAL / VOID",
          bg: "bg-rose-100 text-rose-800 border-rose-300",
        };
      default:
        return {
          label: status.toUpperCase(),
          bg: "bg-slate-100 text-slate-800 border-slate-300",
        };
    }
  }

  function getServiceTypeLabel(type: string) {
    switch (type) {
      case "dine_in":
        return "Dine In";
      case "take_away":
        return "Take Away";
      case "delivery":
        return "Delivery";
      default:
        return type || "Dine In";
    }
  }

  onMount(() => {
    loadOrders();
  });
</script>

<div
  class="h-full bg-brand-surface p-6 flex flex-col space-y-6 overflow-y-auto select-none"
>
  <!-- Top Bar & Title -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h2
        class="text-2xl font-extrabold text-slate-900 flex items-center gap-3"
      >
        <History class="w-7 h-7 text-brand-primary" />
        Histori Pemesanan & Struk Transaksi
      </h2>
      <p class="text-xs text-slate-500 font-semibold mt-1">
        Pantau riwayat seluruh transaksi, verifikasi detail pesanan, dan cetak
        ulang struk pembayaran.
      </p>
    </div>

    <div class="flex items-center gap-2">
      <button
        type="button"
        on:click={loadOrders}
        class="p-2.5 bg-white hover:bg-slate-100 text-slate-700 rounded-xl border border-slate-200 transition shadow-sm"
        title="Muat Ulang Data Pesanan"
      >
        <RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
      </button>
    </div>
  </div>

  <!-- Summary Cards Workspace -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
    <div
      class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4"
    >
      <div class="p-3 bg-brand-primary/10 text-brand-primary rounded-xl">
        <ShoppingBag class="w-6 h-6" />
      </div>
      <div>
        <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">
          Total Pesanan
        </p>
        <h3 class="text-xl font-extrabold text-slate-900">
          {totalOrdersCount} Transaksi
        </h3>
      </div>
    </div>

    <div
      class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4"
    >
      <div class="p-3 bg-emerald-50 text-emerald-600 rounded-xl">
        <DollarSign class="w-6 h-6" />
      </div>
      <div>
        <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">
          Total Omzet
        </p>
        <h3 class="text-xl font-extrabold text-emerald-600">
          {formatPrice(totalRevenue)}
        </h3>
      </div>
    </div>

    <div
      class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4"
    >
      <div class="p-3 bg-cyan-50 text-cyan-600 rounded-xl">
        <CreditCard class="w-6 h-6" />
      </div>
      <div>
        <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">
          Rata-rata Order
        </p>
        <h3 class="text-xl font-extrabold text-slate-900">
          {formatPrice(avgOrderValue)}
        </h3>
      </div>
    </div>

    <div
      class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm flex items-center gap-4"
    >
      <div class="p-3 bg-rose-50 text-rose-600 rounded-xl">
        <XCircle class="w-6 h-6" />
      </div>
      <div>
        <p class="text-xs font-bold text-slate-400 uppercase tracking-wider">
          Transaksi Void/Batal
        </p>
        <h3 class="text-xl font-extrabold text-rose-600">
          {voidOrdersCount} Order
        </h3>
      </div>
    </div>
  </div>

  <!-- Filter & Date Selector Bar -->
  <div
    class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-4"
  >
    <div
      class="flex flex-col md:flex-row md:items-center justify-between gap-3"
    >
      <div
        class="flex items-center gap-2 text-xs font-extrabold text-slate-800 uppercase tracking-wider"
      >
        <Filter class="w-4 h-4 text-brand-primary" />
        Filter & Pencarian Pemesanan
      </div>
      <span class="text-xs text-slate-500 font-semibold">
        {filterMode === "range"
          ? `Periode: ${startDate} s/d ${endDate}`
          : `Tanggal Terpilih: ${new Date(selectedDate + "T00:00:00").toLocaleDateString("id-ID", { weekday: "long", day: "numeric", month: "long", year: "numeric" })}`}
      </span>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-3">
      <!-- Date Presets & Date Inputs -->
      <div class="flex flex-wrap items-center gap-2">
        <div
          class="flex items-center gap-1 p-1 bg-slate-100 rounded-xl border border-slate-200"
        >
          <button
            type="button"
            on:click={() => setPreset("today")}
            class="px-3 py-1.5 rounded-lg text-xs font-bold transition {filterMode ===
            'today'
              ? 'bg-brand-primary text-white shadow-sm'
              : 'text-slate-600 hover:text-slate-900'}"
          >
            Hari Ini
          </button>
          <button
            type="button"
            on:click={() => setPreset("yesterday")}
            class="px-3 py-1.5 rounded-lg text-xs font-bold transition {filterMode ===
            'yesterday'
              ? 'bg-brand-primary text-white shadow-sm'
              : 'text-slate-600 hover:text-slate-900'}"
          >
            Kemarin
          </button>
          <button
            type="button"
            on:click={() => {
              filterMode = "single";
              loadOrders();
            }}
            class="px-3 py-1.5 rounded-lg text-xs font-bold transition {filterMode ===
            'single'
              ? 'bg-brand-primary text-white shadow-sm'
              : 'text-slate-600 hover:text-slate-900'}"
          >
            Tanggal Spesifik
          </button>
          <button
            type="button"
            on:click={() => {
              filterMode = "range";
              loadOrders();
            }}
            class="px-3 py-1.5 rounded-lg text-xs font-bold transition {filterMode ===
            'range'
              ? 'bg-brand-primary text-white shadow-sm'
              : 'text-slate-600 hover:text-slate-900'}"
          >
            Rentang Tanggal
          </button>
        </div>

        {#if filterMode === "single"}
          <div
            class="flex items-center gap-2 bg-slate-50 p-1.5 rounded-xl border border-slate-200 text-xs font-bold text-slate-700"
          >
            <Calendar class="w-4 h-4 text-brand-primary" />
            <input
              type="date"
              bind:value={selectedDate}
              on:change={loadOrders}
              class="bg-transparent border-0 outline-none text-slate-800 font-bold"
            />
          </div>
        {:else if filterMode === "range"}
          <div
            class="flex items-center gap-2 bg-slate-50 p-1.5 rounded-xl border border-slate-200 text-xs font-bold text-slate-700"
          >
            <CalendarDays class="w-4 h-4 text-brand-primary" />
            <input
              type="date"
              bind:value={startDate}
              on:change={loadOrders}
              class="bg-transparent border-0 outline-none text-slate-800 font-bold"
            />
            <span class="text-slate-400">s/d</span>
            <input
              type="date"
              bind:value={endDate}
              on:change={loadOrders}
              class="bg-transparent border-0 outline-none text-slate-800 font-bold"
            />
          </div>
        {/if}
      </div>

      <!-- Search & Status Selectors -->
      <div class="flex flex-wrap items-center gap-2 w-full md:w-auto">
        <select
          bind:value={selectedStatus}
          on:change={loadOrders}
          class="px-3 py-2 bg-slate-100 border border-slate-200 rounded-xl text-xs font-bold text-slate-700 outline-none focus:border-brand-primary"
        >
          <option value="all">Semua Status</option>
          <option value="paid">Lunas / Selesai (Paid)</option>
          <option value="preparing">Diproses (Preparing)</option>
          <option value="ready">Siap (Ready)</option>
          <option value="pending">Pending</option>
          <option value="void">Batal / Void</option>
        </select>

        <div class="relative flex-1 md:w-64">
          <Search class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" />
          <input
            type="text"
            bind:value={searchQuery}
            placeholder="Cari No. Struk, Pelanggan, Meja..."
            class="w-full pl-9 pr-3 py-2 bg-slate-100 border border-slate-200 rounded-xl text-xs font-semibold text-slate-800 outline-none focus:border-brand-primary placeholder:text-slate-400"
          />
        </div>
      </div>
    </div>
  </div>

  <!-- Orders Data Table -->
  <div
    class="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden flex-1 flex flex-col"
  >
    {#if isLoading}
      <div class="p-16 text-center text-slate-400 font-semibold">
        <RefreshCw
          class="w-8 h-8 animate-spin mx-auto mb-2 text-brand-primary"
        />
        Memuat Histori Transaksi...
      </div>
    {:else if filteredOrders.length === 0}
      <div class="p-16 text-center text-slate-400 font-semibold space-y-2">
        <Receipt class="w-12 h-12 mx-auto text-slate-300" />
        <p class="text-sm text-slate-600 font-extrabold">
          Tidak Ada Transaksi Ditemukan
        </p>
        <p class="text-xs text-slate-400">
          Tidak ada histori pesanan yang cocok dengan filter tanggal atau kata
          kunci pencarian ini.
        </p>
      </div>
    {:else}
      <div class="overflow-x-auto flex-1">
        <table class="w-full text-left text-xs border-collapse">
          <thead>
            <tr
              class="bg-slate-100/80 border-b border-slate-200 text-slate-600 font-extrabold uppercase tracking-wider"
            >
              <th class="py-3.5 px-4">No. Struk / Order ID</th>
              <th class="py-3.5 px-4">Waktu Transaksi</th>
              <th class="py-3.5 px-4">Tipe & Meja</th>
              <th class="py-3.5 px-4">Kasir / Pelanggan</th>
              <th class="py-3.5 px-4">Ringkasan Item</th>
              <th class="py-3.5 px-4 text-right">Total Bayar</th>
              <th class="py-3.5 px-4 text-center">Status</th>
              <th class="py-3.5 px-4 text-center">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 font-semibold text-slate-700">
            {#each filteredOrders as order (order.id)}
              {@const badge = getStatusBadge(order.status)}
              {@const items = order.items || order.OrderItems || []}
              {@const voidInfo = getVoidEligibility(order)}
              <tr class="hover:bg-slate-50/80 transition">
                <td class="py-3 px-4 font-mono font-bold text-slate-900">
                  {order.invoice_number ||
                    order.order_number ||
                    `#${order.id.slice(0, 8)}`}
                </td>
                <td class="py-3 px-4 text-slate-500 whitespace-nowrap">
                  {formatDateTime(order.created_at)}
                </td>
                <td class="py-3 px-4">
                  <div class="font-bold text-slate-800">
                    {getServiceTypeLabel(
                      order.service_type || order.service_mode,
                    )}
                  </div>
                  <div class="text-[11px] text-slate-500">
                    {order.table_name ||
                      order.Table?.name ||
                      (order.table_number ? `Meja ${order.table_number}` : "-")}
                  </div>
                </td>
                <td class="py-3 px-4">
                  <div class="font-bold text-slate-800">
                    {order.customer_name || "Umum"}
                  </div>
                  {#if order.User?.name}
                    <div class="text-[11px] text-slate-400">
                      Kasir: {order.User.name}
                    </div>
                  {/if}
                </td>
                <td class="py-3 px-4 max-w-xs truncate text-slate-600">
                  {#if items.length > 0}
                    {items
                      .map(
                        (i) =>
                          `${i.quantity}x ${i.product_name || i.Product?.name || "Menu"}`,
                      )
                      .join(", ")}
                  {:else}
                    <span class="text-slate-400 italic"
                      >Tidak ada rincian item</span
                    >
                  {/if}
                </td>
                <td
                  class="py-3 px-4 text-right font-black text-slate-900 whitespace-nowrap"
                >
                  {formatPrice(order.total_amount)}
                </td>
                <td class="py-3 px-4 text-center">
                  {#if order.status === "open" || order.status === "pending" || order.status === "pending_payment"}
                    <button
                      type="button"
                      on:click={() => openPaymentModalForOrder(order)}
                      class="px-2.5 py-1 rounded-full text-[10px] font-black border tracking-wider uppercase inline-flex items-center gap-1.5 bg-amber-100 text-amber-800 border-amber-300 hover:bg-purple-600 hover:text-white hover:border-purple-600 hover:scale-105 transition shadow-xs cursor-pointer"
                      title="Status Pending. Klik untuk Melanjutkan Pembayaran / QRIS"
                    >
                      PENDING (BAYAR)
                    </button>
                  {:else}
                    <span
                      class="px-2.5 py-1 rounded-full text-[10px] font-black border tracking-wider uppercase inline-block {badge.bg}"
                    >
                      {badge.label}
                    </span>
                  {/if}
                </td>
                <td class="py-3 px-4">
                  <div class="flex items-center justify-center gap-1.5">
                    <button
                      type="button"
                      on:click={() => openDetailModal(order)}
                      class="px-2.5 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-[11px] font-bold transition flex items-center gap-1"
                      title="Cek Detail Order"
                    >
                      <Eye class="w-3.5 h-3.5 text-brand-primary" /> Detail
                    </button>
                    <button
                      type="button"
                      on:click={() => openReceiptModal(order)}
                      class="px-2.5 py-1.5 bg-brand-primary/10 hover:bg-brand-primary/20 text-brand-primary rounded-lg text-[11px] font-bold transition flex items-center gap-1"
                      title="Cetak Ulang Struk"
                    >
                      <Printer class="w-3.5 h-3.5" /> Struk
                    </button>
                    {#if order.status !== "cancelled" && order.status !== "void"}
                      <button
                        type="button"
                        disabled={!voidInfo.allowed}
                        on:click={() => openVoidModal(order)}
                        class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold transition flex items-center gap-1
                        {voidInfo.allowed
                          ? 'bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 shadow-sm'
                          : 'bg-slate-100 text-slate-400 cursor-not-allowed opacity-60'}"
                        title={voidInfo.allowed
                          ? `Batalkan Order (Sisa Waktu: ${voidInfo.remainingMinutes} min)`
                          : voidInfo.reason}
                      >
                        <XCircle class="w-3.5 h-3.5 text-rose-600" /> Batal
                      </button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>

<!-- ORDER DETAIL INSPECTOR MODAL -->
<Modal
  isOpen={isDetailModalOpen}
  title="Rincian Detail Transaksi"
  maxWidth="lg"
  onClose={() => (isDetailModalOpen = false)}
>
  {#if selectedDetailOrder}
    {@const items =
      selectedDetailOrder.items || selectedDetailOrder.OrderItems || []}
    {@const payments = selectedDetailOrder.Payments || []}
    {@const voidInfo = getVoidEligibility(selectedDetailOrder)}
    <div class="space-y-4 text-xs select-none">
      <!-- Order Header Info Card -->
      <div
        class="p-3.5 bg-slate-100 rounded-xl border border-slate-200 space-y-2"
      >
        <div class="flex items-center justify-between">
          <div>
            <p class="text-[10px] uppercase font-bold text-slate-400">
              Nomor Transaksi
            </p>
            <h4 class="text-sm font-black text-slate-900 font-mono">
              {selectedDetailOrder.invoice_number ||
                selectedDetailOrder.order_number ||
                `#${selectedDetailOrder.id}`}
            </h4>
          </div>
          <span
            class="px-2.5 py-1 rounded-full text-[10px] font-black border uppercase {getStatusBadge(
              selectedDetailOrder.status,
            ).bg}"
          >
            {getStatusBadge(selectedDetailOrder.status).label}
          </span>
        </div>

        <!-- Void Time Constraint Banner -->
        {#if selectedDetailOrder.status !== "cancelled" && selectedDetailOrder.status !== "void"}
          <div
            class="p-2.5 rounded-xl border flex items-center justify-between text-xs mt-2
            {voidInfo.allowed
              ? 'bg-amber-50 border-amber-200 text-amber-900'
              : 'bg-slate-100 border-slate-200 text-slate-500'}"
          >
            <div>
              <span class="font-extrabold block"
                >Batas Waktu Pembatalan (30 Menit)</span
              >
              <span class="text-[11px]">
                {#if voidInfo.allowed}
                  ⏱️ Sisa Waktu Pembatalan: <strong
                    class="text-amber-900 font-extrabold"
                    >{voidInfo.remainingMinutes} Menit</strong
                  >
                {:else}
                  ⚠️ {voidInfo.reason}
                {/if}
              </span>
            </div>

            {#if voidInfo.allowed}
              <button
                type="button"
                on:click={() => {
                  if (selectedDetailOrder) openVoidModal(selectedDetailOrder);
                }}
                class="px-3 py-1.5 bg-rose-600 hover:bg-rose-700 text-white rounded-lg font-bold text-xs shadow-md transition flex items-center gap-1"
              >
                <XCircle class="w-3.5 h-3.5" /> Batalkan Pesanan
              </button>
            {/if}
          </div>
        {/if}

        <div
          class="grid grid-cols-2 gap-2 text-slate-600 pt-2 border-t border-slate-200/80"
        >
          <div>
            <span class="text-slate-400">Waktu:</span>
            <strong class="text-slate-800"
              >{formatDateTime(selectedDetailOrder.created_at)}</strong
            >
          </div>
          <div>
            <span class="text-slate-400">Tipe Pesanan:</span>
            <strong class="text-slate-800"
              >{getServiceTypeLabel(
                selectedDetailOrder.service_type ||
                  selectedDetailOrder.service_mode,
              )}</strong
            >
          </div>
          <div>
            <span class="text-slate-400">Pelanggan:</span>
            <strong class="text-slate-800"
              >{selectedDetailOrder.customer_name || "Umum"}</strong
            >
          </div>
          <div>
            <span class="text-slate-400">Meja:</span>
            <strong class="text-slate-800"
              >{selectedDetailOrder.table_name ||
                selectedDetailOrder.Table?.name ||
                (selectedDetailOrder.table_number
                  ? `Meja ${selectedDetailOrder.table_number}`
                  : "-")}</strong
            >
          </div>
        </div>
      </div>

      <!-- Items List -->
      <div class="space-y-2">
        <h5
          class="font-extrabold text-slate-800 uppercase tracking-wider text-[11px]"
        >
          Daftar Menu Dipesan
        </h5>
        <div
          class="bg-white border border-slate-200 rounded-xl overflow-hidden divide-y divide-slate-100"
        >
          {#each items as item}
            {@const price = item.unit_price || item.price || 0}
            <div class="p-3 flex items-center justify-between">
              <div>
                <p class="font-bold text-slate-900">
                  {item.product_name || item.Product?.name || "Menu"}
                </p>
                <p class="text-[11px] text-slate-500">
                  {item.quantity} x {formatPrice(price)}
                </p>
                {#if item.item_notes || item.notes}
                  <p class="text-[10px] text-amber-600 italic">
                    Catatan: {item.item_notes || item.notes}
                  </p>
                {/if}
              </div>
              <div class="font-bold text-slate-900">
                {formatPrice(item.subtotal || price * item.quantity)}
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Financial Calculation breakdown -->
      <div
        class="p-3.5 bg-slate-50 rounded-xl border border-slate-200 space-y-1.5 font-semibold text-slate-700"
      >
        <div class="flex justify-between">
          <span>Subtotal Menu:</span>
          <span>{formatPrice(selectedDetailOrder.subtotal)}</span>
        </div>
        {#if selectedDetailOrder.discount_amount > 0}
          <div class="flex justify-between text-rose-600">
            <span>Diskon:</span>
            <span>- {formatPrice(selectedDetailOrder.discount_amount)}</span>
          </div>
        {/if}
        {#if selectedDetailOrder.tax_amount > 0}
          <div class="flex justify-between text-slate-500">
            <span>Pajak (PB1):</span>
            <span>{formatPrice(selectedDetailOrder.tax_amount)}</span>
          </div>
        {/if}
        {#if selectedDetailOrder.service_amount > 0}
          <div class="flex justify-between text-slate-500">
            <span>Biaya Layanan:</span>
            <span>{formatPrice(selectedDetailOrder.service_amount)}</span>
          </div>
        {/if}
        <div
          class="flex justify-between pt-2 border-t border-slate-200 font-extrabold text-sm text-slate-900"
        >
          <span>Total Pembayaran:</span>
          <span class="text-brand-primary"
            >{formatPrice(selectedDetailOrder.total_amount)}</span
          >
        </div>
      </div>

      <!-- Payment details if available -->
      {#if payments.length > 0}
        <div
          class="p-3 bg-emerald-50/70 border border-emerald-200 rounded-xl space-y-1 text-slate-800"
        >
          <p
            class="text-[11px] font-extrabold text-emerald-800 uppercase tracking-wider flex items-center gap-1.5"
          >
            <CheckCircle2 class="w-3.5 h-3.5 text-emerald-600" /> Informasi Pembayaran
          </p>
          {#each payments as pay}
            <div class="flex justify-between text-[11px]">
              <span
                >Metode: <strong class="uppercase">{pay.payment_method}</strong
                ></span
              >
              <span
                >Dibayar: <strong>{formatPrice(pay.amount_paid)}</strong>
                (Kembalian: {formatPrice(pay.change_amount)})</span
              >
            </div>
          {/each}
        </div>
      {/if}

      <!-- Bottom Actions -->
      <div class="pt-3 flex justify-end gap-2 border-t border-slate-200">
        <button
          type="button"
          on:click={() => (isDetailModalOpen = false)}
          class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold rounded-xl transition"
        >
          Tutup
        </button>
        <button
          type="button"
          on:click={() => {
            if (selectedDetailOrder) {
              const orderToPrint = selectedDetailOrder;
              isDetailModalOpen = false;
              openReceiptModal(orderToPrint);
            }
          }}
          class="px-4 py-2 bg-brand-primary hover:bg-brand-primary/90 text-white font-bold rounded-xl transition flex items-center gap-2 shadow-sm"
        >
          <Printer class="w-4 h-4" /> Cetak Struk
        </button>
      </div>
    </div>
  {/if}
</Modal>

<!-- THERMAL RECEIPT MODAL FOR REPRINTING -->
<ReceiptModal
  isOpen={isReceiptModalOpen}
  order={selectedReceiptOrder}
  onClose={() => (isReceiptModalOpen = false)}
/>

<!-- VOID ORDER REASON MODAL -->
<Modal
  isOpen={isVoidReasonModalOpen}
  title="Konfirmasi Pembatalan Pesanan"
  maxWidth="md"
  onClose={() => (isVoidReasonModalOpen = false)}
>
  {#if voidTargetOrder}
    <div class="space-y-4 text-xs select-none">
      <div
        class="p-3 bg-rose-50 border border-rose-200 text-rose-900 rounded-xl space-y-1.5"
      >
        <p class="font-extrabold text-sm">
          Apakah Anda yakin ingin membatalkan pesanan ini?
        </p>
        <p>
          No. Struk / Order ID: <strong
            >{voidTargetOrder.invoice_number ||
              voidTargetOrder.order_number ||
              `#${voidTargetOrder.id.slice(0, 8)}`}</strong
          >
        </p>
        <p>
          Total Transaksi: <strong
            >{formatPrice(voidTargetOrder.total_amount)}</strong
          >
        </p>
        <p
          class="text-[11px] text-rose-700 font-semibold pt-1 border-t border-rose-200"
        >
          ⚠️ Pembatalan ini akan mengosongkan meja dan memulihkan stok bahan
          baku ke inventaris secara otomatis.
        </p>
      </div>

      <div class="space-y-1.5">
        <label for="void-reason-input" class="block font-bold text-slate-700">
          Alasan Pembatalan / Void Pesanan:
        </label>
        <textarea
          id="void-reason-input"
          bind:value={voidReason}
          placeholder="Contoh: Pelanggan membatalkan pesanan / salah pilih menu..."
          class="w-full p-3 bg-slate-50 border border-slate-300 rounded-xl text-xs focus:ring-2 focus:ring-rose-500 focus:outline-none h-24"
        ></textarea>
      </div>
    </div>
  {/if}

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={() => (isVoidReasonModalOpen = false)}
      class="px-4 py-2 bg-slate-200 text-slate-700 font-bold text-xs rounded-xl hover:bg-slate-300 transition"
    >
      Batal
    </button>
    <button
      type="button"
      disabled={isVoiding}
      on:click={handleConfirmVoidOrder}
      class="px-5 py-2 bg-rose-600 hover:bg-rose-700 text-white font-extrabold text-xs rounded-xl disabled:opacity-50 flex items-center gap-1.5 shadow-md transition"
    >
      {isVoiding ? "Memproses..." : "PROSES PEMBATALAN"}
    </button>
  </svelte:fragment>
</Modal>

<PaymentModal
  isOpen={isPaymentModalOpen}
  existingOrder={selectedPendingOrder}
  onClose={() => {
    isPaymentModalOpen = false;
    selectedPendingOrder = null;
    loadOrders();
  }}
/>
