import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  base: './',
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 4000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8110',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: '../server/resource/public',
    emptyOutDir: true
  }
})
