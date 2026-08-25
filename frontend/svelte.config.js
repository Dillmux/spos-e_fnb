import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      pages: '../backend/public',
      assets: '../backend/public',
      fallback: 'index.html',
      precompress: false,
      strict: false
    })
  }
};

export default config;
