import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path';
export default defineConfig({
    plugins: [vue()],
    define: {
        __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false' // 关闭生产环境 hydration 不匹配详情
    },
    resolve: {
        alias: {
            '@': path.resolve(__dirname, './src'), // 将 '@' 映射到 src 目录
        },
    },
})