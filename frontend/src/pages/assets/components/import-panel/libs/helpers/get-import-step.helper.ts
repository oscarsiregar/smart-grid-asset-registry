import { type ValueOf } from "~/libs/types/types.js";
import { type ImportResultDto } from "~/modules/imports/imports.js";

import { ImportStep } from "../enums/enums.js";

type Options = {
	confirmed: ImportResultDto | undefined;
	validated: ImportResultDto | undefined;
};

/**
 * The step the panel is on, from its results (spec 005 FR-017): a saved
 * result — committed or not — is Done, a checked one Review, otherwise
 * Choose file.
 */
const getImportStep = ({
	confirmed,
	validated,
}: Options): ValueOf<typeof ImportStep> => {
	if (confirmed) {
		return ImportStep.DONE;
	}

	if (validated) {
		return ImportStep.REVIEW;
	}

	return ImportStep.CHOOSE_FILE;
};

export { getImportStep };
