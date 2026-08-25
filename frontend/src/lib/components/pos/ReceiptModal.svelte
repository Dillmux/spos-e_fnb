<script lang="ts">
  import { printer } from "$lib/stores/printer";
  import { config } from "$lib/stores/config";
  import Modal from "$lib/components/common/Modal.svelte";
  import { Printer, Check, ShoppingBag } from "lucide-svelte";
  import type { Order, Outlet, Payment } from "$lib/types";

  export let isOpen: boolean = false;
  export let onClose: () => void = () => {};
  export let order: Order | any = null;
  export let payment: Payment | any = null;
  export let outlet: Outlet | any = null;

  let paperWidth: 58 | 80 = 58;

  $: activeOutlet = outlet || $config.outlet || {
    id: "out-1",
    name: "SMART POS RESTO",
    tagline: "Cita Rasa Kuliner Terbaik",
    address: "Jl. Utama Resto No. 88",
    phone: "0812-3456-7890",
    receipt_footer: "Terima kasih atas kunjungan Anda!",
  };

  function handleSetPaperWidth(width: 58 | 80) {
    paperWidth = width;
    printer.setPaperWidth(width);
  }

  function handlePrintNow() {
    if (order) {
      printer.printReceipt(activeOutlet, order, payment);
    }
  }
</script>

