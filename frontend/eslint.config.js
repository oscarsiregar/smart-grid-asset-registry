import js from "@eslint/js";
import ts from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import { resolve as tsResolver } from "eslint-import-resolver-typescript";
import importPlugin, { configs as importConfigs } from "eslint-plugin-import-x";
import jsdocPlugin from "eslint-plugin-jsdoc";
import jsxA11y from "eslint-plugin-jsx-a11y";
import perfectionist, {
	configs as perfectionistConfigs,
} from "eslint-plugin-perfectionist";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import sonarjs from "eslint-plugin-sonarjs";
import unicorn from "eslint-plugin-unicorn";
import globals from "globals";

const JS_MAX_PARAMS_ALLOWED = 3;

/** @typedef {import("eslint").Linter.Config} */
let Config;
/** @typedef {import("eslint").Linter.ParserModule} */
let ParserModule;

/** @type {Config} */
const filesConfig = {
	files: ["**/*.{js,ts,tsx}"],
};

/** @type {Config} */
const ignoresConfig = {
	ignores: ["build"],
};

/** @type {Config} */
const jsConfig = {
	languageOptions: {
		globals: {
			...globals.node,
			...globals.browser,
			JSX: true,
			React: true,
		},
		parserOptions: {
			ecmaVersion: "latest",
			sourceType: "module",
		},
	},
	rules: {
		...js.configs.recommended.rules,
		"arrow-parens": ["error", "always"],
		curly: ["error", "all"],
		"max-params": ["error", JS_MAX_PARAMS_ALLOWED],
		"no-console": ["error"],
		"no-multiple-empty-lines": [
			"error",
			{
				max: 1,
			},
		],
		"no-restricted-syntax": [
			"error",
			{
				message: "Export/Import all (*) is forbidden.",
				selector: "ExportAllDeclaration,ImportAllDeclaration",
			},
			{
				message: "Exports should be at the end of the file.",
				selector: "ExportNamedDeclaration[declaration!=null]",
			},
			{
				message: "TS features are forbidden",
				selector: "TSEnumDeclaration,ClassDeclaration[abstract=true]",
			},
		],
		quotes: ["error", "double"],
	},
};

/** @type {Config} */
const importConfig = {
	plugins: {
		"import-x": importPlugin,
	},
	rules: {
		...importConfigs.recommended.rules,
		"import-x/exports-last": ["error"],
		"import-x/extensions": [
			"error",
			{
				js: "ignorePackages",
			},
		],
		"import-x/newline-after-import": ["error"],
		"import-x/no-default-export": ["error"],
		"import-x/no-duplicates": ["error"],
	},
	settings: {
		"import-x/parsers": {
			espree: [".js", ".cjs"],
		},
		"import-x/resolver": {
			node: {
				extensions: [".js", ".ts", ".tsx"],
			},
			typescript: tsResolver,
		},
	},
};

// Redux hooks are only reached through the typed wrappers in libs/hooks, so a
// component can never read state or dispatch with an untyped store.
/** @type {Config} */
const reduxHooksImportConfig = {
	files: ["src/**/*.{ts,tsx}"],
	ignores: [
		"src/libs/hooks/use-app-dispatch/*.ts",
		"src/libs/hooks/use-app-selector/*.ts",
	],
	rules: {
		"no-restricted-imports": [
			"error",
			{
				paths: [
					{
						importNames: ["useDispatch", "useSelector"],
						message:
							"Use useAppDispatch / useAppSelector from libs/hooks instead.",
						name: "react-redux",
					},
				],
			},
		],
	},
};

/** @type {Config} */
const sonarConfig = {
	plugins: {
		sonarjs,
	},
	rules: {
		...sonarjs.configs.recommended.rules,
		"sonarjs/no-hardcoded-passwords": "off",
		"sonarjs/todo-tag": "off",
	},
};

/** @type {Config} */
const unicornConfig = {
	plugins: {
		unicorn,
	},
	rules: {
		...unicorn.configs.recommended.rules,
		"unicorn/consistent-boolean-name": [
			"error",
			{
				prefixes: {
					check: true,
				},
			},
		],
		// Class member order is already governed by the explicit
		// perfectionist/sort-classes groups below; the two rules disagree on
		// ordering, so defer to perfectionist.
		"unicorn/consistent-class-member-order": ["off"],
		"unicorn/no-null": ["off"],
	},
};

