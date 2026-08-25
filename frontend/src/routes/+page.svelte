<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { isInitialized } from '$lib/stores/config';
  import { isAuthenticated } from '$lib/stores/auth';
  import { RefreshCw } from 'lucide-svelte';

  onMount(() => {
    if (!$isInitialized) {
      goto('/setup');
    } else if ($isAuthenticated) {
      goto('/pos');
    } else {
      goto('/login');
    }
  });
</script>

<div class="h-full flex flex-col items-center justify-center bg-brand-surface text-slate-700">
  <RefreshCw class="w-10 h-10 text-brand-primary animate-spin mb-4" />
  <p class="font-bold text-lg text-slate-800">Mengarahkan ke Workspace...</p>
  <p class="text-xs text-slate-500 mt-1">Memeriksa status inisialisasi dan sesi pengguna.</p>
</div>
