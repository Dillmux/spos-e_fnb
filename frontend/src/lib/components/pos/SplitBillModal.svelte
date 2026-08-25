<script lang="ts">
  import { cart } from '$lib/stores/cart';
  import { toast } from '$lib/stores/toast';
  import { Split, ArrowRightLeft, CheckCircle2 } from 'lucide-svelte';
  import Modal from '$lib/components/common/Modal.svelte';

  export let isOpen: boolean = false;
  export let onClose: () => void = () => {};

  $: itemsA = $cart.items.filter((i) => i.split_group !== 'B');
  $: itemsB = $cart.items.filter((i) => i.split_group === 'B');

  $: totalA = itemsA.reduce((sum, i) => sum + i.product.price * i.quantity, 0);
  $: totalB = itemsB.reduce((sum, i) => sum + i.product.price * i.quantity, 0);

  function moveToGroup(itemId: string, targetGroup: 'A' | 'B') {
    cart.assignSplitGroup(itemId, targetGroup);
  }

  function handleConfirmSplit() {
    if (itemsB.length === 0) {
      toast.warning('Pindahkan minimal 1 menu ke Bil B untuk melakukan split bill');
      return;
    }
    toast.success('Split Bill berhasil disiapkan! Pembayaran akan diproses secara terpisah.');
    onClose();
  }
</script>

<Modal {isOpen} title="Interactive Split Bill Workspace" maxWidth="4xl" {onClose}>
  <div class="space-y-4">
    <p class="text-xs text-slate-500 font-semibold">
      Klik tombol panah untuk memindahkan item menu antara Struk A dan Struk B secara instan.
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Panel Bill A -->
      <div class="bg-slate-50 p-4 rounded-2xl border-2 border-brand-primary/40 space-y-3 flex flex-col justify-between h-96">
        <div class="space-y-3 overflow-y-auto pr-1">
          <div class="flex items-center justify-between border-b border-slate-200 pb-2">
            <h4 class="font-extrabold text-slate-800 text-base">Struk A (Original)</h4>
            <span class="text-xs font-bold px-2 py-0.5 bg-brand-primary/10 text-brand-primary rounded-md">
              {itemsA.length} Item
            </span>
          </div>

          {#each itemsA as item}
            <div class="p-3 bg-white rounded-xl border border-slate-200 flex items-center justify-between shadow-sm">
              <div>
                <p class="font-bold text-slate-800 text-sm">{item.product.name}</p>
                <p class="text-xs text-slate-500 font-semibold">
                  {item.quantity}x @ Rp {item.product.price.toLocaleString('id-ID')}
                </p>
              </div>
              <button
                type="button"
                on:click={() => moveToGroup(item.id, 'B')}
                class="touch-target p-2 bg-slate-100 hover:bg-brand-primary hover:text-white rounded-xl text-slate-600 transition"
                title="Pindah ke Struk B"
              >
                <ArrowRightLeft class="w-4 h-4" />
              </button>
            </div>
          {/each}
        </div>

        <div class="pt-3 border-t border-slate-200 flex items-center justify-between">
          <span class="text-xs font-bold text-slate-500">Total Struk A:</span>
          <span class="text-lg font-bold text-brand-primary">Rp {totalA.toLocaleString('id-ID')}</span>
        </div>
      </div>

      <!-- Panel Bill B -->
      <div class="bg-slate-50 p-4 rounded-2xl border-2 border-brand-cyan/60 space-y-3 flex flex-col justify-between h-96">
        <div class="space-y-3 overflow-y-auto pr-1">
          <div class="flex items-center justify-between border-b border-slate-200 pb-2">
            <h4 class="font-extrabold text-slate-800 text-base">Struk B (Terpisah)</h4>
            <span class="text-xs font-bold px-2 py-0.5 bg-brand-cyan/20 text-slate-800 rounded-md">
              {itemsB.length} Item
            </span>
          </div>

          {#if itemsB.length === 0}
            <div class="h-48 flex flex-col items-center justify-center text-slate-400 text-xs font-semibold">
              <Split class="w-8 h-8 mb-1 text-slate-300" />
              Pindahkan item di sini untuk Struk B
            </div>
          {:else}
            {#each itemsB as item}
              <div class="p-3 bg-white rounded-xl border border-slate-200 flex items-center justify-between shadow-sm">
                <div>
                  <p class="font-bold text-slate-800 text-sm">{item.product.name}</p>
                  <p class="text-xs text-slate-500 font-semibold">
                    {item.quantity}x @ Rp {item.product.price.toLocaleString('id-ID')}
                  </p>
                </div>
                <button
                  type="button"
                  on:click={() => moveToGroup(item.id, 'A')}
                  class="touch-target p-2 bg-slate-100 hover:bg-brand-primary hover:text-white rounded-xl text-slate-600 transition"
                  title="Kembalikan ke Struk A"
                >
                  <ArrowRightLeft class="w-4 h-4" />
                </button>
              </div>
            {/each}
          {/if}
        </div>

        <div class="pt-3 border-t border-slate-200 flex items-center justify-between">
          <span class="text-xs font-bold text-slate-500">Total Struk B:</span>
          <span class="text-lg font-bold text-slate-800">Rp {totalB.toLocaleString('id-ID')}</span>
        </div>
      </div>
    </div>
  </div>

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={onClose}
      class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
    >
      Batal
    </button>
    <button
      type="button"
      on:click={handleConfirmSplit}
      class="px-6 py-2.5 font-bold text-white bg-brand-primary hover:bg-sky-500 rounded-xl shadow-md flex items-center gap-2 transition"
    >
      <CheckCircle2 class="w-4 h-4" />
      Konfirmasi Split Bill
    </button>
  </svelte:fragment>
</Modal>
