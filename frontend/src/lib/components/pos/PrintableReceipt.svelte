<script lang="ts">
  import { printableReceiptData } from '$lib/stores/printer';

  $: data = $printableReceiptData;
  $: outlet = data?.outlet;
  $: order = data?.order;
  $: payment = data?.payment;
  $: paperWidth = data?.paperWidth || 58;
</script>

{#if data && outlet && order}
  <div
    id="printable-receipt-area"
    class="font-mono text-black p-2 mx-auto bg-white"
    style="width: {paperWidth === 58 ? '58mm' : '80mm'}; max-width: {paperWidth === 58 ? '58mm' : '80mm'};"
  >
    <!-- Header -->
    <div class="text-center mb-2">
      <h1 class="text-base font-bold uppercase tracking-wider">{outlet.name || 'SMART POS RESTO'}</h1>
      {#if outlet.tagline}<p class="text-[10px]">{outlet.tagline}</p>{/if}
      {#if outlet.address}<p class="text-[10px]">{outlet.address}</p>{/if}
      {#if outlet.phone}<p class="text-[10px]">Telp: {outlet.phone}</p>{/if}
    </div>

    <!-- Prominent Queue Number Box -->
    <div class="text-center my-2 p-1.5 border-2 border-black rounded">
      <span class="text-[10px] uppercase font-bold block">NOMOR ANTREAN</span>
      <span class="text-2xl font-black">#{order.queue_number ? String(order.queue_number).padStart(3, '0') : (order.invoice_number?.slice(-4) || '---')}</span>
    </div>

    <div class="border-b border-dashed border-black my-2"></div>

    <!-- Order Meta -->
    <div class="space-y-0.5 text-[11px]">
      <div class="flex justify-between">
        <span>No. Struk:</span>
        <span class="font-bold">{order.invoice_number || order.order_number || `#${order.id?.slice(0, 8)}`}</span>
      </div>
      <div class="flex justify-between">
        <span>Tanggal:</span>
        <span>{new Date(order.created_at || Date.now()).toLocaleDateString('id-ID')} {new Date(order.created_at || Date.now()).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })}</span>
      </div>
      {#if order.customer_name}
        <div class="flex justify-between">
          <span>Pemesan:</span>
          <span class="font-bold">{order.customer_name}</span>
        </div>
      {/if}
      {#if order.table_name || order.table_number}
        <div class="flex justify-between">
          <span>Meja:</span>
          <span>{order.table_name || order.table_number}</span>
        </div>
      {/if}
      <div class="flex justify-between">
        <span>Layanan:</span>
        <span class="uppercase font-bold">
          {order.service_mode === 'dine_in' || order.service_type === 'dine_in' ? 'DINE-IN' :
           order.service_mode === 'take_away' || order.service_type === 'take_away' ? 'TAKEAWAY' :
           order.service_mode === 'delivery' || order.service_type === 'delivery' ? 'DELIVERY' :
           (order.service_mode || order.service_type || 'DINE-IN')}
        </span>
      </div>
    </div>

    <div class="border-b border-dashed border-black my-2"></div>

    <!-- Order Items -->
    <div class="space-y-1.5 text-[11px]">
      {#each order.items || [] as item}
        <div>
          <div class="font-bold">{item.product_name || item.product?.name || 'Item'}</div>
          <div class="flex justify-between pl-2 text-[10px]">
            <span>{item.quantity}x @ Rp {(item.price || item.unit_price || 0).toLocaleString('id-ID')}</span>
            <span>Rp {((item.price || item.unit_price || 0) * item.quantity).toLocaleString('id-ID')}</span>
          </div>
          {#if item.notes || item.item_notes}
            <div class="pl-2 italic text-[9px] text-gray-700">* {item.notes || item.item_notes}</div>
          {/if}
        </div>
      {/each}
    </div>

    <div class="border-b border-dashed border-black my-2"></div>

    <!-- Totals -->
    <div class="space-y-0.5 text-[11px]">
      <div class="flex justify-between">
        <span>Subtotal:</span>
        <span>Rp {(order.subtotal || 0).toLocaleString('id-ID')}</span>
      </div>
      {#if order.discount_amount && order.discount_amount > 0}
        <div class="flex justify-between">
          <span>Diskon:</span>
          <span>-Rp {order.discount_amount.toLocaleString('id-ID')}</span>
        </div>
      {/if}
      {#if order.service_amount && order.service_amount > 0}
        <div class="flex justify-between">
          <span>Service Charge:</span>
          <span>Rp {order.service_amount.toLocaleString('id-ID')}</span>
        </div>
      {/if}
      {#if order.tax_amount && order.tax_amount > 0}
        <div class="flex justify-between">
          <span>Pajak (PB1):</span>
          <span>Rp {order.tax_amount.toLocaleString('id-ID')}</span>
        </div>
      {/if}

      <div class="border-b border-black my-1"></div>

      <div class="flex justify-between font-bold text-xs pt-0.5">
        <span>TOTAL:</span>
        <span>Rp {(order.total_amount || 0).toLocaleString('id-ID')}</span>
      </div>

      {#if payment}
        <div class="flex justify-between pt-1">
          <span>Metode Bayar:</span>
          <span class="uppercase font-bold">{payment.payment_method || 'CASH'}</span>
        </div>
        {#if payment.payment_method === 'cash'}
          <div class="flex justify-between">
            <span>Tunai:</span>
            <span>Rp {(payment.cash_tendered || 0).toLocaleString('id-ID')}</span>
          </div>
          <div class="flex justify-between">
            <span>Kembali:</span>
            <span>Rp {(payment.change_amount || 0).toLocaleString('id-ID')}</span>
          </div>
        {/if}
      {/if}
    </div>

    <div class="border-b border-dashed border-black my-2"></div>

    <!-- Footer -->
    <div class="text-center text-[10px] space-y-1 mt-2">
      <p>{outlet.receipt_footer || 'Terima kasih atas kunjungan Anda!'}</p>
      <p class="text-[9px] text-gray-500">Powered by SPOS-E Engine</p>
    </div>
  </div>
{/if}

<style>
  @media screen {
    #printable-receipt-area {
      display: none !important;
    }
  }

  @media print {
    :global(body *) {
      visibility: hidden !important;
    }
    :global(#printable-receipt-area),
    :global(#printable-receipt-area *) {
      visibility: visible !important;
    }
    :global(#printable-receipt-area) {
      display: block !important;
      position: absolute !important;
      left: 0 !important;
      top: 0 !important;
      width: 100% !important;
      margin: 0 !important;
      padding: 4px !important;
      background: white !important;
      color: black !important;
    }
  }
</style>
