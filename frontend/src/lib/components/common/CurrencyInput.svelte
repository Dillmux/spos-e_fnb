<script lang="ts">
  export let value: number = 0;
  export let placeholder: string = 'Rp 0';
  export let disabled: boolean = false;
  export let className: string = '';

  let displayValue = '';

  $: {
    displayValue = value ? value.toLocaleString('id-ID') : '';
  }

  function handleInput(e: Event) {
    const input = e.target as HTMLInputElement;
    const rawNum = input.value.replace(/\D/g, '');
    value = rawNum ? parseInt(rawNum, 10) : 0;
  }
</script>

<div class="relative flex items-center w-full">
  <span class="absolute left-4 font-bold text-slate-500 select-none">Rp</span>
  <input
    type="text"
    inputmode="numeric"
    value={displayValue}
    on:input={handleInput}
    {placeholder}
    {disabled}
    class="w-full pl-12 pr-4 py-3 min-h-[48px] bg-white border border-slate-300 focus:border-brand-primary focus:ring-2 focus:ring-brand-cyan/50 rounded-xl font-bold text-slate-800 text-lg shadow-sm outline-none transition disabled:bg-slate-100 {className}"
  />
</div>
