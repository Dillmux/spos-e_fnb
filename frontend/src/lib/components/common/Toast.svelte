<script lang="ts">
  import { toast } from "$lib/stores/toast";
  import {
    CheckCircle2,
    AlertCircle,
    AlertTriangle,
    Info,
    X,
  } from "lucide-svelte";

  const icons = {
    success: CheckCircle2,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info,
  };

  const bgClasses = {
    success: "bg-emerald-600 text-white shadow-emerald-500/20",
    error: "bg-brand-danger text-white shadow-red-500/20",
    warning: "bg-amber-500 text-white shadow-amber-500/20",
    info: "bg-brand-primary text-white shadow-blue-500/20",
  };
</script>

<div
  class="fixed bottom-5 right-5 z-[100] flex flex-col gap-2 max-w-sm w-full pointer-events-none"
>
  {#each $toast as t (t.id)}
    <div
      class="pointer-events-auto flex items-center justify-between p-4 rounded-xl shadow-lg border border-white/20 backdrop-blur-md transition-all duration-300 animate-in slide-in-from-bottom-5 {bgClasses[
        t.type
      ]}"
    >
      <div class="flex items-center gap-3">
        <svelte:component this={icons[t.type]} class="w-5 h-5 flex-shrink-0" />
        <p class="text-sm font-semibold leading-tight">{t.text}</p>
      </div>
      <button
        type="button"
        on:click={() => toast.dismiss(t.id)}
        class="p-1 hover:bg-white/20 rounded-lg text-white/80 hover:text-white transition ml-2"
      >
        <X class="w-4 h-4" />
      </button>
    </div>
  {/each}
</div>
