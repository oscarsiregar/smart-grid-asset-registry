import { type ValueOf } from "~/libs/types/types.js";

import { type ImportStep } from "../enums/enums.js";

type ImportStepItem = {
	label: string;
	step: ValueOf<typeof ImportStep>;
};

export { type ImportStepItem };
