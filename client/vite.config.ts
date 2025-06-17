import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import { vanillaExtractPlugin } from '@vanilla-extract/vite-plugin'

const isDocker = process.env.BUILD_TARGET === 'docker';

export default defineConfig({
  plugins: [tailwindcss(), reactRouter(), tsconfigPaths(), vanillaExtractPlugin()],
  server: {
    proxy: {
      '/apis': {
        target: 'http://localhost:4110',
        changeOrigin: true,
      },
      '/images': {
        target: 'http://localhost:4110',
        changeOrigin: true,
      }
    }
  },
  resolve: {
		alias: {
			...(isDocker
        ? { 'react-dom/server': 'react-dom/server.node' }
        : {}),
		},
	},
});