import reactPlugin from "@vitejs/plugin-react";
import { fileURLToPath } from "node:url";
import { type ConfigEnv, loadEnv } from "vite";
import svgr from "vite-plugin-svgr";
import { defineConfig } from "vitest/config";

const config = ({ mode }: ConfigEnv): ReturnType<typeof defineConfig> => {
	const {
		VITE_APP_API_ORIGIN_URL,
		VITE_APP_DEVELOPMENT_PORT,
		VITE_APP_PROXY_SERVER_URL,
	} = loadEnv(mode, process.cwd()) as Record<
		| "VITE_APP_API_ORIGIN_URL"
		| "VITE_APP_DEVELOPMENT_PORT"
		| "VITE_APP_PROXY_SERVER_URL",
		string
	>;

	return defineConfig({
		build: {
			// Bundles live under /static so they never share a path with the
			// app's /assets routes (nginx would otherwise treat /assets as the
			// bundle directory).
			assetsDir: "static",
			outDir: "build",
		},
		plugins: [reactPlugin(), svgr()],
		resolve: {
			alias: [
				{
					find: "~",
					replacement: fileURLToPath(new URL("src", import.meta.url)),
				},
			],
		},
		server: {
			port: Number(VITE_APP_DEVELOPMENT_PORT),
			proxy: {
				[VITE_APP_API_ORIGIN_URL]: {
					changeOrigin: true,
					target: VITE_APP_PROXY_SERVER_URL,
				},
			},
		},
		// Vitest + React Testing Library (constitution, Template Exception 2).
		// Tests sit next to the code they cover as *.test.ts(x).
		test: {
			// fetch needs an absolute URL; the app itself uses the relative /api.
			env: { VITE_APP_API_ORIGIN_URL: "http://localhost/api" },
			environment: "jsdom",
			include: ["src/**/*.test.{ts,tsx}"],
			restoreMocks: true,
			setupFiles: ["src/libs/modules/testing/setup-tests.ts"],
			unstubGlobals: true,
		},
	});
};

export default config;
