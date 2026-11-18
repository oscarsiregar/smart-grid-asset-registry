/** @type {import('lint-staged').Config} */
const config = {
	"*": [
		() => "pnpm run lint:editor",
		() => "pnpm run lint:fs",
		() => "pnpm run lint:trash",
		() => "pnpm run lint:format",
	],
	"**/*.css": [() => "pnpm run lint:css"],
	"**/*.{ts,tsx}": [() => "pnpm run lint:js", () => "pnpm run lint:type"],
};

export default config;
