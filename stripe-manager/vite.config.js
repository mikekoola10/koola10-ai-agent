import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 3001,
    host: '0.0.0.0',
    strictPort: true,
    hmr: false,
    proxy: {
      '/admin': {
        target: 'https://koola10-ai-agent.onrender.com',
        changeOrigin: true,
        secure: true,
      },
    },
  },
});