/** @type {Config} */
const perfectionistConfig = {
	plugins: {
		perfectionist,
	},
	rules: {
		...perfectionistConfigs["recommended-natural"].rules,
		"perfectionist/sort-classes": [
			"error",
			{
				groups: [
					"index-signature",
					// properties
					["static-property", "static-accessor-property"],
					["protected-static-property", "protected-static-accessor-property"],
					["private-static-property", "private-static-accessor-property"],
					["private-property", "private-accessor-property"],
					["public-property", "public-accessor-property"],
					["property", "accessor-property"],

					"constructor",
					// getters-setters
					["static-get-method", "static-set-method"],
					["protected-static-get-method", "protected-static-set-method"],
					["private-static-get-method", "private-static-set-method"],
					["public-static-get-method", "public-static-set-method"],
					["protected-get-method", "protected-set-method"],
					["private-get-method", "private-set-method"],
					["get-method", "set-method"],

					"static-block",

					// methods
					["static-method", "static-function-property"],
					["protected-static-method", "protected-static-function-property"],
					["private-static-method", "private-static-function-property"],
					["protected-method", "protected-function-property"],
					["private-method", "private-function-property"],
					["public-method", "public-function-property"],

					"unknown",
				],
			},
		],
		"perfectionist/sort-named-exports": [
			"error",
			{
				groups: ["type-export", "value-export"],
			},
		],
	},
};

/** @type {Config} */
const typescriptConfig = {
	files: ["**/*.ts", "**/*.tsx"],
	languageOptions: {
		parser: /** @type {ParserModule} */ (tsParser),
		parserOptions: {
			project: "./tsconfig.json",
		},
	},
	plugins: {
		"@typescript-eslint": ts,
	},
	rules: {
		...ts.configs["strict-type-checked"].rules,
		"@typescript-eslint/no-magic-numbers": [
			"error",
			{
				ignoreEnums: true,
				ignoreReadonlyClassProperties: true,
			},
		],
		"@typescript-eslint/no-unnecessary-type-parameters": "off",
		"@typescript-eslint/return-await": ["error", "always"],
	},
};

/** @type {Config} */
const reactConfig = {
	files: ["**/*.tsx"],
	plugins: {
		react,
	},
	rules: {
		...react.configs["jsx-runtime"].rules,
		...react.configs["recommended"].rules,
		"react/jsx-no-bind": ["error", { ignoreRefs: true }],
		"react/react-in-jsx-scope": ["off"],
	},
	settings: {
		react: {
			version: "19.2",
		},
	},
};

/** @type {Config} */
const reactHooksConfig = {
	files: ["**/*.tsx"],
	plugins: {
		"react-hooks": reactHooks,
	},
	rules: reactHooks.configs.recommended.rules,
};

/** @type {Config} */
const jsxA11yConfig = {
	files: ["**/*.tsx"],
	plugins: {
		"jsx-a11y": jsxA11y,
	},
	rules: jsxA11y.configs.recommended.rules,
};

/** @type {Config} */
const jsdocConfig = {
	files: ["eslint.config.js", "lint-staged.config.js", "prettier.config.js"],
	plugins: {
		jsdoc: jsdocPlugin,
	},
	rules: {
		...jsdocPlugin.configs["recommended-typescript-flavor-error"].rules,
		"jsdoc/no-undefined-types": ["error"],
	},
};

/** @type {Config[]} */
const overridesConfigs = [
	{
		files: [
			"commitlint.config.ts",
			"eslint.config.js",
			"knip.config.ts",
			"lint-staged.config.js",
			"packages.d.ts",
			"prettier.config.js",
			"stylelint.config.ts",
			"vite.config.ts",
		],
		rules: {
			"import-x/no-default-export": ["off"],
		},
	},
	{
		files: ["src/vite-env.d.ts"],
		rules: {
			"unicorn/prevent-abbreviations": ["off"],
		},
	},
];

/** @type {Config[]} */
const config = [
	filesConfig,
	ignoresConfig,
	jsConfig,
	importConfig,
	reduxHooksImportConfig,
	sonarConfig,
	unicornConfig,
	perfectionistConfig,
	typescriptConfig,
	reactConfig,
	reactHooksConfig,
	jsxA11yConfig,
	jsdocConfig,
	...overridesConfigs,
];

export default config;
