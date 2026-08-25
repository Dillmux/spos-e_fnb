<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { Order } from "$lib/types";
  import { Clock, CheckCircle2, ChefHat, Flame } from "lucide-svelte";

  export let order: Order;
  export let onUpdateItemStatus: (orderItemId: string, status: string) => void;

  let elapsedMinutes = 0;
  let interval: any;

  function calculateElapsed() {
    if (!order.created_at) return;
    const diffMs = Date.now() - new Date(order.created_at).getTime();
    elapsedMinutes = Math.floor(diffMs / 60000);
  }

  onMount(() => {
    calculateElapsed();
    interval = setInterval(calculateElapsed, 10000);
  });

  onDestroy(() => {
    clearInterval(interval);
  });

  function getNextStatus(curr: string): string {
    if (curr === "queued" || curr === "pending") return "cooking";
    if (curr === "cooking" || curr === "preparing") return "ready";
    if (curr === "ready") return "served";
    return "served";
  }

  function getTableLabel(ord: Order): string {
    if (ord.Table?.table_number) return `Meja: ${ord.Table.table_number}`;
    if ((ord as any).table?.table_number)
      return `Meja: ${(ord as any).table.table_number}`;
    if (ord.table_name) return `Meja: ${ord.table_name}`;
    const mode = ord.service_type || ord.service_mode || "DINE_IN";
    return mode.toUpperCase();
  }

  function getItemName(item: any): string {
    return (
      item.Product?.name || item.product?.name || item.product_name || "Produk"
    );
  }
</script>

<div
  class="bg-slate-900 border border-slate-800 rounded-2xl p-5 text-white space-y-4 shadow-xl flex flex-col justify-between h-full"
>
  <!-- Ticket Header -->
  <div class="flex items-center justify-between border-b border-slate-800 pb-3">
    <div>
      <div class="flex items-center gap-2 mb-0.5">
        <span class="px-2 py-0.5 bg-brand-cyan/20 text-brand-cyan text-xs font-mono font-black rounded border border-brand-cyan/40">
          ANTREAN #{order.queue_number ? String(order.queue_number).padStart(3, '0') : (order.invoice_number?.slice(-4) || '---')}
        </span>
      </div>
      <h4 class="font-extrabold text-lg text-white tracking-wide">
        {order.invoice_number ||
          order.order_number ||
          `#${order.id.slice(0, 6)}`}
      </h4>
      <p class="text-xs text-slate-400 font-semibold">
        {getTableLabel(order)}
      </p>
    </div>

    <!-- Timer Badge -->
    <div
      class="flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-bold
      {elapsedMinutes > 15
        ? 'bg-red-500/20 text-red-400 border border-red-500/40 animate-pulse'
        : 'bg-slate-800 text-slate-300'}"
    >
      <Clock class="w-4 h-4" />
      {elapsedMinutes} mnt lalu
    </div>
  </div>

  <!-- Items list -->
  <div class="space-y-3 flex-1 overflow-y-auto pr-1">
    {#each order.items || order.OrderItems || [] as item (item.id)}
      <div
        class="p-3 bg-slate-950/80 rounded-xl border border-slate-800 flex items-start justify-between"
      >
        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <span class="font-bold text-lg text-white">{item.quantity}x</span>
            <span class="font-bold text-base text-slate-200"
              >{getItemName(item)}</span
            >
          </div>
          {#if item.item_notes || item.notes}
            <p
              class="text-xs italic text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/30"
            >
              Note: {item.item_notes || item.notes}
            </p>
          {/if}
        </div>

        <!-- Kitchen Status Action -->
        <button
          type="button"
          on:click={() =>
            onUpdateItemStatus(item.id, getNextStatus(item.kitchen_status))}
          class="touch-target px-3 py-1.5 rounded-xl text-xs font-bold transition flex items-center gap-1
          {item.kitchen_status === 'queued' || item.kitchen_status === 'pending'
            ? 'bg-amber-500/20 text-amber-300 hover:bg-amber-500/40'
            : ''}
          {item.kitchen_status === 'cooking' ||
          item.kitchen_status === 'preparing'
            ? 'bg-blue-500/20 text-blue-300 hover:bg-blue-500/40'
            : ''}
          {item.kitchen_status === 'ready'
            ? 'bg-emerald-500/20 text-emerald-300 hover:bg-emerald-500/40'
            : ''}
          {item.kitchen_status === 'served'
            ? 'bg-slate-800 text-slate-400'
            : ''}"
        >
          {#if item.kitchen_status === "queued" || item.kitchen_status === "pending"}
            <ChefHat class="w-4 h-4" /> Masak
          {:else if item.kitchen_status === "cooking" || item.kitchen_status === "preparing"}
            <Flame class="w-4 h-4" /> Siap
          {:else if item.kitchen_status === "ready"}
            <CheckCircle2 class="w-4 h-4" /> Sajikan
          {:else}
            <CheckCircle2 class="w-4 h-4" /> Tersajikan
          {/if}
        </button>
      </div>
    {/each}
  </div>
</div>