<Modal isOpen={isOpen} title="Preview & Cetak Struk Transaksi" maxWidth="md" onClose={onClose}>
  {#if !order}
    <div class="py-8 text-center text-slate-500 font-semibold">
      <ShoppingBag class="w-12 h-12 mx-auto mb-2 text-slate-300" />
      Data struk transaksi tidak ditemukan
    </div>
  {:else}
    <div class="space-y-4 select-none">
      <!-- Printer Config Bar (58mm vs 80mm) -->
      <div class="flex items-center justify-between p-2.5 bg-slate-100 rounded-2xl border border-slate-200">
        <span class="text-xs font-bold text-slate-700 flex items-center gap-1.5">
          <Printer class="w-4 h-4 text-brand-primary" />
          Ukuran Kertas Thermal:
        </span>

        <div class="flex items-center gap-1 bg-white p-1 rounded-xl border border-slate-200 shadow-sm text-xs font-bold">
          <button
            type="button"
            on:click={() => handleSetPaperWidth(58)}
            class="px-3 py-1 rounded-lg transition {paperWidth === 58 ? 'bg-brand-primary text-white' : 'text-slate-600 hover:text-slate-900'}"
          >
            58 mm
          </button>
          <button
            type="button"
            on:click={() => handleSetPaperWidth(80)}
            class="px-3 py-1 rounded-lg transition {paperWidth === 80 ? 'bg-brand-primary text-white' : 'text-slate-600 hover:text-slate-900'}"
          >
            80 mm
          </button>
        </div>
      </div>

      <!-- Thermal Paper Visual Box -->
      <div class="bg-slate-200/70 p-4 rounded-2xl border border-slate-300 flex justify-center">
        <div
          class="bg-white p-4 shadow-xl border border-slate-300 text-black font-mono text-xs space-y-2 rounded-sm transition-all duration-300"
          style="width: {paperWidth === 58 ? '260px' : '340px'};"
        >
          <!-- Header -->
          <div class="text-center space-y-0.5">
            <h4 class="font-black text-sm uppercase tracking-wider">{activeOutlet.name || 'SMART POS RESTO'}</h4>
            {#if activeOutlet.tagline}<p class="text-[10px] text-slate-600">{activeOutlet.tagline}</p>{/if}
            {#if activeOutlet.address}<p class="text-[10px] text-slate-600">{activeOutlet.address}</p>{/if}
            {#if activeOutlet.phone}<p class="text-[10px] text-slate-600">Telp: {activeOutlet.phone}</p>{/if}
          </div>

          <!-- Prominent Queue Number Box -->
          <div class="text-center my-2 p-1.5 border-2 border-black rounded bg-slate-50">
            <span class="text-[10px] uppercase font-bold block text-slate-700">NOMOR ANTREAN</span>
            <span class="text-2xl font-black text-black">#{order.queue_number ? String(order.queue_number).padStart(3, '0') : (order.invoice_number?.slice(-4) || '---')}</span>
          </div>

          <div class="border-b border-dashed border-slate-400 my-2"></div>

          <!-- Order Meta -->
          <div class="space-y-0.5 text-[11px]">
            <div class="flex justify-between">
              <span class="text-slate-600">No. Struk:</span>
              <span class="font-bold">{order.invoice_number || order.order_number || `#${order.id?.slice(0, 8)}`}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-600">Waktu:</span>
              <span>{new Date(order.created_at || Date.now()).toLocaleDateString('id-ID')} {new Date(order.created_at || Date.now()).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}</span>
            </div>
            {#if order.customer_name}
              <div class="flex justify-between">
                <span class="text-slate-600">Pemesan:</span>
                <span class="font-bold">{order.customer_name}</span>
              </div>
            {/if}
            {#if order.table_name || order.table_number}
              <div class="flex justify-between">
                <span class="text-slate-600">Meja:</span>
                <span>{order.table_name || order.table_number}</span>
              </div>
            {/if}
            <div class="flex justify-between">
              <span class="text-slate-600">Layanan:</span>
              <span class="uppercase font-bold">
                {order.service_mode === 'dine_in' || order.service_type === 'dine_in' ? 'DINE-IN' :
                 order.service_mode === 'take_away' || order.service_type === 'take_away' ? 'TAKEAWAY' :
                 order.service_mode === 'delivery' || order.service_type === 'delivery' ? 'DELIVERY' :
                 (order.service_mode || order.service_type || 'DINE-IN')}
              </span>
            </div>
          </div>

          <div class="border-b border-dashed border-slate-400 my-2"></div>

          <!-- Order Items -->
          <div class="space-y-1 text-[11px]">
            {#each order.items || [] as item}
              <div>
                <div class="font-bold">{item.product_name || item.product?.name || 'Item'}</div>
                <div class="flex justify-between pl-2 text-[10px] text-slate-700">
                  <span>{item.quantity}x @ Rp {(item.price || item.unit_price || 0).toLocaleString('id-ID')}</span>
                  <span>Rp {((item.price || item.unit_price || 0) * item.quantity).toLocaleString('id-ID')}</span>
                </div>
                {#if item.notes || item.item_notes}
                  <div class="pl-2 italic text-[9px] text-slate-500">* {item.notes || item.item_notes}</div>
                {/if}
              </div>
            {/each}
          </div>

          <div class="border-b border-dashed border-slate-400 my-2"></div>

          <!-- Totals -->
          <div class="space-y-0.5 text-[11px]">
            <div class="flex justify-between text-slate-700">
              <span>Subtotal:</span>
              <span>Rp {(order.subtotal || 0).toLocaleString('id-ID')}</span>
            </div>
            {#if order.discount_amount && order.discount_amount > 0}
              <div class="flex justify-between text-slate-700">
                <span>Diskon:</span>
                <span>-Rp {order.discount_amount.toLocaleString('id-ID')}</span>
              </div>
            {/if}
            {#if order.service_amount && order.service_amount > 0}
              <div class="flex justify-between text-slate-700">
                <span>Service Charge:</span>
                <span>Rp {order.service_amount.toLocaleString('id-ID')}</span>
              </div>
            {/if}
            {#if order.tax_amount && order.tax_amount > 0}
              <div class="flex justify-between text-slate-700">
                <span>Pajak (PB1):</span>
                <span>Rp {order.tax_amount.toLocaleString('id-ID')}</span>
              </div>
            {/if}

            <div class="border-b border-slate-800 my-1"></div>

            <div class="flex justify-between font-extrabold text-xs text-black pt-0.5">
              <span>TOTAL:</span>
              <span>Rp {(order.total_amount || 0).toLocaleString('id-ID')}</span>
            </div>

            {#if payment}
              <div class="flex justify-between pt-1 text-slate-700">
                <span>Metode Bayar:</span>
                <span class="uppercase font-bold text-black">{payment.payment_method || 'CASH'}</span>
              </div>
              {#if payment.payment_method === 'cash'}
                <div class="flex justify-between text-slate-700">
                  <span>Tunai:</span>
                  <span>Rp {(payment.cash_tendered || 0).toLocaleString('id-ID')}</span>
                </div>
                <div class="flex justify-between text-slate-700">
                  <span>Kembali:</span>
                  <span>Rp {(payment.change_amount || 0).toLocaleString('id-ID')}</span>
                </div>
              {/if}
            {/if}
          </div>

          <div class="border-b border-dashed border-slate-400 my-2"></div>

          <!-- Footer -->
          <div class="text-center text-[10px] space-y-1 text-slate-600 mt-2">
            <p>{activeOutlet.receipt_footer || 'Terima kasih atas kunjungan Anda!'}</p>
            <p class="text-[9px] text-slate-400">Powered by SPOS-E Engine</p>
          </div>
        </div>
      </div>
    </div>
  {/if}

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={onClose}
      class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
    >
      Tutup
    </button>

    {#if order}
      <button
        type="button"
        on:click={handlePrintNow}
        class="px-6 py-2.5 font-extrabold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-lg shadow-sky-500/20 transition flex items-center gap-2"
      >
        <Printer class="w-4 h-4" />
        Cetak Struk Sekarang
      </button>
    {/if}
  </svelte:fragment>
</Modal>
