import vue from '@vitejs/plugin-vue2';
import path from 'path';
import {defineConfig} from 'vite';

export default defineConfig(() => {
  return {
    plugins: [vue()],
    resolve: {
      alias: {
        'vue': 'vue/dist/vue.esm.js',
        '~@fortawesome': path.resolve(__dirname, './node_modules/@fortawesome'),
        '~sweetalert2': path.resolve(__dirname, './node_modules/sweetalert2'),
        '~emoji-datasource': path.resolve(__dirname, './node_modules/emoji-datasource'),
        '~': path.resolve(__dirname, './src'),
        '@': path.resolve(__dirname, './src'),
      },
      extensions: ['.js', '.ts', '.jsx', '.tsx', '.json', '.vue'],
    },
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `@import "~/assets/styles/_variables.scss";`,
        },
      },
    },
    esbuild: {
      tsconfigRaw: {
        compilerOptions: {
          experimentalDecorators: true,
        },
      },
    },
    server: {
      port: 3000,
      host: '0.0.0.0',
      allowedHosts: 'all',
      // HMR is disabled in AI Studio via DISABLE_HMR env var.
      hmr: process.env.DISABLE_HMR !== 'true',
      watch: process.env.DISABLE_HMR === 'true' ? null : {},
    },
  };
});
