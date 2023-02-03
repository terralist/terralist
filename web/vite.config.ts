import { defineConfig } from 'vite'
import { resolve } from 'path'
import { svelte, vitePreprocess } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte({
      preprocess: [vitePreprocess()],
      onwarn: (warning, handler) => {
          if (warning.code.startsWith('a11y-')) {
              return; // silence a11y warnings
          }
          handler(warning);
      },
    }),
  ],
  envPrefix: "TERRALIST",
  base: "./",
  build: {
    rollupOptions: {
      input: {
        index: resolve(__dirname, 'index.html'),
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
  resolve: {
    alias: {
      '@': __dirname + '/src',
    }
  },
});
