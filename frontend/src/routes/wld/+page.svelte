<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { posApi } from "$lib/api/pos";
  import DisabledFeatureState from "$lib/components/common/DisabledFeatureState.svelte";
  import { featureFlags, outletProfile } from "$lib/stores/config";
  import { toast } from "$lib/stores/toast";
  import {
    Tv,
    Volume2,
    VolumeX,
    Maximize2,
    Minimize2,
    Clock,
    CheckCircle2,
    Flame,
    AlertCircle,
    Check,
  } from "lucide-svelte";

  interface WldOrder {
    id: string;
    invoice_number: string;
    order_number: string;
    queue_number?: number;
    customer_name?: string;
    table_name?: string;
    display_name: string;
    service_mode?: string;
    service_type?: string;
    status: string;
    kitchen_status: "preparing" | "ready";
    items_count: number;
    updated_at?: string;
  }

  let orders: WldOrder[] = [];
  let tablesMap = new Map<string | number, string>();
  let isLoading = true;
  let refreshInterval: any;
  let clockTime = "";
  let isSoundEnabled = true;
  let isFullscreen = false;
  let previousReadyIds: Set<string> = new Set();
  let completingOrderIds: Set<string> = new Set();

  function updateClock() {
    const now = new Date();
    clockTime = now.toLocaleTimeString("id-ID", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  }

  function playChimeSound() {
    if (!isSoundEnabled) return;
    try {
      const AudioContext =
        window.AudioContext || (window as any).webkitAudioContext;
      if (!AudioContext) return;
      const ctx = new AudioContext();

      // First tone (G5)
      const osc1 = ctx.createOscillator();
      const gain1 = ctx.createGain();
      osc1.type = "sine";
      osc1.frequency.setValueAtTime(783.99, ctx.currentTime); // G5
      gain1.gain.setValueAtTime(0.3, ctx.currentTime);
      gain1.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.5);
      osc1.connect(gain1);
      gain1.connect(ctx.destination);
      osc1.start(ctx.currentTime);
      osc1.stop(ctx.currentTime + 0.5);

      // Second tone (C6) after 150ms
      setTimeout(() => {
        const osc2 = ctx.createOscillator();
        const gain2 = ctx.createGain();
        osc2.type = "sine";
        osc2.frequency.setValueAtTime(1046.5, ctx.currentTime); // C6
        gain2.gain.setValueAtTime(0.3, ctx.currentTime);
        gain2.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.6);
        osc2.connect(gain2);
        gain2.connect(ctx.destination);
        osc2.start(ctx.currentTime);
        osc2.stop(ctx.currentTime + 0.6);
      }, 150);
    } catch {
      // Ignore Web Audio errors if audio context blocked before interaction
    }
  }

  async function loadTables() {
    try {
      const res = await posApi.getTables();
      const tables = res.data || [];
      const newMap = new Map<string | number, string>();
      for (const t of tables) {
        if (t.id && t.table_number) {
          newMap.set(t.id, t.table_number);
          newMap.set(Number(t.id), t.table_number);
          newMap.set(String(t.id), t.table_number);
        }
      }
      tablesMap = newMap;
    } catch {
      // Ignore table fetch error fallback
    }
  }

  async function loadWldOrders() {
    try {
      const res = await posApi.getKdsOrders();
      const rawOrders = res.data || [];

      const parsedOrders: WldOrder[] = rawOrders
        .map((o: any) => {
          const items = o.items || o.OrderItems || [];
          const allItemsServed =
            items.length > 0 &&
            items.every((i: any) => i.kitchen_status === "served");

          if (allItemsServed) return null;

          const allItemsReady =
            items.length > 0 &&
            items.every(
              (i: any) =>
                i.kitchen_status === "served" || i.kitchen_status === "ready",
            );

          const invNumber =
            o.invoice_number ||
            o.order_number ||
            `#${o.id.substring(0, 6).toUpperCase()}`;

          const custName = o.customer_name || o.CustomerName || o.name || "";
          const mode = o.service_mode || o.service_type || "dine_in";

          // Comprehensive Table Name Resolution
          const tableFromMap = o.table_id
            ? tablesMap.get(o.table_id) ||
              tablesMap.get(Number(o.table_id)) ||
              tablesMap.get(String(o.table_id))
            : "";

          const rawTable =
            o.table_name ||
            o.table_number ||
            o.table?.table_number ||
            o.table?.name ||
            o.Table?.table_number ||
            o.Table?.name ||
            tableFromMap ||
            "";

          let tblName = "";
          if (mode === "take_away" || mode === "takeaway") {
            tblName = "Takeaway";
          } else if (mode === "delivery") {
            tblName = "Delivery";
          } else if (rawTable) {
            const cleanStr = String(rawTable).trim();
            tblName = cleanStr.toLowerCase().startsWith("meja")
              ? cleanStr
              : `Meja ${cleanStr}`;
          }

          let displayName = "";
          if (tblName && custName) {
            displayName = `${custName} | ${tblName}`;
          } else if (tblName) {
            displayName = tblName;
          } else if (custName) {
            displayName = custName;
          } else {
            displayName = invNumber;
          }

          return {
            id: o.id,
            invoice_number: invNumber,
            order_number: o.order_number || invNumber,
            queue_number: o.queue_number || o.QueueNumber || 0,
            customer_name: custName,
            table_name: tblName,
            display_name: displayName,
            service_mode: mode,
            status: o.status,
            kitchen_status: allItemsReady
              ? ("ready" as const)
              : ("preparing" as const),
            items_count: items.length,
          };
        })
        .filter(Boolean) as WldOrder[];

      // Check for new READY orders to trigger sound
      const currentReadyIds = new Set(
        parsedOrders
          .filter((o) => o.kitchen_status === "ready")
          .map((o) => o.id),
      );
      if (previousReadyIds.size > 0) {
        let hasNewReady = false;
        for (const readyId of currentReadyIds) {
          if (!previousReadyIds.has(readyId)) {
            hasNewReady = true;
            break;
          }
        }
        if (hasNewReady) {
          playChimeSound();
        }
      }
      previousReadyIds = currentReadyIds;
      orders = parsedOrders;
    } catch {
      orders = [];
    } finally {
      isLoading = false;
    }
  }

  async function handleConfirmPickup(orderId: string, displayName: string) {
    completingOrderIds.add(orderId);
    completingOrderIds = new Set(completingOrderIds);
    try {
      await posApi.completeOrder(orderId);
      toast.success(`Pesanan "${displayName}" berhasil diserahkan!`);
      orders = orders.filter((o) => o.id !== orderId);
    } catch (err: any) {
      toast.error(err.message || "Gagal mengonfirmasi penyerahan pesanan");
    } finally {
      completingOrderIds.delete(orderId);
      completingOrderIds = new Set(completingOrderIds);
    }
  }

  function handleFullscreenChange() {
    isFullscreen = !!document.fullscreenElement;
  }

  function toggleFullscreen() {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(() => {});
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen().catch(() => {});
      }
    }
  }

  let eventSource: EventSource | null = null;

  onMount(() => {
    updateClock();
    const clockInterval = setInterval(updateClock, 1000);
    document.addEventListener("fullscreenchange", handleFullscreenChange);

    (async () => {
      await loadTables();
      await loadWldOrders();
    })();

    refreshInterval = setInterval(loadWldOrders, 10000);

    if (typeof window !== 'undefined' && 'EventSource' in window) {
      try {
        eventSource = new EventSource('/api/v1/events/stream');
        eventSource.onmessage = () => {
          loadWldOrders();
        };
        eventSource.addEventListener('order_created', () => loadWldOrders());
        eventSource.addEventListener('order_updated', () => loadWldOrders());
        eventSource.addEventListener('kitchen_status_updated', () => loadWldOrders());
      } catch {}
    }

    return () => {
      clearInterval(clockInterval);
      if (refreshInterval) clearInterval(refreshInterval);
      if (eventSource) eventSource.close();
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
    };
  });

  $: preparingOrders = orders.filter((o) => (o.kitchen_status as any) === "preparing" || (o.kitchen_status as any) === "queued" || (o.kitchen_status as any) === "pending");
  $: readyOrders = orders.filter((o) => o.kitchen_status === "ready");
