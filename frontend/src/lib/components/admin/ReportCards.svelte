<script lang="ts">
  import type { DailyReport } from '$lib/types';
  import { DollarSign, TrendingUp, ShoppingBag, PieChart, Award } from 'lucide-svelte';

  export let report: DailyReport;

  $: totalRevenue = report?.total_revenue ?? report?.total_sales ?? 0;
  $: totalOrders = report?.total_orders ?? report?.total_transactions ?? 0;
  $: grossProfit = report?.gross_profit ?? 0;
  $: totalCogs = report?.total_cogs ?? 0;

  $: cashSales = report?.sales_by_payment?.cash ?? 
    (report?.payment_breakdowns?.find((p) => p.payment_method === 'cash')?.total_amount ?? 0);

  $: qrisSales = report?.sales_by_payment?.qris ?? 
    (report?.payment_breakdowns?.filter((p) => p.payment_method.includes('qris') || p.payment_method.includes('midtrans'))
      ?.reduce((acc, curr) => acc + curr.total_amount, 0) ?? 0);
  function getTopProductInfo(item: any) {
    return {
      name: item.product_name || item.name || '',
      qty: item.total_sold ?? item.qty ?? 0,
      revenue: item.total_revenue ?? item.revenue ?? 0,
    };
  }
</script>

<div class="space-y-6">
  <!-- Key Metrics Cards -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
    <!-- Total Sales -->
    <div class="p-5 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-2">
      <div class="flex items-center justify-between text-slate-500">
        <span class="text-xs font-bold uppercase tracking-wider">Total Omzet Penjualan</span>
        <div class="p-2 bg-blue-100 text-brand-primary rounded-xl">
          <DollarSign class="w-5 h-5" />
        </div>
      </div>
      <h3 class="text-2xl font-extrabold text-slate-900">
        Rp {totalRevenue.toLocaleString('id-ID')}
      </h3>
      <p class="text-xs text-slate-500 font-semibold">{totalOrders} Transaksi Lunas</p>
    </div>

    <!-- Gross Profit -->
    <div class="p-5 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-2">
      <div class="flex items-center justify-between text-slate-500">
        <span class="text-xs font-bold uppercase tracking-wider">Laba Kotor (Gross Profit)</span>
        <div class="p-2 bg-emerald-100 text-emerald-600 rounded-xl">
          <TrendingUp class="w-5 h-5" />
        </div>
      </div>
      <h3 class="text-2xl font-extrabold text-emerald-600">
        Rp {grossProfit.toLocaleString('id-ID')}
      </h3>
      <p class="text-xs text-slate-500 font-semibold">Omzet - Total HPP/COGS</p>
    </div>

    <!-- Total COGS -->
    <div class="p-5 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-2">
      <div class="flex items-center justify-between text-slate-500">
        <span class="text-xs font-bold uppercase tracking-wider">Total Modal Bahan (COGS)</span>
        <div class="p-2 bg-amber-100 text-amber-600 rounded-xl">
          <ShoppingBag class="w-5 h-5" />
        </div>
      </div>
      <h3 class="text-2xl font-extrabold text-slate-800">
        Rp {totalCogs.toLocaleString('id-ID')}
      </h3>
      <p class="text-xs text-slate-500 font-semibold">Berdasarkan Kalkulasi BOM</p>
    </div>

    <!-- Payment Split -->
    <div class="p-5 bg-white rounded-2xl border border-slate-200 shadow-sm space-y-2">
      <div class="flex items-center justify-between text-slate-500">
        <span class="text-xs font-bold uppercase tracking-wider">Rasio Metode Bayar</span>
        <div class="p-2 bg-purple-100 text-purple-600 rounded-xl">
          <PieChart class="w-5 h-5" />
        </div>
      </div>
      <div class="space-y-1 text-xs font-bold pt-1">
        <div class="flex justify-between text-slate-700">
          <span>Tunai (Cash):</span>
          <span>Rp {cashSales.toLocaleString('id-ID')}</span>
        </div>
        <div class="flex justify-between text-purple-700">
          <span>QRIS / Dynamic:</span>
          <span>Rp {qrisSales.toLocaleString('id-ID')}</span>
        </div>
      </div>
    </div>
  </div>

  <!-- Top Products Table -->
  {#if report?.top_products && report.top_products.length > 0}
    <div class="bg-white p-5 rounded-2xl border border-slate-200 shadow-sm space-y-4">
      <h4 class="font-bold text-slate-800 text-base flex items-center gap-2">
        <Award class="w-5 h-5 text-amber-500" />
        Produk Terlaris Harian (Rank Sales)
      </h4>

      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 text-xs font-bold text-slate-500 uppercase">
              <th class="py-3 px-4">Nama Produk</th>
              <th class="py-3 px-4 text-center">Jumlah Terjual</th>
              <th class="py-3 px-4 text-right">Total Pendapatan</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 text-sm font-semibold text-slate-800">
            {#each report.top_products as item}
              {@const info = getTopProductInfo(item)}
              <tr class="hover:bg-slate-50">
                <td class="py-3 px-4">{info.name}</td>
                <td class="py-3 px-4 text-center font-bold text-brand-primary">{info.qty} porsi</td>
                <td class="py-3 px-4 text-right">Rp {info.revenue.toLocaleString('id-ID')}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>


