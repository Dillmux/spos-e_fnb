<script lang="ts">
  import { goto } from "$app/navigation";
  import { auth, isAuthenticated } from "$lib/stores/auth";
  import { toast } from "$lib/stores/toast";
  import {
    Store,
    Lock,
    User as UserIcon,
    LogIn,
    AlertCircle,
  } from "lucide-svelte";

  let username = "";
  let password = "";
  let isSubmitting = false;
  let loginErrorMessage = "";

  $: if ($isAuthenticated) {
    goto("/pos");
  }

  function clearError() {
    if (loginErrorMessage) loginErrorMessage = "";
  }

  async function handleLogin() {
    loginErrorMessage = "";
    if (!username || !password) {
      loginErrorMessage = "Username dan password wajib diisi!";
      toast.warning("Masukkan username dan password!");
      return;
    }

    isSubmitting = true;
    try {
      await auth.login(username, password);
      toast.success("Login Berhasil! Selamat datang.");
      goto("/pos");
    } catch (err: any) {
      loginErrorMessage =
        err.message ||
        "Username atau password yang Anda masukkan salah. Periksa kembali kredensial Anda.";
      toast.error(loginErrorMessage);
    } finally {
      isSubmitting = false;
    }
  }
</script>

<div
  class="min-h-full bg-brand-surface py-12 px-4 flex flex-col items-center justify-center"
>
  <div
    class="w-full max-w-md bg-white rounded-3xl shadow-2xl overflow-hidden border border-slate-100 p-8 space-y-6"
  >
    <!-- Header -->
    <div class="text-center space-y-2">
      <div
        class="inline-flex p-4 bg-brand-primary text-white rounded-2xl shadow-lg mb-1"
      >
        <Store class="w-10 h-10" />
      </div>
      <h2 class="text-2xl font-extrabold text-slate-900">
        Masuk Kasir & Backoffice
      </h2>
      <p class="text-xs text-slate-500 font-semibold">
        Smart POS & Resto Hospitality Engine
      </p>
    </div>

    <!-- Login Form -->
    <form on:submit|preventDefault={handleLogin} class="space-y-4">
      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-600 uppercase"
          for="login-username">Username</label
        >
        <div class="relative">
          <input
            id="login-username"
            type="text"
            bind:value={username}
            on:input={clearError}
            placeholder="username"
            required
            class="w-full pl-11 pr-4 py-3 min-h-[48px] bg-slate-50 border border-slate-300 focus:border-brand-primary focus:bg-white rounded-xl font-bold text-slate-800 outline-none transition"
          />
          <UserIcon class="w-5 h-5 absolute left-3.5 top-3.5 text-slate-400" />
        </div>
      </div>

      <div class="space-y-1">
        <label
          class="block text-xs font-bold text-slate-600 uppercase"
          for="login-password">Password</label
        >
        <div class="relative">
          <input
            id="login-password"
            type="password"
            bind:value={password}
            on:input={clearError}
            placeholder="••••••••"
            required
            class="w-full pl-11 pr-4 py-3 min-h-[48px] bg-slate-50 border border-slate-300 focus:border-brand-primary focus:bg-white rounded-xl font-bold text-slate-800 outline-none transition"
          />
          <Lock class="w-5 h-5 absolute left-3.5 top-3.5 text-slate-400" />
        </div>
      </div>

      <!-- Alert Bar above Login Button -->
      {#if loginErrorMessage}
        <div
          class="p-3.5 bg-red-50 border border-red-200 text-red-700 rounded-xl text-xs font-semibold flex items-start gap-2.5 animate-in fade-in duration-200 shadow-sm"
        >
          <AlertCircle class="w-4 h-4 text-red-600 flex-shrink-0 mt-0.5" />
          <div class="flex-1">
            <strong class="block font-extrabold text-red-800 text-xs"
              >Gagal Masuk Workspace</strong
            >
            <span>{loginErrorMessage}</span>
          </div>
        </div>
      {/if}

      <button
        type="submit"
        disabled={isSubmitting}
        class="w-full py-4 min-h-[48px] bg-brand-primary hover:bg-sky-500 active:scale-95 text-white font-extrabold text-base rounded-2xl shadow-xl shadow-sky-500/20 flex items-center justify-center gap-2 transition"
      >
        <LogIn class="w-5 h-5" />
        {isSubmitting ? "Authenticating..." : "MASUK WORKSPACE"}
      </button>
    </form>
  </div>
</div>
