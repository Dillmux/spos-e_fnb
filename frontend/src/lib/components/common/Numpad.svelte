<script lang="ts">
  import { Delete } from "lucide-svelte";

  export let value: string = "0";
  export let maxLength: number = 10;
  export let onChange: (val: string) => void = () => {};

  function addDigit(digit: string) {
    if (value === "0") {
      value = digit;
    } else if (value.length < maxLength) {
      value += digit;
    }
    onChange(value);
  }

  function addZeros() {
    if (value !== "0" && value.length + 3 <= maxLength) {
      value += "000";
      onChange(value);
    }
  }

  function backspace() {
    if (value.length <= 1) {
      value = "0";
    } else {
      value = value.slice(0, -1);
    }
    onChange(value);
  }

  function clear() {
    value = "0";
    onChange(value);
  }
</script>

<div
  class="grid grid-cols-3 gap-2 bg-slate-50 p-3 rounded-2xl border border-slate-200 shadow-inner"
>
  {#each ["1", "2", "3", "4", "5", "6", "7", "8", "9"] as digit}
    <button
      type="button"
      on:click={() => addDigit(digit)}
      class="touch-target h-12 text-xl font-bold text-slate-800 bg-white hover:bg-slate-100 active:bg-brand-primary active:text-white rounded-xl border border-slate-200 shadow-sm transition-all flex items-center justify-center"
    >
      {digit}
    </button>
  {/each}
  <button
    type="button"
    on:click={addZeros}
    class="touch-target h-12 text-xs font-extrabold text-brand-primary bg-brand-primary/10 hover:bg-brand-primary/20 rounded-xl transition-all flex items-center justify-center"
  >
    +000
  </button>
  <button
    type="button"
    on:click={() => addDigit("0")}
    class="touch-target h-12 text-xl font-bold text-slate-800 bg-white hover:bg-slate-100 active:bg-brand-primary active:text-white rounded-xl border border-slate-200 shadow-sm transition-all flex items-center justify-center"
  >
    0
  </button>
  <button
    type="button"
    on:click={backspace}
    class="touch-target h-12 text-slate-700 bg-amber-100 hover:bg-amber-200 active:bg-amber-300 rounded-xl transition-all flex items-center justify-center"
    title="Hapus Digit Terakhir"
  >
    <Delete class="w-5 h-5 text-amber-800" />
  </button>
  <button
    type="button"
    on:click={clear}
    class="col-span-3 touch-target h-4 text-xs font-extrabold text-slate-700 bg-slate-200 hover:bg-red-200 hover:text-red-900 rounded-xl transition-all flex items-center justify-center uppercase tracking-wider"
  >
    Bersihkan Input (C)
  </button>
</div>