</script>

<svelte:head>
  <title
    >Waiting Line Display (WLD) - {$outletProfile?.name || "Smart POS"}</title
  >
</svelte:head>

{#if !$featureFlags.enable_wld}
  <DisabledFeatureState
    title="Modul Waiting Line Display (WLD) Nonaktif"
    description="Fitur layar antrean panggil pemesan (WLD) belum diaktifkan di Setup Wizard / Konfigurasi Modul. Silakan aktifkan modul WLD di menu pengaturan."
    icon={Tv}
    actionText="Kembali ke Kasir POS"
    actionHref="/pos"
  />
{:else}
  <div
    class="h-full bg-slate-950 text-white flex flex-col overflow-hidden select-none font-sans"
  >
    <!-- WLD Top Banner / Header -->
    <header
      class="bg-slate-900 px-6 py-4 border-b border-slate-800 flex items-center justify-between flex-shrink-0 shadow-lg"
    >
      <div class="flex items-center gap-3">
        <div class="p-2.5 bg-brand-primary rounded-2xl text-white shadow-md">
          <Tv class="w-7 h-7" />
        </div>
        <div>
          <h1 class="text-2xl font-black tracking-wider text-white uppercase">
            {$outletProfile?.name || "SMART POS RESTO"}
          </h1>
          <p
            class="text-xs text-brand-cyan font-bold tracking-widest uppercase"
          >
            Waiting Line Customer Screen • Monitor Antrean Pesanan Pelanggan
          </p>
        </div>
      </div>

      <!-- Controls & Digital Clock -->
      <div class="flex items-center gap-4">
        <div
          class="flex items-center gap-2 px-4 py-2 bg-slate-800 rounded-2xl border border-slate-700 text-brand-cyan font-mono font-extrabold text-xl shadow-inner"
        >
          <Clock class="w-5 h-5 text-brand-primary" />
          {clockTime}
        </div>

        <button
          type="button"
          on:click={toggleFullscreen}
          class="p-3 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-2xl border border-slate-700 transition"
          title="Layar Penuh TV"
        >
          {#if isFullscreen}
            <Minimize2 class="w-5 h-5" />
          {:else}
            <Maximize2 class="w-5 h-5" />
          {/if}
        </button>
      </div>
    </header>

    <!-- Main 2-Column Board -->
    <div
      class="flex-1 grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-6 p-3 sm:p-6 overflow-y-auto md:overflow-hidden"
    >
      <!-- COLUMN 1: PREPARING / SEDANG DIPROSES -->
      <div
        class="flex flex-col bg-slate-900/60 rounded-3xl border border-slate-800/80 overflow-hidden shadow-2xl"
      >
        <!-- Header -->
        <div
          class="px-4 sm:px-6 py-3.5 sm:py-4 bg-amber-500/10 border-b border-amber-500/20 flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <div class="p-2 bg-amber-500/20 text-amber-400 rounded-xl">
              <Flame class="w-6 h-6 animate-pulse" />
            </div>
            <h2
              class="text-lg sm:text-xl font-black text-amber-400 uppercase tracking-wider"
            >
              Sedang Diproses <span
                class="text-xs font-semibold text-amber-300/70"
                >(PREPARING)</span
              >
            </h2>
          </div>
          <span
            class="px-3.5 py-1 bg-amber-500/20 text-amber-300 font-extrabold text-xs sm:text-sm rounded-full border border-amber-500/40"
          >
            {preparingOrders.length} Pesanan
          </span>
        </div>

        <!-- Order Grid -->
        <div
          class="p-4 sm:p-6 overflow-y-auto flex-1 grid grid-cols-1 sm:grid-cols-2 gap-4 align-content-start"
        >
          {#if preparingOrders.length === 0}
            <div
              class="col-span-full py-16 sm:py-20 text-center text-slate-600 font-semibold space-y-2"
            >
              <Flame class="w-10 h-10 mx-auto opacity-30" />
              <p class="text-sm">Tidak ada pesanan sedang dimasak</p>
            </div>
          {:else}
            {#each preparingOrders as order (order.id)}
              <div
                class="p-4 sm:p-5 bg-slate-900 border border-slate-800 rounded-2xl flex flex-col justify-between space-y-3 shadow-sm"
              >
                <div>
                  <div class="flex items-center justify-between gap-2 mb-2">
                    <span
                      class="px-3 py-1 bg-amber-500/20 text-amber-300 border border-amber-500/40 rounded-xl font-mono font-black text-lg shadow-inner"
                    >
                      ANTREAN #{order.queue_number ? String(order.queue_number).padStart(3, '0') : (order.invoice_number?.slice(-4) || '---')}
                    </span>
                    <span
                      class="px-2.5 py-1 bg-slate-800 text-amber-300 text-[10px] uppercase font-bold rounded-lg shrink-0"
                    >
                      {order.service_mode}
                    </span>
                  </div>

                  <span
                    class="text-xl sm:text-2xl font-black text-white tracking-wide block truncate"
                  >
                    {order.display_name}
                  </span>
                </div>

                <div
                  class="pt-2 border-t border-slate-800/80 flex items-center justify-between text-xs text-slate-400 font-bold"
                >
                  <span class="truncate font-mono"
                    >Invoice: {order.invoice_number}</span
                  >
                </div>
              </div>
            {/each}
          {/if}
        </div>
      </div>

      <!-- COLUMN 2: READY FOR PICKUP / SIAP DIAMBIL -->
      <div
        class="flex flex-col bg-slate-900/60 rounded-3xl border border-slate-800/80 overflow-hidden shadow-2xl"
      >
        <!-- Header -->
        <div
          class="px-4 sm:px-6 py-3.5 sm:py-4 bg-emerald-500/10 border-b border-emerald-500/20 flex items-center justify-between"
        >
          <div class="flex items-center gap-3">
            <div class="p-2 bg-emerald-500/20 text-emerald-400 rounded-xl">
              <CheckCircle2 class="w-6 h-6 animate-bounce" />
            </div>
            <h2
              class="text-lg sm:text-xl font-black text-emerald-400 uppercase tracking-wider"
            >
              Siap Diambil <span
                class="text-xs font-semibold text-emerald-300/70"
                >(READY FOR PICKUP)</span
              >
            </h2>
          </div>
          <span
            class="px-3.5 py-1 bg-emerald-500/20 text-emerald-300 font-extrabold text-xs sm:text-sm rounded-full border border-emerald-500/40"
          >
            {readyOrders.length} Pesanan
          </span>
        </div>

        <!-- Order Grid -->
        <div
          class="p-4 sm:p-6 overflow-y-auto flex-1 grid grid-cols-1 sm:grid-cols-2 gap-4 align-content-start"
        >
          {#if readyOrders.length === 0}
            <div
              class="col-span-full py-16 sm:py-20 text-center text-slate-600 font-semibold space-y-2"
            >
              <CheckCircle2 class="w-10 h-10 mx-auto opacity-30" />
              <p class="text-sm">Belum ada pesanan siap diambil</p>
            </div>
          {:else}
            {#each readyOrders as order (order.id)}
              <div
                class="p-4 sm:p-5 bg-gradient-to-br from-emerald-950/90 to-slate-900 border-2 border-emerald-500/60 rounded-2xl flex flex-col justify-between space-y-3 shadow-lg shadow-emerald-950/50"
              >
                <div>
                  <div class="flex items-center justify-between gap-2 mb-2">
                    <span
                      class="px-3 py-1 bg-emerald-500 text-slate-950 rounded-xl font-mono font-black text-xl shadow-md"
                    >
                      ANTREAN #{order.queue_number ? String(order.queue_number).padStart(3, '0') : (order.invoice_number?.slice(-4) || '---')}
                    </span>
                    <span
                      class="px-2.5 py-1 bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 text-[10px] font-black uppercase rounded-lg shrink-0"
                    >
                      PICKUP
                    </span>
                  </div>

                  <span
                    class="text-xl sm:text-2xl font-black text-white tracking-wide block truncate"
                  >
                    {order.display_name}
                  </span>
                  <span
                    class="text-xs font-semibold text-slate-400 block mt-1 font-mono truncate"
                  >
                    Invoice: {order.invoice_number}
                  </span>
                </div>

                <!-- Konfirmasi Pengambilan Button -->
                <button
                  type="button"
                  disabled={completingOrderIds.has(order.id)}
                  on:click={() =>
                    handleConfirmPickup(order.id, order.display_name)}
                  class="w-full py-3 min-h-[48px] bg-emerald-500 hover:bg-emerald-400 active:scale-95 text-slate-950 font-extrabold text-xs sm:text-sm rounded-xl shadow-md flex items-center justify-center gap-2 transition disabled:opacity-50 cursor-pointer"
                >
                  <Check class="w-5 h-5 stroke-[3]" />
                  {completingOrderIds.has(order.id)
                    ? "Memproses..."
                    : "Konfirmasi Diambil / Selesai"}
                </button>
              </div>
            {/each}
          {/if}
        </div>
      </div>
    </div>
  </div>
{/if}
