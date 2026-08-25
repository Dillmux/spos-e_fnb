<script lang="ts">
  import Modal from './Modal.svelte';
  import { AlertTriangle, Trash2, CheckCircle, ShieldAlert } from 'lucide-svelte';

  export let isOpen: boolean = false;
  export let title: string = 'Konfirmasi Hapus Data';
  export let message: string = 'Apakah Anda yakin ingin menghapus data ini? Tindakan ini tidak dapat dibatalkan.';
  export let confirmText: string = 'Ya, Hapus Data';
  export let cancelText: string = 'Batal';
  export let isDanger: boolean = true;
  export let isLoading: boolean = false;
  export let onConfirm: () => void = () => {};
  export let onClose: () => void = () => {};
</script>

<Modal {isOpen} {title} maxWidth="md" {onClose}>
  <div class="py-2 flex items-start gap-4">
    <div class="p-3 rounded-2xl flex-shrink-0 {isDanger ? 'bg-red-100 text-red-600' : 'bg-amber-100 text-amber-600'}">
      {#if isDanger}
        <Trash2 class="w-6 h-6" />
      {:else}
        <AlertTriangle class="w-6 h-6" />
      {/if}
    </div>

    <div class="space-y-1">
      <h4 class="font-extrabold text-slate-900 text-base">{title}</h4>
      <p class="text-xs text-slate-600 font-semibold leading-relaxed">
        {message}
      </p>
    </div>
  </div>

  <svelte:fragment slot="footer">
    <button
      type="button"
      on:click={onClose}
      class="px-5 py-2.5 font-bold text-slate-600 bg-white border border-slate-300 rounded-xl hover:bg-slate-100 transition"
    >
      {cancelText}
    </button>

    <button
      type="button"
      disabled={isLoading}
      on:click={onConfirm}
      class="px-6 py-2.5 font-bold text-white rounded-xl shadow-md transition flex items-center gap-2
      {isDanger ? 'bg-red-600 hover:bg-red-700' : 'bg-brand-primary hover:bg-sky-500'}"
    >
      {#if isLoading}
        Memproses...
      {:else}
        {confirmText}
      {/if}
    </button>
  </svelte:fragment>
</Modal>
