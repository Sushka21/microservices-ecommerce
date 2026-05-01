import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // LOMS (8081) — более специфичные пути первыми
      '/v1/product': { target: 'http://localhost:8081', changeOrigin: true },
      '/v1/stock': { target: 'http://localhost:8081', changeOrigin: true },
      '/v1/order': { target: 'http://localhost:8081', changeOrigin: true },
      // Cart (8080)
      '/v1/cart': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
})
