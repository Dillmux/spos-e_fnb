<script lang="ts">
  import { onMount } from 'svelte';
  import { reportsApi } from '$lib/api/reports';
  import { toast } from '$lib/stores/toast';
  import ReportCards from '$lib/components/admin/ReportCards.svelte';
  import type { DailyReport } from '$lib/types';
  import { BarChart3, Calendar, RefreshCw, Filter, CalendarDays, Download } from 'lucide-svelte';

  function getTodayString() {
    return new Date().toISOString().split('T')[0];
  }

  function getYesterdayString() {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    return d.toISOString().split('T')[0];
  }

  let filterMode: 'today' | 'yesterday' | 'single' | 'range' = 'today';
  let selectedDate: string = getTodayString();
  let startDate: string = getTodayString();
  let endDate: string = getTodayString();

  let reportData: DailyReport | null = null;
  let isLoading = true;

  async function loadReports() {
    isLoading = true;
    try {
      if (filterMode === 'today') {
        selectedDate = getTodayString();
      } else if (filterMode === 'yesterday') {
        selectedDate = getYesterdayString();
      }

      let salesRes: any;
      let profitRes: any;
      let topRes: any;

      if (filterMode === 'range') {
        [salesRes, profitRes, topRes] = await Promise.all([
          reportsApi.getDailySales(endDate),
          reportsApi.getGrossProfit(startDate, endDate),
          reportsApi.getTopSelling(undefined, startDate, endDate),
        ]);
      } else {
        [salesRes, profitRes, topRes] = await Promise.all([
          reportsApi.getDailySales(selectedDate),
          reportsApi.getGrossProfit(selectedDate, selectedDate),
          reportsApi.getTopSelling(selectedDate),
        ]);
      }

      const sales = salesRes.data || {};
      const profit = profitRes.data || {};
      const top = topRes.data || [];

      reportData = {
        date: selectedDate,
        total_revenue: profit.gross_revenue || sales.total_revenue || 0,
        total_cogs: profit.total_cogs || 0,
        gross_profit: profit.gross_profit || 0,
        total_transactions: sales.total_transactions || 0,
        payment_breakdowns: sales.payment_breakdowns || [],
        top_products: top.map((t: any) => ({
          name: t.name || t.product_name || 'Menu',
          qty: t.qty || t.quantity_sold || t.total_qty || 0,
          revenue: t.revenue || t.total_revenue || 0,
        })),
      };
    } catch (err: any) {
      toast.error(err.message || 'Gagal memuat data laporan');
    } finally {
      isLoading = false;
    }
  }

  function exportCSVReports() {
    if (!reportData) {
      toast.warning('Data laporan belum dimuat.');
      return;
    }

    const rows = [
      ['=== RINGKASAN LAPORAN PENJUALAN & LABA KOTOR ==='],
      ['Periode/Tanggal', filterMode === 'range' ? `${startDate} s/d ${endDate}` : selectedDate],
      ['Total Omset / Gross Revenue', reportData.total_revenue || 0],
      ['Total Modal / COGS', reportData.total_cogs || 0],
      ['Total Laba Kotor / Gross Profit', reportData.gross_profit || 0],
      ['Total Transaksi Selesai', reportData.total_transactions || 0],
      [],
      ['=== DETAIL PRODUK TERLARIS (TOP SELLING) ==='],
      ['Nama Menu', 'Jumlah Terjual (Qty)', 'Total Omset (Rp)'],
    ];

    if (reportData.top_products && reportData.top_products.length > 0) {
      for (const p of reportData.top_products) {
        rows.push([`"${p.name.replace(/"/g, '""')}"`, p.qty, p.revenue]);
      }
    } else {
      rows.push(['Tidak ada data penjualan produk', 0, 0]);
    }

    const csvContent = 'data:text/csv;charset=utf-8,\uFEFF' + rows.map((e) => e.join(',')).join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `Laporan_SPOS_${filterMode}_${selectedDate}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    toast.success('Laporan berhasil di-ekspor ke berkas CSV!');
  }

  function setPreset(mode: 'today' | 'yesterday') {
    filterMode = mode;
    if (mode === 'today') selectedDate = getTodayString();
    if (mode === 'yesterday') selectedDate = getYesterdayString();
    loadReports();
  }

  onMount(() => {
    loadReports();
  });
</script>

<div class="h-full bg-brand-surface p-6 flex flex-col space-y-6 overflow-y-auto">
  <!-- Top Bar & Title -->
  <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
    <div>
      <h2 class="text-2xl font-extrabold text-slate-900 flex items-center gap-3">
        <BarChart3 class="w-7 h-7 text-brand-primary" />
        Laporan Penjualan & Analisis Laba Kotor
      </h2>
      <p class="text-xs text-slate-500 font-semibold mt-1">
        Ringkasan performa finansial harian, perbandingan metode bayar, dan rasio HPP/COGS.
      </p>
    </div>

    <div class="flex items-center gap-2">
      <button
        type="button"
        on:click={exportCSVReports}
        class="px-4 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-bold rounded-xl shadow-sm flex items-center gap-2 transition cursor-pointer"
        title="Ekspor Laporan ke Berkas CSV"
      >
        <Download class="w-4 h-4" />
        Ekspor CSV
      </button>
      <button
        type="button"
        on:click={loadReports}
        class="p-2.5 bg-white hover:bg-slate-100 text-slate-700 rounded-xl border border-slate-200 transition"
        title="Muat Ulang Laporan"
      >
        <RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
      </button>
    </div>
  </div>

  <!-- Filter & Date Selector Workspace -->
  <div class="p-4 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-3">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2 text-xs font-extrabold text-slate-800 uppercase tracking-wider">
        <Filter class="w-4 h-4 text-brand-primary" />
        Filter Tanggal Laporan
      </div>
      <span class="text-xs text-slate-500 font-semibold">
        {filterMode === 'range' 
          ? `Menampilkan Periode: ${startDate} s/d ${endDate}`
          : `Tanggal Terpilih: ${new Date(selectedDate + 'T00:00:00').toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}`}
      </span>
    </div>

    <div class="flex flex-wrap items-center gap-3 pt-1">
      <!-- Quick Preset Buttons -->
      <div class="flex items-center gap-1.5 p-1 bg-slate-100 rounded-xl border border-slate-200">
        <button
          type="button"
          on:click={() => setPreset('today')}
          class="px-3 py-1.5 rounded-lg text-xs font-bold transition
          {filterMode === 'today' ? 'bg-brand-primary text-white shadow-sm' : 'text-slate-600 hover:text-slate-900'}"
        >
          Hari Ini
        </button>
        <button
          type="button"
          on:click={() => setPreset('yesterday')}
          class="px-3 py-1.5 rounded-lg text-xs font-bold transition
          {filterMode === 'yesterday' ? 'bg-brand-primary text-white shadow-sm' : 'text-slate-600 hover:text-slate-900'}"
        >
          Kemarin
        </button>
        <button
          type="button"
          on:click={() => (filterMode = 'single')}
          class="px-3 py-1.5 rounded-lg text-xs font-bold transition
          {filterMode === 'single' ? 'bg-brand-primary text-white shadow-sm' : 'text-slate-600 hover:text-slate-900'}"
        >
          Pilih Tanggal
        </button>
        <button
          type="button"
          on:click={() => (filterMode = 'range')}
          class="px-3 py-1.5 rounded-lg text-xs font-bold transition
          {filterMode === 'range' ? 'bg-brand-primary text-white shadow-sm' : 'text-slate-600 hover:text-slate-900'}"
        >
          Rentang Tanggal
        </button>
      </div>

      <!-- Single Date Picker Input -->
      {#if filterMode === 'single'}
        <div class="flex items-center gap-2 animate-in fade-in duration-200">
          <label for="report-single-date" class="text-xs font-bold text-slate-600 flex items-center gap-1">
            <CalendarDays class="w-4 h-4 text-brand-primary" /> Tanggal:
          </label>
          <input
            id="report-single-date"
            type="date"
            bind:value={selectedDate}
            on:change={loadReports}
            class="px-3 py-1.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-xs font-bold text-slate-800 outline-none"
          />
        </div>
      {/if}

      <!-- Date Range Picker Inputs -->
      {#if filterMode === 'range'}
        <div class="flex flex-wrap items-center gap-2 animate-in fade-in duration-200">
          <label for="report-start-date" class="text-xs font-bold text-slate-600 flex items-center gap-1">
            <CalendarDays class="w-4 h-4 text-brand-primary" /> Dari:
          </label>
          <input
            id="report-start-date"
            type="date"
            bind:value={startDate}
            on:change={loadReports}
            class="px-3 py-1.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-xs font-bold text-slate-800 outline-none"
          />

          <label for="report-end-date" class="text-xs font-bold text-slate-600">Sampai:</label>
          <input
            id="report-end-date"
            type="date"
            bind:value={endDate}
            on:change={loadReports}
            class="px-3 py-1.5 bg-slate-50 border border-slate-300 focus:border-brand-primary rounded-xl text-xs font-bold text-slate-800 outline-none"
          />

          <button
            type="button"
            on:click={loadReports}
            class="px-4 py-1.5 bg-brand-primary hover:bg-sky-500 text-white text-xs font-bold rounded-xl transition"
          >
            Terapkan Filter
          </button>
        </div>
      {/if}
    </div>
  </div>

  {#if isLoading}
    <div class="py-16 text-center text-slate-400 font-semibold">Mengkalkulasi Laporan Penjualan...</div>
  {:else if reportData}
    <ReportCards report={reportData} />
  {/if}
</div>

