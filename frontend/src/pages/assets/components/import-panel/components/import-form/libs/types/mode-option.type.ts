import { type ValueOf } from "~/libs/types/types.js";
import { type ImportMode } from "~/modules/imports/imports.js";

type ModeOption = {
	description: string;
	label: string;
	value: ValueOf<typeof ImportMode>;
};

export { type ModeOption };
