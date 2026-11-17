// The phone recorder's page (Amendment 21), built apart from the desktop's so
// it never enters the dist folder the desktop embeds.

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  // Relative, so the page works wherever it is served from.
  base: './',
  build: {
    outDir: 'dist-recorder',
    emptyOutDir: true,
    rollupOptions: { input: 'recorder.html' },
  },
})
