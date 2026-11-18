import { type KnipConfig } from "knip";

const config: KnipConfig = {
	prettier: ["./prettier.config.js"],
	project: ["src/**/*.{ts,tsx}"],
	stylelint: ["./stylelint.config.ts"],
};

export default config;
