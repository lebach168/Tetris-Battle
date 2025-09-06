import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';
import tanstackRouter from '@tanstack/router-plugin/vite';

export default defineConfig({
  server: {
    port: 3000,
    strictPort: true,
  },
  plugins: [
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
      // mặc định: routesDirectory: './src/routes', generatedRouteTree: './src/routeTree.gen.ts'
    }),
    tailwindcss(),
    react(),
  ],
});
