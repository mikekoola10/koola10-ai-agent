import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 3000,
    host: '0.0.0.0',
    strictPort: true,
    hmr: false,
    proxy: {
      '/api/koola10': {
        target: 'https://koola10-ai-agent.onrender.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/koola10/, ''),
        secure: true,
      },
      '/api/spiral': {
        target: 'https://spiral-ai-agent.onrender.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/spiral/, ''),
        secure: true,
      },
      '/api/apex': {
        target: 'https://apex-grok-edition.onrender.com',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api\/apex/, ''),
        secure: true,
      },
    },
  },
});
