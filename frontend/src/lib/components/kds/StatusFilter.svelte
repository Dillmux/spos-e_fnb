<script lang="ts">
  export let selectedStatus: string = 'all';
  export let onChangeStatus: (status: string) => void = () => {};
</script>

<div class="flex items-center gap-1.5 bg-slate-900 p-1.5 rounded-2xl border border-slate-800 overflow-x-auto no-scrollbar max-w-full">
  <button
    type="button"
    on:click={() => onChangeStatus('all')}
    class="px-4 py-2.5 rounded-xl text-xs font-bold transition whitespace-nowrap min-h-[44px] flex items-center
    {selectedStatus === 'all' ? 'bg-brand-primary text-white shadow-md font-extrabold' : 'text-slate-400 hover:text-white'}"
  >
    Semua Antrean
  </button>
  <button
    type="button"
    on:click={() => onChangeStatus('pending')}
    class="px-4 py-2.5 rounded-xl text-xs font-bold transition whitespace-nowrap min-h-[44px] flex items-center
    {selectedStatus === 'pending' ? 'bg-amber-500 text-white shadow-md font-extrabold' : 'text-slate-400 hover:text-white'}"
  >
    Belum Dimasak
  </button>
  <button
    type="button"
    on:click={() => onChangeStatus('preparing')}
    class="px-4 py-2.5 rounded-xl text-xs font-bold transition whitespace-nowrap min-h-[44px] flex items-center
    {selectedStatus === 'preparing' ? 'bg-blue-500 text-white shadow-md font-extrabold' : 'text-slate-400 hover:text-white'}"
  >
    Proses Dimasak
  </button>
  <button
    type="button"
    on:click={() => onChangeStatus('ready')}
    class="px-4 py-2.5 rounded-xl text-xs font-bold transition whitespace-nowrap min-h-[44px] flex items-center
    {selectedStatus === 'ready' ? 'bg-emerald-500 text-white shadow-md font-extrabold' : 'text-slate-400 hover:text-white'}"
  >
    Siap Saji
  </button>
</div>
