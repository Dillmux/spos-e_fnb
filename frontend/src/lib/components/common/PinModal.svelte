<script lang="ts">
  import { activePinRequest } from '$lib/stores/auth';
  import { authApi } from '$lib/api/auth';
  import { toast } from '$lib/stores/toast';
  import { ShieldAlert, X, Delete } from 'lucide-svelte';

  let pin = '';
  let isSubmitting = false;
  let errorMessage = '';

  $: request = $activePinRequest;

  function addDigit(digit: string) {
    if (pin.length < 6) {
      pin += digit;
      errorMessage = '';
    }
  }

  function backspace() {
    pin = pin.slice(0, -1);
    errorMessage = '';
  }

  function clearPin() {
    pin = '';
    errorMessage = '';
  }

  async function handleVerify() {
    if (pin.length !== 6) {
      errorMessage = 'PIN supervisor harus 6 digit';
      return;
    }

    isSubmitting = true;
    errorMessage = '';

    try {
      const res = await authApi.verifyPin(pin);
      const supervisorName = res.data?.verified_by || res.data?.supervisor_name || 'Supervisor';
      if (res.success && (res.data?.verified_by || res.data?.valid)) {
        toast.success(`Otorisasi Berhasil: ${supervisorName}`);
        request?.onSuccess(pin);
        pin = '';
      } else {
        errorMessage = 'PIN Supervisor Tidak Valid';
        pin = '';
      }
    } catch (err: any) {
      errorMessage = err.message || 'Gagal memverifikasi PIN';
      pin = '';
    } finally {
      isSubmitting = false;
    }
  }

  function handleCancel() {
    pin = '';
    errorMessage = '';
    request?.onCancel();
  }

  function portal(node: HTMLElement) {
    document.body.appendChild(node);
    return {
      destroy() {
        if (node.parentNode) {
          node.parentNode.removeChild(node);
        }
      },
    };
  }
</script>

{#if request}
  <div use:portal class="fixed inset-0 z-[9999] flex items-center justify-center bg-slate-900/75 backdrop-blur-sm p-4 animate-in fade-in duration-200">
    <div class="w-full max-w-md bg-white rounded-2xl shadow-2xl overflow-hidden border border-slate-100 flex flex-col">
      <!-- Header -->
      <div class="bg-brand-dark p-5 text-white flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-amber-500/20 text-amber-400 rounded-lg">
            <ShieldAlert class="w-6 h-6" />
          </div>
          <div>
            <h3 class="font-bold text-lg leading-tight">Otorisasi Supervisor</h3>
            <p class="text-xs text-slate-300">{request.title}</p>
          </div>
        </div>
        <button
          type="button"
          on:click={handleCancel}
          class="p-2 hover:bg-slate-800 rounded-full text-slate-400 hover:text-white transition"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- PIN Mask Display -->
      <div class="p-6 text-center flex flex-col items-center bg-slate-50 border-b border-slate-100">
        <p class="text-sm font-semibold text-slate-600 mb-3">Masukkan 6-Digit PIN Supervisor</p>
        <div class="flex justify-center gap-3 my-2">
          {#each Array(6) as _, i}
            <div
              class="w-10 h-12 rounded-xl border-2 flex items-center justify-center text-xl font-bold transition-all shadow-sm
              {i < pin.length ? 'border-brand-primary bg-brand-primary/10 text-brand-primary' : 'border-slate-300 bg-white'}"
            >
              {i < pin.length ? '•' : ''}
            </div>
          {/each}
        </div>
        {#if errorMessage}
          <p class="text-xs font-semibold text-brand-danger mt-2 animate-bounce">{errorMessage}</p>
        {/if}
      </div>

      <!-- Numpad Keypad -->
      <div class="p-5 grid grid-cols-3 gap-3 bg-white">
        {#each ['1', '2', '3', '4', '5', '6', '7', '8', '9'] as digit}
          <button
            type="button"
            on:click={() => addDigit(digit)}
            class="touch-target h-14 text-xl font-bold text-slate-800 bg-slate-50 hover:bg-slate-100 active:bg-brand-primary active:text-white rounded-xl border border-slate-200 shadow-sm transition"
          >
            {digit}
          </button>
        {/each}
        <button
          type="button"
          on:click={clearPin}
          class="touch-target h-14 font-semibold text-slate-500 bg-slate-100 hover:bg-slate-200 rounded-xl transition"
        >
          C
        </button>
        <button
          type="button"
          on:click={() => addDigit('0')}
          class="touch-target h-14 text-xl font-bold text-slate-800 bg-slate-50 hover:bg-slate-100 active:bg-brand-primary active:text-white rounded-xl border border-slate-200 shadow-sm transition"
        >
          0
        </button>
        <button
          type="button"
          on:click={backspace}
          class="touch-target h-14 flex items-center justify-center text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-xl transition"
        >
          <Delete class="w-6 h-6" />
        </button>
      </div>

      <!-- Actions -->
      <div class="p-4 bg-slate-50 border-t border-slate-100 flex gap-3">
        <button
          type="button"
          on:click={handleCancel}
          class="flex-1 py-3 font-semibold text-slate-700 bg-white hover:bg-slate-100 rounded-xl border border-slate-200 transition"
        >
          Batal
        </button>
        <button
          type="button"
          disabled={pin.length !== 6 || isSubmitting}
          on:click={handleVerify}
          class="flex-1 py-3 font-bold text-white bg-brand-primary hover:opacity-90 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed rounded-xl transition shadow-md"
        >
          {isSubmitting ? 'Verifikasi...' : 'Verifikasi PIN'}
        </button>
      </div>
    </div>
  </div>
{/if}
