<script lang="ts">
  import { goto } from '$app/navigation';
  import { AlertTriangle, ArrowLeft, ShieldAlert } from 'lucide-svelte';

  export let title: string = 'Modul Fitur Nonaktif';
  export let description: string = 'Fitur ini belum diaktifkan di Setup Wizard atau Konfigurasi Modul.';
  export let icon: any = ShieldAlert;
  export let actionText: string = 'Kembali ke Kasir POS';
  export let actionHref: string = '/pos';

  function handleAction() {
    if (actionHref) {
      goto(actionHref);
    }
  }
</script>

<div class="w-full h-full min-h-[450px] bg-gradient-to-b from-slate-950 via-slate-900 to-slate-950 p-6 flex flex-col items-center justify-center text-center">
  <div class="w-full max-w-md bg-slate-900/90 backdrop-blur-md rounded-3xl border border-slate-800 p-8 shadow-2xl space-y-6 flex flex-col items-center relative overflow-hidden">
    <!-- Ambient Glow Background -->
    <div class="absolute -top-24 -left-24 w-48 h-48 bg-brand-primary/10 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute -bottom-24 -right-24 w-48 h-48 bg-purple-500/10 rounded-full blur-3xl pointer-events-none"></div>

    <!-- Status Badge -->
    <span class="inline-flex items-center gap-1.5 px-3 py-1 bg-amber-500/10 border border-amber-500/20 text-amber-400 rounded-full text-xs font-bold uppercase tracking-wider">
      <AlertTriangle class="w-3.5 h-3.5" />
      Modul Nonaktif
    </span>

    <!-- Feature Icon Badge -->
    <div class="p-5 bg-gradient-to-br from-slate-800 to-slate-900 border border-slate-700/60 rounded-3xl text-brand-cyan shadow-xl shadow-sky-500/5 ring-1 ring-white/10">
      <svelte:component this={icon} class="w-12 h-12" />
    </div>

    <!-- Title & Description -->
    <div class="space-y-2">
      <h3 class="text-xl sm:text-2xl font-black text-white tracking-wide">
        {title}
      </h3>
      <p class="text-xs sm:text-sm text-slate-400 font-medium leading-relaxed max-w-sm">
        {description}
      </p>
    </div>

    <!-- Action Button -->
    <div class="pt-2 w-full">
      <button
        type="button"
        on:click={handleAction}
        class="w-full py-3.5 px-6 min-h-[48px] bg-brand-primary hover:bg-sky-500 active:scale-95 text-white font-extrabold text-sm rounded-2xl shadow-lg shadow-sky-500/20 flex items-center justify-center gap-2 transition-all duration-200"
      >
        <ArrowLeft class="w-4 h-4" />
        {actionText}
      </button>
    </div>
  </div>
</div>
