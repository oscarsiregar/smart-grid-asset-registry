import { type DefaultValues } from "react-hook-form";

import {
	ImportMode,
	type ImportRequestDto,
} from "~/modules/imports/imports.js";

/** Partial is the default mode (spec FR-021). No file is chosen yet. */
const DEFAULT_IMPORT_PAYLOAD: DefaultValues<ImportRequestDto> = {
	mode: ImportMode.PARTIAL,
};

export { DEFAULT_IMPORT_PAYLOAD };
