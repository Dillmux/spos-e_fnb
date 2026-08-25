<script lang="ts">
  import { X } from "lucide-svelte";

  export let isOpen: boolean = false;
  export let title: string = "";
  export let maxWidth: "sm" | "md" | "lg" | "xl" | "2xl" | "4xl" = "md";
  export let onClose: () => void = () => {};

  const widthClasses = {
    sm: "max-w-sm",
    md: "max-w-md",
    lg: "max-w-lg",
    xl: "max-w-xl",
    "2xl": "max-w-2xl",
    "4xl": "max-w-4xl",
  };

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

{#if isOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    use:portal
    on:click={onClose}
    class="fixed inset-0 z-[75] flex items-center justify-center p-2.5 sm:p-4 bg-slate-900/75 backdrop-blur-sm animate-in fade-in duration-200 overflow-y-auto"
  >
    <div
      on:click|stopPropagation
      class="w-full {widthClasses[
        maxWidth
      ]} bg-white rounded-2xl sm:rounded-3xl shadow-2xl overflow-hidden border border-slate-100 flex flex-col max-h-[92vh] my-auto"
    >
      <!-- Header -->
      <div
        class="px-4 sm:px-6 py-3.5 bg-brand-surface border-b border-slate-200 flex items-center justify-between"
      >
        <h3
          class="font-extrabold text-base sm:text-lg text-slate-800 leading-tight"
        >
          {title}
        </h3>
        <button
          type="button"
          on:click={onClose}
          class="p-2 hover:bg-slate-200 rounded-xl text-slate-500 hover:text-slate-800 transition min-h-[48px] min-w-[48px] flex items-center justify-center"
          aria-label="Tutup Dialog"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Body -->
      <div class="p-4 sm:p-6 overflow-y-auto flex-1">
        <slot />
      </div>

      <!-- Footer slot -->
      {#if $$slots.footer}
        <div
          class="px-4 sm:px-6 py-3.5 bg-slate-50 border-t border-slate-200 flex items-center justify-end gap-2.5 sm:gap-3"
        >
          <slot name="footer" />
        </div>
      {/if}
    </div>
  </div>
{/if}
