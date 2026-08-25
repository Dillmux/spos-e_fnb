<script lang="ts">
  import "../app.css";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { config, featureFlags, isInitialized } from "$lib/stores/config";
  import { auth, isAuthenticated, currentUser } from "$lib/stores/auth";
  import {
    printer,
    isPrinterConnected,
    connectedPrinterName,
  } from "$lib/stores/printer";
  import { toast } from "$lib/stores/toast";
  import Toast from "$lib/components/common/Toast.svelte";
  import PinModal from "$lib/components/common/PinModal.svelte";
  import ShiftModal from "$lib/components/pos/ShiftModal.svelte";
  import PrintableReceipt from "$lib/components/pos/PrintableReceipt.svelte";
  import {
    Store,
    Monitor,
    Tv,
    LayoutGrid,
    Coffee,
    Package,
    History,
    BarChart3,
    Users,
    Printer,
    Wallet,
    LogOut,
    User as UserIcon,
    Menu,
    X,
  } from "lucide-svelte";

  let isShiftModalOpen = false;
  let isFullscreen = false;
  let isMobileNavOpen = false;

  onMount(() => {
    const handleFullscreenChange = () => {
      isFullscreen = !!document.fullscreenElement;
    };
    document.addEventListener("fullscreenchange", handleFullscreenChange);

    const init = async () => {
      // Check system setup initialization status on application mount
      await config.checkStatus();
      // Initialize/Restore user session if token exists
      await auth.initSession();

      const currentPath = $page.url.pathname;
      const initialized = $isInitialized;

      if (!initialized && currentPath !== "/setup") {
        goto("/setup");
      } else if (initialized && currentPath === "/setup") {
        if (!$isAuthenticated) {
          goto("/login");
        } else {
          goto("/pos");
        }
      } else if (initialized && !$isAuthenticated && currentPath !== "/login") {
        goto("/login");
      } else if (initialized && $isAuthenticated && currentPath === "/login") {
        goto("/pos");
      }
    };

    init();

    const handleOpenShiftEvent = () => {
      isShiftModalOpen = true;
    };
    window.addEventListener("open-shift-modal", handleOpenShiftEvent);

    return () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange);
      window.removeEventListener("open-shift-modal", handleOpenShiftEvent);
    };
  });

  $: {
    // Close mobile nav on route change
    if ($page.url.pathname) {
      isMobileNavOpen = false;
    }

    // Pengawal rute reaktif saat rute atau status otentikasi berubah
    if (
      typeof window !== "undefined" &&
      !$config.checkingStatus &&
      !$auth.isLoading
    ) {
      const path = $page.url.pathname;
      if (!$isInitialized && path !== "/setup") {
        goto("/setup");
      } else if ($isInitialized && path === "/setup") {
        if (!$isAuthenticated) {
          goto("/login");
        } else {
          goto("/pos");
        }
      } else if ($isInitialized && !$isAuthenticated && path !== "/login") {
        goto("/login");
      } else if ($isInitialized && $isAuthenticated && path === "/login") {
        goto("/pos");
      }
    }
  }

  function handlePrinterConnect() {
    if ($isPrinterConnected) {
      printer.disconnect();
    } else {
      printer.connect();
    }
  }

  function handleLogout() {
    auth.logout();
    toast.info("Berhasil keluar dari akun.");
    goto("/login");
  }
</script>

<div
  class="h-full flex flex-col bg-brand-surface text-slate-900 overflow-hidden font-sans"
