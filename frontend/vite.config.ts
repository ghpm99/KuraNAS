import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import legacy from '@vitejs/plugin-legacy';
import path from 'path';

export default defineConfig(() => {
    const env = process.env.VITE_API_URL;

    return {
        plugins: [
            react(),
            legacy({
                targets: ['defaults', '> 0.2%', 'not dead', 'Opera >= 50'],
                modernPolyfills: true,
            }),
        ],
        resolve: {
            alias: {
                '@': path.resolve(__dirname, './src'),
            },
        },
        define: {
            'process.env.VITE_API_URL': JSON.stringify(env ?? ''),
        },
        build: {
            rollupOptions: {
                output: {
                    manualChunks(id) {
                        if (!id.includes('node_modules')) return;

                        if (id.includes('@mui/x-charts') || id.includes('@mui/x-data-grid')) {
                            return 'vendor-mui-x';
                        }
                        if (id.includes('@mui') || id.includes('@emotion')) {
                            return 'vendor-mui';
                        }
                        if (id.includes('framer-motion')) {
                            return 'vendor-motion';
                        }
                        if (id.includes('@tanstack')) {
                            return 'vendor-query';
                        }
                    },
                },
            },
        },
    };
});
