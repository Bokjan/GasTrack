import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { VitePWA } from 'vite-plugin-pwa';
import { resolve } from 'path';
export default defineConfig({
    plugins: [
        react(),
        VitePWA({
            registerType: 'autoUpdate',
            includeAssets: ['favicon.svg', 'icons/*.png'],
            manifest: {
                name: 'GasTrack - 车辆能耗与费用管理',
                short_name: 'GasTrack',
                description: '记录加油/充电，追踪维保开销，分析能耗趋势，辅助用车决策',
                theme_color: '#1677ff',
                background_color: '#ffffff',
                display: 'standalone',
                orientation: 'portrait',
                start_url: '/',
                scope: '/',
                categories: ['finance', 'utilities'],
                icons: [
                    {
                        src: '/icons/icon-192x192.png',
                        sizes: '192x192',
                        type: 'image/png',
                    },
                    {
                        src: '/icons/icon-512x512.png',
                        sizes: '512x512',
                        type: 'image/png',
                    },
                    {
                        src: '/icons/icon-512x512.png',
                        sizes: '512x512',
                        type: 'image/png',
                        purpose: 'maskable',
                    },
                ],
            },
            workbox: {
                globPatterns: ['**/*.{js,css,html,ico,png,svg,woff2}'],
                maximumFileSizeToCacheInBytes: 4 * 1024 * 1024, // 4 MiB
                runtimeCaching: [
                    {
                        // API 请求 — 网络优先，失败回退缓存
                        urlPattern: /\/api\/.*/i,
                        handler: 'NetworkFirst',
                        options: {
                            cacheName: 'api-cache',
                            expiration: {
                                maxEntries: 100,
                                maxAgeSeconds: 5 * 60, // 5 min
                            },
                            cacheableResponse: {
                                statuses: [0, 200],
                            },
                        },
                    },
                    {
                        // 图片资源 — 缓存优先
                        urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp)$/i,
                        handler: 'CacheFirst',
                        options: {
                            cacheName: 'image-cache',
                            expiration: {
                                maxEntries: 60,
                                maxAgeSeconds: 30 * 24 * 60 * 60, // 30 days
                            },
                        },
                    },
                ],
            },
        }),
    ],
    resolve: {
        alias: {
            '@': resolve(__dirname, 'src'),
        },
    },
    server: {
        port: 3000,
        proxy: {
            '/api': {
                target: 'http://localhost:8098',
                changeOrigin: true,
            },
        },
    },
});
