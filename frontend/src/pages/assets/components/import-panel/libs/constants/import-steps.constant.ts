import { ImportPanelMessage, ImportStep } from "../enums/enums.js";
import { type ImportStepItem } from "../types/types.js";

/** The steps in order (spec 005 FR-017). */
const IMPORT_STEPS: readonly ImportStepItem[] = [
	{ label: ImportPanelMessage.STEP_CHOOSE_FILE, step: ImportStep.CHOOSE_FILE },
	{ label: ImportPanelMessage.STEP_REVIEW, step: ImportStep.REVIEW },
	{ label: ImportPanelMessage.STEP_DONE, step: ImportStep.DONE },
];

export { IMPORT_STEPS };
