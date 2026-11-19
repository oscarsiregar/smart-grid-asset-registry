import { type ValueOf } from "~/libs/types/types.js";

import { type OutcomeTone } from "../enums/enums.js";

type OutcomeSummary = {
	message: string;
	tone: ValueOf<typeof OutcomeTone>;
};

export { type OutcomeSummary };