>
  <!-- Top Navigation Header (Hidden during Setup Wizard, Login, or Fullscreen WDL) -->
  {#if $page.url.pathname !== "/" && !$page.url.pathname.startsWith("/setup") && !$page.url.pathname.startsWith("/login") && !($page.url.pathname.startsWith("/wld") && isFullscreen)}
    <header
      class="bg-brand-dark text-white px-3 sm:px-5 py-2.5 flex items-center justify-between shadow-md relative z-[50] border-b border-slate-800"
    >
      <!-- Brand & Outlet Name -->
      <div class="flex items-center gap-3">
        <!-- Mobile Menu Toggle Button -->
        <button
          type="button"
          on:click={() => (isMobileNavOpen = !isMobileNavOpen)}
          class="lg:hidden p-2 text-slate-300 hover:text-white hover:bg-slate-800 rounded-xl transition min-h-[48px] min-w-[48px] flex items-center justify-center"
          aria-label="Buka Menu Navigasi"
        >
          {#if isMobileNavOpen}
            <X class="w-6 h-6 text-brand-cyan" />
          {:else}
            <Menu class="w-6 h-6" />
          {/if}
        </button>

        <a href="/pos" class="flex items-center gap-2.5 group">
          <div
            class="p-2 bg-brand-primary rounded-xl text-white shadow-sm group-hover:scale-105 transition"
          >
            <Store class="w-5 h-5" />
          </div>
          <div>
            <h1
              class="font-extrabold text-sm sm:text-base leading-tight tracking-wide text-white truncate max-w-[160px] sm:max-w-none"
            >
              {$config.outlet?.name || "Smart POS & Resto"}
            </h1>
            <p class="text-[10px] sm:text-[11px] text-brand-cyan font-semibold">
              SPOS-Engine
            </p>
          </div>
        </a>

        <!-- Desktop Navigation Tabs -->
        <nav
          class="hidden lg:flex items-center gap-1.5 ml-4 pl-4 border-l border-slate-800"
        >
          <a
            href="/pos"
            class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
            {$page.url.pathname === '/pos'
              ? 'bg-brand-primary text-white shadow-md'
              : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
          >
            <LayoutGrid class="w-4 h-4" /> POS
          </a>

          {#if $featureFlags.enable_kds}
            <a
              href="/kds"
              class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
              {$page.url.pathname === '/kds'
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <Monitor class="w-4 h-4" /> KDS
            </a>
          {/if}

          {#if $featureFlags.enable_wld}
            <a
              href="/wld"
              class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
              {$page.url.pathname === '/wld'
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <Tv class="w-4 h-4" /> WLD
            </a>
          {/if}

          <a
            href="/admin/menu"
            class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
            {$page.url.pathname.startsWith('/admin/menu')
              ? 'bg-brand-primary text-white shadow-md'
              : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
          >
            <Coffee class="w-4 h-4" /> Menu & BOM
          </a>

          <a
            href="/admin/inventory"
            class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
            {$page.url.pathname.startsWith('/admin/inventory')
              ? 'bg-brand-primary text-white shadow-md'
              : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
          >
            <Package class="w-4 h-4" /> Stok
          </a>

          <a
            href="/admin/orders"
            class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
            {$page.url.pathname.startsWith('/admin/orders')
              ? 'bg-brand-primary text-white shadow-md'
              : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
          >
            <History class="w-4 h-4" /> Riwayat
          </a>

          <a
            href="/admin/reports"
            class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
            {$page.url.pathname.startsWith('/admin/reports')
              ? 'bg-brand-primary text-white shadow-md'
              : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
          >
            <BarChart3 class="w-4 h-4" /> Laporan
          </a>

          <a
            href="/admin/users"
            class="px-3 py-2 rounded-xl text-xs font-bold transition flex items-center gap-1.5 min-h-[44px]
            {$page.url.pathname.startsWith('/admin/users')
              ? 'bg-brand-primary text-white shadow-md'
              : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
          >
            <Users class="w-4 h-4" /> Pengguna
          </a>
        </nav>
      </div>

      <!-- Right Utility Controls -->
      <div class="flex items-center gap-2 sm:gap-3">
        <!-- Shift Drawer Control Button -->
        <button
          type="button"
          on:click={() => (isShiftModalOpen = true)}
          class="px-3 py-2 bg-slate-800 hover:bg-slate-700 text-brand-cyan border border-slate-700 rounded-xl text-xs font-bold transition flex items-center gap-1.5 shadow-sm min-h-[44px]"
        >
          <Wallet class="w-4 h-4 text-brand-cyan" />
          <span class="hidden sm:inline">Shift & Kas Drawer</span>
          <span class="sm:hidden">Shift</span>
        </button>

        <!-- User Profile & Logout -->
        {#if $isAuthenticated}
          <div
            class="flex items-center gap-2 pl-2 sm:pl-3 border-l border-slate-800"
          >
            <div
              class="w-8 h-8 rounded-full bg-brand-primary text-white font-bold text-xs flex items-center justify-center shadow-sm"
            >
              {$currentUser?.name ? $currentUser.name[0].toUpperCase() : "A"}
            </div>
            <div class="hidden xl:block text-left">
              <p class="text-xs font-bold leading-tight text-white">
                {$currentUser?.name || "Kasir"}
              </p>
              <p class="text-[10px] text-slate-400 uppercase font-semibold">
                {$currentUser?.role || "User"}
              </p>
            </div>
            <button
              type="button"
              on:click={handleLogout}
              class="p-2 text-slate-400 hover:text-brand-danger transition rounded-lg min-h-[44px] min-w-[44px] flex items-center justify-center"
              title="Keluar Akun"
            >
              <LogOut class="w-4 h-4" />
            </button>
          </div>
        {:else}
          <a
            href="/login"
            class="px-3.5 py-2 bg-brand-primary text-white rounded-xl text-xs font-bold hover:bg-sky-500 transition min-h-[44px] flex items-center"
          >
            Login Kasir
          </a>
        {/if}
      </div>
    </header>

    <!-- Mobile Slide-over Drawer Backdrop & Navigation Panel -->
    {#if isMobileNavOpen}
      <button
        type="button"
        aria-label="Tutup Navigasi Mobile"
        class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-40 lg:hidden transition-opacity cursor-default w-full h-full border-none p-0"
        on:click={() => (isMobileNavOpen = false)}
      />
      <div
        class="fixed top-[57px] left-0 bottom-0 w-72 bg-brand-dark border-r border-slate-800 text-white z-50 p-4 overflow-y-auto flex flex-col justify-between shadow-2xl lg:hidden animate-in slide-in-from-left duration-200"
      >
        <div class="space-y-4">
          <div
            class="px-2 py-1 text-xs font-bold text-slate-400 uppercase tracking-wider"
          >
            Navigasi Aplikasi
          </div>
          <nav class="space-y-1">
            <a
              href="/pos"
              class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
              {$page.url.pathname === '/pos'
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <LayoutGrid class="w-5 h-5" /> Kasir POS
            </a>

            {#if $featureFlags.enable_kds}
              <a
                href="/kds"
                class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
                {$page.url.pathname === '/kds'
                  ? 'bg-brand-primary text-white shadow-md'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
              >
                <Monitor class="w-5 h-5" /> KDS Dapur
              </a>
            {/if}

            {#if $featureFlags.enable_wld}
              <a
                href="/wld"
                class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
                {$page.url.pathname === '/wld'
                  ? 'bg-brand-primary text-white shadow-md'
                  : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
              >
                <Tv class="w-5 h-5" /> WLD Antrean
              </a>
            {/if}

            <a
              href="/admin/menu"
              class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
              {$page.url.pathname.startsWith('/admin/menu')
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <Coffee class="w-5 h-5" /> Menu & Resep BOM
            </a>

            <a
              href="/admin/inventory"
              class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
              {$page.url.pathname.startsWith('/admin/inventory')
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <Package class="w-5 h-5" /> Stok Bahan Baku
            </a>

            <a
              href="/admin/orders"
              class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
              {$page.url.pathname.startsWith('/admin/orders')
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <History class="w-5 h-5" /> Riwayat Transaksi
            </a>

            <a
              href="/admin/reports"
              class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
              {$page.url.pathname.startsWith('/admin/reports')
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <BarChart3 class="w-5 h-5" /> Laporan Analitik
            </a>

            <a
              href="/admin/users"
              class="px-4 py-3 rounded-xl text-sm font-bold transition flex items-center gap-3 min-h-[48px]
              {$page.url.pathname.startsWith('/admin/users')
                ? 'bg-brand-primary text-white shadow-md'
                : 'text-slate-300 hover:bg-slate-800 hover:text-white'}"
            >
              <Users class="w-5 h-5" /> Manajemen Pengguna
            </a>
          </nav>
        </div>

        <div class="pt-4 border-t border-slate-800 space-y-3">
          <div class="flex items-center gap-3 px-2">
            <div
              class="w-10 h-10 rounded-full bg-brand-primary text-white font-bold text-sm flex items-center justify-center"
            >
              {$currentUser?.name ? $currentUser.name[0].toUpperCase() : "A"}
            </div>
            <div>
              <p class="text-sm font-bold text-white">
                {$currentUser?.name || "Kasir"}
              </p>
              <p class="text-xs text-slate-400 uppercase font-semibold">
                {$currentUser?.role || "User"}
              </p>
            </div>
          </div>

          <button
            type="button"
            on:click={handleLogout}
            class="w-full py-3 px-4 bg-slate-800 hover:bg-brand-danger text-white rounded-xl text-xs font-bold transition flex items-center justify-center gap-2 min-h-[48px]"
          >
            <LogOut class="w-4 h-4" /> Keluar Akun
          </button>
        </div>
      </div>
    {/if}
  {/if}

  <!-- Main View Container -->
  <main class="flex-1 overflow-hidden relative z-30">
    <slot />
  </main>

  <!-- Global Modals & Notifications -->
  <Toast />
  <PinModal />
  <ShiftModal
    isOpen={isShiftModalOpen}
    onClose={() => (isShiftModalOpen = false)}
  />
  <PrintableReceipt />
</div>
