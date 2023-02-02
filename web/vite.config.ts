import { defineConfig } from 'vite'
import { resolve } from 'path'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  build: {
    rollupOptions: {
      input: {
        artifact: resolve(__dirname, 'artifact.html'),
        login: resolve(__dirname, 'login.html'),
        main: resolve(__dirname, 'index.html'),
        management: resolve(__dirname, 'management.html'),
        runtime: resolve(__dirname, 'src', 'runtime.ts'),
      },
      output: {
        entryFileNames: (chunkInfo) => {
          if (chunkInfo.name === 'runtime') {
            return 'runtime.js'
          }

          return 'assets/[name]-[hash].js'
        },
      },
    },
  },
});
